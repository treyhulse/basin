package api

import (
	"database/sql"
	"net/http"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"
	"go-rbac-api/internal/models"

	"go-rbac-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type TenantHandler struct {
	db  *db.DB
	cfg *config.Config
}

func NewTenantHandler(db *db.DB, cfg *config.Config) *TenantHandler {
	return &TenantHandler{
		db:  db,
		cfg: cfg,
	}
}

// CreateTenant handles POST /tenants requests
// @Summary      Create Tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        body  body   models.CreateTenantRequest true "Tenant creation payload"
// @Success      201   {object} models.TenantResponse
// @Failure      400   {object} map[string]string
// @Failure      409   {object} map[string]string
// @Router       /tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var createReq models.CreateTenantRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Check if tenant slug already exists
	existingTenant, err := h.db.Queries.GetTenantBySlug(c.Request.Context(), createReq.Slug)
	if err == nil && existingTenant.ID != uuid.Nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tenant with this slug already exists"})
		return
	}

	// Generate UUID for new tenant
	tenantID := uuid.New()

	// Create tenant in database
	tenant, err := h.db.Queries.CreateTenant(c.Request.Context(), sqlc.CreateTenantParams{
		ID:       tenantID,
		Name:     createReq.Name,
		Slug:     createReq.Slug,
		Domain:   sql.NullString{String: createReq.Domain, Valid: createReq.Domain != ""},
		Settings: pqtype.NullRawMessage{Valid: false},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, models.TenantResponse{
		Message: "Tenant created successfully",
		Tenant: models.Tenant{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			Domain:    tenant.Domain.String,
			IsActive:  tenant.IsActive.Bool,
			CreatedAt: tenant.CreatedAt.Time,
			UpdatedAt: tenant.UpdatedAt.Time,
		},
	})
}

// GetTenants handles GET /tenants requests
// @Summary      Get All Tenants
// @Tags         tenants
// @Produce      json
// @Success      200 {array} models.Tenant
// @Failure      500 {object} map[string]string
// @Router       /tenants [get]
func (h *TenantHandler) GetTenants(c *gin.Context) {
	tenants, err := h.db.Queries.GetAllTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}

	var response []models.Tenant
	for _, tenant := range tenants {
		response = append(response, models.Tenant{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			Domain:    tenant.Domain.String,
			IsActive:  tenant.IsActive.Bool,
			CreatedAt: tenant.CreatedAt.Time,
			UpdatedAt: tenant.UpdatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetTenant handles GET /tenants/:id requests
// @Summary      Get Tenant by ID
// @Tags         tenants
// @Produce      json
// @Param        id    path     string true "Tenant ID"
// @Success      200   {object} models.Tenant
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, models.Tenant{
		ID:        tenant.ID,
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Domain:    tenant.Domain.String,
		IsActive:  tenant.IsActive.Bool,
		CreatedAt: tenant.CreatedAt.Time,
		UpdatedAt: tenant.UpdatedAt.Time,
	})
}

// UpdateTenant handles PUT /tenants/:id requests
// @Summary      Update Tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id    path     string true "Tenant ID"
// @Param        body  body     models.UpdateTenantRequest true "Update payload"
// @Success      200   {object} models.Tenant
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var updateReq models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get existing tenant
	existingTenant, err := h.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Update fields if provided
	if updateReq.Name != nil {
		existingTenant.Name = *updateReq.Name
	}
	if updateReq.Slug != nil {
		existingTenant.Slug = *updateReq.Slug
	}
	if updateReq.Domain != nil {
		existingTenant.Domain.String = *updateReq.Domain
		existingTenant.Domain.Valid = *updateReq.Domain != ""
	}

	// Update tenant in database
	updatedTenant, err := h.db.Queries.UpdateTenant(c.Request.Context(), sqlc.UpdateTenantParams{
		ID:       tenantID,
		Name:     existingTenant.Name,
		Slug:     existingTenant.Slug,
		Domain:   existingTenant.Domain,
		Settings: existingTenant.Settings,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, models.Tenant{
		ID:        updatedTenant.ID,
		Name:      updatedTenant.Name,
		Slug:      updatedTenant.Slug,
		Domain:    updatedTenant.Domain.String,
		IsActive:  updatedTenant.IsActive.Bool,
		CreatedAt: updatedTenant.CreatedAt.Time,
		UpdatedAt: updatedTenant.UpdatedAt.Time,
	})
}

// DeleteTenant handles DELETE /tenants/:id requests
// @Summary      Delete Tenant
// @Tags         tenants
// @Produce      json
// @Param        id    path     string true "Tenant ID"
// @Success      200   {object} models.DeleteTenantResponse
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Check if tenant exists
	_, err = h.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Delete tenant
	err = h.db.Queries.DeleteTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	c.JSON(http.StatusOK, models.DeleteTenantResponse{
		Message: "Tenant deleted successfully",
	})
}

