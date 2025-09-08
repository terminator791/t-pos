package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/infrastructure/logger"
)

// Logger middleware logs HTTP requests
func Logger(logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status
		status := c.Writer.Status()

		// Log the request
		logger.Info("HTTP Request: %s %s - Status: %d - Latency: %v - IP: %s",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
			c.ClientIP(),
		)
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

// Auth middleware handles authentication (placeholder for now)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For now, we'll just pass through
		// TODO: Implement JWT token validation
		
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// For development, we'll accept any token for now
		// In production, validate JWT token here
		if authHeader != "Bearer test-token" {
			// For now, allow any Bearer token for development
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				c.JSON(401, gin.H{
					"error":   "unauthorized",
					"message": "Invalid authorization header format",
				})
				c.Abort()
				return
			}
		}

		// TODO: Extract user information from token and set in context
		// c.Set("user_id", userID)
		// c.Set("user_role", userRole)

		c.Next()
	}
}

// RateLimit middleware handles rate limiting (placeholder)
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic
		c.Next()
	}
}

// Validate middleware handles request validation
func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement request validation logic
		c.Next()
	}
}