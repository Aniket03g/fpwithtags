package middleware

import (
	"net/http"

	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/utils"
	"github.com/gin-gonic/gin"
)

// RequireAuthWeb middleware checks for jwt-token cookie and validates it
func RequireAuthWeb(userRepo *repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from the cookie
		cookie, err := c.Cookie("jwt-token")
		if err != nil {
			// No cookie found, redirect to login
			c.Redirect(http.StatusFound, "/web/login")
			c.Abort()
			return
		}

		// Validate the token
		claims, err := utils.ValidateToken(cookie)
		if err != nil {
			// Invalid token, redirect to login
			c.Redirect(http.StatusFound, "/web/login")
			c.Abort()
			return
		}

		// Extract user ID from claims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			// Invalid token claims, redirect to login
			c.Redirect(http.StatusFound, "/web/login")
			c.Abort()
			return
		}
		userID := int(userIDFloat)

		// Check if the user exists in the database
		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			// User not found or db error, redirect to login
			c.Redirect(http.StatusFound, "/web/login")
			c.Abort()
			return
		}

		// Set user info in context for use in handlers
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("email", user.Email)
		c.Set("role", user.Role)

		c.Next()
	}
}

// OptionalAuthWeb middleware checks for jwt-token cookie but doesn't require it
func OptionalAuthWeb() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("jwt-token")
		if err == nil {
			claims, err := utils.ValidateToken(cookie)
			if err == nil {
				// Valid token found, set user info in context
				if userID, ok := claims["user_id"].(float64); ok {
					c.Set("user_id", int(userID))
				}
				if username, ok := claims["username"].(string); ok {
					c.Set("username", username)
				}
				if email, ok := claims["email"].(string); ok {
					c.Set("email", email)
				}
				if role, ok := claims["role"].(string); ok {
					c.Set("role", role)
				}
			}
		}
		c.Next()
	}
}
