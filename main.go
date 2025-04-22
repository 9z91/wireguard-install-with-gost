package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
)

var logger = logrus.New()
var config *Config

// WireGuard configuration
const (
	wgConfigDir = "/etc/wireguard"
	wgInterface = "wg0"
)

func init() {
	// Set up logging
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	config = loadConfig()
}

func getWireGuardPort() string {
	if port := os.Getenv("WG_PORT"); port != "" {
		return port
	}
	return "51820" // Default port
}

func main() {
	// Check if running as root
	if os.Geteuid() != 0 {
		logger.Fatal("This application must be run as root")
	}

	// Initialize router
	r := gin.Default()

	// Apply authentication middleware to all routes
	r.Use(authMiddleware(config))

	// Routes
	r.POST("/clients", createClient)
	r.GET("/clients", listClients)
	r.GET("/clients/:name", getClient)
	r.DELETE("/clients/:name", deleteClient)

	// Start server
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server: ", err)
	}
}

func createClient(c *gin.Context) {
	var client struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get client configuration from environment variables or use defaults
	clientIPv4 := os.Getenv("CLIENT_IPV4")
	if clientIPv4 == "" {
		clientIPv4 = fmt.Sprintf("10.66.66.%d", getNextClientIP())
	}

	clientIPv6 := os.Getenv("CLIENT_IPV6")
	if clientIPv6 == "" {
		clientIPv6 = fmt.Sprintf("fd42:42:42::%d", getNextClientIP())
	}

	// Generate private key for client
	privateKey, err := exec.Command("wg", "genkey").Output()
	if err != nil {
		logger.Error("Failed to generate private key: ", err)
		c.JSON(500, gin.H{"error": "Failed to generate private key"})
		return
	}

	// Generate public key from private key
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(string(privateKey))
	publicKey, err := cmd.Output()
	if err != nil {
		logger.Error("Failed to generate public key: ", err)
		c.JSON(500, gin.H{"error": "Failed to generate public key"})
		return
	}

	// Get server public key
	serverPubKey, err := exec.Command("wg", "show", wgInterface, "public-key").Output()
	if err != nil {
		logger.Error("Failed to get server public key: ", err)
		c.JSON(500, gin.H{"error": "Failed to get server public key"})
		return
	}

	// Get server endpoint
	var serverEndpoint string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error("Failed to get server endpoint: ", err)
		c.JSON(500, gin.H{"error": "Failed to get server endpoint"})
		return
	}

	// Find the first non-loopback IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				serverEndpoint = ipnet.IP.String()
				break
			}
		}
	}

	if serverEndpoint == "" {
		logger.Error("No suitable IP address found")
		c.JSON(500, gin.H{"error": "No suitable IP address found"})
		return
	}

	// Create client configuration
	clientConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32,%s/128
DNS = 1.1.1.1, 1.0.0.1

