package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func authMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		if parts[1] != config.BearerToken {
			c.JSON(401, gin.H{"error": "Invalid bearer token"})
			c.Abort()
			return
		}

		c.Next()
	}
} 