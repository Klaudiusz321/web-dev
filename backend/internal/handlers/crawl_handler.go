package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"web-crawler-backend/internal/models"
	"web-crawler-backend/internal/services"
)

type CrawlHandler struct {
	crawlerService *services.CrawlerService
}

func NewCrawlHandler(crawlerService *services.CrawlerService) *CrawlHandler {
	return &CrawlHandler{crawlerService: crawlerService}
}

// StartCrawl handles POST /api/v1/crawl/:id
func (h *CrawlHandler) StartCrawl(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL ID",
			"message": "ID must be a valid number",
		})
		return
	}

	// Start crawling in background
	go h.crawlerService.StartCrawl(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Crawling started",
		"url_id":  id,
	})
}

// GetCrawlStatus handles GET /api/v1/crawl/status/:id
func (h *CrawlHandler) GetCrawlStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL ID",
			"message": "ID must be a valid number",
		})
		return
	}

	status, err := h.crawlerService.GetCrawlStatus(uint(id))
	if err != nil {
		if err.Error() == "URL not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "URL not found",
				"message": "The requested URL does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get crawl status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}

// BulkRerunCrawls handles POST /api/v1/crawl/bulk-rerun
func (h *CrawlHandler) BulkRerunCrawls(c *gin.Context) {
	var req models.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No IDs provided",
			"message": "At least one URL ID must be provided",
		})
		return
	}

	if err := h.crawlerService.BulkRerunCrawls(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to rerun crawls",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Crawls restarted successfully",
	})
} 