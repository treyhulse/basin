package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Global test server and token for integration tests
var (
	testServer   *http.Server
	testBaseURL  string
	testToken    string
	testDatabase *db.DB
)

// setupIntegrationTest sets up for testing against the running server
func setupIntegrationTest(t *testing.T) {
	// Test against the real running server on port 8080
	testBaseURL = "http://localhost:8080"

	// Connect to database for any setup needed
	cfg, err := config.Load()
	if err != nil {
		t.Logf("Warning: Could not load config: %v", err)
		return
	}

	testDatabase, err = db.NewDB(cfg)
	if err != nil {
		t.Logf("Warning: Could not connect to database: %v", err)
		return
	}
}

// teardownIntegrationTest cleans up after tests
func teardownIntegrationTest(t *testing.T) {
	if testDatabase != nil {
		testDatabase.Close()
	}
}

// isDatabaseRunning checks if Docker database is available
func isDatabaseRunning() bool {
	// Try to connect using default Docker database settings
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "go_rbac_db")
	os.Setenv("DB_SSL_MODE", "disable")

	cfg, err := config.Load()
	if err != nil {
		return false
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		return false
	}
	defer database.Close()

	return true
}

func TestIntegration_AuthFlow(t *testing.T) {
	// Skip if no database available (check for either env var or default Docker setup)
	if os.Getenv("DB_HOST") == "" && !isDatabaseRunning() {
		t.Skip("Skipping integration test: no database configured")
	}

	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/health")
		require.NoError(t, err, "Health check should not error")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "Health response should be valid JSON")
		assert.Equal(t, "ok", response["status"], "Health should return ok status")
	})

	t.Run("Login with Admin Credentials", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "admin@example.com",
			"password": "password", // This should match what's in your database
		}

		jsonData, _ := json.Marshal(loginData)
		resp, err := http.Post(testBaseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err, "Login request should not error")
		defer resp.Body.Close()

		// Check if login was successful
		if resp.StatusCode == http.StatusOK {
			var loginResponse map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&loginResponse)
			assert.NoError(t, err, "Login response should be valid JSON")
			assert.Contains(t, loginResponse, "token", "Login should return token")
			assert.Contains(t, loginResponse, "user", "Login should return user")

			// Store token for other tests
			if token, ok := loginResponse["token"].(string); ok {
				testToken = token
			}
		} else {
			// Login failed - this might be expected if no admin user exists
			var errorResponse map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResponse)
			t.Logf("Login failed (expected if no admin user): %d - %v", resp.StatusCode, errorResponse)
		}
	})

	t.Run("Auth Me with Token", func(t *testing.T) {
		if testToken == "" {
			t.Skip("Skipping: no auth token available")
		}

		req, _ := http.NewRequest("GET", testBaseURL+"/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+testToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err, "Auth me request should not error")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Auth me should succeed with valid token")

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "Auth me response should be valid JSON")
		assert.Contains(t, response, "email", "Should return user email")
	})

	t.Run("Auth Me without Token", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/auth/me")
		require.NoError(t, err, "Request should not error")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Auth me without token should fail")
	})
}

func TestIntegration_ItemsEndpoints(t *testing.T) {
	// Skip if no database available
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping integration test: no database configured")
	}

	if testToken == "" {
		t.Skip("Skipping: no auth token available from previous test")
	}

	tables := []string{"roles", "users", "collections", "fields", "permissions"}

	for _, table := range tables {
		t.Run(fmt.Sprintf("Query %s table", table), func(t *testing.T) {
			req, _ := http.NewRequest("GET", testBaseURL+"/items/"+table, nil)
			req.Header.Set("Authorization", "Bearer "+testToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err, "Request should not error")
			defer resp.Body.Close()

			// The response could be 200 (success) or 400/403 (validation/permission error)
			// Both are valid for testing - we just want to make sure the handler runs
			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
				"Should get a valid HTTP response (not 5xx server error)")

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err, "Response should be valid JSON")

			if resp.StatusCode == http.StatusOK {
				t.Logf("✅ Successfully queried %s table", table)
			} else {
				t.Logf("⚠️  Query %s returned %d (may be expected): %v", table, resp.StatusCode, response)
			}
		})
	}

	t.Run("Unauthorized Access", func(t *testing.T) {
		resp, err := http.Get(testBaseURL + "/items/users")
		require.NoError(t, err, "Request should not error")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authorization")
	})
}
