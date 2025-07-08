package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"web-crawler-backend/internal/models"
)

func setupURLTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate all models
	err = db.AutoMigrate(&models.URL{}, &models.Crawl{}, &models.Link{}, &models.User{})
	require.NoError(t, err)

	return db
}

// Mock crawler service for testing
type mockCrawlerService struct {
	startCrawlCalled bool
	lastURLID        uint
}

func (m *mockCrawlerService) StartCrawl(urlID uint) {
	m.startCrawlCalled = true
	m.lastURLID = urlID
}

func (m *mockCrawlerService) GetCrawlStatus(urlID uint) (*models.CrawlStatusResponse, error) {
	return &models.CrawlStatusResponse{
		ID:     1,
		URL:    "https://example.com",
		Status: "completed",
	}, nil
}

func (m *mockCrawlerService) BulkRerunCrawls(urlIDs []uint) error {
	return nil
}

func TestNewURLService(t *testing.T) {
	db := setupURLTestDB(t)
	crawlerService := &mockCrawlerService{}
	service := NewURLService(db, crawlerService)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, crawlerService, service.crawlerService)
}

func TestURLService_CreateURL(t *testing.T) {
	t.Run("successful URL creation", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		url, err := service.CreateURL("https://example.com")
		require.NoError(t, err)
		assert.NotNil(t, url)
		assert.Equal(t, "https://example.com", url.URL)
		assert.Equal(t, "pending", url.Status)
		assert.NotZero(t, url.ID)

		// Verify crawler was started
		time.Sleep(10 * time.Millisecond) // Allow goroutine to execute
		assert.True(t, crawlerService.startCrawlCalled)
		assert.Equal(t, url.ID, crawlerService.lastURLID)
	})

	t.Run("duplicate URL handling", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create first URL
		url1, err := service.CreateURL("https://example.com")
		require.NoError(t, err)

		// Reset mock
		crawlerService.startCrawlCalled = false

		// Try to create duplicate URL
		url2, err := service.CreateURL("https://example.com")
		require.NoError(t, err)
		assert.Equal(t, url1.ID, url2.ID)
		assert.Equal(t, "pending", url2.Status)

		// Verify crawler was restarted
		time.Sleep(10 * time.Millisecond)
		assert.True(t, crawlerService.startCrawlCalled)
	})

	t.Run("restore soft-deleted URL", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create and delete URL
		url, err := service.CreateURL("https://example.com")
		require.NoError(t, err)
		
		err = service.DeleteURL(url.ID)
		require.NoError(t, err)

		// Reset mock
		crawlerService.startCrawlCalled = false

		// Try to create the same URL again
		restoredURL, err := service.CreateURL("https://example.com")
		require.NoError(t, err)
		assert.Equal(t, url.ID, restoredURL.ID)
		assert.Equal(t, "pending", restoredURL.Status)

		// Verify it's not soft-deleted anymore
		var urlRecord models.URL
		err = db.First(&urlRecord, url.ID).Error
		require.NoError(t, err)
		assert.False(t, urlRecord.DeletedAt.Valid)
	})
}

func TestURLService_GetURLs(t *testing.T) {
	t.Run("basic pagination", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URLs
		urls := []*models.URL{
			{URL: "https://example1.com", Title: "Example 1", Status: "completed"},
			{URL: "https://example2.com", Title: "Example 2", Status: "pending"},
			{URL: "https://example3.com", Title: "Example 3", Status: "completed"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}

		// Get first page
		result, total, err := service.GetURLs(2, 0, "", "", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 2)

		// Get second page
		result, total, err = service.GetURLs(2, 2, "", "", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 1)
	})

	t.Run("search filtering", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URLs
		urls := []*models.URL{
			{URL: "https://google.com", Title: "Google Search", Status: "completed"},
			{URL: "https://github.com", Title: "GitHub", Status: "pending"},
			{URL: "https://golang.org", Title: "Go Programming", Status: "completed"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}

		// Search by URL
		result, total, err := service.GetURLs(10, 0, "google", "", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0].URL, "google")

		// Search by title
		result, total, err = service.GetURLs(10, 0, "programming", "", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0].Title, "Programming")
	})

	t.Run("status filtering", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URLs with different statuses
		urls := []*models.URL{
			{URL: "https://completed1.com", Status: "completed"},
			{URL: "https://completed2.com", Status: "completed"},
			{URL: "https://pending1.com", Status: "pending"},
			{URL: "https://error1.com", Status: "error"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}

		// Filter by completed status
		result, total, err := service.GetURLs(10, 0, "", "completed", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
		for _, url := range result {
			assert.Equal(t, "completed", url.Status)
		}

		// Filter by pending status
		result, total, err = service.GetURLs(10, 0, "", "pending", "created_at", "desc")
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "pending", result[0].Status)
	})

	t.Run("sorting", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URLs with different titles
		urls := []*models.URL{
			{URL: "https://c.com", Title: "C Title"},
			{URL: "https://a.com", Title: "A Title"},
			{URL: "https://b.com", Title: "B Title"},
		}
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
		}

		// Sort by title ascending
		result, _, err := service.GetURLs(10, 0, "", "", "title", "asc")
		require.NoError(t, err)
		require.Len(t, result, 3)
		
		// Check that titles are sorted alphabetically
		titles := []string{result[0].Title, result[1].Title, result[2].Title}
		assert.Equal(t, "A Title", titles[0])
		assert.Equal(t, "B Title", titles[1])
		assert.Equal(t, "C Title", titles[2])
	})
}