// AddUserToTenant handles POST /tenants/:id/users requests
// @Summary      Add User to Tenant
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id    path     string true "Tenant ID"
// @Param        body  body     models.AddUserToTenantRequest true "Add user payload"
// @Success      200   {object} models.UserTenantResponse
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /tenants/{id}/users [post]
func (h *TenantHandler) AddUserToTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var addReq models.AddUserToTenantRequest
	if err := c.ShouldBindJSON(&addReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify tenant exists
	_, err = h.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Verify user exists
	_, err = h.db.Queries.GetUserByID(c.Request.Context(), addReq.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Add user to tenant
	err = h.db.Queries.AddUserToTenant(c.Request.Context(), sqlc.AddUserToTenantParams{
		UserID:   addReq.UserID,
		TenantID: tenantID,
		RoleID:   uuid.NullUUID{UUID: addReq.RoleID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to tenant"})
		return
	}

	c.JSON(http.StatusOK, models.UserTenantResponse{
		Message: "User added to tenant successfully",
	})
}

// RemoveUserFromTenant handles DELETE /tenants/:id/users/:user_id requests
// @Summary      Remove User from Tenant
// @Tags         tenants
// @Produce      json
// @Param        id       path     string true "Tenant ID"
// @Param        user_id  path     string true "User ID"
// @Success      200      {object} models.UserTenantResponse
// @Failure      400      {object} map[string]string
// @Router       /tenants/{id}/users/{user_id} [delete]
func (h *TenantHandler) RemoveUserFromTenant(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Remove user from tenant
	err = h.db.Queries.RemoveUserFromTenant(c.Request.Context(), sqlc.RemoveUserFromTenantParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from tenant"})
		return
	}

	c.JSON(http.StatusOK, models.UserTenantResponse{
		Message: "User removed from tenant successfully",
	})
}

// JoinTenant handles POST /tenants/:id/join requests
// @Summary      Join Tenant (Current User)
// @Tags         tenants
// @Accept       json
// @Produce      json
// @Param        id    path     string true "Tenant ID"
// @Success      200   {object} models.UserTenantResponse
// @Failure      400   {object} map[string]string
// @Failure      404   {object} map[string]string
// @Router       /tenants/{id}/join [post]
func (h *TenantHandler) JoinTenant(c *gin.Context) {
	// Get current user from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Verify tenant exists
	_, err = h.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Check if user is already in this tenant
	_, err = h.db.Queries.GetUserTenant(c.Request.Context(), sqlc.GetUserTenantParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err == nil {
		// User is already in this tenant
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this tenant"})
		return
	}

	// Add user to tenant without a specific role (role_id will be NULL)
	err = h.db.Queries.AddUserToTenant(c.Request.Context(), sqlc.AddUserToTenantParams{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   uuid.NullUUID{Valid: false}, // No role assigned initially
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join tenant"})
		return
	}

	c.JSON(http.StatusOK, models.UserTenantResponse{
		Message: "Successfully joined tenant",
	})
}
