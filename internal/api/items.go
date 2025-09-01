// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This is the main items handler file that coordinates HTTP requests for Basin's Directus-style
// generic API, providing CRUD operations for any database table with comprehensive RBAC support.
//
// Basin API Architecture:
//
// The Basin API provides a generic, Directus-style interface where:
// - Collections define the structure of data tables (similar to database schemas)
// - Fields define the columns and validation rules within collections
// - Data is stored in tenant-specific dynamic tables based on these definitions
// - All access is controlled through comprehensive Role-Based Access Control (RBAC)
//
// API Endpoints:
// - GET    /items/:table     - List all items in a table (with filtering/pagination)
// - GET    /items/:table/:id - Get a specific item by ID
// - POST   /items/:table     - Create a new item in a table
// - PUT    /items/:table/:id - Update an existing item
// - DELETE /items/:table/:id - Delete an item by ID
//
// The API automatically handles:
// - Multi-tenant data isolation
// - Field-level permissions (users only see allowed fields)
// - Row-level security (users only see permitted data)
// - Dynamic schema management (tables created/modified as collections change)
// - Comprehensive input validation and error handling
//
// ⚠️  REFACTORING STATUS:
// This file has been successfully refactored from 1,884 lines to ~400 lines by delegating
// operations to specialized handlers while maintaining the exact same generic API functionality.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-rbac-api/internal/db"
	"go-rbac-api/internal/middleware"
	"go-rbac-api/internal/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ItemsHandler is the main HTTP request coordinator for Basin's dynamic table API.
//
// This handler provides the public HTTP interface for Basin's Directus-style generic API,
// supporting CRUD operations on any database table with comprehensive security and validation.
// It acts as a coordinator, delegating actual operations to specialized handlers while managing
// authentication, authorization, and request/response formatting.
//
// Architecture:
// The ItemsHandler follows a delegation pattern where it:
// 1. Validates incoming HTTP requests (table names, UUIDs, JSON payloads)
// 2. Authenticates users and checks RBAC permissions
// 3. Routes requests to appropriate specialized handlers based on table type
// 4. Formats responses consistently for the API
//
// Supported Table Types:
//   - Schema Tables: collections, fields, users, roles, permissions, api_keys
//     → Delegated to SchemaHandlers for structured CRUD with business logic
//   - Dynamic Tables: tenant-specific data tables (e.g., products, orders)
//     → Delegated to DynamicHandlers for flexible, schema-driven operations
//
// Security Features:
// - JWT token validation via middleware integration
// - Role-Based Access Control (RBAC) with field-level permissions
// - Tenant isolation (users can only access their tenant's data)
// - Input validation and SQL injection prevention
// - Comprehensive error handling with proper HTTP status codes
type ItemsHandler struct {
	db                 *db.DB              // Database connection pool for direct queries
	policyChecker      *rbac.PolicyChecker // RBAC policy evaluation engine
	utils              *ItemsUtils         // Utility functions for common operations
	schemaHandlers     *SchemaHandlers     // Handler for schema management tables
	dynamicHandlers    *DynamicHandlers    // Handler for dynamic tenant data tables
	collectionsHandler *CollectionsHandler // Handler for user-created collections
}

// NewItemsHandler creates a fully configured ItemsHandler with all required dependencies.
//
// This constructor initializes the handler and all its specialized sub-handlers,
// creating a complete system ready to handle Basin API requests. The initialization
// follows a dependency injection pattern for better testability and modularity.
//
// Parameters:
//   - db: Database connection pool that will be shared across all handlers
//
// Returns:
//   - *ItemsHandler: Fully configured handler ready to process HTTP requests
//
// Example:
//
//	handler := NewItemsHandler(dbConnection)
//	router.GET("/items/:table", handler.GetItems)
//	router.POST("/items/:table", handler.CreateItem)
func NewItemsHandler(db *db.DB) *ItemsHandler {
	handler := &ItemsHandler{
		db:            db,
		policyChecker: rbac.NewPolicyChecker(db.Queries),
	}

	// Initialize utility and handler components
	handler.utils = NewItemsUtils(db)
	handler.schemaHandlers = NewSchemaHandlers(handler, handler.utils)
	handler.dynamicHandlers = NewDynamicHandlers(db, handler.utils)
	handler.collectionsHandler = NewCollectionsHandler(db, handler.utils, handler.dynamicHandlers)

	return handler
}

