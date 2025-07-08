package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"golang.org/x/net/html"
	"web-crawler-backend/internal/models"
)

type CrawlerService struct {
	db *gorm.DB
}

// Ensure CrawlerService implements CrawlerServiceInterface
var _ CrawlerServiceInterface = (*CrawlerService)(nil)

func NewCrawlerService(db *gorm.DB) *CrawlerService {
	return &CrawlerService{db: db}
}

// StartCrawl initiates the crawling process for a URL
func (s *CrawlerService) StartCrawl(urlID uint) {
	// Get URL record
	var urlRecord models.URL
	if err := s.db.First(&urlRecord, urlID).Error; err != nil {
		log.Printf("Failed to find URL record %d: %v", urlID, err)
		return
	}

	// Create crawl record
	crawl := &models.Crawl{
		URLID:     urlID,
		Status:    "running",
		StartedAt: &time.Time{},
	}
	*crawl.StartedAt = time.Now()

	if err := s.db.Create(crawl).Error; err != nil {
		log.Printf("Failed to create crawl record: %v", err)
		return
	}

	// Update URL status
	s.db.Model(&urlRecord).Update("status", "running")

	// Perform crawling
	s.performCrawl(&urlRecord, crawl)
}

// performCrawl does the actual crawling work
func (s *CrawlerService) performCrawl(urlRecord *models.URL, crawl *models.Crawl) {
	defer func() {
		// Complete crawl
		now := time.Now()
		crawl.CompletedAt = &now
		s.db.Save(crawl)

		// Update URL status
		urlRecord.Status = crawl.Status
		s.db.Save(urlRecord)
	}()

	// Make HTTP request
	resp, err := http.Get(urlRecord.URL)
	if err != nil {
		crawl.Status = "error"
		crawl.ErrorMessage = fmt.Sprintf("HTTP request failed: %v", err)
		log.Printf("Failed to fetch URL %s: %v", urlRecord.URL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		crawl.Status = "error"
		crawl.ErrorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		log.Printf("URL %s returned status %d", urlRecord.URL, resp.StatusCode)
		return
	}

	// Parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		crawl.Status = "error"
		crawl.ErrorMessage = fmt.Sprintf("HTML parsing failed: %v", err)
		log.Printf("Failed to parse HTML for URL %s: %v", urlRecord.URL, err)
		return
	}

	// Extract data
	data := s.extractData(doc, urlRecord.URL)

	// Update URL record
	urlRecord.Title = data.Title
	urlRecord.HTMLVersion = data.HTMLVersion
	urlRecord.HasLoginForm = data.HasLoginForm

	// Update crawl record
	crawl.InternalLinks = data.InternalLinks
	crawl.ExternalLinks = data.ExternalLinks
	crawl.BrokenLinks = data.BrokenLinks
	
	headingCountsJSON, _ := json.Marshal(data.HeadingCounts)
	crawl.HeadingCounts = string(headingCountsJSON)
	crawl.Status = "completed"

	// Save links
	for _, link := range data.Links {
		link.URLID = urlRecord.ID
		link.CrawlID = crawl.ID
		s.db.Create(&link)
	}
}

// CrawlData holds extracted data from crawling
type CrawlData struct {
	Title         string
	HTMLVersion   string
	HasLoginForm  bool
	HeadingCounts models.HeadingCounts
	InternalLinks int
	ExternalLinks int
	BrokenLinks   int
	Links         []models.Link
}

// extractData extracts relevant data from HTML document
func (s *CrawlerService) extractData(doc *html.Node, baseURL string) *CrawlData {
	data := &CrawlData{
		HTMLVersion:   "HTML5", // Default assumption
		HeadingCounts: models.HeadingCounts{},
		Links:         []models.Link{},
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("Failed to parse base URL %s: %v", baseURL, err)
		return data
	}

	s.traverseHTML(doc, data, parsedBaseURL)
	s.checkLinkAccessibility(data)

	return data
}

// traverseHTML recursively traverses HTML nodes to extract data
func (s *CrawlerService) traverseHTML(n *html.Node, data *CrawlData, baseURL *url.URL) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if data.Title == "" && n.FirstChild != nil {
				data.Title = strings.TrimSpace(n.FirstChild.Data)
			}
		case "h1":
			data.HeadingCounts.H1++
		case "h2":
			data.HeadingCounts.H2++
		case "h3":
			data.HeadingCounts.H3++
		case "h4":
			data.HeadingCounts.H4++
		case "h5":
			data.HeadingCounts.H5++
		case "h6":
			data.HeadingCounts.H6++
		case "a":
			s.processLink(n, data, baseURL)
		case "form":
			s.checkLoginForm(n, data)
		case "html":
			s.detectHTMLVersion(n, data)
		}
	}

	// Traverse children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.traverseHTML(c, data, baseURL)
	}
}

