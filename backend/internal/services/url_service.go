package services

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"web-crawler-backend/internal/models"
)

// CrawlerServiceInterface defines the interface for crawler service
type CrawlerServiceInterface interface {
	StartCrawl(urlID uint)
	GetCrawlStatus(urlID uint) (*models.CrawlStatusResponse, error)
	BulkRerunCrawls(urlIDs []uint) error
}

type URLService struct {
	db             *gorm.DB
	crawlerService CrawlerServiceInterface
}

func NewURLService(db *gorm.DB, crawlerService CrawlerServiceInterface) *URLService {
	return &URLService{
		db:             db,
		crawlerService: crawlerService,
	}
}

// CreateURL creates a new URL record and starts crawling
func (s *URLService) CreateURL(url string) (*models.URL, error) {
	// Try to create new URL first
	urlRecord := &models.URL{
		URL:    url,
		Status: "pending",
	}

	err := s.db.Create(urlRecord).Error
	if err == nil {
		// Successfully created new URL, start crawling
		go s.crawlerService.StartCrawl(urlRecord.ID)
		return urlRecord, nil
	}

	// Check if error is due to duplicate key constraint
	if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
		// URL already exists (might be soft-deleted), try to fetch it including deleted records
		var existingURL models.URL
		if fetchErr := s.db.Unscoped().Where("url = ?", url).First(&existingURL).Error; fetchErr != nil {
			return nil, fmt.Errorf("failed to fetch existing URL after duplicate error: %w", fetchErr)
		}

		// If URL was soft-deleted, restore it
		if existingURL.DeletedAt.Valid {
			existingURL.DeletedAt = gorm.DeletedAt{}
		}

		// Update status and restart crawling
		existingURL.Status = "pending"
		if updateErr := s.db.Unscoped().Save(&existingURL).Error; updateErr != nil {
			return nil, fmt.Errorf("failed to update existing URL status: %w", updateErr)
		}
		
		// Restart crawling process
		go s.crawlerService.StartCrawl(existingURL.ID)
		
		return &existingURL, nil
	}

	// Some other database error occurred
	return nil, fmt.Errorf("failed to create URL record: %w", err)
}

// GetURLs retrieves URLs with pagination, filtering, and sorting
func (s *URLService) GetURLs(limit, offset int, search, status, sortBy, sortOrder string) ([]*models.URL, int64, error) {
	var urls []*models.URL
	var total int64

	// Build query
	query := s.db.Model(&models.URL{})

	// Apply search filter
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(url) LIKE ? OR LOWER(title) LIKE ?", searchPattern, searchPattern)
	}

	// Apply status filter
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count URLs: %w", err)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortOrder))
	query = query.Order(orderClause)

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	// Preload related data
	query = query.Preload("Crawls", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Preload("Links")

	// Execute query
	if err := query.Find(&urls).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch URLs: %w", err)
	}

	return urls, total, nil
}

// GetURL retrieves a single URL by ID with full details
func (s *URLService) GetURL(id uint) (*models.URL, error) {
	var url models.URL

	if err := s.db.
		Preload("Crawls", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("Links", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_accessible = ?", false)
		}).
		First(&url, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("URL not found")
		}
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	return &url, nil
}

// DeleteURL soft deletes a URL by ID
func (s *URLService) DeleteURL(id uint) error {
	if err := s.db.Delete(&models.URL{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	return nil
}

// BulkDeleteURLs soft deletes multiple URLs
func (s *URLService) BulkDeleteURLs(ids []uint) error {
	if err := s.db.Delete(&models.URL{}, ids).Error; err != nil {
		return fmt.Errorf("failed to bulk delete URLs: %w", err)
	}
	return nil
}

// GetURLLinks retrieves links for a specific URL with filtering
func (s *URLService) GetURLLinks(urlID uint, linkType string, limit, offset int) ([]*models.Link, int64, error) {
	var links []*models.Link
	var total int64

	// Verify URL exists
	var url models.URL
	if err := s.db.First(&url, urlID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("URL not found")
		}
		return nil, 0, fmt.Errorf("failed to verify URL: %w", err)
	}

	// Build query
	query := s.db.Model(&models.Link{}).Where("url_id = ?", urlID)

	// Apply link type filter
	switch linkType {
	case "internal":
		query = query.Where("link_type = ?", "internal")
	case "external":
		query = query.Where("link_type = ?", "external")
	case "broken":
		query = query.Where("is_accessible = ?", false)
	case "accessible":
		query = query.Where("is_accessible = ?", true)
	// "all" or empty - no additional filter
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count links: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC").Limit(limit).Offset(offset)

	// Execute query
	if err := query.Find(&links).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch links: %w", err)
	}

	return links, total, nil
} 