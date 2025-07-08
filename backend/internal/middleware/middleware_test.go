package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"web-crawler-backend/internal/models"
	"web-crawler-backend/internal/services"
)

// Setup test environment
func setupMiddlewareTest() (*gin.Engine, *services.AuthService) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	// Auto migrate models
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		panic(err)
	}

	// Create auth service
	authService := services.NewAuthService(db)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Logger(), ErrorHandler(), Recovery())

	return router, authService
}

func TestLogger(t *testing.T) {
	t.Run("logs request correctly", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		
		// Capture log output
		var logBuffer bytes.Buffer
		gin.DefaultWriter = &logBuffer
		
		router.Use(Logger())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		// Make request
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify log contains request information
		logOutput := logBuffer.String()
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/test")
		assert.Contains(t, logOutput, "200")
		assert.Contains(t, logOutput, "test-agent")
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("handles bind errors", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(ErrorHandler())
		
		router.POST("/test", func(c *gin.Context) {
			var req struct {
				Required string `json:"required" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.Error(err).SetType(gin.ErrorTypeBind)
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		// Send invalid JSON
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Validation error", response["error"])
	})
	
	t.Run("handles public errors", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			c.Error(fmt.Errorf("public error")).SetType(gin.ErrorTypePublic)
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
	})
	
	t.Run("handles generic errors", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			c.Error(fmt.Errorf("generic error"))
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "Something went wrong", response["message"])
	})
	
	t.Run("no errors - passes through", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(ErrorHandler())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})
}

func TestRecovery(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(Recovery())
		
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Should recover and return 500
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "Server encountered an unexpected error", response["message"])
	})
	
	t.Run("normal operation - no panic", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(Recovery())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthRequired(t *testing.T) {
	t.Run("missing authorization header", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(AuthRequired(authService))
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Unauthorized", response["error"])
		assert.Equal(t, "Authorization header is required", response["message"])
	})
	
	t.Run("invalid authorization header format", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(AuthRequired(authService))
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Unauthorized", response["error"])
		assert.Equal(t, "Invalid authorization header format", response["message"])
	})
	
	t.Run("invalid token", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(AuthRequired(authService))
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Unauthorized", response["error"])
	})
	
	t.Run("valid token", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		
		// Create a test user
		user, err := authService.Register(&models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		})
		require.NoError(t, err)
		
		// Login to get token
		authResponse, err := authService.Login(&models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		})
		require.NoError(t, err)
		token := authResponse.Token
		
		router.Use(AuthRequired(authService))
		router.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			username, _ := c.Get("username")
			isAdmin, _ := c.Get("is_admin")
			
			c.JSON(http.StatusOK, gin.H{
				"message":  "success",
				"user_id":  userID,
				"username": username,
				"is_admin": isAdmin,
			})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.Equal(t, float64(user.ID), response["user_id"])
		assert.Equal(t, "testuser", response["username"])
		assert.False(t, response["is_admin"].(bool))
	})
}

func TestAdminRequired(t *testing.T) {
	t.Run("missing user context", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		router.Use(AdminRequired())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", response["error"])
		assert.Equal(t, "User authentication context not found", response["message"])
	})
	
	t.Run("non-admin user", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		
		// Setup middleware that sets user context
		router.Use(func(c *gin.Context) {
			c.Set("is_admin", false)
			c.Next()
		})
		router.Use(AdminRequired())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Forbidden", response["error"])
		assert.Equal(t, "Admin access required", response["message"])
	})
	
	t.Run("admin user", func(t *testing.T) {
		router, _ := setupMiddlewareTest()
		
		// Setup middleware that sets admin context
		router.Use(func(c *gin.Context) {
			c.Set("is_admin", true)
			c.Next()
		})
		router.Use(AdminRequired())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})
}

func TestOptionalAuth(t *testing.T) {
	t.Run("no authorization header - continues", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(OptionalAuth(authService))
		
		router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{
				"message":         "success",
				"authenticated":   exists,
				"user_id":        userID,
			})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.False(t, response["authenticated"].(bool))
		assert.Nil(t, response["user_id"])
	})
	
	t.Run("invalid token - continues without auth", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(OptionalAuth(authService))
		
		router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{
				"message":         "success",
				"authenticated":   exists,
				"user_id":        userID,
			})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.False(t, response["authenticated"].(bool))
	})
	
	t.Run("valid token - sets user context", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		
		// Create a test user
		user, err := authService.Register(&models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		})
		require.NoError(t, err)
		
		// Login to get token
		authResponse, err := authService.Login(&models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		})
		require.NoError(t, err)
		token := authResponse.Token
		
		router.Use(OptionalAuth(authService))
		router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{
				"message":         "success",
				"authenticated":   exists,
				"user_id":        userID,
				"username":       username,
			})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.True(t, response["authenticated"].(bool))
		assert.Equal(t, float64(user.ID), response["user_id"])
		assert.Equal(t, "testuser", response["username"])
	})
	
	t.Run("invalid header format - continues without auth", func(t *testing.T) {
		router, authService := setupMiddlewareTest()
		router.Use(OptionalAuth(authService))
		
		router.GET("/test", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{
				"message":         "success",
				"authenticated":   exists,
				"user_id":        userID,
			})
		})
		
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.False(t, response["authenticated"].(bool))
	})
} 