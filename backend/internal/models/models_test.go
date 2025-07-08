package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserModel(t *testing.T) {
	t.Run("User creation with valid data", func(t *testing.T) {
		user := User{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "hashedpassword",
			FirstName: "Test",
			LastName:  "User",
			IsActive:  true,
			IsAdmin:   false,
		}

		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test", user.FirstName)
		assert.Equal(t, "User", user.LastName)
		assert.True(t, user.IsActive)
		assert.False(t, user.IsAdmin)
	})

	t.Run("User JSON serialization excludes password", func(t *testing.T) {
		user := User{
			ID:        1,
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "secret",
			FirstName: "Test",
			LastName:  "User",
			IsActive:  true,
			IsAdmin:   false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		jsonData, err := json.Marshal(user)
		require.NoError(t, err)

		var userMap map[string]interface{}
		err = json.Unmarshal(jsonData, &userMap)
		require.NoError(t, err)

		// Password should not be in JSON
		_, passwordExists := userMap["password"]
		assert.False(t, passwordExists, "Password should be excluded from JSON")

		// Other fields should be present
		assert.Equal(t, "testuser", userMap["username"])
		assert.Equal(t, "test@example.com", userMap["email"])
	})
}

func TestURLModel(t *testing.T) {
	t.Run("URL creation with valid data", func(t *testing.T) {
		url := URL{
			URL:         "https://example.com",
			Title:       "Example Site",
			HTMLVersion: "HTML5",
			Status:      "pending",
			HasLoginForm: false,
		}

		assert.Equal(t, "https://example.com", url.URL)
		assert.Equal(t, "Example Site", url.Title)
		assert.Equal(t, "HTML5", url.HTMLVersion)
		assert.Equal(t, "pending", url.Status)
		assert.False(t, url.HasLoginForm)
	})

	t.Run("URL JSON serialization", func(t *testing.T) {
		url := URL{
			ID:          1,
			URL:         "https://example.com",
			Title:       "Example Site",
			Status:      "completed",
			HasLoginForm: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		jsonData, err := json.Marshal(url)
		require.NoError(t, err)

		var urlMap map[string]interface{}
		err = json.Unmarshal(jsonData, &urlMap)
		require.NoError(t, err)

		assert.Equal(t, "https://example.com", urlMap["url"])
		assert.Equal(t, "Example Site", urlMap["title"])
		assert.Equal(t, "completed", urlMap["status"])
		assert.True(t, urlMap["has_login_form"].(bool))
	})
}

func TestCrawlModel(t *testing.T) {
	t.Run("Crawl creation with valid data", func(t *testing.T) {
		startTime := time.Now()
		crawl := Crawl{
			URLID:         1,
			Status:        "running",
			StartedAt:     &startTime,
			InternalLinks: 5,
			ExternalLinks: 3,
			BrokenLinks:   1,
			HeadingCounts: `{"h1":1,"h2":3,"h3":5}`,
		}

		assert.Equal(t, uint(1), crawl.URLID)
		assert.Equal(t, "running", crawl.Status)
		assert.Equal(t, &startTime, crawl.StartedAt)
		assert.Equal(t, 5, crawl.InternalLinks)
		assert.Equal(t, 3, crawl.ExternalLinks)
		assert.Equal(t, 1, crawl.BrokenLinks)
		assert.Equal(t, `{"h1":1,"h2":3,"h3":5}`, crawl.HeadingCounts)
	})

	t.Run("Crawl with completed status", func(t *testing.T) {
		startTime := time.Now()
		completedTime := time.Now().Add(time.Minute * 5)
		
		crawl := Crawl{
			URLID:         1,
			Status:        "completed",
			StartedAt:     &startTime,
			CompletedAt:   &completedTime,
			InternalLinks: 10,
			ExternalLinks: 7,
			BrokenLinks:   0,
		}

		assert.Equal(t, "completed", crawl.Status)
		assert.NotNil(t, crawl.StartedAt)
		assert.NotNil(t, crawl.CompletedAt)
		assert.True(t, crawl.CompletedAt.After(*crawl.StartedAt))
		assert.Equal(t, 0, crawl.BrokenLinks)
	})
}

func TestLinkModel(t *testing.T) {
	t.Run("Link creation with valid data", func(t *testing.T) {
		link := Link{
			URLID:        1,
			CrawlID:      1,
			LinkURL:      "https://example.com/page1",
			LinkText:     "Page 1",
			LinkType:     "internal",
			StatusCode:   200,
			IsAccessible: true,
		}

		assert.Equal(t, uint(1), link.URLID)
		assert.Equal(t, uint(1), link.CrawlID)
		assert.Equal(t, "https://example.com/page1", link.LinkURL)
		assert.Equal(t, "Page 1", link.LinkText)
		assert.Equal(t, "internal", link.LinkType)
		assert.Equal(t, 200, link.StatusCode)
		assert.True(t, link.IsAccessible)
	})

	t.Run("External broken link", func(t *testing.T) {
		link := Link{
			URLID:        1,
			CrawlID:      1,
			LinkURL:      "https://broken-site.com",
			LinkText:     "Broken Link",
			LinkType:     "external",
			StatusCode:   404,
			IsAccessible: false,
		}

		assert.Equal(t, "external", link.LinkType)
		assert.Equal(t, 404, link.StatusCode)
		assert.False(t, link.IsAccessible)
	})
}

func TestHeadingCounts(t *testing.T) {
	t.Run("HeadingCounts JSON serialization", func(t *testing.T) {
		headingCounts := HeadingCounts{
			H1: 1,
			H2: 3,
			H3: 5,
			H4: 2,
			H5: 1,
			H6: 0,
		}

		jsonData, err := json.Marshal(headingCounts)
		require.NoError(t, err)

		var counts map[string]interface{}
		err = json.Unmarshal(jsonData, &counts)
		require.NoError(t, err)

		assert.Equal(t, float64(1), counts["h1"])
		assert.Equal(t, float64(3), counts["h2"])
		assert.Equal(t, float64(5), counts["h3"])
		assert.Equal(t, float64(2), counts["h4"])
		assert.Equal(t, float64(1), counts["h5"])
		assert.Equal(t, float64(0), counts["h6"])
	})
}

func TestCrawlRequest(t *testing.T) {
	t.Run("Valid CrawlRequest", func(t *testing.T) {
		request := CrawlRequest{
			URL: "https://example.com",
		}

		assert.Equal(t, "https://example.com", request.URL)
	})

	t.Run("CrawlRequest JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"url":"https://test.com"}`
		
		var request CrawlRequest
		err := json.Unmarshal([]byte(jsonData), &request)
		require.NoError(t, err)

		assert.Equal(t, "https://test.com", request.URL)
	})
}

func TestAuthModels(t *testing.T) {
	t.Run("LoginRequest validation", func(t *testing.T) {
		loginReq := LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}

		assert.Equal(t, "testuser", loginReq.Username)
		assert.Equal(t, "testpass", loginReq.Password)
	})

	t.Run("RegisterRequest validation", func(t *testing.T) {
		registerReq := RegisterRequest{
			Username:  "newuser",
			Email:     "new@example.com",
			Password:  "securepass",
			FirstName: "New",
			LastName:  "User",
		}

		assert.Equal(t, "newuser", registerReq.Username)
		assert.Equal(t, "new@example.com", registerReq.Email)
		assert.Equal(t, "securepass", registerReq.Password)
		assert.Equal(t, "New", registerReq.FirstName)
		assert.Equal(t, "User", registerReq.LastName)
	})

	t.Run("AuthResponse structure", func(t *testing.T) {
		user := &User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		}

		authResp := AuthResponse{
			Token: "jwt-token-here",
			User:  user,
		}

		assert.Equal(t, "jwt-token-here", authResp.Token)
		assert.Equal(t, user, authResp.User)
		assert.Equal(t, uint(1), authResp.User.ID)
	})
}