// processLink processes anchor tags and categorizes links
func (s *CrawlerService) processLink(n *html.Node, data *CrawlData, baseURL *url.URL) {
	var href, linkText string

	// Extract href attribute
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			href = attr.Val
			break
		}
	}

	if href == "" {
		return
	}

	// Extract link text
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		linkText = strings.TrimSpace(n.FirstChild.Data)
	}

	// Parse and resolve URL
	linkURL, err := url.Parse(href)
	if err != nil {
		return
	}

	resolvedURL := baseURL.ResolveReference(linkURL)
	
	// Determine link type
	linkType := "external"
	if resolvedURL.Host == baseURL.Host {
		linkType = "internal"
	}

	link := models.Link{
		LinkURL:      resolvedURL.String(),
		LinkText:     linkText,
		LinkType:     linkType,
		StatusCode:   0, // Will be set during accessibility check
		IsAccessible: true,
	}

	data.Links = append(data.Links, link)

	if linkType == "internal" {
		data.InternalLinks++
	} else {
		data.ExternalLinks++
	}
}

// checkLoginForm checks if the form might be a login form
func (s *CrawlerService) checkLoginForm(n *html.Node, data *CrawlData) {
	// Look for common login form indicators
	loginIndicators := []string{"login", "signin", "email", "username", "password"}
	
	// Check form attributes and content
	formHTML := s.nodeToString(n)
	lowerHTML := strings.ToLower(formHTML)
	
	for _, indicator := range loginIndicators {
		if strings.Contains(lowerHTML, indicator) {
			data.HasLoginForm = true
			return
		}
	}
}

// detectHTMLVersion detects HTML version from doctype or html tag
func (s *CrawlerService) detectHTMLVersion(n *html.Node, data *CrawlData) {
	// Simple HTML5 detection (most modern websites)
	data.HTMLVersion = "HTML5"
}

// checkLinkAccessibility checks if links are accessible
func (s *CrawlerService) checkLinkAccessibility(data *CrawlData) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i := range data.Links {
		link := &data.Links[i]
		
		// Skip checking internal links for now (to avoid self-crawling)
		if link.LinkType == "internal" {
			link.StatusCode = 200
			continue
		}

		// Make HEAD request to check accessibility
		resp, err := client.Head(link.LinkURL)
		if err != nil {
			link.StatusCode = 0
			link.IsAccessible = false
			data.BrokenLinks++
			continue
		}

		link.StatusCode = resp.StatusCode
		if resp.StatusCode >= 400 {
			link.IsAccessible = false
			data.BrokenLinks++
		}
	}
}

// nodeToString converts HTML node to string (simplified)
func (s *CrawlerService) nodeToString(n *html.Node) string {
	var buf strings.Builder
	s.renderNode(&buf, n)
	return buf.String()
}

func (s *CrawlerService) renderNode(buf *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.ElementNode:
		buf.WriteString("<" + n.Data)
		for _, attr := range n.Attr {
			buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
		}
		buf.WriteString(">")
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			s.renderNode(buf, c)
		}
		
		buf.WriteString("</" + n.Data + ">")
	case html.TextNode:
		buf.WriteString(n.Data)
	}
}

// GetCrawlStatus returns the status of a crawl
func (s *CrawlerService) GetCrawlStatus(urlID uint) (*models.CrawlStatusResponse, error) {
	var url models.URL
	if err := s.db.Preload("Crawls", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).First(&url, urlID).Error; err != nil {
		return nil, fmt.Errorf("URL not found")
	}

	if len(url.Crawls) == 0 {
		return &models.CrawlStatusResponse{
			ID:     urlID,
			URL:    url.URL,
			Status: "pending",
		}, nil
	}

	crawl := url.Crawls[0]
	
	// Parse heading counts
	var headingCounts models.HeadingCounts
	if crawl.HeadingCounts != "" {
		json.Unmarshal([]byte(crawl.HeadingCounts), &headingCounts)
	}

	return &models.CrawlStatusResponse{
		ID:            crawl.ID,
		URL:           url.URL,
		Status:        crawl.Status,
		InternalLinks: crawl.InternalLinks,
		ExternalLinks: crawl.ExternalLinks,
		BrokenLinks:   crawl.BrokenLinks,
		HeadingCounts: &headingCounts,
		StartedAt:     crawl.StartedAt,
		CompletedAt:   crawl.CompletedAt,
		ErrorMessage:  crawl.ErrorMessage,
	}, nil
}

// BulkRerunCrawls restarts crawling for multiple URLs
func (s *CrawlerService) BulkRerunCrawls(urlIDs []uint) error {
	for _, urlID := range urlIDs {
		go s.StartCrawl(urlID)
	}
	return nil
} 