package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// Mock crawler service for testing
type mockCrawlerServiceHandler struct {
	startCrawlCalled bool
	lastURLID        uint
}

func (m *mockCrawlerServiceHandler) StartCrawl(urlID uint) {
	m.startCrawlCalled = true
	m.lastURLID = urlID
}

func (m *mockCrawlerServiceHandler) GetCrawlStatus(urlID uint) (*models.CrawlStatusResponse, error) {
	return &models.CrawlStatusResponse{
		ID:     1,
		URL:    "https://example.com",
		Status: "completed",
	}, nil
}

func (m *mockCrawlerServiceHandler) BulkRerunCrawls(urlIDs []uint) error {
	return nil
}

func setupURLHandlerTest() (*gin.Engine, *URLHandler, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	
	// Setup database
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	db.AutoMigrate(&models.URL{}, &models.Crawl{}, &models.Link{})
	
	// Setup services
	crawlerService := &mockCrawlerServiceHandler{}
	urlService := services.NewURLService(db, crawlerService)
	handler := NewURLHandler(urlService)
	
	// Create test router
	router := gin.New()
	
	return router, handler, db
}

func TestURLHandler_GetURLs(t *testing.T) {
	t.Run("successful retrieval with default params", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test data
		urls := []*models.URL{
			{URL: "https://example1.com", Title: "Example 1", Status: "completed"},
			{URL: "https://example2.com", Title: "Example 2", Status: "pending"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		pagination := response["pagination"].(map[string]interface{})
		
		assert.Len(t, data, 2)
		assert.Equal(t, float64(2), pagination["total"])
		assert.Equal(t, float64(20), pagination["limit"])
		assert.Equal(t, float64(0), pagination["offset"])
	})
	
	t.Run("with pagination parameters", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test data
		for i := 1; i <= 5; i++ {
			url := &models.URL{
				URL:   fmt.Sprintf("https://example%d.com", i),
				Title: fmt.Sprintf("Example %d", i),
			}
			require.NoError(t, db.Create(url).Error)
		}
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls?limit=2&offset=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		pagination := response["pagination"].(map[string]interface{})
		
		assert.Len(t, data, 2)
		assert.Equal(t, float64(5), pagination["total"])
		assert.Equal(t, float64(2), pagination["limit"])
		assert.Equal(t, float64(2), pagination["offset"])
	})
	
	t.Run("with search parameter", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test data
		urls := []*models.URL{
			{URL: "https://google.com", Title: "Google Search"},
			{URL: "https://github.com", Title: "GitHub"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls?search=google", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		pagination := response["pagination"].(map[string]interface{})
		
		assert.Len(t, data, 1)
		assert.Equal(t, float64(1), pagination["total"])
	})
	
	t.Run("with status filter", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test data
		urls := []*models.URL{
			{URL: "https://completed.com", Status: "completed"},
			{URL: "https://pending.com", Status: "pending"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls?status=completed", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
		
		urlData := data[0].(map[string]interface{})
		assert.Equal(t, "completed", urlData["status"])
	})
	
	t.Run("with sorting parameters", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test data with different titles
		urls := []*models.URL{
			{URL: "https://c.com", Title: "C Title"},
			{URL: "https://a.com", Title: "A Title"},
			{URL: "https://b.com", Title: "B Title"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls?sortBy=title&sortOrder=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		assert.Len(t, data, 3)
		
		// Check sorted order
		firstURL := data[0].(map[string]interface{})
		assert.Equal(t, "A Title", firstURL["title"])
	})
	
	t.Run("invalid limit parameter", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		url := &models.URL{URL: "https://example.com"}
		require.NoError(t, db.Create(url).Error)
		
		router.GET("/urls", handler.GetURLs)
		
		req := httptest.NewRequest("GET", "/urls?limit=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(20), pagination["limit"]) // Should fallback to default
	})
}

func TestURLHandler_CreateURL(t *testing.T) {
	t.Run("successful URL creation", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.POST("/urls", handler.CreateURL)
		
		requestBody := `{"url": "https://example.com"}`
		req := httptest.NewRequest("POST", "/urls", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "URL created and crawling started", response["message"])
		
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "https://example.com", data["url"])
		assert.Equal(t, "pending", data["status"])
	})
	
	t.Run("invalid JSON body", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.POST("/urls", handler.CreateURL)
		
		requestBody := `{invalid json}`
		req := httptest.NewRequest("POST", "/urls", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid request body", response["error"])
	})
	
	t.Run("missing URL field", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.POST("/urls", handler.CreateURL)
		
		requestBody := `{}`
		req := httptest.NewRequest("POST", "/urls", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid request body", response["error"])
	})
}

func TestURLHandler_GetURL(t *testing.T) {
	t.Run("successful URL retrieval", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test URL
		url := &models.URL{
			URL:         "https://example.com",
			Title:       "Example Site",
			Status:      "completed",
			HTMLVersion: "HTML5",
		}
		require.NoError(t, db.Create(url).Error)
		
		router.GET("/urls/:id", handler.GetURL)
		
		req := httptest.NewRequest("GET", "/urls/"+strconv.Itoa(int(url.ID)), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "https://example.com", data["url"])
		assert.Equal(t, "Example Site", data["title"])
		assert.Equal(t, "completed", data["status"])
	})
	
	t.Run("invalid ID parameter", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.GET("/urls/:id", handler.GetURL)
		
		req := httptest.NewRequest("GET", "/urls/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid URL ID", response["error"])
		assert.Equal(t, "ID must be a valid number", response["message"])
	})
	
	t.Run("URL not found", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.GET("/urls/:id", handler.GetURL)
		
		req := httptest.NewRequest("GET", "/urls/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "URL not found", response["error"])
		assert.Equal(t, "The requested URL does not exist", response["message"])
	})
}

