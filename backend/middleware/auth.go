package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/FeaturePlus/backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens and sets user_id in the context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from Authorization header first
		var tokenString string
		authHeader := c.GetHeader("Authorization")

		if authHeader != "" {
			// Properly extract the token from the Authorization header
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no valid token in Authorization header, try to get from cookie
		if tokenString == "" {
			cookieToken, err := c.Cookie("token")
			if err == nil && cookieToken != "" {
				tokenString = cookieToken
			}
		}

		// If no token found in either Authorization header or cookie
		if tokenString == "" {
			c.Header("WWW-Authenticate", "Bearer realm=\"FeaturePlus API\"")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "missing_auth",
			})
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// Check if token is expired
			if strings.Contains(err.Error(), "token is expired") {
				c.Header("WWW-Authenticate", "Bearer error=\"expired_token\"")
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
					"code":  "token_expired",
				})
			} else {
				c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\"")
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token: " + err.Error(),
					"code":  "invalid_token",
				})
			}
			c.Abort()
			return
		}

		// Extract user ID from claims and convert to uint
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.Header("WWW-Authenticate", "Bearer error=\"invalid_token\"")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token payload: missing or invalid user_id",
				"code":  "invalid_token_payload",
			})
			c.Abort()
			return
		}

		// Debug log: print the user ID extracted from token
		log.Printf("[AuthMiddleware] Successfully extracted user_id %v from token", uint(userID))

		c.Set("user_id", uint(userID))

		// Debug log: confirm what is being set in context before passing to next middleware
		log.Printf("[AuthMiddleware] Setting user_id in context: %v", uint(userID))

		c.Next()
	}
}