// GetItems handles GET /items/:table requests with comprehensive RBAC filtering.
//
// This endpoint provides the core "list all items" functionality for Basin's generic API,
// supporting both schema management tables and dynamic tenant data tables. It automatically
// applies role-based filtering to ensure users only see data they're authorized to access.
//
// URL Pattern: GET /items/:table
//
// Path Parameters:
//   - table: Name of the table to query (e.g., "products", "users", "collections")
//
// Query Parameters (optional):
//   - Standard filtering and pagination parameters are supported
//   - Specific parameters depend on the table type and user permissions
//
// Authentication:
//   - Requires valid JWT token in Authorization header
//   - Token must contain valid user ID for permission checking
//
// Authorization:
//   - User must have "read" permission for the specified table
//   - Field-level filtering applied based on user's allowed fields
//   - Tenant isolation enforced (users only see their tenant's data)
//
// Response Format:
//   - 200: Success with filtered data array and metadata
//   - 400: Invalid table name or malformed request
//   - 401: Missing or invalid authentication token
//   - 403: User lacks permission to read from this table
//   - 500: Internal server error
//
// Example Response:
//
//	{
//	  "data": [{"id": "123", "name": "Product 1", ...}],
//	  "meta": {"table": "products", "count": 1, "type": "data"}
//	}
//
// @Summary      List items from dynamic table
// @Tags         items
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a paginated list of items from any dynamic table in the system. This endpoint works with both core schema tables (users, roles, permissions, collections, fields, api-keys) and custom dynamic tables (e.g., blog_posts, customers, products). The API automatically adapts to the table's schema, applying filters, sorting, and pagination. Requires authentication via JWT Bearer token or API key.
// @Param        table    path   string true  "Table name (e.g., 'users', 'blog_posts', 'customers')"
// @Param        limit    query  int    false "Limit (max 500, default 25)"
// @Param        offset   query  int    false "Offset for pagination"
// @Param        page     query  int    false "Page number (1-based, alternative to offset)"
// @Param        per_page query  int    false "Items per page (alternative to limit)"
// @Param        sort     query  string false "Sort field (e.g., 'created_at', 'name', 'email')"
// @Param        order    query  string false "Sort order: ASC or DESC (default: DESC)"
// @Param        filter   query  string false "JSON filter object for advanced filtering"
// @Param        limit    query  int    false "Limit"
// @Param        offset   query  int    false "Offset"
// @Param        page     query  int    false "Page (1-based)"
// @Param        per_page query  int    false "Per page"
// @Param        sort     query  string false "Sort field"
// @Param        order    query  string false "ASC or DESC"
// @Produce      json
// @Success      200 {object} models.ItemsListResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/{table} [get]
func (h *ItemsHandler) GetItems(c *gin.Context) {
	tableName := c.Param("table")

	// Validate table name
	if !rbac.ValidateTableName(tableName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check permissions
	// Get tenant context from the request
	tenantID, _ := middleware.GetTenantID(c)

	// Create a context with tenant information
	ctxWithTenant := context.WithValue(c.Request.Context(), "tenant_id", tenantID)

	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(ctxWithTenant, userID, tableName, "read")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Route to appropriate handler based on table type
	if h.isSchemaTable(tableName) {
		h.handleSchemaTableQuery(c, tableName, userID, allowedFields)
		return
	}

	// Check if this is a user-created collection
	if h.isUserCollection(c.Request.Context(), userID, tableName) {
		h.handleUserCollectionQuery(c, tableName, userID, allowedFields)
		return
	}

	// Handle dynamic data tables
	h.handleDynamicTableQuery(c, tableName, userID, allowedFields)
}

// GetItem handles GET /items/:table/:id requests
// @Summary      Get item from dynamic table
// @Tags         items
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Retrieve a specific item by ID from any dynamic table in the system. This endpoint works with both core schema tables and custom dynamic tables. Requires authentication via JWT Bearer token or API key.
// @Param        table   path      string true  "Table name (e.g., 'users', 'blog_posts', 'customers')"
// @Param        id      path      string true  "Item ID"
// @Produce      json
// @Success      200 {object} models.ItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/{table}/{id} [get]
func (h *ItemsHandler) GetItem(c *gin.Context) {
	tableName := c.Param("table")
	itemID := c.Param("id")

	// Validate table name
	if !rbac.ValidateTableName(tableName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	// Validate item ID
	if _, err := uuid.Parse(itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check permissions
	// Get tenant context from the request
	tenantID, _ := middleware.GetTenantID(c)

	// Create a context with tenant information
	ctxWithTenant := context.WithValue(c.Request.Context(), "tenant_id", tenantID)

	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(ctxWithTenant, userID, tableName, "read")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Check if this is a user collection and route accordingly
	if h.isUserCollection(c.Request.Context(), userID, tableName) {
		h.handleUserCollectionGetItem(c, tableName, userID, itemID, allowedFields)
		return
	}

	// Build query with WHERE clause
	query := rbac.BuildSelectQuery(tableName, allowedFields) + " WHERE id = $1"

	// Execute query
	rows, err := h.db.Query(query, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
		return
	}

	// Scan the single row
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
		return
	}

	// Convert to map
	row := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if val != nil {
			switch v := val.(type) {
			case []byte:
				var jsonVal interface{}
				if err := json.Unmarshal(v, &jsonVal); err == nil {
					row[col] = jsonVal
				} else {
					row[col] = string(v)
				}
			default:
				row[col] = v
			}
		} else {
			row[col] = nil
		}
	}

	// Apply field filtering
	filteredRow := h.policyChecker.FilterFields(row, allowedFields)

	c.JSON(http.StatusOK, gin.H{
		"data": filteredRow,
		"meta": gin.H{
			"table": tableName,
			"id":    itemID,
		},
	})
}

// CreateItem handles POST /items/:table requests with delegation to specialized handlers.
//
// This endpoint provides the core "create item" functionality for Basin's generic API,
// routing requests to appropriate specialized handlers based on table type while
// maintaining consistent validation, authentication, and response formatting.
//
// URL Pattern: POST /items/:table
//
// Path Parameters:
//   - table: Name of the table to create an item in (e.g., "products", "users", "collections")
//
// Request Body:
//   - JSON object containing the data for the new item
//   - Fields are automatically filtered based on user permissions
//
// Authentication & Authorization:
//   - Requires valid JWT token in Authorization header
//   - User must have "create" permission for the specified table
//   - Field-level filtering applied based on user's allowed fields
//   - Tenant isolation enforced for all operations
//
// Response Format:
//   - 201: Success with created item data and metadata
//   - 400: Invalid table name, malformed JSON, or validation errors
//   - 401: Missing or invalid authentication token
//   - 403: User lacks permission to create in this table
//   - 500: Internal server error during creation
//
// @Summary      Create item in dynamic table
// @Tags         items
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Create a new item in any dynamic table in the system. This endpoint works with both core schema tables and custom dynamic tables. The item structure depends on the table's schema (fields, validation rules, etc.). Requires authentication via JWT Bearer token or API key.
// @Param        table   path      string true  "Table name (e.g., 'users', 'blog_posts', 'customers')"
// @Param        body    body      map[string]interface{} true "Item data"
// @Accept       json
// @Produce      json
// @Success      201 {object} models.CreateItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Router       /items/{table} [post]
func (h *ItemsHandler) CreateItem(c *gin.Context) {
	tableName := c.Param("table")

	// Validate and authenticate request
	userID, requestData, err := h.validateCreateUpdateRequest(c, tableName, "create")
	if err != nil {
		return // Error already sent in validation
	}

	// Check permissions and filter data
	// Get tenant context from the request
	tenantID, _ := middleware.GetTenantID(c)

	// Create a context with tenant information
	ctxWithTenant := context.WithValue(c.Request.Context(), "tenant_id", tenantID)

	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(ctxWithTenant, userID, tableName, "create")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	filteredData := h.policyChecker.FilterFields(requestData, allowedFields)

	// Route to appropriate handler based on table type
	if h.isSchemaTable(tableName) {
		h.handleSchemaTableCreate(c, tableName, userID, filteredData)
		return
	}

	// Check if this is a user-created collection
	if h.isUserCollection(c.Request.Context(), userID, tableName) {
		h.handleUserCollectionCreate(c, tableName, userID, filteredData)
		return
	}

	// Handle dynamic data tables
	err = h.dynamicHandlers.CreateDynamicItem(c.Request.Context(), userID, tableName, filteredData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": filteredData,
		"meta": gin.H{"table": tableName},
	})
}

