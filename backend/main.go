package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"web-crawler-backend/internal/config"
	"web-crawler-backend/internal/database"
	"web-crawler-backend/internal/handlers"
	"web-crawler-backend/internal/middleware"
	"web-crawler-backend/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations (use GORM AutoMigrate for development, file-based for production)
	if cfg.Environment == "production" {
		if err := database.RunMigrationsWithFiles(cfg.DatabaseURL); err != nil {
			log.Printf("File-based migrations failed, falling back to AutoMigrate: %v", err)
			if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
				log.Fatal("Failed to run migrations:", err)
			}
		}
	} else {
		if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatal("Failed to run migrations:", err)
		}
	}

	// Initialize services
	authService := services.NewAuthService(db)
	crawlerService := services.NewCrawlerService(db)
	urlService := services.NewURLService(db, crawlerService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	urlHandler := handlers.NewURLHandler(urlService)
	crawlHandler := handlers.NewCrawlHandler(crawlerService)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	// Setup middleware
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// Setup routes
	setupRoutes(router, authHandler, authService, urlHandler, crawlHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, authService *services.AuthService, urlHandler *handlers.URLHandler, crawlHandler *handlers.CrawlHandler) {
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Auth endpoints (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			// Protected auth endpoints
			auth.GET("/profile", middleware.AuthRequired(authService), authHandler.GetProfile)
			auth.POST("/logout", middleware.AuthRequired(authService), authHandler.Logout)
			auth.GET("/validate", middleware.AuthRequired(authService), authHandler.ValidateToken)
		}

		// URL endpoints (protected)
		urls := api.Group("/urls")
		urls.Use(middleware.AuthRequired(authService))
		{
			urls.GET("", urlHandler.GetURLs)
			urls.POST("", urlHandler.CreateURL)
			urls.GET("/:id", urlHandler.GetURL)
			urls.GET("/:id/links", urlHandler.GetURLLinks)
			urls.DELETE("/:id", urlHandler.DeleteURL)
			urls.POST("/bulk-delete", urlHandler.BulkDeleteURLs)
		}

		// Crawl endpoints (protected)
		crawl := api.Group("/crawl")
		crawl.Use(middleware.AuthRequired(authService))
		{
			crawl.POST("/:id", crawlHandler.StartCrawl)
			crawl.GET("/status/:id", crawlHandler.GetCrawlStatus)
			crawl.POST("/bulk-rerun", crawlHandler.BulkRerunCrawls)
		}
	}
} 