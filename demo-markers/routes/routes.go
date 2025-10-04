package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the HTTP routes for the application
func RegisterRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// MARKER:API_ROUTES

	// Static file serving
	r.Static("/static", "./static")

	// Not found handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Route not found",
		})
	})
}

// UserHandler handles user-related requests
func UserHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "User endpoint",
	})
}

// ProductHandler handles product-related requests
func ProductHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Product endpoint",
	})
}

// AuthMiddleware is a middleware that checks for authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			return
		}
		c.Next()
	}
}
