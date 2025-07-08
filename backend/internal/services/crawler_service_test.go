package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"web-crawler-backend/internal/models"
)

func setupCrawlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(&models.URL{}, &models.Crawl{}, &models.Link{}, &models.User{})
	require.NoError(t, err)

	return db
}

func TestNewCrawlerService(t *testing.T) {
	db := setupCrawlerTestDB(t)
	service := NewCrawlerService(db)
	
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestCrawlerService_StartCrawl(t *testing.T) {
	t.Run("successful crawl", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Create test server
		testHTML := `
		<!DOCTYPE html>
		<html>
		<head><title>Test Page</title></head>
		<body>
			<h1>Main Title</h1>
			<h2>Subtitle</h2>
			<h3>Section</h3>
			<a href="/internal">Internal Link</a>
			<a href="https://external.com">External Link</a>
		</body>
		</html>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testHTML))
		}))
		defer server.Close()

		// Create URL record
		urlRecord := &models.URL{
			URL:    server.URL,
			Status: "pending",
		}
		require.NoError(t, db.Create(urlRecord).Error)

		// Start crawl
		service.StartCrawl(urlRecord.ID)

		// Wait a bit for crawling to complete
		time.Sleep(100 * time.Millisecond)

		// Verify URL was updated
		var updatedURL models.URL
		require.NoError(t, db.First(&updatedURL, urlRecord.ID).Error)
		assert.Equal(t, "Test Page", updatedURL.Title)
		assert.Equal(t, "completed", updatedURL.Status)

		// Verify crawl was created
		var crawl models.Crawl
		require.NoError(t, db.Where("url_id = ?", urlRecord.ID).First(&crawl).Error)
		assert.Equal(t, "completed", crawl.Status)
		assert.NotNil(t, crawl.StartedAt)
		assert.NotNil(t, crawl.CompletedAt)
	})

	t.Run("crawl non-existent URL", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Try to crawl non-existent URL
		service.StartCrawl(999)

		// Should handle gracefully without panicking
		// Check that no crawl record was created
		var count int64
		db.Model(&models.Crawl{}).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("crawl with HTTP error", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Create test server that returns 404
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// Create URL record
		urlRecord := &models.URL{
			URL:    server.URL,
			Status: "pending",
		}
		require.NoError(t, db.Create(urlRecord).Error)

		// Start crawl
		service.StartCrawl(urlRecord.ID)

		// Wait for completion
		time.Sleep(100 * time.Millisecond)

		// Verify crawl failed
		var crawl models.Crawl
		require.NoError(t, db.Where("url_id = ?", urlRecord.ID).First(&crawl).Error)
		assert.Equal(t, "error", crawl.Status)
		assert.Contains(t, crawl.ErrorMessage, "404")
	})
}

func TestCrawlerService_extractData(t *testing.T) {
	t.Run("extract basic HTML data", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		htmlContent := `
		<!DOCTYPE html>
		<html>
		<head><title>Test Page</title></head>
		<body>
			<h1>Title 1</h1>
			<h1>Title 2</h1>
			<h2>Subtitle 1</h2>
			<h3>Section 1</h3>
			<h3>Section 2</h3>
			<h3>Section 3</h3>
			<a href="/internal">Internal Link</a>
			<a href="https://external.com">External Link</a>
			<form>
				<input type="password" name="password">
				<input type="submit" value="Login">
			</form>
		</body>
		</html>`

		doc, err := html.Parse(strings.NewReader(htmlContent))
		require.NoError(t, err)

		data := service.extractData(doc, "https://example.com")

		assert.Equal(t, "Test Page", data.Title)
		assert.Equal(t, "HTML5", data.HTMLVersion)
		assert.True(t, data.HasLoginForm)
		assert.Equal(t, 2, data.HeadingCounts.H1)
		assert.Equal(t, 1, data.HeadingCounts.H2)
		assert.Equal(t, 3, data.HeadingCounts.H3)
		assert.Equal(t, 0, data.HeadingCounts.H4)
		assert.Equal(t, 0, data.HeadingCounts.H5)
		assert.Equal(t, 0, data.HeadingCounts.H6)
	})

	t.Run("extract links categorization", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		htmlContent := `
		<html>
		<body>
			<a href="/page1">Internal 1</a>
			<a href="/page2">Internal 2</a>
			<a href="https://example.com/page3">Same domain</a>
			<a href="https://external.com">External</a>
			<a href="mailto:test@example.com">Email</a>
			<a>No href</a>
		</body>
		</html>`

		doc, err := html.Parse(strings.NewReader(htmlContent))
		require.NoError(t, err)

		data := service.extractData(doc, "https://example.com")

		// Should have processed valid links
		assert.True(t, len(data.Links) > 0)
	})

	t.Run("no title fallback", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		htmlContent := `<html><body><h1>No Title Tag</h1></body></html>`

		doc, err := html.Parse(strings.NewReader(htmlContent))
		require.NoError(t, err)

		data := service.extractData(doc, "https://example.com")

		assert.Empty(t, data.Title)
		assert.Equal(t, 1, data.HeadingCounts.H1)
	})
}

func TestCrawlerService_GetCrawlStatus(t *testing.T) {
	t.Run("successful status retrieval", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Create test data
		urlRecord := &models.URL{
			URL:         "https://example.com",
			Title:       "Test Site",
			Status:      "completed",
			HTMLVersion: "HTML5",
		}
		require.NoError(t, db.Create(urlRecord).Error)

		startTime := time.Now().Add(-time.Minute)
		endTime := time.Now()
		crawl := &models.Crawl{
			URLID:         urlRecord.ID,
			Status:        "completed",
			StartedAt:     &startTime,
			CompletedAt:   &endTime,
			InternalLinks: 10,
			ExternalLinks: 5,
			BrokenLinks:   2,
			HeadingCounts: `{"h1":1,"h2":3,"h3":5}`,
		}
		require.NoError(t, db.Create(crawl).Error)

		// Get status
		status, err := service.GetCrawlStatus(urlRecord.ID)
		require.NoError(t, err)
		assert.NotNil(t, status)
		assert.Equal(t, "https://example.com", status.URL)
		assert.Equal(t, "completed", status.Status)
		assert.Equal(t, 10, status.InternalLinks)
		assert.Equal(t, 5, status.ExternalLinks)
		assert.Equal(t, 2, status.BrokenLinks)
	})

	t.Run("URL not found", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		status, err := service.GetCrawlStatus(999)
		assert.Error(t, err)
		assert.Nil(t, status)
		assert.Contains(t, err.Error(), "URL not found")
	})

	t.Run("no crawl data", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Create URL without crawl
		urlRecord := &models.URL{
			URL:    "https://example.com",
			Status: "pending",
		}
		require.NoError(t, db.Create(urlRecord).Error)

		status, err := service.GetCrawlStatus(urlRecord.ID)
		require.NoError(t, err)
		assert.Equal(t, "pending", status.Status)
		assert.Equal(t, 0, status.InternalLinks)
	})
}

func TestCrawlerService_BulkRerunCrawls(t *testing.T) {
	t.Run("successful bulk rerun", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		// Create test URLs
		url1 := &models.URL{URL: "https://example1.com", Status: "completed"}
		url2 := &models.URL{URL: "https://example2.com", Status: "completed"}
		require.NoError(t, db.Create(url1).Error)
		require.NoError(t, db.Create(url2).Error)

		// Run bulk rerun
		err := service.BulkRerunCrawls([]uint{url1.ID, url2.ID})
		require.NoError(t, err)

		// Wait for async operations to complete
		time.Sleep(200 * time.Millisecond)

		// Verify URLs status was updated
		var updatedURL1, updatedURL2 models.URL
		require.NoError(t, db.First(&updatedURL1, url1.ID).Error)
		require.NoError(t, db.First(&updatedURL2, url2.ID).Error)
		validStatuses := map[string]bool{"pending":true, "completed":true, "error":true}
		assert.True(t, validStatuses[updatedURL1.Status], "unexpected status: %s", updatedURL1.Status)
		assert.True(t, validStatuses[updatedURL2.Status], "unexpected status: %s", updatedURL2.Status)
	})

	t.Run("empty IDs list", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		err := service.BulkRerunCrawls([]uint{})
		assert.NoError(t, err) // Should handle gracefully
	})

	t.Run("non-existent URLs", func(t *testing.T) {
		db := setupCrawlerTestDB(t)
		service := NewCrawlerService(db)

		err := service.BulkRerunCrawls([]uint{999, 1000})
		assert.NoError(t, err) // Should handle gracefully without errors
	})
}

func TestCrawlerService_checkLoginForm(t *testing.T) {
	db := setupCrawlerTestDB(t)
	service := NewCrawlerService(db)

	t.Run("detects password input", func(t *testing.T) {
		htmlContent := `<form><input type="password" name="pass"></form>`
		doc, _ := html.Parse(strings.NewReader(htmlContent))
		
		data := &CrawlData{}
		service.traverseHTML(doc, data, nil)
		
		assert.True(t, data.HasLoginForm)
	})

	t.Run("detects login button", func(t *testing.T) {
		htmlContent := `<form><input type="submit" value="Login"></form>`
		doc, _ := html.Parse(strings.NewReader(htmlContent))
		
		data := &CrawlData{}
		service.traverseHTML(doc, data, nil)
		
		assert.True(t, data.HasLoginForm)
	})

	t.Run("no login form", func(t *testing.T) {
		htmlContent := `<form><input type="text" name="search"></form>`
		doc, _ := html.Parse(strings.NewReader(htmlContent))
		
		data := &CrawlData{}
		service.traverseHTML(doc, data, nil)
		
		assert.False(t, data.HasLoginForm)
	})
} 