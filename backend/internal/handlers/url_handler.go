package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"web-crawler-backend/internal/models"
	"web-crawler-backend/internal/services"
)

type URLHandler struct {
	urlService *services.URLService
}

func NewURLHandler(urlService *services.URLService) *URLHandler {
	return &URLHandler{urlService: urlService}
}

// GetURLs handles GET /api/v1/urls
func (h *URLHandler) GetURLs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	search := c.Query("search")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sortBy", "updated_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Validate sort parameters
	validSortColumns := map[string]bool{
		"url":          true,
		"title":        true,
		"status":       true,
		"html_version": true,
		"created_at":   true,
		"updated_at":   true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "updated_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Get URLs from service
	urls, total, err := h.urlService.GetURLs(limit, offset, search, status, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch URLs",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": urls,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// CreateURL handles POST /api/v1/urls
func (h *URLHandler) CreateURL(c *gin.Context) {
	var req models.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	// Create URL and start crawling
	url, err := h.urlService.CreateURL(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create URL",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    url,
		"message": "URL created and crawling started",
	})
}

// GetURL handles GET /api/v1/urls/:id
func (h *URLHandler) GetURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL ID",
			"message": "ID must be a valid number",
		})
		return
	}

	url, err := h.urlService.GetURL(uint(id))
	if err != nil {
		if err.Error() == "URL not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "URL not found",
				"message": "The requested URL does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch URL",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": url,
	})
}

// DeleteURL handles DELETE /api/v1/urls/:id
func (h *URLHandler) DeleteURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL ID",
			"message": "ID must be a valid number",
		})
		return
	}

	if err := h.urlService.DeleteURL(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete URL",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
	})
}

// BulkDeleteURLs handles POST /api/v1/urls/bulk-delete
func (h *URLHandler) BulkDeleteURLs(c *gin.Context) {
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

	if err := h.urlService.BulkDeleteURLs(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete URLs",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URLs deleted successfully",
	})
}

// GetURLLinks handles GET /api/v1/urls/:id/links
func (h *URLHandler) GetURLLinks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid URL ID",
			"message": "ID must be a valid number",
		})
		return
	}

	// Parse query parameters
	linkType := c.Query("type")     // all, internal, external, broken
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	links, total, err := h.urlService.GetURLLinks(uint(id), linkType, limit, offset)
	if err != nil {
		if err.Error() == "URL not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "URL not found",
				"message": "The requested URL does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch links",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": links,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
} 