func TestURLHandler_DeleteURL(t *testing.T) {
	t.Run("successful URL deletion", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)
		
		router.DELETE("/urls/:id", handler.DeleteURL)
		
		req := httptest.NewRequest("DELETE", "/urls/"+strconv.Itoa(int(url.ID)), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "URL deleted successfully", response["message"])
		
		// Verify URL is soft deleted
		var deletedURL models.URL
		err = db.First(&deletedURL, url.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("invalid ID parameter", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.DELETE("/urls/:id", handler.DeleteURL)
		
		req := httptest.NewRequest("DELETE", "/urls/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid URL ID", response["error"])
	})
}

func TestURLHandler_BulkDeleteURLs(t *testing.T) {
	t.Run("successful bulk deletion", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test URLs
		urls := []*models.URL{
			{URL: "https://example1.com", Status: "completed"},
			{URL: "https://example2.com", Status: "pending"},
		}
		var ids []uint
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
			ids = append(ids, url.ID)
		}
		
		router.POST("/urls/bulk-delete", handler.BulkDeleteURLs)
		
		requestBody := fmt.Sprintf(`{"ids": [%d, %d]}`, ids[0], ids[1])
		req := httptest.NewRequest("POST", "/urls/bulk-delete", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "URLs deleted successfully", response["message"])
		
		// Verify URLs are soft deleted
		var count int64
		db.Model(&models.URL{}).Count(&count)
		assert.Equal(t, int64(0), count)
	})
	
	t.Run("invalid JSON body", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.POST("/urls/bulk-delete", handler.BulkDeleteURLs)
		
		requestBody := `{invalid json}`
		req := httptest.NewRequest("POST", "/urls/bulk-delete", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid request body", response["error"])
	})
	
	t.Run("empty IDs list", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.POST("/urls/bulk-delete", handler.BulkDeleteURLs)
		
		requestBody := `{"ids": []}`
		req := httptest.NewRequest("POST", "/urls/bulk-delete", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "No IDs provided", response["error"])
		assert.Equal(t, "At least one URL ID must be provided", response["message"])
	})
}

func TestURLHandler_GetURLLinks(t *testing.T) {
	t.Run("successful links retrieval", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)
		
		// Create test links
		links := []*models.Link{
			{URLID: url.ID, LinkURL: "https://example.com/page1", LinkType: "internal", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://external.com", LinkType: "external", IsAccessible: false},
		}
		for _, link := range links {
			require.NoError(t, db.Create(link).Error)
		}
		
		router.GET("/urls/:id/links", handler.GetURLLinks)
		
		req := httptest.NewRequest("GET", "/urls/"+strconv.Itoa(int(url.ID))+"/links", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		pagination := response["pagination"].(map[string]interface{})
		
		assert.Len(t, data, 2)
		assert.Equal(t, float64(2), pagination["total"])
	})
	
	t.Run("with link type filter", func(t *testing.T) {
		router, handler, db := setupURLHandlerTest()
		
		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)
		
		// Create test links
		links := []*models.Link{
			{URLID: url.ID, LinkURL: "https://example.com/page1", LinkType: "internal", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://external.com", LinkType: "external", IsAccessible: true},
		}
		for _, link := range links {
			require.NoError(t, db.Create(link).Error)
		}
		
		router.GET("/urls/:id/links", handler.GetURLLinks)
		
		req := httptest.NewRequest("GET", "/urls/"+strconv.Itoa(int(url.ID))+"/links?type=internal", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data := response["data"].([]interface{})
		pagination := response["pagination"].(map[string]interface{})
		
		assert.Len(t, data, 1)
		assert.Equal(t, float64(1), pagination["total"])
		
		linkData := data[0].(map[string]interface{})
		assert.Equal(t, "internal", linkData["link_type"])
	})
	
	t.Run("invalid URL ID", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.GET("/urls/:id/links", handler.GetURLLinks)
		
		req := httptest.NewRequest("GET", "/urls/invalid/links", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "Invalid URL ID", response["error"])
	})
	
	t.Run("URL not found", func(t *testing.T) {
		router, handler, _ := setupURLHandlerTest()
		
		router.GET("/urls/:id/links", handler.GetURLLinks)
		
		req := httptest.NewRequest("GET", "/urls/999/links", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "URL not found", response["error"])
	})
} 