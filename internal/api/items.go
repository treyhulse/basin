package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"
	"go-rbac-api/internal/middleware"
	"go-rbac-api/internal/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type ItemsHandler struct {
	db            *db.DB
	policyChecker *rbac.PolicyChecker
}

func NewItemsHandler(db *db.DB) *ItemsHandler {
	return &ItemsHandler{
		db:            db,
		policyChecker: rbac.NewPolicyChecker(db.Queries),
	}
}

// GetItems handles GET /items/:table requests with RBAC filtering
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
	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(c.Request.Context(), userID, tableName, "read")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Handle schema management tables specially
	if tableName == "collections" || tableName == "fields" || tableName == "users" || tableName == "roles" || tableName == "permissions" || tableName == "api_keys" {
		// These are schema tables - use direct queries with tenant filtering
		query := rbac.BuildSelectQuery(tableName, allowedFields)

		var rows *sql.Rows
		var err error

		// Only add tenant filtering for tables that have tenant_id column
		if tableName == "api_keys" {
			// API keys table doesn't have tenant_id, filter by user_id instead
			query += " WHERE user_id = $1"
			rows, err = h.db.Query(query, userID)
		} else {
			// Add tenant filtering for multi-tenant support
			userTenantID, err := h.getUserTenantID(c.Request.Context(), userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tenant"})
				return
			}

			if userTenantID != uuid.Nil {
				query += " WHERE tenant_id = $1"
				rows, err = h.db.Query(query, userTenantID)
			} else {
				rows, err = h.db.Query(query)
			}
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
			return
		}
		defer rows.Close()

		// Process results (existing logic)
		results := h.scanRowsToMaps(rows)
		filteredResults := make([]map[string]interface{}, len(results))
		for i, result := range results {
			filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
		}

		c.JSON(http.StatusOK, gin.H{
			"data": filteredResults,
			"meta": gin.H{
				"table": tableName,
				"count": len(filteredResults),
				"type":  "schema",
			},
		})
		return
	}

	// Handle dynamic data tables with tenant support
	userTenantID, err := h.getUserTenantID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user tenant"})
		return
	}

	// Get tenant schema
	tenantSchema, err := h.getTenantSchema(c.Request.Context(), userTenantID)
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
	tableExists, err := h.tableExists(dataTableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check table existence"})
		return
	}

	if !tableExists {
		// Table doesn't exist - return empty result or create it
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

	// Execute query
	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	// Process results
	results := h.scanRowsToMaps(rows)
	filteredResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredResults,
		"meta": gin.H{
			"table": tableName,
			"count": len(filteredResults),
			"type":  "data",
		},
	})
}

// Helper method to scan rows to maps (extracted from existing logic)
func (h *ItemsHandler) scanRowsToMaps(rows *sql.Rows) []map[string]interface{} {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// Convert to map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Handle specific types
				switch v := val.(type) {
				case []byte:
					// Try to unmarshal as JSON, fallback to string
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

		results = append(results, row)
	}

	return results
}

