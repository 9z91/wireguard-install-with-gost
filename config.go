package main

import (
	"os"
)

type Config struct {
	BearerToken string
}

func loadConfig() *Config {
	token := os.Getenv("WG_AUTH_TOKEN")
	if token == "" {
		token = "default-token" // Default token for development
	}
	return &Config{
		BearerToken: token,
	}
} 