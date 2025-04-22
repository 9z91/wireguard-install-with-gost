package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()
var config *Config

func init() {
	// Set up logging
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	config = loadConfig()
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

	// Execute wireguard-install.sh with the client name
	cmd := exec.Command("./wireguard-install.sh")
	cmd.Stdin = strings.NewReader(client.Name + "\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to create client: ", err)
		c.JSON(500, gin.H{"error": "Failed to create client"})
		return
	}

	c.JSON(200, gin.H{"message": "Client created successfully", "output": string(output)})
}

func listClients(c *gin.Context) {
	// Execute wireguard-install.sh to list clients
	cmd := exec.Command("./wireguard-install.sh")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to list clients: ", err)
		c.JSON(500, gin.H{"error": "Failed to list clients"})
		return
	}

	c.JSON(200, gin.H{"clients": string(output)})
}

func deleteClient(c *gin.Context) {
	clientName := c.Param("name")
	if clientName == "" {
		c.JSON(400, gin.H{"error": "Client name is required"})
		return
	}

	// Execute wireguard-install.sh to delete the client
	cmd := exec.Command("./wireguard-install.sh")
	cmd.Stdin = strings.NewReader("3\n" + clientName + "\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to delete client: ", err)
		c.JSON(500, gin.H{"error": "Failed to delete client"})
		return
	}

	c.JSON(200, gin.H{"message": "Client deleted successfully", "output": string(output)})
} 