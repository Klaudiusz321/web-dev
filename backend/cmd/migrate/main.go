package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"web-crawler-backend/internal/config"
	"web-crawler-backend/internal/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse command line flags
	var (
		action = flag.String("action", "up", "Migration action: up, down, version")
		steps  = flag.Int("steps", 1, "Number of steps for down migration")
	)
	flag.Parse()

	// Initialize configuration
	cfg := config.Load()

	switch *action {
	case "up":
		if err := database.RunMigrationsWithFiles(cfg.DatabaseURL); err != nil {
			log.Fatal("Failed to run migrations up:", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		for i := 0; i < *steps; i++ {
			if err := database.RollbackMigration(cfg.DatabaseURL); err != nil {
				log.Fatal("Failed to rollback migration:", err)
			}
		}
		fmt.Printf("Rolled back %d migration(s) successfully\n", *steps)

	case "version":
		version, dirty, err := database.GetMigrationVersion(cfg.DatabaseURL)
		if err != nil {
			log.Fatal("Failed to get migration version:", err)
		}
		fmt.Printf("Current migration version: %d\n", version)
		if dirty {
			fmt.Println("Warning: Migration state is dirty")
		}

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: up, down, version")
		os.Exit(1)
	}
} 