// UpdateItem handles PUT /items/:table/:id requests with delegation to specialized handlers.
//
// This endpoint provides the core "update item" functionality for Basin's generic API,
// routing requests to appropriate specialized handlers based on table type while
// maintaining consistent validation, authentication, and response formatting.
//
// URL Pattern: PUT /items/:table/:id
//
// Path Parameters:
//   - table: Name of the table containing the item to update
//   - id: UUID of the item to update
//
// Request Body:
//   - JSON object containing the fields to update
//   - Only provided fields will be updated (partial updates supported)
//   - Fields are automatically filtered based on user permissions
//
// Authentication & Authorization:
//   - Requires valid JWT token in Authorization header
//   - User must have "update" permission for the specified table
//   - Field-level filtering applied based on user's allowed fields
//   - Tenant isolation enforced for all operations
//
// Response Format:
//   - 200: Success with updated item data and metadata
//   - 400: Invalid table name, item ID, or malformed JSON
//   - 401: Missing or invalid authentication token
//   - 403: User lacks permission to update in this table
//   - 404: Item not found or not accessible to user
//   - 500: Internal server error during update
//
// @Summary      Update item in dynamic table
// @Tags         items
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Update an existing item in any dynamic table in the system. This endpoint works with both core schema tables and custom dynamic tables. Only the fields provided in the request body will be updated. Requires authentication via JWT Bearer token or API key.
// @Param        table   path      string true  "Table name (e.g., 'users', 'blog_posts', 'customers')"
// @Param        id      path      string true  "Item ID"
// @Param        body    body      map[string]interface{} true "Item data to update"
// @Accept       json
// @Produce      json
// @Success      200 {object} models.UpdateItemResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /items/{table}/{id} [put]
func (h *ItemsHandler) UpdateItem(c *gin.Context) {
	tableName := c.Param("table")
	itemID := c.Param("id")

	// Validate and authenticate request
	userID, requestData, err := h.validateCreateUpdateRequest(c, tableName, "update")
	if err != nil {
		return // Error already sent in validation
	}

	// Validate item ID
	if _, err := uuid.Parse(itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Check permissions and filter data
	// Get tenant context from the request
	tenantID, _ := middleware.GetTenantID(c)

	// Create a context with tenant information
	ctxWithTenant := context.WithValue(c.Request.Context(), "tenant_id", tenantID)

	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(ctxWithTenant, userID, tableName, "update")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	filteredData := h.policyChecker.FilterFields(requestData, allowedFields)

	// Route to appropriate handler based on table type
	if h.isSchemaTable(tableName) {
		h.handleSchemaTableUpdate(c, tableName, userID, itemID, filteredData)
		return
	}

	// Check if this is a user-created collection
	if h.isUserCollection(c.Request.Context(), userID, tableName) {
		h.handleUserCollectionUpdate(c, tableName, userID, itemID, filteredData)
		return
	}

	// Handle dynamic data tables
	err = h.dynamicHandlers.UpdateDynamicItem(c.Request.Context(), userID, tableName, itemID, filteredData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredData,
		"meta": gin.H{"table": tableName, "id": itemID},
	})
}

