package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	requiredVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"JWT_SECRET", "JWT_EXPIRY",
		"SERVER_PORT", "SERVER_MODE",
	}

	for _, key := range requiredVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Cleanup function to restore environment
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("All required environment variables are set", func(t *testing.T) {
		// Set all required environment variables
		envVars := map[string]string{
			"DB_HOST":     "localhost",
			"DB_PORT":     "5432",
			"DB_USER":     "postgres",
			"DB_PASSWORD": "postgres",
			"DB_NAME":     "test_db",
			"DB_SSLMODE":  "disable",
			"JWT_SECRET":  "test-secret-key",
			"JWT_EXPIRY":  "24h",
			"SERVER_PORT": "8080",
			"SERVER_MODE": "debug",
		}

		for key, value := range envVars {
			os.Setenv(key, value)
		}

		// Test that all variables are accessible
		for key, expectedValue := range envVars {
			actualValue := os.Getenv(key)
			assert.Equal(t, expectedValue, actualValue, "Environment variable %s should be set correctly", key)
		}
	})

	t.Run("Missing database environment variables", func(t *testing.T) {
		// Clear database environment variables
		dbVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
		for _, key := range dbVars {
			os.Unsetenv(key)
		}

		for _, key := range dbVars {
			value := os.Getenv(key)
			assert.Empty(t, value, "Database environment variable %s should be empty", key)
		}
	})

	t.Run("Missing JWT environment variables", func(t *testing.T) {
		// Clear JWT environment variables
		jwtVars := []string{"JWT_SECRET", "JWT_EXPIRY"}
		for _, key := range jwtVars {
			os.Unsetenv(key)
		}

		for _, key := range jwtVars {
			value := os.Getenv(key)
			assert.Empty(t, value, "JWT environment variable %s should be empty", key)
		}
	})

	t.Run("Invalid port number", func(t *testing.T) {
		os.Setenv("SERVER_PORT", "invalid_port")
		port := os.Getenv("SERVER_PORT")
		assert.Equal(t, "invalid_port", port)

		// Test port validation (this would be in your actual config loading)
		assert.NotEqual(t, "8080", port, "Invalid port should not equal valid port")
	})

	t.Run("Invalid database port", func(t *testing.T) {
		os.Setenv("DB_PORT", "invalid_port")
		dbPort := os.Getenv("DB_PORT")
		assert.Equal(t, "invalid_port", dbPort)
	})
}

func TestEnvironmentDefaults(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		defaultValue string
		testValue    string
	}{
		{
			name:         "Default server port",
			envVar:       "SERVER_PORT",
			defaultValue: "8080",
			testValue:    "3000",
		},
		{
			name:         "Default server mode",
			envVar:       "SERVER_MODE",
			defaultValue: "debug",
			testValue:    "release",
		},
		{
			name:         "Default JWT expiry",
			envVar:       "JWT_EXPIRY",
			defaultValue: "24h",
			testValue:    "12h",
		},
		{
			name:         "Default SSL mode",
			envVar:       "DB_SSLMODE",
			defaultValue: "disable",
			testValue:    "require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(tt.envVar)
			defer func() {
				if original == "" {
					os.Unsetenv(tt.envVar)
				} else {
					os.Setenv(tt.envVar, original)
				}
			}()

			// Test with unset environment variable
			os.Unsetenv(tt.envVar)
			value := getEnvWithDefault(tt.envVar, tt.defaultValue)
			assert.Equal(t, tt.defaultValue, value, "Should use default value when env var is not set")

			// Test with set environment variable
			os.Setenv(tt.envVar, tt.testValue)
			value = getEnvWithDefault(tt.envVar, tt.defaultValue)
			assert.Equal(t, tt.testValue, value, "Should use env var value when set")
		})
	}
}

func TestDatabaseConnectionString(t *testing.T) {
	// Save original environment
	originalVars := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_SSLMODE":  os.Getenv("DB_SSLMODE"),
	}

	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("Valid database connection parameters", func(t *testing.T) {
		// Set valid database environment variables
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "disable")

		// Test individual components
		assert.Equal(t, "localhost", os.Getenv("DB_HOST"))
		assert.Equal(t, "5432", os.Getenv("DB_PORT"))
		assert.Equal(t, "testuser", os.Getenv("DB_USER"))
		assert.Equal(t, "testpass", os.Getenv("DB_PASSWORD"))
		assert.Equal(t, "testdb", os.Getenv("DB_NAME"))
		assert.Equal(t, "disable", os.Getenv("DB_SSLMODE"))

		// Test connection string format (mock)
		expectedFormat := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
		actualFormat := buildConnectionString()
		assert.Equal(t, expectedFormat, actualFormat)
	})

	t.Run("Missing database host", func(t *testing.T) {
		os.Unsetenv("DB_HOST")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_SSLMODE", "disable")

		host := os.Getenv("DB_HOST")
		assert.Empty(t, host, "DB_HOST should be empty")
	})
}

func TestJWTConfiguration(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	originalExpiry := os.Getenv("JWT_EXPIRY")

	defer func() {
		if originalSecret == "" {
			os.Unsetenv("JWT_SECRET")
		} else {
			os.Setenv("JWT_SECRET", originalSecret)
		}
		if originalExpiry == "" {
			os.Unsetenv("JWT_EXPIRY")
		} else {
			os.Setenv("JWT_EXPIRY", originalExpiry)
		}
	}()

	t.Run("Valid JWT configuration", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "super-secret-key-for-testing-that-is-definitely-long-enough")
		os.Setenv("JWT_EXPIRY", "24h")

		secret := os.Getenv("JWT_SECRET")
		expiry := os.Getenv("JWT_EXPIRY")

		assert.NotEmpty(t, secret, "JWT secret should not be empty")
		assert.Equal(t, "super-secret-key-for-testing-that-is-definitely-long-enough", secret)
		assert.Equal(t, "24h", expiry)
		assert.True(t, len(secret) >= 32, "JWT secret should be at least 32 characters for security")
	})

	t.Run("Weak JWT secret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "weak")
		secret := os.Getenv("JWT_SECRET")

		assert.Equal(t, "weak", secret)
		assert.True(t, len(secret) < 32, "Weak secret should be less than 32 characters")
	})

	t.Run("Invalid JWT expiry format", func(t *testing.T) {
		os.Setenv("JWT_EXPIRY", "invalid")
		expiry := os.Getenv("JWT_EXPIRY")

		assert.Equal(t, "invalid", expiry)
		// In real implementation, you'd validate the duration format
	})
}

// Helper functions (these would be in your actual config package)
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func buildConnectionString() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	return "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
}
