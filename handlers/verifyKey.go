package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
