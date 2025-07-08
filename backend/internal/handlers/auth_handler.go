package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"web-crawler-backend/internal/models"
	"web-crawler-backend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "username or email already exists" {
			statusCode = http.StatusConflict
		}
		
		c.JSON(statusCode, gin.H{
			"error":   "Registration failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid credentials" {
			statusCode = http.StatusUnauthorized
		}
		
		c.JSON(statusCode, gin.H{
			"error":   "Login failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// RefreshToken handles token refresh
// @Summary Refresh JWT token
// @Description Generate a new JWT token using existing valid token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Current token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	authResponse, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// GetProfile returns current user profile
// @Summary Get user profile
// @Description Get current authenticated user's profile
// @Tags auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from middleware context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User context not found",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uint))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}
		
		c.JSON(statusCode, gin.H{
			"error":   "Failed to get profile",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Logout handles user logout (client-side token invalidation)
// @Summary Logout user
// @Description Logout user (client should discard token)
// @Tags auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT implementation, logout is typically handled client-side
	// by discarding the token. For enhanced security, you could implement a
	// token blacklist on the server side.
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ValidateToken validates if the current token is valid
// @Summary Validate token
// @Description Check if the current JWT token is valid
// @Tags auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// If we reach here, it means the token is valid (middleware passed)
	username, _ := c.Get("username")
	isAdmin, _ := c.Get("is_admin")
	userID, _ := c.Get("user_id")

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"user_id":  userID,
		"username": username,
		"is_admin": isAdmin,
		"message":  "Token is valid",
	})
} 