// Helper method to check if table exists
func (h *ItemsHandler) tableExists(tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = split_part($1, '.', 1)
			AND table_name = split_part($1, '.', 2)
		)
	`
	var exists bool
	err := h.db.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

// GetItem handles GET /items/:table/:id requests
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
	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(c.Request.Context(), userID, tableName, "read")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
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

// CreateItem handles POST /items/:table requests
func (h *ItemsHandler) CreateItem(c *gin.Context) {
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
	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(c.Request.Context(), userID, tableName, "create")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Parse request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Filter fields based on permissions
	filteredData := h.policyChecker.FilterFields(requestData, allowedFields)

	// Handle schema tables specially - use sqlc queries
	if tableName == "collections" {
		collection, err := h.createCollection(c.Request.Context(), userID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"data": collection,
			"meta": gin.H{
				"table": tableName,
			},
		})
		return
	}

	if tableName == "fields" {
		field, err := h.createField(c.Request.Context(), userID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create field: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"data": field,
			"meta": gin.H{
				"table": tableName,
			},
		})
		return
	}

	if tableName == "users" {
		user, err := h.createUser(c.Request.Context(), userID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"data": user,
			"meta": gin.H{
				"table": tableName,
			},
		})
		return
	}

	if tableName == "api_keys" {
		apiKey, err := h.createAPIKey(c.Request.Context(), userID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"data": apiKey,
			"meta": gin.H{
				"table": tableName,
			},
		})
		return
	}

	// For dynamic data tables, insert into the actual table
	err = h.createDynamicItem(c.Request.Context(), userID, tableName, filteredData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": filteredData,
		"meta": gin.H{
			"table": tableName,
		},
	})
}

// UpdateItem handles PUT /items/:table/:id requests
func (h *ItemsHandler) UpdateItem(c *gin.Context) {
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
	hasPermission, allowedFields, err := h.policyChecker.CheckPermission(c.Request.Context(), userID, tableName, "update")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Parse request body
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Filter fields based on permissions
	filteredData := h.policyChecker.FilterFields(requestData, allowedFields)

	// Handle schema tables specially - use sqlc queries
	if tableName == "api_keys" {
		updatedAPIKey, err := h.updateAPIKey(c.Request.Context(), userID, itemID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": updatedAPIKey,
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "collections" {
		updatedCollection, err := h.updateCollection(c.Request.Context(), userID, itemID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": updatedCollection,
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "fields" {
		updatedField, err := h.updateField(c.Request.Context(), userID, itemID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update field: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": updatedField,
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "users" {
		updatedUser, err := h.updateUser(c.Request.Context(), userID, itemID, filteredData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data": updatedUser,
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	// For dynamic data tables, update in the actual table
	err = h.updateDynamicItem(c.Request.Context(), userID, tableName, itemID, filteredData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredData,
		"meta": gin.H{
			"table": tableName,
			"id":    itemID,
		},
	})
}

// DeleteItem handles DELETE /items/:table/:id requests
func (h *ItemsHandler) DeleteItem(c *gin.Context) {
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
	hasPermission, _, err := h.policyChecker.CheckPermission(c.Request.Context(), userID, tableName, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Handle schema tables specially - use sqlc queries
	if tableName == "api_keys" {
		err := h.deleteAPIKey(c.Request.Context(), userID, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "collections" {
		err := h.deleteCollection(c.Request.Context(), userID, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "fields" {
		err := h.deleteField(c.Request.Context(), userID, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete field: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	if tableName == "users" {
		err := h.deleteUser(c.Request.Context(), userID, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"meta": gin.H{
				"table": tableName,
				"id":    itemID,
			},
		})
		return
	}

	// For dynamic data tables, delete from the actual table
	err = h.deleteDynamicItem(c.Request.Context(), userID, tableName, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"table": tableName,
			"id":    itemID,
		},
	})
}

// Helper method to get user's tenant ID
func (h *ItemsHandler) getUserTenantID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT tenant_id FROM users WHERE id = $1`
	var tenantID uuid.UUID
	err := h.db.QueryRowContext(ctx, query, userID).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, err
	}
	return tenantID, nil
}

// Helper method to get tenant schema name
func (h *ItemsHandler) getTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `SELECT slug FROM tenants WHERE id = $1`
	var schema string
	err := h.db.QueryRowContext(ctx, query, tenantID).Scan(&schema)
	if err != nil {
		return "default", err // Fallback to default schema
	}
	return schema, nil
}

