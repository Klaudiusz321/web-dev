package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"web-crawler-backend/internal/models"
)

var (
	jwtSecret = []byte("your-secret-key") // In production, use environment variable
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Register creates a new user account
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		IsAdmin:   false, // Default to non-admin
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Don't return password in response
	user.Password = ""
	return &user, nil
}

// Login authenticates a user and returns JWT token
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	var user models.User
	
	// Find user by username
	if err := s.db.Where("username = ? AND is_active = ?", req.Username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateJWTToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Don't return password in response
	user.Password = ""

	return &models.AuthResponse{
		Token: token,
		User:  &user,
	}, nil
}

// ValidateToken validates JWT token and returns user claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt < time.Now().Unix() {
			return nil, errors.New("token has expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Don't return password
	user.Password = ""
	return &user, nil
}

// RefreshToken generates a new JWT token for the user
func (s *AuthService) RefreshToken(tokenString string) (*models.AuthResponse, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get current user data
	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new token
	newToken, err := s.generateJWTToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %v", err)
	}

	return &models.AuthResponse{
		Token: newToken,
		User:  user,
	}, nil
}

// generateJWTToken creates a JWT token for the user
func (s *AuthService) generateJWTToken(user *models.User) (string, error) {
	// Set token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := &models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
} 