func TestURLService_GetURL(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URL with crawls and links
		url := &models.URL{
			URL:         "https://example.com",
			Title:       "Example Site",
			Status:      "completed",
			HTMLVersion: "HTML5",
		}
		require.NoError(t, db.Create(url).Error)

		// Create crawl
		crawl := &models.Crawl{
			URLID:  url.ID,
			Status: "completed",
		}
		require.NoError(t, db.Create(crawl).Error)

		// Create links (including broken ones)
		links := []*models.Link{
			{URLID: url.ID, LinkURL: "https://working.com", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://broken.com", IsAccessible: false},
		}
		for _, link := range links {
			require.NoError(t, db.Create(link).Error)
		}

		// Get URL
		result, err := service.GetURL(url.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, url.URL, result.URL)
		assert.Equal(t, url.Title, result.Title)
		assert.Len(t, result.Crawls, 1)
		
		// Check links if any exist
		if len(result.Links) > 0 {
			assert.False(t, result.Links[0].IsAccessible)
		}
	})

	t.Run("URL not found", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		result, err := service.GetURL(999)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "URL not found")
	})
}

func TestURLService_DeleteURL(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)

		// Delete URL
		err := service.DeleteURL(url.ID)
		require.NoError(t, err)

		// Verify soft deletion
		var deletedURL models.URL
		err = db.First(&deletedURL, url.ID).Error
		assert.Error(t, err) // Should not find it in normal query
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// But should find it with Unscoped
		err = db.Unscoped().First(&deletedURL, url.ID).Error
		require.NoError(t, err)
		assert.True(t, deletedURL.DeletedAt.Valid)
	})

	t.Run("delete non-existent URL", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		err := service.DeleteURL(999)
		assert.NoError(t, err) // Soft delete doesn't fail for non-existent records
	})
}

func TestURLService_BulkDeleteURLs(t *testing.T) {
	t.Run("successful bulk deletion", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URLs
		urls := []*models.URL{
			{URL: "https://example1.com", Status: "completed"},
			{URL: "https://example2.com", Status: "pending"},
			{URL: "https://example3.com", Status: "error"},
		}
		var ids []uint
		for _, url := range urls {
			require.NoError(t, db.Create(url).Error)
			ids = append(ids, url.ID)
		}

		// Bulk delete
		err := service.BulkDeleteURLs(ids)
		require.NoError(t, err)

		// Verify all are soft deleted
		var count int64
		db.Model(&models.URL{}).Count(&count)
		assert.Equal(t, int64(0), count)

		// But should find them with Unscoped
		db.Unscoped().Model(&models.URL{}).Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("empty IDs list", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		err := service.BulkDeleteURLs([]uint{})
		assert.Error(t, err) // Should fail with empty list
		assert.Contains(t, err.Error(), "WHERE conditions required")
	})
}

func TestURLService_GetURLLinks(t *testing.T) {
	t.Run("successful link retrieval", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)

		// Create test links
		links := []*models.Link{
			{URLID: url.ID, LinkURL: "https://example.com/page1", LinkType: "internal", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://example.com/page2", LinkType: "internal", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://external.com", LinkType: "external", IsAccessible: true},
			{URLID: url.ID, LinkURL: "https://broken.com", LinkType: "external", IsAccessible: false},
		}
		for _, link := range links {
			require.NoError(t, db.Create(link).Error)
		}

		// Get all links
		result, total, err := service.GetURLLinks(url.ID, "all", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
		assert.Len(t, result, 4)

		// Get internal links only
		result, total, err = service.GetURLLinks(url.ID, "internal", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
		for _, link := range result {
			assert.Equal(t, "internal", link.LinkType)
		}

		// Get external links only
		result, total, err = service.GetURLLinks(url.ID, "external", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
		for _, link := range result {
			assert.Equal(t, "external", link.LinkType)
		}

		// Get broken links only
		result, total, err = service.GetURLLinks(url.ID, "broken", 10, 0)
		require.NoError(t, err)
		brokenCount := 0
		for _, link := range result {
			if !link.IsAccessible {
				brokenCount++
			}
		}
		assert.Equal(t, int64(brokenCount), total)
		
		// Get accessible links only
		result, total, err = service.GetURLLinks(url.ID, "accessible", 10, 0)
		require.NoError(t, err)
		accessibleCount := 0
		for _, link := range result {
			if link.IsAccessible {
				accessibleCount++
			}
		}
		assert.Equal(t, int64(accessibleCount), total)
	})

	t.Run("pagination", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		// Create test URL
		url := &models.URL{URL: "https://example.com", Status: "completed"}
		require.NoError(t, db.Create(url).Error)

		// Create many links
		for i := 0; i < 5; i++ {
			link := &models.Link{
				URLID:        url.ID,
				LinkURL:      fmt.Sprintf("https://example.com/page%d", i),
				LinkType:     "internal",
				IsAccessible: true,
			}
			require.NoError(t, db.Create(link).Error)
		}

		// Get first page
		result, total, err := service.GetURLLinks(url.ID, "all", 2, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 2)

		// Get second page
		result, total, err = service.GetURLLinks(url.ID, "all", 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 2)

		// Get third page
		result, total, err = service.GetURLLinks(url.ID, "all", 2, 4)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, result, 1)
	})

	t.Run("URL not found", func(t *testing.T) {
		db := setupURLTestDB(t)
		crawlerService := &mockCrawlerService{}
		service := NewURLService(db, crawlerService)

		result, total, err := service.GetURLLinks(999, "all", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), total)
		assert.Contains(t, err.Error(), "URL not found")
	})
} 