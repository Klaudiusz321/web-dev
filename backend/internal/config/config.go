package config

import (
	"os"
)

type Config struct {
	Environment string
	DatabaseURL string
	Port        string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "root:password@tcp(localhost:3306)/webcrawler?charset=utf8mb4&parseTime=True&loc=Local"),
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-here"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 