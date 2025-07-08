package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"web-crawler-backend/internal/services"
)

// Logger provides request logging middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// ErrorHandler provides centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Error: %v", err.Err)

			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation error",
					"message": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "Something went wrong",
				})
			}
		}
	}
}

// Recovery provides panic recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Server encountered an unexpected error",
		})
	})
}

// AuthRequired provides JWT authentication middleware
func AuthRequired(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Set user info in context for later use
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminRequired provides admin-only access middleware
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after AuthRequired
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "User authentication context not found",
			})
			c.Abort()
			return
		}

		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth provides optional authentication (doesn't fail if no token)
func OptionalAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := authHeader[7:]
			
			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err == nil {
				// Set user info in context if token is valid
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("is_admin", claims.IsAdmin)
				c.Set("claims", claims)
			}
			// Continue regardless of token validity
		}

		c.Next()
	}
} 