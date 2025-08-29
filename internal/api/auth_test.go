package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	"go-rbac-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_SignUp(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock config
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	// Create a mock database (this will be nil, but we're just testing the handler structure)
	var mockDB *db.DB

	// Create the auth handler
	handler := NewAuthHandler(mockDB, cfg)
	assert.NotNil(t, handler)

	// Test invalid request body
	t.Run("Invalid Request Body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/auth/signup", handler.SignUp)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test missing required fields
	t.Run("Missing Required Fields", func(t *testing.T) {
		signUpReq := models.SignUpRequest{
			Email: "test@example.com",
			// Missing password, first_name, last_name
		}

		reqBody, _ := json.Marshal(signUpReq)
		req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := gin.New()
		router.POST("/auth/signup", handler.SignUp)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test handler creation and basic structure
	t.Run("Handler Structure", func(t *testing.T) {
		assert.NotNil(t, handler.db)
		assert.NotNil(t, handler.cfg)
		assert.Equal(t, cfg, handler.cfg)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock config
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	// Create a mock database
	var mockDB *db.DB

	// Create the auth handler
	handler := NewAuthHandler(mockDB, cfg)
	assert.NotNil(t, handler)

	// Test handler creation and basic structure
	t.Run("Handler Structure", func(t *testing.T) {
		assert.NotNil(t, handler.db)
		assert.NotNil(t, handler.cfg)
		assert.Equal(t, cfg, handler.cfg)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock config
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}

	// Create a mock database
	var mockDB *db.DB

	// Create the auth handler
	handler := NewAuthHandler(mockDB, cfg)
	assert.NotNil(t, handler)

	// Test handler creation and basic structure
	t.Run("Handler Structure", func(t *testing.T) {
		assert.NotNil(t, handler.db)
		assert.NotNil(t, handler.cfg)
		assert.Equal(t, cfg, handler.cfg)
	})
}