[Peer]
PublicKey = %s
Endpoint = %s:%s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`, strings.TrimSpace(string(privateKey)), clientIPv4, clientIPv6, strings.TrimSpace(string(serverPubKey)), serverEndpoint, getWireGuardPort())

	// Save client configuration
	clientConfigPath := filepath.Join(wgConfigDir, fmt.Sprintf("%s.conf", client.Name))
	if err := os.WriteFile(clientConfigPath, []byte(clientConfig), 0600); err != nil {
		logger.Error("Failed to save client configuration: ", err)
		c.JSON(500, gin.H{"error": "Failed to save client configuration"})
		return
	}

	// Add peer to WireGuard interface
	cmd = exec.Command("wg", "set", wgInterface, "peer", strings.TrimSpace(string(publicKey)), "allowed-ips", fmt.Sprintf("%s/32", clientIPv4))
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to add peer: ", err)
		c.JSON(500, gin.H{"error": "Failed to add peer"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Client created successfully",
		"config":  clientConfig,
	})
}

func listClients(c *gin.Context) {
	// List all client configurations
	files, err := os.ReadDir(wgConfigDir)
	if err != nil {
		logger.Error("Failed to list clients: ", err)
		c.JSON(500, gin.H{"error": "Failed to list clients"})
		return
	}

	var clients []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			clients = append(clients, strings.TrimSuffix(file.Name(), ".conf"))
		}
	}

	c.JSON(200, gin.H{"clients": clients})
}

func deleteClient(c *gin.Context) {
	clientName := c.Param("name")
	if clientName == "" {
		c.JSON(400, gin.H{"error": "Client name is required"})
		return
	}

	// Read client configuration to get public key
	clientConfigPath := filepath.Join(wgConfigDir, fmt.Sprintf("%s.conf", clientName))
	config, err := os.ReadFile(clientConfigPath)
	if err != nil {
		logger.Error("Failed to read client configuration: ", err)
		c.JSON(500, gin.H{"error": "Failed to read client configuration"})
		return
	}

	// Extract public key from configuration
	lines := strings.Split(string(config), "\n")
	var publicKey string
	for _, line := range lines {
		if strings.HasPrefix(line, "PublicKey = ") {
			publicKey = strings.TrimPrefix(line, "PublicKey = ")
			break
		}
	}

	if publicKey == "" {
		logger.Error("Failed to find public key in configuration")
		c.JSON(500, gin.H{"error": "Failed to find public key in configuration"})
		return
	}

	// Remove peer from WireGuard interface
	cmd := exec.Command("wg", "set", wgInterface, "peer", publicKey, "remove")
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to remove peer: ", err)
		c.JSON(500, gin.H{"error": "Failed to remove peer"})
		return
	}

	// Delete client configuration file
	if err := os.Remove(clientConfigPath); err != nil {
		logger.Error("Failed to delete client configuration: ", err)
		c.JSON(500, gin.H{"error": "Failed to delete client configuration"})
		return
	}

	c.JSON(200, gin.H{"message": "Client deleted successfully"})
}

func getNextClientIP() int {
	// Get list of existing clients
	files, err := os.ReadDir(wgConfigDir)
	if err != nil {
		return 2 // Start from 2 if we can't read the directory
	}

	// Find the highest IP in use
	highestIP := 1
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			config, err := os.ReadFile(filepath.Join(wgConfigDir, file.Name()))
			if err != nil {
				continue
			}
			lines := strings.Split(string(config), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Address = 10.66.66.") {
					ip := strings.TrimPrefix(line, "Address = 10.66.66.")
					ip = strings.TrimSuffix(ip, "/24")
					if ipNum := atoi(ip); ipNum > highestIP {
						highestIP = ipNum
					}
				}
			}
		}
	}

	return highestIP + 1
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func getClient(c *gin.Context) {
	clientName := c.Param("name")
	if clientName == "" {
		c.JSON(400, gin.H{"error": "Client name is required"})
		return
	}

	// Read client configuration
	clientConfigPath := filepath.Join(wgConfigDir, fmt.Sprintf("%s.conf", clientName))
	config, err := os.ReadFile(clientConfigPath)
	if err != nil {
		logger.Error("Failed to read client configuration: ", err)
		c.JSON(404, gin.H{"error": "Client not found"})
		return
	}

	// Check if raw config is requested
	if c.Query("raw") == "true" {
		c.Header("Content-Type", "text/plain")
		c.String(200, string(config))
		return
	}

	// Check if QR code is requested
	if c.Query("qr") == "true" {
		qr, err := qrcode.Encode(string(config), qrcode.Medium, 256)
		if err != nil {
			logger.Error("Failed to generate QR code: ", err)
			c.JSON(500, gin.H{"error": "Failed to generate QR code"})
			return
		}
		c.Header("Content-Type", "image/png")
		c.Data(200, "image/png", qr)
		return
	}

	c.JSON(200, gin.H{
		"name": clientName,
		"config": string(config),
	})
} 