func TestJWTClaims(t *testing.T) {
	t.Run("JWT Claims creation", func(t *testing.T) {
		claims := JWTClaims{
			UserID:   1,
			Username: "testuser",
			IsAdmin:  false,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
				Issuer:    "web-crawler",
			},
		}

		assert.Equal(t, uint(1), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.False(t, claims.IsAdmin)
		assert.Equal(t, "web-crawler", claims.Issuer)
		assert.True(t, claims.ExpiresAt > time.Now().Unix())
	})
}

func TestBulkRequest(t *testing.T) {
	t.Run("BulkRequest with multiple IDs", func(t *testing.T) {
		bulkReq := BulkRequest{
			IDs: []uint{1, 2, 3, 4, 5},
		}

		assert.Len(t, bulkReq.IDs, 5)
		assert.Contains(t, bulkReq.IDs, uint(1))
		assert.Contains(t, bulkReq.IDs, uint(5))
	})

	t.Run("BulkRequest JSON unmarshaling", func(t *testing.T) {
		jsonData := `{"ids":[1,2,3]}`
		
		var bulkReq BulkRequest
		err := json.Unmarshal([]byte(jsonData), &bulkReq)
		require.NoError(t, err)

		assert.Len(t, bulkReq.IDs, 3)
		assert.Equal(t, []uint{1, 2, 3}, bulkReq.IDs)
	})
}

func TestCrawlStatusResponse(t *testing.T) {
	t.Run("CrawlStatusResponse creation", func(t *testing.T) {
		startTime := time.Now()
		completedTime := time.Now().Add(time.Minute)
		headingCounts := &HeadingCounts{H1: 1, H2: 3, H3: 5}

		response := CrawlStatusResponse{
			ID:            1,
			URL:           "https://example.com",
			Status:        "completed",
			InternalLinks: 10,
			ExternalLinks: 5,
			BrokenLinks:   1,
			HeadingCounts: headingCounts,
			StartedAt:     &startTime,
			CompletedAt:   &completedTime,
		}

		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "https://example.com", response.URL)
		assert.Equal(t, "completed", response.Status)
		assert.Equal(t, 10, response.InternalLinks)
		assert.Equal(t, 5, response.ExternalLinks)
		assert.Equal(t, 1, response.BrokenLinks)
		assert.NotNil(t, response.HeadingCounts)
		assert.Equal(t, 1, response.HeadingCounts.H1)
	})
} 