// Helper method to create a collection
func (h *ItemsHandler) createCollection(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate ID if not provided
	collectionID := uuid.New()
	if id, ok := data["id"].(string); ok {
		if parsedID, err := uuid.Parse(id); err == nil {
			collectionID = parsedID
		}
	}

	// Create collection using sqlc
	collection, err := h.db.Queries.CreateCollection(ctx, sqlc.CreateCollectionParams{
		ID:          collectionID,
		Name:        data["name"].(string),
		DisplayName: sql.NullString{String: getStringFromMap(data, "display_name"), Valid: true},
		Description: sql.NullString{String: getStringFromMap(data, "description"), Valid: true},
		Icon:        sql.NullString{String: getStringFromMap(data, "icon"), Valid: true},
		IsSystem:    sql.NullBool{Bool: getBoolFromMap(data, "is_system"), Valid: true},
		TenantID:    uuid.NullUUID{UUID: userTenantID, Valid: true},
		CreatedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := map[string]interface{}{
		"id":           collection.ID.String(),
		"name":         collection.Name,
		"display_name": collection.DisplayName.String,
		"description":  collection.Description.String,
		"icon":         collection.Icon.String,
		"is_system":    collection.IsSystem.Bool,
		"tenant_id":    collection.TenantID.UUID.String(),
		"created_by":   collection.CreatedBy.UUID.String(),
		"created_at":   collection.CreatedAt.Time,
		"updated_at":   collection.UpdatedAt.Time,
	}

	return result, nil
}

// Helper method to create a field
func (h *ItemsHandler) createField(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate ID if not provided
	fieldID := uuid.New()
	if id, ok := data["id"].(string); ok {
		if parsedID, err := uuid.Parse(id); err == nil {
			fieldID = parsedID
		}
	}

	// Parse collection_id
	collectionID, err := uuid.Parse(data["collection_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid collection_id")
	}

	// Create field using sqlc
	field, err := h.db.Queries.CreateField(ctx, sqlc.CreateFieldParams{
		ID:              fieldID,
		CollectionID:    uuid.NullUUID{UUID: collectionID, Valid: true},
		Name:            data["name"].(string),
		DisplayName:     sql.NullString{String: getStringFromMap(data, "display_name"), Valid: true},
		Type:            data["type"].(string),
		IsPrimary:       sql.NullBool{Bool: getBoolFromMap(data, "is_primary"), Valid: true},
		IsRequired:      sql.NullBool{Bool: getBoolFromMap(data, "is_required"), Valid: true},
		IsUnique:        sql.NullBool{Bool: getBoolFromMap(data, "is_unique"), Valid: true},
		DefaultValue:    sql.NullString{String: getStringFromMap(data, "default_value"), Valid: true},
		ValidationRules: pqtype.NullRawMessage{},
		RelationConfig:  pqtype.NullRawMessage{},
		SortOrder:       sql.NullInt32{Int32: int32(getIntFromMap(data, "sort_order")), Valid: true},
		TenantID:        uuid.NullUUID{UUID: userTenantID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := map[string]interface{}{
		"id":            field.ID.String(),
		"collection_id": field.CollectionID.UUID.String(),
		"name":          field.Name,
		"display_name":  field.DisplayName.String,
		"type":          field.Type,
		"is_primary":    field.IsPrimary.Bool,
		"is_required":   field.IsRequired.Bool,
		"is_unique":     field.IsUnique.Bool,
		"default_value": field.DefaultValue.String,
		"sort_order":    field.SortOrder.Int32,
		"tenant_id":     field.TenantID.UUID.String(),
		"created_at":    field.CreatedAt.Time,
		"updated_at":    field.UpdatedAt.Time,
	}

	return result, nil
}

// Helper method to create a user
func (h *ItemsHandler) createUser(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate ID if not provided
	newUserID := uuid.New()
	if id, ok := data["id"].(string); ok {
		if parsedID, err := uuid.Parse(id); err == nil {
			newUserID = parsedID
		}
	}

	// Hash password if provided
	passwordHash := ""
	if password, ok := data["password"].(string); ok {
		// You'll need to implement password hashing here
		passwordHash = password // TODO: Hash this properly
	}

	// Create user using sqlc
	user, err := h.db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           newUserID,
		Email:        data["email"].(string),
		PasswordHash: passwordHash,
		FirstName:    sql.NullString{String: getStringFromMap(data, "first_name"), Valid: true},
		LastName:     sql.NullString{String: getStringFromMap(data, "last_name"), Valid: true},
		TenantID:     uuid.NullUUID{UUID: userTenantID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert to map (don't return password hash)
	result := map[string]interface{}{
		"id":         user.ID.String(),
		"email":      user.Email,
		"first_name": user.FirstName.String,
		"last_name":  user.LastName.String,
		"is_active":  user.IsActive.Bool,
		"tenant_id":  user.TenantID.UUID.String(),
		"created_at": user.CreatedAt.Time,
		"updated_at": user.UpdatedAt.Time,
	}

	return result, nil
}

// Helper method to create dynamic item in data table
func (h *ItemsHandler) createDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, data map[string]interface{}) error {
	// Get tenant schema
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := h.getTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	tableExists, err := h.tableExists(dataTableName)
	if err != nil {
		return err
	}

	if !tableExists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Build INSERT query dynamically
	var columns []string
	var placeholders []string
	var values []interface{}

	// Add standard columns
	columns = append(columns, "created_by", "updated_by")
	placeholders = append(placeholders, "$1", "$2")
	values = append(values, userID, userID)

	paramIndex := 3
	for key, value := range data {
		if key != "id" && key != "created_at" && key != "updated_at" {
			columns = append(columns, fmt.Sprintf(`"%s"`, key))
			placeholders = append(placeholders, fmt.Sprintf("$%d", paramIndex))
			values = append(values, value)
			paramIndex++
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		dataTableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = h.db.ExecContext(ctx, query, values...)
	return err
}

// Helper functions to safely extract values from map
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolFromMap(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getIntFromMap(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// Helper method to create an API key
func (h *ItemsHandler) createAPIKey(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get target user ID (can create API keys for other users if admin)
	targetUserID := userID // Default to current user
	if targetUserStr, ok := data["user_id"].(string); ok {
		if parsedID, err := uuid.Parse(targetUserStr); err == nil {
			targetUserID = parsedID
		}
	}

	// Generate a secure API key
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the API key for storage
	keyHash := hashAPIKey(apiKey)

	// Set expiration (default 1 year from now, or use provided value)
	expiresAt := time.Now().AddDate(1, 0, 0)
	if expStr, ok := data["expires_at"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, expStr); err == nil {
			expiresAt = parsedTime
		}
	}

	// Get name for the API key
	name := "API Key"
	if nameStr, ok := data["name"].(string); ok {
		name = nameStr
	}

	// Create API key using sqlc
	createdKey, err := h.db.Queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		UserID:    targetUserID,
		Name:      name,
		KeyHash:   keyHash,
		ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert to map (include the plain API key only in creation response)
	result := map[string]interface{}{
		"id":           createdKey.ID.String(),
		"user_id":      createdKey.UserID.String(),
		"name":         createdKey.Name,
		"api_key":      apiKey, // Only returned on creation!
		"is_active":    createdKey.IsActive.Bool,
		"expires_at":   createdKey.ExpiresAt.Time,
		"last_used_at": nil,
		"created_at":   createdKey.CreatedAt.Time,
		"updated_at":   createdKey.UpdatedAt.Time,
	}

	return result, nil
}

// generateAPIKey generates a secure random API key
func generateAPIKey() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert to hex string with prefix
	return "basin_" + hex.EncodeToString(bytes), nil
}

// hashAPIKey creates a SHA-256 hash of the API key for secure storage
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// Helper method to update an API key
func (h *ItemsHandler) updateAPIKey(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	apiKeyID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid API key ID: %w", err)
	}

	// Check if user owns this API key (unless admin)
	existingKey, err := h.db.Queries.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Only allow users to update their own keys (unless admin)
	if existingKey.UserID != userID {
		// Check if user is admin
		hasAdminAccess, _, _ := h.policyChecker.CheckPermission(ctx, userID, "users", "read")
		if !hasAdminAccess {
			return nil, fmt.Errorf("unauthorized: can only update your own API keys")
		}
	}

	// Extract fields with defaults
	name := existingKey.Name
	if nameVal, ok := data["name"].(string); ok {
		name = nameVal
	}

	isActive := existingKey.IsActive.Bool
	if activeVal, ok := data["is_active"].(bool); ok {
		isActive = activeVal
	}

	expiresAt := existingKey.ExpiresAt
	if expVal, ok := data["expires_at"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, expVal); err == nil {
			expiresAt = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}

	// Update API key using sqlc
	updatedKey, err := h.db.Queries.UpdateAPIKey(ctx, sqlc.UpdateAPIKeyParams{
		ID:        apiKeyID,
		Name:      name,
		IsActive:  sql.NullBool{Bool: isActive, Valid: true},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}

	// Convert to map (don't include the actual API key in updates)
	result := map[string]interface{}{
		"id":           updatedKey.ID.String(),
		"user_id":      updatedKey.UserID.String(),
		"name":         updatedKey.Name,
		"is_active":    updatedKey.IsActive.Bool,
		"expires_at":   updatedKey.ExpiresAt.Time,
		"last_used_at": nil,
		"created_at":   updatedKey.CreatedAt.Time,
		"updated_at":   updatedKey.UpdatedAt.Time,
	}

	if updatedKey.LastUsedAt.Valid {
		result["last_used_at"] = updatedKey.LastUsedAt.Time
	}

	return result, nil
}

// Helper method to delete an API key
func (h *ItemsHandler) deleteAPIKey(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	apiKeyID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid API key ID: %w", err)
	}

	// Check if user owns this API key (unless admin)
	existingKey, err := h.db.Queries.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		return fmt.Errorf("API key not found: %w", err)
	}

	// Only allow users to delete their own keys (unless admin)
	if existingKey.UserID != userID {
		// Check if user is admin
		hasAdminAccess, _, _ := h.policyChecker.CheckPermission(ctx, userID, "users", "read")
		if !hasAdminAccess {
			return fmt.Errorf("unauthorized: can only delete your own API keys")
		}
	}

	// Delete API key using sqlc
	return h.db.Queries.DeleteAPIKey(ctx, apiKeyID)
}

// Helper method to update a collection
func (h *ItemsHandler) updateCollection(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	collectionID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing collection
	existingCollection, err := h.db.Queries.GetCollection(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check tenant access
	if existingCollection.TenantID.Valid && existingCollection.TenantID.UUID != userTenantID {
		return nil, fmt.Errorf("unauthorized: collection not accessible")
	}

	// Extract fields with defaults
	displayName := existingCollection.DisplayName
	if displayVal, ok := data["display_name"].(string); ok {
		displayName = sql.NullString{String: displayVal, Valid: true}
	}

	description := existingCollection.Description
	if descVal, ok := data["description"].(string); ok {
		description = sql.NullString{String: descVal, Valid: true}
	}

	icon := existingCollection.Icon
	if iconVal, ok := data["icon"].(string); ok {
		icon = sql.NullString{String: iconVal, Valid: true}
	}

	// Update collection using sqlc
	updatedCollection, err := h.db.Queries.UpdateCollection(ctx, sqlc.UpdateCollectionParams{
		ID:          collectionID,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		UpdatedBy:   uuid.NullUUID{UUID: userID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := map[string]interface{}{
		"id":           updatedCollection.ID.String(),
		"name":         updatedCollection.Name,
		"display_name": updatedCollection.DisplayName.String,
		"description":  updatedCollection.Description.String,
		"icon":         updatedCollection.Icon.String,
		"tenant_id":    nil,
		"created_by":   nil,
		"updated_by":   nil,
		"created_at":   updatedCollection.CreatedAt.Time,
		"updated_at":   updatedCollection.UpdatedAt.Time,
	}

	if updatedCollection.TenantID.Valid {
		result["tenant_id"] = updatedCollection.TenantID.UUID.String()
	}
	if updatedCollection.CreatedBy.Valid {
		result["created_by"] = updatedCollection.CreatedBy.UUID.String()
	}
	if updatedCollection.UpdatedBy.Valid {
		result["updated_by"] = updatedCollection.UpdatedBy.UUID.String()
	}

	return result, nil
}

// Helper method to delete a collection
func (h *ItemsHandler) deleteCollection(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	collectionID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid collection ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing collection to check access
	existingCollection, err := h.db.Queries.GetCollection(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Check tenant access
	if existingCollection.TenantID.Valid && existingCollection.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: collection not accessible")
	}

	// Delete collection using sqlc (this will trigger the database trigger to drop the data table)
	return h.db.Queries.DeleteCollection(ctx, collectionID)
}

// Helper method to update a field
func (h *ItemsHandler) updateField(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	fieldID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing field
	existingField, err := h.db.Queries.GetField(ctx, fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %w", err)
	}

	// Check tenant access
	if existingField.TenantID.Valid && existingField.TenantID.UUID != userTenantID {
		return nil, fmt.Errorf("unauthorized: field not accessible")
	}

	// Extract fields with defaults
	displayName := existingField.DisplayName
	if displayVal, ok := data["display_name"].(string); ok {
		displayName = sql.NullString{String: displayVal, Valid: true}
	}

	fieldType := existingField.Type
	if typeVal, ok := data["type"].(string); ok {
		fieldType = typeVal
	}

	isPrimary := existingField.IsPrimary
	if primaryVal, ok := data["is_primary"].(bool); ok {
		isPrimary = sql.NullBool{Bool: primaryVal, Valid: true}
	}

	isRequired := existingField.IsRequired
	if reqVal, ok := data["is_required"].(bool); ok {
		isRequired = sql.NullBool{Bool: reqVal, Valid: true}
	}

	isUnique := existingField.IsUnique
	if uniqueVal, ok := data["is_unique"].(bool); ok {
		isUnique = sql.NullBool{Bool: uniqueVal, Valid: true}
	}

	defaultValue := existingField.DefaultValue
	if defVal, ok := data["default_value"].(string); ok {
		defaultValue = sql.NullString{String: defVal, Valid: true}
	}

	sortOrder := existingField.SortOrder
	if sortInt := getIntFromMap(data, "sort_order"); sortInt > 0 {
		sortOrder = sql.NullInt32{Int32: int32(sortInt), Valid: true}
	}

	// Update field using sqlc
	updatedField, err := h.db.Queries.UpdateField(ctx, sqlc.UpdateFieldParams{
		ID:              fieldID,
		DisplayName:     displayName,
		Type:            fieldType,
		IsPrimary:       isPrimary,
		IsRequired:      isRequired,
		IsUnique:        isUnique,
		DefaultValue:    defaultValue,
		ValidationRules: existingField.ValidationRules,
		RelationConfig:  existingField.RelationConfig,
		SortOrder:       sortOrder,
	})
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := map[string]interface{}{
		"id":            updatedField.ID.String(),
		"collection_id": nil,
		"name":          updatedField.Name,
		"display_name":  updatedField.DisplayName.String,
		"type":          updatedField.Type,
		"is_primary":    updatedField.IsPrimary.Bool,
		"is_required":   updatedField.IsRequired.Bool,
		"is_unique":     updatedField.IsUnique.Bool,
		"default_value": updatedField.DefaultValue.String,
		"sort_order":    updatedField.SortOrder.Int32,
		"tenant_id":     nil,
		"created_at":    updatedField.CreatedAt.Time,
		"updated_at":    updatedField.UpdatedAt.Time,
	}

	if updatedField.CollectionID.Valid {
		result["collection_id"] = updatedField.CollectionID.UUID.String()
	}
	if updatedField.TenantID.Valid {
		result["tenant_id"] = updatedField.TenantID.UUID.String()
	}

	return result, nil
}

// Helper method to delete a field
func (h *ItemsHandler) deleteField(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	fieldID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid field ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing field to check access
	existingField, err := h.db.Queries.GetField(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	// Check tenant access
	if existingField.TenantID.Valid && existingField.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: field not accessible")
	}

	// Delete field using sqlc
	return h.db.Queries.DeleteField(ctx, fieldID)
}

// Helper method to update a user
func (h *ItemsHandler) updateUser(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	targetUserID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing user
	existingUser, err := h.db.Queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check tenant access
	if existingUser.TenantID.Valid && existingUser.TenantID.UUID != userTenantID {
		return nil, fmt.Errorf("unauthorized: user not accessible")
	}

	// Extract fields with defaults
	email := existingUser.Email
	if emailVal, ok := data["email"].(string); ok {
		email = emailVal
	}

	firstName := existingUser.FirstName
	if firstVal, ok := data["first_name"].(string); ok {
		firstName = sql.NullString{String: firstVal, Valid: true}
	}

	lastName := existingUser.LastName
	if lastVal, ok := data["last_name"].(string); ok {
		lastName = sql.NullString{String: lastVal, Valid: true}
	}

	isActive := existingUser.IsActive
	if activeVal, ok := data["is_active"].(bool); ok {
		isActive = sql.NullBool{Bool: activeVal, Valid: true}
	}

	// Update user using sqlc
	updatedUser, err := h.db.Queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        targetUserID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  isActive,
	})
	if err != nil {
		return nil, err
	}

	// Convert to map (don't include password hash)
	result := map[string]interface{}{
		"id":         updatedUser.ID.String(),
		"email":      updatedUser.Email,
		"first_name": updatedUser.FirstName.String,
		"last_name":  updatedUser.LastName.String,
		"is_active":  updatedUser.IsActive.Bool,
		"tenant_id":  nil,
		"created_at": updatedUser.CreatedAt.Time,
		"updated_at": updatedUser.UpdatedAt.Time,
	}

	if updatedUser.TenantID.Valid {
		result["tenant_id"] = updatedUser.TenantID.UUID.String()
	}

	return result, nil
}

// Helper method to delete a user
func (h *ItemsHandler) deleteUser(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	targetUserID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Prevent self-deletion
	if targetUserID == userID {
		return fmt.Errorf("cannot delete your own user account")
	}

	// Get tenant ID for filtering
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing user to check access
	existingUser, err := h.db.Queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check tenant access
	if existingUser.TenantID.Valid && existingUser.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: user not accessible")
	}

	// Delete user using sqlc
	return h.db.Queries.DeleteUser(ctx, targetUserID)
}

// Helper method to update dynamic item in data table
func (h *ItemsHandler) updateDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, itemID string, data map[string]interface{}) error {
	// Get tenant schema
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := h.getTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	exists, err := h.tableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = h.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}

	// Build dynamic UPDATE query
	if len(data) == 0 {
		return fmt.Errorf("no data provided for update")
	}

	setParts := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)+1)
	argIndex := 1

	for field, value := range data {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = CURRENT_TIMESTAMP WHERE id = $%d",
		dataTableName, strings.Join(setParts, ", "), argIndex)
	args = append(args, itemID)

	// Execute update
	result, err := h.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found or no changes made")
	}

	return nil
}

// Helper method to delete dynamic item from data table
func (h *ItemsHandler) deleteDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, itemID string) error {
	// Get tenant schema
	userTenantID, err := h.getUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := h.getTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	exists, err := h.tableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = h.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}

	// Execute delete
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", dataTableName)
	result, err := h.db.Exec(query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found")
	}

	return nil
}
