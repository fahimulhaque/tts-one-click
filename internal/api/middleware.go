package api

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets permissive CORS headers and handles OPTIONS preflight.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RequestLogger is a placeholder for richer per-request logging.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
