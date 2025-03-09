package config

import (
	"fmt"
	"os"
	"strconv"
)

// AppConfig holds all application configurations
type AppConfig struct {
	Server      ServerConfig
	Database    DatabaseConfig
	RateLimiter RateLimiterConfig
}

// ServerConfig holds HTTP server related configurations
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database related configurations
type DatabaseConfig struct {
	URL string
}

// RateLimiterConfig holds rate limiter related configurations
type RateLimiterConfig struct {
	Enabled     bool
	MaxRequests int
}

// LoadConfig loads application configurations from environment variables
func LoadConfig() (*AppConfig, error) {
	config := &AppConfig{
		Server: ServerConfig{
			Port: getEnvWithDefault("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		RateLimiter: RateLimiterConfig{},
	}

	rateLimitStatus := getEnvWithDefault("RATE_LIMITER", "enabled")
	if rateLimitStatus == "enabled" {
		config.RateLimiter.Enabled = true
	} else {
		config.RateLimiter.Enabled = false
	}

	maxRequests := getEnvWithDefault("RATE_LIMITER_MAX_REQUESTS", "10")
	if parsed, err := strconv.Atoi(maxRequests); err == nil && parsed > 0 {
		config.RateLimiter.MaxRequests = parsed
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateConfig checks if all required configurations are set
func validateConfig(config *AppConfig) error {
	if config.Database.URL == "" {
		return fmt.Errorf("database URL is required (set DATABASE_URL environment variable)")
	}

	if config.RateLimiter.Enabled && config.RateLimiter.MaxRequests == 0 {
		return fmt.Errorf("rate limiter max requests must be greater than zero")
	}

	return nil
}
