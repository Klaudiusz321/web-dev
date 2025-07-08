package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"type:varchar(191);uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"type:varchar(191);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"` // Hidden from JSON responses
	FirstName string    `json:"first_name" gorm:"type:varchar(191)"`
	LastName  string    `json:"last_name" gorm:"type:varchar(191)"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// URL represents a website URL to be crawled
type URL struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	URL         string    `json:"url" gorm:"not null;unique"`
	Title       string    `json:"title"`
	HTMLVersion string    `json:"html_version"`
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, running, completed, error
	HasLoginForm bool     `json:"has_login_form" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Crawls []Crawl `json:"crawls,omitempty" gorm:"foreignKey:URLID"`
	Links  []Link  `json:"links,omitempty" gorm:"foreignKey:URLID"`
}

// Crawl represents a crawling session for a URL
type Crawl struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	URLID         uint       `json:"url_id" gorm:"not null"`
	Status        string     `json:"status" gorm:"default:'queued'"` // queued, running, completed, error
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ErrorMessage  string     `json:"error_message"`
	InternalLinks int        `json:"internal_links" gorm:"default:0"`
	ExternalLinks int        `json:"external_links" gorm:"default:0"`
	BrokenLinks   int        `json:"broken_links" gorm:"default:0"`
	HeadingCounts string     `json:"heading_counts"` // JSON string: {"h1":1,"h2":3,...}
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relationships
	URL   URL    `json:"url,omitempty" gorm:"foreignKey:URLID"`
	Links []Link `json:"links,omitempty" gorm:"foreignKey:CrawlID"`
}

// Link represents a link found during crawling
type Link struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	URLID       uint   `json:"url_id" gorm:"not null"`
	CrawlID     uint   `json:"crawl_id" gorm:"not null"`
	LinkURL     string `json:"link_url" gorm:"not null"`
	LinkText    string `json:"link_text"`
	LinkType    string `json:"link_type"` // internal, external
	StatusCode  int    `json:"status_code"`
	IsAccessible bool  `json:"is_accessible" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	URL   URL   `json:"url,omitempty" gorm:"foreignKey:URLID"`
	Crawl Crawl `json:"crawl,omitempty" gorm:"foreignKey:CrawlID"`
}

// HeadingCounts represents the count of heading tags
type HeadingCounts struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

// CrawlRequest represents the request to start crawling
type CrawlRequest struct {
	URL string `json:"url" binding:"required"`
}

// CrawlStatusResponse represents the crawl status response
type CrawlStatusResponse struct {
	ID            uint           `json:"id"`
	URL           string         `json:"url"`
	Status        string         `json:"status"`
	InternalLinks int            `json:"internal_links"`
	ExternalLinks int            `json:"external_links"`
	BrokenLinks   int            `json:"broken_links"`
	HeadingCounts *HeadingCounts `json:"heading_counts"`
	StartedAt     *time.Time     `json:"started_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	ErrorMessage  string         `json:"error_message,omitempty"`
}

// BulkRequest represents bulk action requests
type BulkRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// Authentication-related structs
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// JWT Claims structure
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.StandardClaims
} 