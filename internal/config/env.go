package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type DeploymentMode string

const (
	DeploymentModeLocal   DeploymentMode = "local"
	DeploymentModeRailway DeploymentMode = "railway"
	DeploymentModeAuto    DeploymentMode = "auto"
)

type Config struct {
	DeploymentMode DeploymentMode

	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DatabaseURL       string // For Railway compatibility
	DatabasePublicURL string // Railway provides this for external access

	JWTSecret string
	JWTExpiry time.Duration

	ServerPort int
	ServerMode string
}

func Load() (*Config, error) {
	// Always try to load .env file first (for both local and Railway)
	// This ensures DATABASE_URL is available for auto-detection
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	// Determine deployment mode AFTER loading .env
	deploymentMode := getDeploymentMode()

	fmt.Printf("=== DEPLOYMENT MODE: %s ===\n", deploymentMode)

	config := &Config{
		DeploymentMode: deploymentMode,

		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvAsInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "go_rbac_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),        // Railway provides this
		DatabasePublicURL: getEnv("DATABASE_PUBLIC_URL", ""), // Railway provides this for external access

		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		JWTExpiry: getEnvAsDuration("JWT_EXPIRY", 24*time.Hour),

		ServerPort: getEnvAsInt("SERVER_PORT", 8080),
		ServerMode: getEnv("SERVER_MODE", "debug"),
	}

	// Debug: Print all environment variables at startup
	fmt.Printf("=== ENVIRONMENT VARIABLES DEBUG ===\n")
	fmt.Printf("DATABASE_URL: %s\n", os.Getenv("DATABASE_URL"))
	fmt.Printf("DATABASE_PUBLIC_URL: %s\n", os.Getenv("DATABASE_PUBLIC_URL"))
	fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("DB_PORT: %s\n", os.Getenv("DB_PORT"))
	fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_PASSWORD: %s\n", os.Getenv("DB_PASSWORD"))
	fmt.Printf("DB_NAME: %s\n", os.Getenv("DB_NAME"))
	fmt.Printf("DB_SSLMODE: %s\n", os.Getenv("DB_SSLMODE"))
	fmt.Printf("PORT: %s\n", os.Getenv("PORT"))
	fmt.Printf("===================================\n")

	// Handle database configuration based on deployment mode
	if err := config.configureDatabase(); err != nil {
		return nil, fmt.Errorf("failed to configure database: %w", err)
	}

	// Debug: Print the actual values being loaded
	fmt.Printf("DEBUG: Final configuration:\n")
	fmt.Printf("  DEPLOYMENT_MODE: %s\n", config.DeploymentMode)
	fmt.Printf("  DB_HOST=%s, DB_PORT=%d, DB_USER=%s, DB_NAME=%s, DB_SSLMODE=%s\n",
		config.DBHost, config.DBPort, config.DBUser, config.DBName, config.DBSSLMode)

	return config, nil
}

// getDeploymentMode determines the current deployment environment
func getDeploymentMode() DeploymentMode {
	// Check for explicit override
	if mode := os.Getenv("DEPLOYMENT_MODE"); mode != "" {
		switch strings.ToLower(mode) {
		case "local":
			return DeploymentModeLocal
		case "railway":
			return DeploymentModeRailway
		}
	}

	// Auto-detect based on environment
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("DATABASE_URL") != "" {
		return DeploymentModeRailway
	}

	// Default to local if no indicators found
	return DeploymentModeLocal
}

// configureDatabase sets up database configuration based on deployment mode
func (c *Config) configureDatabase() error {
	switch c.DeploymentMode {
	case DeploymentModeRailway:
		return c.configureRailwayDatabase()
	case DeploymentModeLocal:
		return c.configureLocalDatabase()
	default:
		return fmt.Errorf("unknown deployment mode: %s", c.DeploymentMode)
	}
}

// configureRailwayDatabase sets up Railway-specific database configuration
func (c *Config) configureRailwayDatabase() error {
	fmt.Printf("Configuring database for Railway deployment...\n")

	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required for Railway deployment")
	}

	// Parse DATABASE_URL and override individual settings
	if err := c.parseDatabaseURL(); err != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Force SSL mode for Railway
	c.DBSSLMode = "require"

	fmt.Printf("Railway database configured: host=%s, port=%d, db=%s, ssl=%s\n",
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)

	return nil
}

// configureLocalDatabase sets up local development database configuration
func (c *Config) configureLocalDatabase() error {
	fmt.Printf("Configuring database for local development...\n")

	// Use individual environment variables or defaults
	// .env file should have been loaded if it exists

	fmt.Printf("Local database configured: host=%s, port=%d, db=%s, ssl=%s\n",
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)

	return nil
}

// parseDatabaseURL parses Railway's DATABASE_URL format
func (c *Config) parseDatabaseURL() error {
	parsedURL, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// Extract host and port
	if parsedURL.Host != "" {
		hostParts := strings.Split(parsedURL.Host, ":")
		c.DBHost = hostParts[0]
		if len(hostParts) > 1 {
			if port, err := strconv.Atoi(hostParts[1]); err == nil {
				c.DBPort = port
			}
		}
	}

	// Extract username and password
	if parsedURL.User != nil {
		c.DBUser = parsedURL.User.Username()
		if password, ok := parsedURL.User.Password(); ok {
			c.DBPassword = password
		}
	}

	// Extract database name
	if parsedURL.Path != "" {
		c.DBName = strings.TrimPrefix(parsedURL.Path, "/")
	}

	// Set SSL mode based on scheme
	if parsedURL.Scheme == "postgresql" || parsedURL.Scheme == "postgres" {
		// Railway typically requires SSL
		c.DBSSLMode = "require"
	}

	return nil
}

func (c *Config) GetDBConnString() string {
	fmt.Printf("DEBUG: GetDBConnString called with DEPLOYMENT_MODE=%s\n", c.DeploymentMode)

	// For Railway deployment, we should always have DATABASE_URL
	if c.DeploymentMode == DeploymentModeRailway {
		if c.DatabaseURL != "" {
			fmt.Printf("DEBUG: Using Railway DATABASE_URL (internal)\n")
			return c.DatabaseURL
		}
		// If we're in Railway mode but don't have DATABASE_URL, this is an error
		fmt.Printf("ERROR: Railway mode requires DATABASE_URL but none was provided\n")
		return ""
	}

	// For local development, try DATABASE_PUBLIC_URL first (for testing Railway from local)
	if c.DatabasePublicURL != "" {
		fmt.Printf("DEBUG: Using Railway DATABASE_PUBLIC_URL (external)\n")
		return c.DatabasePublicURL
	}

	// Fallback to local database configuration
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
	fmt.Printf("DEBUG: Local connection string: %s\n", connStr)
	return connStr
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
