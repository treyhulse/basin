package api

import (
	"database/sql"
	"net/http"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"
	"go-rbac-api/internal/middleware"
	"go-rbac-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
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

// SignUp handles POST /auth/signup requests
// @Summary      Sign Up
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body   models.SignUpRequest true "Sign up payload"
// @Success      201   {object} models.SignUpResponse
// @Failure      400   {object} map[string]string
// @Failure      409   {object} map[string]string
// @Router       /auth/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	var signUpReq models.SignUpRequest
	if err := c.ShouldBindJSON(&signUpReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Check if user already exists
	existingUser, err := h.db.Queries.GetUserByEmail(c.Request.Context(), signUpReq.Email)
	if err == nil && existingUser.ID != uuid.Nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	passwordHash, err := models.HashPassword(signUpReq.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Generate UUID for new user
	userID := uuid.New()

	// Create user in database
	user, err := h.db.Queries.CreateUser(c.Request.Context(), sqlc.CreateUserParams{
		ID:           userID,
		Email:        signUpReq.Email,
		PasswordHash: passwordHash,
		FirstName:    sql.NullString{String: signUpReq.FirstName, Valid: true},
		LastName:     sql.NullString{String: signUpReq.LastName, Valid: true},
		TenantID:     uuid.NullUUID{}, // No tenant for now, can be updated later
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Handle tenant creation or joining if specified
	if signUpReq.TenantSlug != "" {
		// Check if tenant exists
		tenant, err := h.db.Queries.GetTenantBySlug(c.Request.Context(), signUpReq.TenantSlug)
		if err != nil {
			// Tenant doesn't exist, create it
			tenantID := uuid.New()
			tenant, err = h.db.Queries.CreateTenant(c.Request.Context(), sqlc.CreateTenantParams{
				ID:       tenantID,
				Name:     signUpReq.TenantSlug, // Use slug as name initially
				Slug:     signUpReq.TenantSlug,
				Domain:   sql.NullString{Valid: false},
				Settings: pqtype.NullRawMessage{}, // No settings initially
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
				return
			}
		}

		// TODO: Add user to tenant with default role
		// This requires the AddUserToTenant query to be implemented
		// For now, we'll just create the tenant and user separately
		_ = tenant // Use tenant variable to avoid linter error
	}

	// Return success response
	c.JSON(http.StatusCreated, models.SignUpResponse{
		Message: "User created successfully",
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

// UpdateUser handles PUT /auth/users/:id requests
// @Summary      Update User
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        id    path     string true "User ID"
// @Param        body  body     models.UpdateUserRequest true "Update payload"
// @Success      200   {object} models.User
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Failure      500   {object} map[string]string
// @Router       /auth/users/{id} [put]
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateReq models.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get existing user
	existingUser, err := h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields if provided
	if updateReq.FirstName != nil {
		existingUser.FirstName.String = *updateReq.FirstName
		existingUser.FirstName.Valid = true
	}
	if updateReq.LastName != nil {
		existingUser.LastName.String = *updateReq.LastName
		existingUser.LastName.Valid = true
	}
	if updateReq.IsActive != nil {
		existingUser.IsActive.Bool = *updateReq.IsActive
		existingUser.IsActive.Valid = true
	}

	// Update user in database
	updatedUser, err := h.db.Queries.UpdateUser(c.Request.Context(), sqlc.UpdateUserParams{
		ID:        userID,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		IsActive:  existingUser.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, models.User{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		FirstName: updatedUser.FirstName.String,
		LastName:  updatedUser.LastName.String,
		IsActive:  updatedUser.IsActive.Bool,
		CreatedAt: updatedUser.CreatedAt.Time,
		UpdatedAt: updatedUser.UpdatedAt.Time,
	})
}

// DeleteUser handles DELETE /auth/users/:id requests
// @Summary      Delete User
// @Tags         auth
// @Produce      json
// @Param        id    path     string true "User ID"
// @Success      200   {object} models.DeleteUserResponse
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Failure      500   {object} map[string]string
// @Router       /auth/users/{id} [delete]
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user exists
	_, err = h.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	err = h.db.Queries.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, models.DeleteUserResponse{
		Message: "User deleted successfully",
	})
}