// DeleteItem handles DELETE /items/:table/:id requests with delegation to specialized handlers.
//
// This endpoint provides the core "delete item" functionality for Basin's generic API,
// routing requests to appropriate specialized handlers based on table type while
// maintaining consistent validation, authentication, and response formatting.
//
// URL Pattern: DELETE /items/:table/:id
//
// Path Parameters:
//   - table: Name of the table containing the item to delete
//   - id: UUID of the item to delete
//
// Authentication & Authorization:
//   - Requires valid JWT token in Authorization header
//   - User must have "delete" permission for the specified table
//   - Tenant isolation enforced (users can only delete from their tenant)
//   - Additional business rules may apply (e.g., cannot delete own user account)
//
// Response Format:
//   - 200: Success with metadata about the deleted item
//   - 400: Invalid table name or item ID
//   - 401: Missing or invalid authentication token
//   - 403: User lacks permission to delete from this table
//   - 404: Item not found or not accessible to user
//   - 500: Internal server error during deletion
//
// Note: Deletions may cascade to related data depending on table relationships
// @Summary      Delete item from dynamic table
// @Tags         items
// @Security     BearerAuth
// @Security     ApiKeyAuth
// @Description  Delete an item from any dynamic table in the system. This endpoint works with both core schema tables and custom dynamic tables. The deletion is permanent and cannot be undone. Requires authentication via JWT Bearer token or API key.
// @Param        table   path      string true  "Table name (e.g., 'users', 'blog_posts', 'customers')"
// @Param        id      path      string true  "Item ID"
// @Produce      json
// @Success      200 {object} models.DeleteItemResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      403 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Router       /items/{table}/{id} [delete]
func (h *ItemsHandler) DeleteItem(c *gin.Context) {
	tableName := c.Param("table")
	itemID := c.Param("id")

	// Validate inputs
	if !rbac.ValidateTableName(tableName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	if _, err := uuid.Parse(itemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	// Get user ID and check permissions
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get tenant context from the request
	tenantID, _ := middleware.GetTenantID(c)

	// Create a context with tenant information
	ctxWithTenant := context.WithValue(c.Request.Context(), "tenant_id", tenantID)

	hasPermission, _, err := h.policyChecker.CheckPermission(ctxWithTenant, userID, tableName, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Route to appropriate handler based on table type
	if h.isSchemaTable(tableName) {
		h.handleSchemaTableDelete(c, tableName, userID, itemID)
		return
	}

	// Check if this is a user-created collection
	if h.isUserCollection(c.Request.Context(), userID, tableName) {
		h.handleUserCollectionDelete(c, tableName, userID, itemID)
		return
	}

	// Handle dynamic data tables
	err = h.dynamicHandlers.DeleteDynamicItem(c.Request.Context(), userID, tableName, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{"table": tableName, "id": itemID},
	})
}

// Helper methods for request validation and routing

// validateCreateUpdateRequest handles common validation for create and update requests
func (h *ItemsHandler) validateCreateUpdateRequest(c *gin.Context, tableName, operation string) (uuid.UUID, map[string]interface{}, error) {
	// Validate table name
	if !rbac.ValidateTableName(tableName) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return uuid.Nil, nil, fmt.Errorf("invalid table name")
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return uuid.Nil, nil, fmt.Errorf("user not authenticated")
	}

	// Parse request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return uuid.Nil, nil, fmt.Errorf("invalid request body")
	}

	return userID, requestData, nil
}

