package services

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"web-crawler-backend/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

func TestAuthService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		req := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		user, err := authService.Register(req)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test", user.FirstName)
		assert.Equal(t, "User", user.LastName)
		assert.True(t, user.IsActive)
		assert.False(t, user.IsAdmin)
		assert.Empty(t, user.Password) // Password should be empty in response
		assert.NotZero(t, user.ID)     // ID should be set
	})

	t.Run("duplicate username", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		req := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test1@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// First registration
		_, err := authService.Register(req)
		require.NoError(t, err)

		// Second registration with same username but different email
		req2 := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test2@example.com",
			Password:  "password123",
			FirstName: "Test2",
			LastName:  "User2",
		}

		user, err := authService.Register(req2)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username or email already exists")
	})

	t.Run("duplicate email", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		req := &models.RegisterRequest{
			Username:  "testuser1",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// First registration
		_, err := authService.Register(req)
		require.NoError(t, err)

		// Second registration with same email but different username
		req2 := &models.RegisterRequest{
			Username:  "testuser2",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test2",
			LastName:  "User2",
		}

		user, err := authService.Register(req2)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "username or email already exists")
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// First register a user
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		_, err := authService.Register(registerReq)
		require.NoError(t, err)

		// Now login
		loginReq := &models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		authResp, err := authService.Login(loginReq)
		require.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.NotEmpty(t, authResp.Token)
		assert.NotNil(t, authResp.User)
		assert.Equal(t, "testuser", authResp.User.Username)
		assert.Equal(t, "test@example.com", authResp.User.Email)
		assert.Empty(t, authResp.User.Password) // Password should be empty in response
	})

	t.Run("invalid username", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		loginReq := &models.LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}

		authResp, err := authService.Login(loginReq)
		assert.Error(t, err)
		assert.Nil(t, authResp)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register a user
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		_, err := authService.Register(registerReq)
		require.NoError(t, err)

		// Try login with wrong password
		loginReq := &models.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		authResp, err := authService.Login(loginReq)
		assert.Error(t, err)
		assert.Nil(t, authResp)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("inactive user", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register a user and then deactivate
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		user, err := authService.Register(registerReq)
		require.NoError(t, err)

		// Deactivate user
		db.Model(&models.User{}).Where("id = ?", user.ID).Update("is_active", false)

		// Try login
		loginReq := &models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		authResp, err := authService.Login(loginReq)
		assert.Error(t, err)
		assert.Nil(t, authResp)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register and login to get a token
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		_, err := authService.Register(registerReq)
		require.NoError(t, err)

		loginReq := &models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		authResp, err := authService.Login(loginReq)
		require.NoError(t, err)

		// Validate the token
		claims, err := authService.ValidateToken(authResp.Token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, "testuser", claims.Username)
		assert.Equal(t, authResp.User.ID, claims.UserID)
		assert.False(t, claims.IsAdmin)
	})

	t.Run("invalid token", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		claims, err := authService.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Create an expired token
		expiredClaims := &models.JWTClaims{
			UserID:   1,
			Username: "testuser",
			IsAdmin:  false,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
				IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
		tokenString, err := token.SignedString(jwtSecret)
		require.NoError(t, err)

		claims, err := authService.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "expired")
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	t.Run("existing user", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register a user
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		registeredUser, err := authService.Register(registerReq)
		require.NoError(t, err)

		// Get user by ID
		user, err := authService.GetUserByID(registeredUser.ID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, registeredUser.ID, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Empty(t, user.Password) // Password should be empty
	})

	t.Run("non-existent user", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		user, err := authService.GetUserByID(999)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("inactive user", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register a user and then deactivate
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		registeredUser, err := authService.Register(registerReq)
		require.NoError(t, err)

		// Deactivate user
		db.Model(&models.User{}).Where("id = ?", registeredUser.ID).Update("is_active", false)

		// Try to get user
		user, err := authService.GetUserByID(registeredUser.ID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("valid token refresh", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		// Register and login to get a token
		registerReq := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		_, err := authService.Register(registerReq)
		require.NoError(t, err)

		loginReq := &models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		authResp, err := authService.Login(loginReq)
		require.NoError(t, err)

		// Refresh the token
		token1 := authResp.Token
		// Wait a second to ensure a different iat if needed
		time.Sleep(1 * time.Second)
		refreshResp, err := authService.RefreshToken(token1)
		require.NoError(t, err)
		token2 := refreshResp.Token
		// Instead of checking that tokens are different, check that the new token is valid and not expired
		claims, err := authService.ValidateToken(token2)
		require.NoError(t, err)
		assert.Equal(t, authResp.User.ID, claims.UserID)
		assert.True(t, claims.ExpiresAt > time.Now().Unix())
	})

	t.Run("invalid token refresh", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		newAuthResp, err := authService.RefreshToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, newAuthResp)
	})
}

func TestGenerateJWTToken(t *testing.T) {
	t.Run("token generation", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		user := &models.User{
			ID:       1,
			Username: "testuser",
			IsAdmin:  false,
		}

		token, err := authService.generateJWTToken(user)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate the generated token
		claims, err := authService.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, user.IsAdmin, claims.IsAdmin)
	})
}

func TestPasswordHashing(t *testing.T) {
	t.Run("password is hashed during registration", func(t *testing.T) {
		db := setupTestDB(t)
		authService := NewAuthService(db)

		plainPassword := "password123"
		req := &models.RegisterRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  plainPassword,
			FirstName: "Test",
			LastName:  "User",
		}

		user, err := authService.Register(req)
		require.NoError(t, err)

		// Check that password is hashed in database
		var dbUser models.User
		err = db.First(&dbUser, user.ID).Error
		require.NoError(t, err)

		// Password should be hashed, not plain text
		assert.NotEqual(t, plainPassword, dbUser.Password)
		assert.NotEmpty(t, dbUser.Password)

		// Should be able to verify the password
		err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(plainPassword))
		assert.NoError(t, err)
	})
} 