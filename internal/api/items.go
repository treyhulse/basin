package api

import (
	"encoding/json"
	"net/http"

	"go-rbac-api/internal/db"
	"go-rbac-api/internal/middleware"
	"go-rbac-api/internal/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// Build query based on allowed fields
	query := rbac.BuildSelectQuery(tableName, allowedFields)

	// Execute query
	rows, err := h.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns"})
		return
	}

	// Scan results
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
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

	// Apply field filtering
	filteredResults := make([]map[string]interface{}, len(results))
	for i, result := range results {
		filteredResults[i] = h.policyChecker.FilterFields(result, allowedFields)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredResults,
		"meta": gin.H{
			"table": tableName,
			"count": len(filteredResults),
		},
	})
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

	// For now, we'll return the filtered data
	// In a real implementation, you'd insert this into the database
	c.JSON(http.StatusCreated, gin.H{
		"data": filteredData,
		"meta": gin.H{
			"table":   tableName,
			"message": "Item created (demo mode - not actually inserted)",
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

	// For now, we'll return the filtered data
	// In a real implementation, you'd update this in the database
	c.JSON(http.StatusOK, gin.H{
		"data": filteredData,
		"meta": gin.H{
			"table":   tableName,
			"id":      itemID,
			"message": "Item updated (demo mode - not actually updated)",
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

	// For now, we'll return success
	// In a real implementation, you'd delete this from the database
	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"table":   tableName,
			"id":      itemID,
			"message": "Item deleted (demo mode - not actually deleted)",
		},
	})
}