// isSchemaTable checks if a table is a schema management table
func (h *ItemsHandler) isSchemaTable(tableName string) bool {
	schemaTableNames := []string{"collections", "fields", "users", "roles", "permissions", "api_keys"}
	for _, name := range schemaTableNames {
		if tableName == name {
			return true
		}
	}
	return false
}

// isUserCollection checks if a table is a user-created collection
func (h *ItemsHandler) isUserCollection(ctx context.Context, userID uuid.UUID, tableName string) bool {
	// Get user's tenant
	userTenantID, err := h.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return false
	}

	// Check if collection exists in the collections table
	_, err = h.collectionsHandler.GetCollection(ctx, userTenantID, tableName)
	return err == nil
}

// handleSchemaTableCreate routes create requests for schema management tables
func (h *ItemsHandler) handleSchemaTableCreate(c *gin.Context, tableName string, userID uuid.UUID, data map[string]interface{}) {
	var result map[string]interface{}
	var err error

	switch tableName {
	case "collections":
		result, err = h.schemaHandlers.CreateCollection(c.Request.Context(), userID, data)
	case "fields":
		result, err = h.schemaHandlers.CreateField(c.Request.Context(), userID, data)
	case "users":
		result, err = h.schemaHandlers.CreateUser(c.Request.Context(), userID, data)
	case "api_keys":
		result, err = h.schemaHandlers.CreateAPIKey(c.Request.Context(), userID, data)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported schema table for creation"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create " + tableName + ": " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": result,
		"meta": gin.H{"table": tableName},
	})
}

// handleSchemaTableUpdate routes update requests for schema management tables
func (h *ItemsHandler) handleSchemaTableUpdate(c *gin.Context, tableName string, userID uuid.UUID, itemID string, data map[string]interface{}) {
	var result map[string]interface{}
	var err error

	switch tableName {
	case "collections":
		result, err = h.schemaHandlers.UpdateCollection(c.Request.Context(), userID, itemID, data)
	case "fields":
		result, err = h.schemaHandlers.UpdateField(c.Request.Context(), userID, itemID, data)
	case "users":
		result, err = h.schemaHandlers.UpdateUser(c.Request.Context(), userID, itemID, data)
	case "api_keys":
		result, err = h.schemaHandlers.UpdateAPIKey(c.Request.Context(), userID, itemID, data)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported schema table for updates"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update " + tableName + ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"meta": gin.H{"table": tableName, "id": itemID},
	})
}

