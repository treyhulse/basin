// Package api provides HTTP handlers for the Basin API's core schema management functionality.
// This file contains comprehensive Swagger documentation for all core schema tables.
package api

import (
	"github.com/gin-gonic/gin"
)

// Core Tables API Documentation
// These endpoints provide CRUD operations for the core schema management tables.
// All endpoints require authentication and appropriate permissions.
// These tables are accessed via the /items/:table endpoints, not separate /core/ endpoints.

// =============================================================================
// USERS MANAGEMENT
// =============================================================================

// GetUsers handles GET /items/users requests
// @Summary      List users
// @Tags         core-users
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of users in the system. Requires authentication and user management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (email, first_name, last_name, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        email    query  string false "Filter by email"
// @Param        is_active query bool   false "Filter by active status"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/users [get]
func GetUsers(c *gin.Context) {
	// This is just for Swagger documentation
	// The actual implementation is handled by ItemsHandler.GetItems
}

// GetUser handles GET /items/users/:id requests
// @Summary      Get user
// @Tags         core-users
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific user by ID. Requires authentication and user management permissions.
// @Param        id   path      string true "User ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/users/{id} [get]
func GetUser(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreateUser handles POST /items/users requests
// @Summary      Create user
// @Tags         core-users
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new user in the system. Requires authentication and user creation permissions.
// @Param        body body map[string]interface{} true "User data (email, first_name, last_name, password, role_id, tenant_id)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/users [post]
func CreateUser(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdateUser handles PUT /items/users/:id requests
// @Summary      Update user
// @Tags         core-users
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing user. Requires authentication and user update permissions.
// @Param        id   path      string true "User ID (UUID)"
// @Param        body body map[string]interface{} true "User data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/users/{id} [put]
func UpdateUser(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeleteUser handles DELETE /items/users/:id requests
// @Summary      Delete user
// @Tags         core-users
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete a user from the system. Requires authentication and user deletion permissions.
// @Param        id   path      string true "User ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/users/{id} [delete]
func DeleteUser(c *gin.Context) {
	// This is just for Swagger documentation
}

// =============================================================================
// ROLES MANAGEMENT
// =============================================================================

// GetRoles handles GET /items/roles requests
// @Summary      List roles
// @Tags         core-roles
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of roles in the system. Requires authentication and role management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (name, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        name     query  string false "Filter by role name"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/roles [get]
func GetRoles(c *gin.Context) {
	// This is just for Swagger documentation
}

// GetRole handles GET /items/roles/:id requests
// @Summary      Get role
// @Tags         core-roles
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific role by ID. Requires authentication and role management permissions.
// @Param        id   path      string true "Role ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/roles/{id} [get]
func GetRole(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreateRole handles POST /items/roles requests
// @Summary      Create role
// @Tags         core-roles
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new role in the system. Requires authentication and role creation permissions.
// @Param        body body map[string]interface{} true "Role data (name, description, tenant_id)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/roles [post]
func CreateRole(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdateRole handles PUT /items/roles/:id requests
// @Summary      Update role
// @Tags         core-roles
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing role. Requires authentication and role update permissions.
// @Param        id   path      string true "Role ID (UUID)"
// @Param        body body map[string]interface{} true "Role data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/roles/{id} [put]
func UpdateRole(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeleteRole handles DELETE /items/roles/:id requests
// @Summary      Delete role
// @Tags         core-roles
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete a role from the system. Requires authentication and role deletion permissions.
// @Param        id   path      string true "Role ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/roles/{id} [delete]
func DeleteRole(c *gin.Context) {
	// This is just for Swagger documentation
}

// =============================================================================
// PERMISSIONS MANAGEMENT
// =============================================================================

// GetPermissions handles GET /items/permissions requests
// @Summary      List permissions
// @Tags         core-permissions
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of permissions in the system. Requires authentication and permission management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (table_name, action, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        table_name query string false "Filter by table name"
// @Param        action   query  string false "Filter by action (read, create, update, delete)"
// @Param        role_id  query  string false "Filter by role ID"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/permissions [get]
func GetPermissions(c *gin.Context) {
	// This is just for Swagger documentation
}

// GetPermission handles GET /items/permissions/:id requests
// @Summary      Get permission
// @Tags         core-permissions
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific permission by ID. Requires authentication and permission management permissions.
// @Param        id   path      string true "Permission ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/permissions/{id} [get]
func GetPermission(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreatePermission handles POST /items/permissions requests
// @Summary      Create permission
// @Tags         core-permissions
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new permission in the system. Requires authentication and permission creation permissions.
// @Param        body body map[string]interface{} true "Permission data (role_id, table_name, action, tenant_id)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/permissions [post]
func CreatePermission(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdatePermission handles PUT /items/permissions/:id requests
// @Summary      Update permission
// @Tags         core-permissions
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing permission. Requires authentication and permission update permissions.
// @Param        id   path      string true "Permission ID (UUID)"
// @Param        body body map[string]interface{} true "Permission data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/permissions/{id} [put]
func UpdatePermission(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeletePermission handles DELETE /items/permissions/:id requests
// @Summary      Delete permission
// @Tags         core-permissions
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete a permission from the system. Requires authentication and permission deletion permissions.
// @Param        id   path      string true "Permission ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/permissions/{id} [delete]
func DeletePermission(c *gin.Context) {
	// This is just for Swagger documentation
}

// =============================================================================
// COLLECTIONS MANAGEMENT
// =============================================================================

// GetCollections handles GET /items/collections requests
// @Summary      List collections
// @Tags         core-collections
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of collections in the system. Requires authentication and collection management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (name, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        name     query  string false "Filter by collection name"
// @Param        icon     query  string false "Filter by icon"
// @Param        is_primary query bool false "Filter by primary status"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/collections [get]
func GetCollections(c *gin.Context) {
	// This is just for Swagger documentation
}

// GetCollection handles GET /items/collections/:id requests
// @Summary      Get collection
// @Tags         core-collections
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific collection by ID. Requires authentication and collection management permissions.
// @Param        id   path      string true "Collection ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/collections/{id} [get]
func GetCollection(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreateCollection handles POST /items/collections requests
// @Summary      Create collection
// @Tags         core-collections
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new collection in the system. Requires authentication and collection creation permissions.
// @Param        body body map[string]interface{} true "Collection data (name, description, icon, is_primary, tenant_id)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/collections [post]
func CreateCollection(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdateCollection handles PUT /items/collections/:id requests
// @Summary      Update collection
// @Tags         core-collections
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing collection. Requires authentication and collection update permissions.
// @Param        id   path      string true "Collection ID (UUID)"
// @Param        body body map[string]interface{} true "Collection data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/collections/{id} [put]
func UpdateCollection(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeleteCollection handles DELETE /items/collections/:id requests
// @Summary      Delete collection
// @Tags         core-collections
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete a collection from the system. Requires authentication and collection deletion permissions.
// @Param        id   path      string true "Collection ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/collections/{id} [delete]
func DeleteCollection(c *gin.Context) {
	// This is just for Swagger documentation
}

// =============================================================================
// FIELDS MANAGEMENT
// =============================================================================

// GetFields handles GET /items/fields requests
// @Summary      List fields
// @Tags         core-fields
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of fields in the system. Requires authentication and field management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (name, collection_id, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        name     query  string false "Filter by field name"
// @Param        collection_id query string false "Filter by collection ID"
// @Param        field_type query string false "Filter by field type (text, integer, boolean, jsonb, timestamp, uuid)"
// @Param        is_primary query bool false "Filter by primary status"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/fields [get]
func GetFields(c *gin.Context) {
	// This is just for Swagger documentation
}

// GetField handles GET /items/fields/:id requests
// @Summary      Get field
// @Tags         core-fields
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific field by ID. Requires authentication and field management permissions.
// @Param        id   path      string true "Field ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/fields/{id} [get]
func GetField(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreateField handles POST /items/fields requests
// @Summary      Create field
// @Tags         core-fields
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new field in the system. Requires authentication and field creation permissions.
// @Param        body body map[string]interface{} true "Field data (name, collection_id, field_type, is_required, is_primary, validation_rules, tenant_id)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/fields [post]
func CreateField(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdateField handles PUT /items/fields/:id requests
// @Summary      Update field
// @Tags         core-fields
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing field. Requires authentication and field update permissions.
// @Param        id   path      string true "Field ID (UUID)"
// @Param        body body map[string]interface{} true "Field data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/fields/{id} [put]
func UpdateField(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeleteField handles DELETE /items/fields/:id requests
// @Summary      Delete field
// @Tags         core-fields
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete a field from the system. Requires authentication and field deletion permissions.
// @Param        id   path      string true "Field ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/fields/{id} [delete]
func DeleteField(c *gin.Context) {
	// This is just for Swagger documentation
}

// =============================================================================
// API KEYS MANAGEMENT
// =============================================================================

// GetAPIKeys handles GET /items/api-keys requests
// @Summary      List API keys
// @Tags         core-api-keys
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a list of API keys in the system. Requires authentication and API key management permissions.
// @Param        limit    query  int    false "Limit (max 500)"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field (name, user_id, created_at)"
// @Param        order    query  string false "ASC or DESC"
// @Param        name     query  string false "Filter by API key name"
// @Param        user_id  query  string false "Filter by user ID"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/api-keys [get]
func GetAPIKeys(c *gin.Context) {
	// This is just for Swagger documentation
}

// GetAPIKey handles GET /items/api-keys/:id requests
// @Summary      Get API key
// @Tags         core-api-keys
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific API key by ID. Requires authentication and API key management permissions.
// @Param        id   path      string true "API Key ID (UUID)"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/api-keys/{id} [get]
func GetAPIKey(c *gin.Context) {
	// This is just for Swagger documentation
}

// CreateAPIKey handles POST /items/api-keys requests
// @Summary      Create API key
// @Tags         core-api-keys
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new API key in the system. Requires authentication and API key creation permissions.
// @Param        body body map[string]interface{} true "API Key data (name, user_id, permissions)"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/api-keys [post]
func CreateAPIKey(c *gin.Context) {
	// This is just for Swagger documentation
}

// UpdateAPIKey handles PUT /items/api-keys/:id requests
// @Summary      Update API key
// @Tags         core-api-keys
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing API key. Requires authentication and API key update permissions.
// @Param        id   path      string true "API Key ID (UUID)"
// @Param        body body map[string]interface{} true "API Key data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/api-keys/{id} [put]
func UpdateAPIKey(c *gin.Context) {
	// This is just for Swagger documentation
}

// DeleteAPIKey handles DELETE /items/api-keys/:id requests
// @Summary      Delete API key
// @Tags         core-api-keys
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete an API key from the system. Requires authentication and API key deletion permissions.
// @Param        id   path      string true "API Key ID (UUID)"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/api-keys/{id} [delete]
func DeleteAPIKey(c *gin.Context) {
	// This is just for Swagger documentation
}
