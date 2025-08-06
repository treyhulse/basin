package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Basin API is running",
		})
	})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response["message"], "Basin API")
}

// TestAuthFlow tests the complete authentication flow
func TestAuthFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router with auth endpoints
	router := gin.New()

	// Mock login endpoint
	router.POST("/auth/login", func(c *gin.Context) {
		var loginReq struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock successful authentication
		if loginReq.Email == "admin@example.com" && loginReq.Password == "password" {
			c.JSON(http.StatusOK, gin.H{
				"token": "mock-jwt-token",
				"user": gin.H{
					"id":         "123e4567-e89b-12d3-a456-426614174000",
					"email":      loginReq.Email,
					"first_name": "Admin",
					"last_name":  "User",
					"is_active":  true,
				},
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	})

	// Mock /auth/me endpoint
	router.GET("/auth/me", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || authHeader != "Bearer mock-jwt-token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         "123e4567-e89b-12d3-a456-426614174000",
			"email":      "admin@example.com",
			"first_name": "Admin",
			"last_name":  "User",
			"is_active":  true,
		})
	})

	t.Run("Successful login", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "admin@example.com",
			"password": "password",
		}

		jsonData, _ := json.Marshal(loginData)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")
		assert.Equal(t, "mock-jwt-token", response["token"])
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "admin@example.com",
			"password": "wrongpassword",
		}

		jsonData, _ := json.Marshal(loginData)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Get current user with valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req.Header.Set("Authorization", "Bearer mock-jwt-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "admin@example.com", response["email"])
	})

	t.Run("Get current user without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestItemsEndpoints tests the items API endpoints
func TestItemsEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Mock items endpoints
	router.GET("/items/:table", func(c *gin.Context) {
		table := c.Param("table")

		// Validate table name
		validTables := map[string]bool{
			"products":  true,
			"customers": true,
			"orders":    true,
		}

		if !validTables[table] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
			return
		}

		// Mock response
		mockData := []map[string]interface{}{
			{
				"id":   "123e4567-e89b-12d3-a456-426614174000",
				"name": "Test Product",
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"data": mockData,
			"meta": gin.H{
				"total": 1,
				"page":  1,
			},
		})
	})

	router.GET("/items/:table/:id", func(c *gin.Context) {
		table := c.Param("table")
		id := c.Param("id")

		// Basic validation
		if table == "" || id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing table or id"})
			return
		}

		// Mock single item response
		c.JSON(http.StatusOK, gin.H{
			"data": map[string]interface{}{
				"id":   id,
				"name": "Test Item",
			},
		})
	})

	t.Run("Get items from valid table", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/items/products", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "meta")
	})

	t.Run("Get items from invalid table", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/items/invalid_table", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Get single item", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/items/products/123e4567-e89b-12d3-a456-426614174000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "data")
	})
}

// TestCORSHeaders tests that CORS headers are properly set
func TestCORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	t.Run("CORS headers are set", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("OPTIONS request returns 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