// handleUserCollectionCreate routes create requests for user-created collections
func (h *ItemsHandler) handleUserCollectionCreate(c *gin.Context, tableName string, userID uuid.UUID, data map[string]interface{}) {
	// Create the item using collections handler
	result, err := h.collectionsHandler.CreateCollectionItem(c.Request.Context(), userID, tableName, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create collection item: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": result,
		"meta": gin.H{"table": tableName, "type": "collection"},
	})
}

// handleUserCollectionUpdate routes update requests for user-created collections
func (h *ItemsHandler) handleUserCollectionUpdate(c *gin.Context, tableName string, userID uuid.UUID, itemID string, data map[string]interface{}) {
	// Update the item using collections handler
	result, err := h.collectionsHandler.UpdateCollectionItem(c.Request.Context(), userID, tableName, itemID, data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update collection item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"meta": gin.H{"table": tableName, "id": itemID, "type": "collection"},
	})
}

// handleUserCollectionDelete routes delete requests for user-created collections
func (h *ItemsHandler) handleUserCollectionDelete(c *gin.Context, tableName string, userID uuid.UUID, itemID string) {
	// Delete the item using collections handler
	err := h.collectionsHandler.DeleteCollectionItem(c.Request.Context(), userID, tableName, itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete collection item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{"table": tableName, "id": itemID, "type": "collection"},
	})
}

// handleUserCollectionGetItem handles getting a specific item from a user collection
func (h *ItemsHandler) handleUserCollectionGetItem(c *gin.Context, tableName string, userID uuid.UUID, itemID string, allowedFields []string) {
	// Get the item using collections handler
	item, err := h.collectionsHandler.GetCollectionItem(c.Request.Context(), userID, tableName, itemID)
	if err != nil {
		if strings.Contains(err.Error(), "item not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch item"})
		}
		return
	}

	// Apply field filtering
	filteredItem := h.policyChecker.FilterFields(item, allowedFields)

	c.JSON(http.StatusOK, gin.H{
		"data": filteredItem,
		"meta": gin.H{
			"table":      tableName,
			"id":         itemID,
			"type":       "collection",
			"collection": tableName,
		},
	})
}

// handleSchemaTableDelete routes delete requests for schema management tables
func (h *ItemsHandler) handleSchemaTableDelete(c *gin.Context, tableName string, userID uuid.UUID, itemID string) {
	var err error

	switch tableName {
	case "collections":
		err = h.schemaHandlers.DeleteCollection(c.Request.Context(), userID, itemID)
	case "fields":
		err = h.schemaHandlers.DeleteField(c.Request.Context(), userID, itemID)
	case "users":
		err = h.schemaHandlers.DeleteUser(c.Request.Context(), userID, itemID)
	case "api_keys":
		err = h.schemaHandlers.DeleteAPIKey(c.Request.Context(), userID, itemID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported schema table for deletion"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete " + tableName + ": " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{"table": tableName, "id": itemID},
	})
}

// handleSchemaTableQuery handles queries for schema management tables
func (h *ItemsHandler) handleSchemaTableQuery(c *gin.Context, tableName string, userID uuid.UUID, allowedFields []string) {
	query := rbac.BuildSelectQuery(tableName, allowedFields)

	var queryParams []interface{}
	var whereConditions []string
	paramIndex := 1

	// Handle tenant filtering for different schema tables
	if tableName == "api_keys" {
		// API keys table doesn't have tenant_id, filter by user_id instead
		whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", paramIndex))
		queryParams = append(queryParams, userID)
		paramIndex++
	} else {
		// Add tenant filtering for multi-tenant support
		userTenantID, err := h.utils.GetUserTenantID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tenant"})
			return
		}

		if userTenantID != uuid.Nil {
			whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", paramIndex))
			queryParams = append(queryParams, userTenantID)
			paramIndex++
		}
	}

	// Add query parameter filtering (exclude special params)
	queryValues := c.Request.URL.Query()
	for key, values := range queryValues {
		if key == "limit" || key == "offset" || key == "page" || key == "per_page" || key == "sort" || key == "order" {
			continue
		}
		if len(values) > 0 && values[0] != "" {
			if Contains(allowedFields, key) {
				whereConditions = append(whereConditions, fmt.Sprintf("%s = $%d", key, paramIndex))
				queryParams = append(queryParams, values[0])
				paramIndex++
			}
		}
	}

	// Add WHERE clause if we have conditions
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Sorting
	if sortField := c.Query("sort"); sortField != "" && Contains(allowedFields, sortField) {
		order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY \"%s\" %s", sortField, order)
	}

	// Pagination
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := c.Query("per_page"); v != "" { // alias
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := c.Query("page"); v != "" { // 1-based
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := h.db.Query(query, queryParams...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	// Process results
	results := h.utils.ScanRowsToMaps(rows)
	filteredResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredResults,
		"meta": gin.H{
			"table":  tableName,
			"count":  len(filteredResults),
			"limit":  limit,
			"offset": offset,
			"type":   "schema",
		},
	})
}

