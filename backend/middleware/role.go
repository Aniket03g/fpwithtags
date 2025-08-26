package middleware

import (
	"log"
	"net/http"

	"github.com/FeaturePlus/backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleMiddleware reads user_id from context (set by AuthMiddleware),
// queries the database for that user, and sets user_role in the request context.
// If requiredRoles are provided, it will enforce that the user has one of those roles.
func RoleMiddleware(db *gorm.DB, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Query the database for the user
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Debug log to print the role found in database
		log.Printf("[RoleMiddleware] User ID %v has role '%s' in database", userID, user.Role)

		// Set user_role in context for downstream handlers
		c.Set("user_role", user.Role)

		// Debug log to confirm what role is being set in context
		userRole, _ := c.Get("user_role")
		log.Printf("[RoleMiddleware] Setting user_role in context: '%v'", userRole)

		// If specific roles are required, check them
		if len(requiredRoles) > 0 {
			allowed := false
			for _, role := range requiredRoles {
				if user.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CreateRoleMiddleware returns a RoleMiddleware with the provided database instance
func CreateRoleMiddleware(db *gorm.DB) func(...string) gin.HandlerFunc {
	return func(requiredRoles ...string) gin.HandlerFunc {
		return RoleMiddleware(db, requiredRoles...)
	}
}
