package api

import (
	"context"
	"database/sql"
	"fmt"
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

// CreateTenant handles POST /tenants requests with full initialization
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

	// Get the creating user
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Generate UUID for new tenant
	tenantID := uuid.New()

	// Start a database transaction for atomicity
	tx, err := h.db.DB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

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

	// Initialize tenant with default roles, permissions, and collections
	if err := h.initializeTenant(c.Request.Context(), tenantID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize tenant: " + err.Error()})
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, models.TenantResponse{
		Message: "Tenant created and initialized successfully",
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

// initializeTenant sets up a new tenant with default roles, permissions, and collections
func (h *TenantHandler) initializeTenant(ctx context.Context, tenantID uuid.UUID, creatorUserID uuid.UUID) error {
	// 1. Create default roles
	roles, err := h.createDefaultRoles(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to create default roles: %w", err)
	}

	// 2. Add creator as admin to the tenant
	adminRole := roles["admin"]
	if err := h.db.Queries.AddUserToTenant(ctx, sqlc.AddUserToTenantParams{
		UserID:   creatorUserID,
		TenantID: tenantID,
		RoleID:   uuid.NullUUID{UUID: adminRole.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("failed to add creator to tenant: %w", err)
	}

	// 3. Add admin role to user_roles table
	if err := h.db.Queries.AddUserRole(ctx, sqlc.AddUserRoleParams{
		UserID: creatorUserID,
		RoleID: adminRole.ID,
	}); err != nil {
		return fmt.Errorf("failed to assign admin role to user: %w", err)
	}

	// 4. Create default permissions for system tables
	if err := h.createDefaultPermissions(ctx, tenantID, roles); err != nil {
		return fmt.Errorf("failed to create default permissions: %w", err)
	}

	// 5. Create default collections
	if err := h.createDefaultCollections(ctx, tenantID, creatorUserID); err != nil {
		return fmt.Errorf("failed to create default collections: %w", err)
	}

	return nil
}

// createDefaultRoles creates the standard roles for a new tenant
func (h *TenantHandler) createDefaultRoles(ctx context.Context, tenantID uuid.UUID) (map[string]sqlc.Role, error) {
	roles := make(map[string]sqlc.Role)

	defaultRoles := []struct {
		name        string
		description string
	}{
		{"admin", "Full system access and management"},
		{"manager", "Can manage users, content, and settings"},
		{"editor", "Can create and edit content"},
		{"viewer", "Can view content and data"},
	}

	for _, roleData := range defaultRoles {
		roleID := uuid.New()
		role, err := h.db.Queries.CreateRole(ctx, sqlc.CreateRoleParams{
			ID:          roleID,
			Name:        roleData.name,
			Description: sql.NullString{String: roleData.description, Valid: true},
			TenantID:    uuid.NullUUID{UUID: tenantID, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create role %s: %w", roleData.name, err)
		}
		roles[roleData.name] = role
	}

	return roles, nil
}

// createDefaultPermissions creates standard permissions for system tables
func (h *TenantHandler) createDefaultPermissions(ctx context.Context, tenantID uuid.UUID, roles map[string]sqlc.Role) error {
	// System tables that need permissions
	systemTables := []string{
		"users", "roles", "permissions", "collections", "fields",
		"tenants", "api_keys", "user_tenants", "user_roles",
	}

	// Define role permissions
	rolePermissions := map[string][]string{
		"admin":   {"create", "read", "update", "delete"},
		"manager": {"create", "read", "update"},
		"editor":  {"create", "read", "update"},
		"viewer":  {"read"},
	}

	for roleName, role := range roles {
		permissions, exists := rolePermissions[roleName]
		if !exists {
			continue
		}

		for _, table := range systemTables {
			for _, action := range permissions {
				permissionID := uuid.New()
				_, err := h.db.Queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
					ID:            permissionID,
					RoleID:        uuid.NullUUID{UUID: role.ID, Valid: true},
					TableName:     table,
					Action:        action,
					FieldFilter:   pqtype.NullRawMessage{Valid: false},
					AllowedFields: []string{"*"}, // Full field access for system tables
					TenantID:      uuid.NullUUID{UUID: tenantID, Valid: true},
				})
				if err != nil {
					return fmt.Errorf("failed to create permission %s:%s for role %s: %w",
						table, action, roleName, err)
				}
			}
		}
	}

	return nil
}

// createDefaultCollections creates some useful default collections for the tenant
func (h *TenantHandler) createDefaultCollections(ctx context.Context, tenantID uuid.UUID, creatorUserID uuid.UUID) error {
	defaultCollections := []struct {
		name        string
		displayName string
		description string
		icon        string
	}{
		{
			name:        "customers",
			displayName: "Customers",
			description: "Customer information and contact details",
			icon:        "👥",
		},
		{
			name:        "products",
			displayName: "Products",
			description: "Product catalog and inventory",
			icon:        "📦",
		},
		{
			name:        "orders",
			displayName: "Orders",
			description: "Customer orders and transactions",
			icon:        "📋",
		},
	}

	for _, collectionData := range defaultCollections {
		collectionID := uuid.New()
		_, err := h.db.Queries.CreateCollection(ctx, sqlc.CreateCollectionParams{
			ID:          collectionID,
			Name:        collectionData.displayName, // Display name (e.g., "Customers")
			Slug:        collectionData.name,        // URL-friendly slug (e.g., "customers")
			DisplayName: sql.NullString{String: collectionData.displayName, Valid: true},
			Description: sql.NullString{String: collectionData.description, Valid: true},
			Icon:        sql.NullString{String: collectionData.icon, Valid: true},
			IsSystem:    sql.NullBool{Bool: false, Valid: true},
			TenantID:    uuid.NullUUID{UUID: tenantID, Valid: true},
			CreatedBy:   uuid.NullUUID{UUID: creatorUserID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", collectionData.name, err)
		}

		// Add default fields for each collection
		if err := h.createDefaultFields(ctx, collectionID, collectionData.name, tenantID); err != nil {
			return fmt.Errorf("failed to create fields for collection %s: %w", collectionData.name, err)
		}
	}

	return nil
}

// createDefaultFields creates standard fields for a collection
func (h *TenantHandler) createDefaultFields(ctx context.Context, collectionID uuid.UUID, collectionName string, tenantID uuid.UUID) error {
	// Define default fields based on collection type
	var defaultFields []struct {
		name        string
		displayName string
		type_       string
		isRequired  bool
		isPrimary   bool
		sortOrder   int32
	}

	switch collectionName {
	case "customers":
		defaultFields = []struct {
			name        string
			displayName string
			type_       string
			isRequired  bool
			isPrimary   bool
			sortOrder   int32
		}{
			{"name", "Name", "string", true, true, 1},
			{"email", "Email", "string", true, false, 2},
			{"phone", "Phone", "string", false, false, 3},
			{"address", "Address", "text", false, false, 4},
		}
	case "products":
		defaultFields = []struct {
			name        string
			displayName string
			type_       string
			isRequired  bool
			isPrimary   bool
			sortOrder   int32
		}{
			{"name", "Product Name", "string", true, true, 1},
			{"description", "Description", "text", false, false, 2},
			{"price", "Price", "decimal", true, false, 3},
			{"sku", "SKU", "string", true, false, 4},
			{"stock", "Stock Quantity", "integer", false, false, 5},
		}
	case "orders":
		defaultFields = []struct {
			name        string
			displayName string
			type_       string
			isRequired  bool
			isPrimary   bool
			sortOrder   int32
		}{
			{"order_number", "Order Number", "string", true, true, 1},
			{"customer_id", "Customer", "uuid", true, false, 2},
			{"total_amount", "Total Amount", "decimal", true, false, 3},
			{"status", "Status", "string", true, false, 4},
			{"order_date", "Order Date", "datetime", true, false, 5},
		}
	default:
		// Generic fields for any collection
		defaultFields = []struct {
			name        string
			displayName string
			type_       string
			isRequired  bool
			isPrimary   bool
			sortOrder   int32
		}{
			{"name", "Name", "string", true, true, 1},
			{"description", "Description", "text", false, false, 2},
		}
	}

	for _, fieldData := range defaultFields {
		fieldID := uuid.New()
		_, err := h.db.Queries.CreateField(ctx, sqlc.CreateFieldParams{
			ID:              fieldID,
			CollectionID:    uuid.NullUUID{UUID: collectionID, Valid: true},
			Name:            fieldData.name,
			DisplayName:     sql.NullString{String: fieldData.displayName, Valid: true},
			Type:            fieldData.type_,
			IsPrimary:       sql.NullBool{Bool: fieldData.isPrimary, Valid: true},
			IsRequired:      sql.NullBool{Bool: fieldData.isRequired, Valid: true},
			IsUnique:        sql.NullBool{Bool: false, Valid: true},
			DefaultValue:    sql.NullString{Valid: false},
			ValidationRules: pqtype.NullRawMessage{Valid: false},
			RelationConfig:  pqtype.NullRawMessage{Valid: false},
			SortOrder:       sql.NullInt32{Int32: fieldData.sortOrder, Valid: true},
			TenantID:        uuid.NullUUID{UUID: tenantID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create field %s: %w", fieldData.name, err)
		}
	}

	return nil
}