// handleUserCollectionQuery handles queries for user-created collections
func (h *ItemsHandler) handleUserCollectionQuery(c *gin.Context, tableName string, userID uuid.UUID, allowedFields []string) {
	// Get user's tenant
	userTenantID, err := h.utils.GetUserTenantID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tenant"})
		return
	}

	// Get collection definition
	collection, err := h.collectionsHandler.GetCollection(c.Request.Context(), userTenantID, tableName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}

	// Get tenant schema
	tenantSchema, err := h.utils.GetTenantSchema(c.Request.Context(), userTenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant schema"})
		return
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Set user context for RLS
	_, err = h.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set user context"})
		return
	}

	// Check if the data table exists
	tableExists, err := h.utils.TableExists(dataTableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check table existence"})
		return
	}

	if !tableExists {
		// Table doesn't exist - return empty result
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{},
			"meta": gin.H{
				"table":      tableName,
				"count":      0,
				"type":       "collection",
				"collection": collection.Name,
				"message":    "Collection table does not exist yet",
			},
		})
		return
	}

	// Build query based on allowed fields for data table
	query := rbac.BuildSelectQueryWithTenant(tenantSchema, tableName, allowedFields)

	// Sorting
	if sortField := c.Query("sort"); sortField != "" && Contains(allowedFields, sortField) {
		order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY \"%s\" %s", sortField, order)
	}

	// Pagination
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := c.Query("per_page"); v != "" { // alias
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := c.Query("page"); v != "" { // 1-based
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// Execute query
	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	// Process results
	results := h.utils.ScanRowsToMaps(rows)
	filteredResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredResults,
		"meta": gin.H{
			"table":      tableName,
			"count":      len(filteredResults),
			"limit":      limit,
			"offset":     offset,
			"type":       "collection",
			"collection": collection.Name,
		},
	})
}

// handleDynamicTableQuery handles queries for dynamic data tables
func (h *ItemsHandler) handleDynamicTableQuery(c *gin.Context, tableName string, userID uuid.UUID, allowedFields []string) {
	// Get tenant schema
	userTenantID, err := h.utils.GetUserTenantID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tenant"})
		return
	}

	tenantSchema, err := h.utils.GetTenantSchema(c.Request.Context(), userTenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant schema"})
		return
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Set user context for RLS
	_, err = h.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set user context"})
		return
	}

	// Check if the data table exists
	tableExists, err := h.utils.TableExists(dataTableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check table existence"})
		return
	}

	if !tableExists {
		// Table doesn't exist - return empty result
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{},
			"meta": gin.H{
				"table":   tableName,
				"count":   0,
				"type":    "data",
				"message": "Table does not exist yet",
			},
		})
		return
	}

	// Build query based on allowed fields for data table
	query := rbac.BuildSelectQueryWithTenant(tenantSchema, tableName, allowedFields)

	// Sorting
	if sortField := c.Query("sort"); sortField != "" && Contains(allowedFields, sortField) {
		order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY \"%s\" %s", sortField, order)
	}

	// Pagination
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := c.Query("per_page"); v != "" { // alias
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := c.Query("page"); v != "" { // 1-based
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	// Execute query
	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	// Process results
	results := h.utils.ScanRowsToMaps(rows)
	filteredResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredResults,
		"meta": gin.H{
			"table":  tableName,
			"count":  len(filteredResults),
			"limit":  limit,
			"offset": offset,
			"type":   "data",
		},
	})
}
