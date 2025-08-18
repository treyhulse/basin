package api

import (
	"net/http"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	"go-rbac-api/internal/middleware"
	"go-rbac-api/internal/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	db  *db.DB
	cfg *config.Config
}

func NewAuthHandler(db *db.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:  db,
		cfg: cfg,
	}
}

// Login handles POST /auth/login requests
// @Summary      Login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body   models.LoginRequest true "Login payload"
// @Success      200   {object} models.LoginResponse
// @Failure      400   {object} map[string]string
// @Failure      401   {object} map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get user from database
	user, err := h.db.Queries.GetUserByEmail(c.Request.Context(), loginReq.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive.Bool {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Verify password
	if !models.CheckPassword(loginReq.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(models.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User: models.User{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName.String,
			LastName:  user.LastName.String,
			IsActive:  user.IsActive.Bool,
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		},
	})
}

// Me handles GET /auth/me requests to get current user info
// @Summary      Get current user
// @Tags         auth
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve information about the currently authenticated user. Requires valid JWT Bearer token or API key.
// @Produce      json
// @Success      200 {object} models.User
// @Failure      401 {object} models.ErrorResponse
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user from database
	user, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		IsActive:  user.IsActive.Bool,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	})
}
