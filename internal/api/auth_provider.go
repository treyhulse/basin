package api

import (
	"fmt"
	"net/http"
	"time"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"
	"go-rbac-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthProviderService handles authentication sessions and provides centralized auth context
type AuthProviderService struct {
	db  *db.DB
	cfg *config.Config
}

// NewAuthProviderService creates a new auth provider service
func NewAuthProviderService(db *db.DB, cfg *config.Config) *AuthProviderService {
	return &AuthProviderService{
		db:  db,
		cfg: cfg,
	}
}

// CreateSession creates a new tenant-scoped authentication session
func (s *AuthProviderService) CreateSession(userID, tenantID uuid.UUID) (*middleware.Session, error) {
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(s.cfg.JWTExpiry)

	// Store session in database (you could add a sessions table for this)
	// For now, we'll just return the session object
	session := &middleware.Session{
		ID:        sessionID,
		UserID:    userID,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	return session, nil
}

// GetSession retrieves the current session from the context
func (s *AuthProviderService) GetSession(c *gin.Context) (*middleware.AuthProvider, bool) {
	return middleware.GetAuthProvider(c)
}

// GetCurrentUser retrieves the current authenticated user
func (s *AuthProviderService) GetCurrentUser(c *gin.Context) (*sqlc.User, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	user, err := s.db.Queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetCurrentTenant retrieves the current tenant context
func (s *AuthProviderService) GetCurrentTenant(c *gin.Context) (*sqlc.Tenant, error) {
	tenantID, exists := middleware.GetTenantID(c)
	if !exists || tenantID == uuid.Nil {
		return nil, fmt.Errorf("no tenant context")
	}

	tenant, err := s.db.Queries.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// SwitchTenant allows a user to switch to a different tenant they have access to
func (s *AuthProviderService) SwitchTenant(c *gin.Context, newTenantID uuid.UUID) (*middleware.AuthProvider, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	// Check if user has access to the new tenant
	_, err := s.db.Queries.GetUserTenant(c.Request.Context(), sqlc.GetUserTenantParams{
		UserID:   userID,
		TenantID: newTenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("user does not have access to this tenant")
	}

	// Get the new tenant
	tenant, err := s.db.Queries.GetTenantByID(c.Request.Context(), newTenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Get user roles and permissions for the new tenant
	userRoles, err := s.db.Queries.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if user is admin
	isAdmin := false
	roles := make([]string, 0, len(userRoles))
	for _, role := range userRoles {
		roles = append(roles, role.Name)
		if role.Name == "admin" {
			isAdmin = true
		}
	}

	// Get user permissions for the new tenant
	var permissions []string
	userPermissions, err := s.db.Queries.GetPermissionsByUserAndTenant(c.Request.Context(), sqlc.GetPermissionsByUserAndTenantParams{
		UserID:   userID,
		TenantID: uuid.NullUUID{UUID: newTenantID, Valid: true},
	})
	if err == nil {
		permissions = make([]string, 0, len(userPermissions))
		for _, perm := range userPermissions {
			permissions = append(permissions, fmt.Sprintf("%s:%s", perm.TableName, perm.Action))
		}
	}

	// Create new auth provider
	authProvider := &middleware.AuthProvider{
		UserID:      userID,
		TenantID:    newTenantID,
		TenantSlug:  tenant.Slug,
		IsAdmin:     isAdmin,
		Roles:       roles,
		Permissions: permissions,
		SessionID:   uuid.New().String(),
		ExpiresAt:   time.Now().Add(s.cfg.JWTExpiry),
	}

	// Update context with new tenant info
	c.Set("auth", authProvider)
	c.Set("tenant_id", newTenantID)
	c.Set("tenant_slug", tenant.Slug)

	return authProvider, nil
}

// GetUserTenants retrieves all tenants the current user has access to
func (s *AuthProviderService) GetUserTenants(c *gin.Context) ([]sqlc.Tenant, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	tenants, err := s.db.Queries.GetUserTenants(c.Request.Context(), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	return tenants, nil
}

// ValidatePermission checks if the current user has a specific permission
func (s *AuthProviderService) ValidatePermission(c *gin.Context, tableName, action string) bool {
	auth, exists := middleware.GetAuthProvider(c)
	if !exists {
		return false
	}

	// Admin bypass
	if auth.IsAdmin {
		return true
	}

	// Check for specific permission
	requiredPermission := fmt.Sprintf("%s:%s", tableName, action)
	for _, permission := range auth.Permissions {
		if permission == requiredPermission {
			return true
		}
	}

	return false
}

// ValidateRole checks if the current user has a specific role
func (s *AuthProviderService) ValidateRole(c *gin.Context, roleName string) bool {
	auth, exists := middleware.GetAuthProvider(c)
	if !exists {
		return false
	}

	// Admin bypass
	if auth.IsAdmin {
		return true
	}

	// Check for specific role
	for _, role := range auth.Roles {
		if role == roleName {
			return true
		}
	}

	return false
}

// RequirePermission creates a middleware that requires a specific permission
func (s *AuthProviderService) RequirePermission(tableName, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.ValidatePermission(c, tableName, action) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Insufficient permissions: %s:%s", tableName, action),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func (s *AuthProviderService) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.ValidateRole(c, roleName) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Insufficient role: %s", roleName),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireTenant creates a middleware that requires a tenant context
func (s *AuthProviderService) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := middleware.GetTenantID(c)
		if !exists || tenantID == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAuthContext returns a comprehensive auth context object
func (s *AuthProviderService) GetAuthContext(c *gin.Context) (map[string]interface{}, error) {
	auth, exists := middleware.GetAuthProvider(c)
	if !exists {
		return nil, fmt.Errorf("user not authenticated")
	}

	user, err := s.GetCurrentUser(c)
	if err != nil {
		return nil, err
	}

	var tenant *sqlc.Tenant
	if auth.TenantID != uuid.Nil {
		tenant, err = s.GetCurrentTenant(c)
		if err != nil {
			return nil, err
		}
	}

	context := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName.String,
			"last_name":  user.LastName.String,
			"is_active":  user.IsActive.Bool,
			"created_at": user.CreatedAt.Time,
			"updated_at": user.UpdatedAt.Time,
		},
		"session": map[string]interface{}{
			"id":         auth.SessionID,
			"expires_at": auth.ExpiresAt,
		},
		"auth": map[string]interface{}{
			"is_admin":    auth.IsAdmin,
			"roles":       auth.Roles,
			"permissions": auth.Permissions,
		},
	}

	if tenant != nil {
		context["tenant"] = map[string]interface{}{
			"id":         tenant.ID,
			"name":       tenant.Name,
			"slug":       tenant.Slug,
			"domain":     tenant.Domain.String,
			"is_active":  tenant.IsActive.Bool,
			"created_at": tenant.CreatedAt.Time,
			"updated_at": tenant.UpdatedAt.Time,
		}
	}

	return context, nil
}
