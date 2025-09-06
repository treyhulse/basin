// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This file contains handlers for dynamic data table operations - the tenant-specific tables
// that store actual user data based on the collections and fields defined in the schema.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-rbac-api/internal/db"

	"github.com/google/uuid"
)

// DynamicHandlers provides CRUD operations for tenant-specific data tables.
//
// In Basin's architecture, each tenant has their own PostgreSQL schema containing
// dynamically created data tables (e.g., "tenant_abc.data_products"). These tables
// store the actual user data based on collections and fields defined in the schema.
//
// This handler manages operations on these dynamic tables:
// - Respects tenant isolation (users can only access their tenant's data)
// - Implements Row-Level Security (RLS) for additional data protection
// - Handles dynamic SQL generation for arbitrary table structures
// - Provides proper error handling for missing tables/data
//
// Key Features:
// - Tenant-aware table resolution (schema.data_tablename format)
// - RLS context setting for secure data access
// - Dynamic INSERT/UPDATE/DELETE query generation
// - Proper transaction handling and error reporting
// - Table existence validation before operations
type DynamicHandlers struct {
	db    *db.DB      // Database connection for direct queries
	utils *ItemsUtils // Utility functions for tenant/table management
}

// NewDynamicHandlers creates a new DynamicHandlers instance with required dependencies.
//
// Parameters:
//   - db: Database connection for direct queries
//   - utils: ItemsUtils instance providing utility functions
//
// Returns:
//   - *DynamicHandlers: Configured dynamic handler ready for use
//
// Example:
//
//	dynamicHandler := NewDynamicHandlers(db, utils)
//	err := dynamicHandler.CreateDynamicItem(ctx, userID, "products", productData)
func NewDynamicHandlers(db *db.DB, utils *ItemsUtils) *DynamicHandlers {
	return &DynamicHandlers{
		db:    db,
		utils: utils,
	}
}

// CreateDynamicItem creates a new item in a dynamic data table
func (d *DynamicHandlers) CreateDynamicItem(ctx context.Context, userID uuid.UUID, collectionSlug string, data map[string]interface{}) error {
	// Get tenant ID
	userTenantID, err := d.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get the actual data table name from the collections table
	var dataTableName string
	query := `SELECT data_table_name FROM collections WHERE slug = $1 AND tenant_id = $2`
	err = d.db.QueryRowContext(ctx, query, collectionSlug, userTenantID).Scan(&dataTableName)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Use the data schema
	fullTableName := fmt.Sprintf(`data.%s`, dataTableName)

	// Check if table exists
	tableExists, err := d.utils.TableExists(fullTableName)
	if err != nil {
		return err
	}

	if !tableExists {
		return fmt.Errorf("table %s does not exist", fullTableName)
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
		fullTableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = d.db.ExecContext(ctx, query, values...)
	return err
}

// GetDynamicItem retrieves a specific item from a dynamic data table by ID
func (d *DynamicHandlers) GetDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, itemID string) (map[string]interface{}, error) {
	// Get tenant schema
	userTenantID, err := d.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tenantSchema, err := d.utils.GetTenantSchema(ctx, userTenantID)
	if err != nil {
		return nil, err
	}

	dataTableName := fmt.Sprintf(`"%s".data_%s`, tenantSchema, tableName)

	// Check if table exists
	tableExists, err := d.utils.TableExists(dataTableName)
	if err != nil {
		return nil, err
	}

	if !tableExists {
		return nil, fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = d.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to set user context: %w", err)
	}

	// Query the item
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", dataTableName)
	rows, err := d.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query item: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Create slice to hold values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Check if we have a row
	if !rows.Next() {
		return nil, fmt.Errorf("item not found")
	}

	// Scan the row
	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Convert to map
	result := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if val != nil {
			// Handle JSONB columns
			if jsonBytes, ok := val.([]byte); ok {
				var jsonData interface{}
				if err := json.Unmarshal(jsonBytes, &jsonData); err == nil {
					result[col] = jsonData
				} else {
					result[col] = string(jsonBytes)
				}
			} else {
				result[col] = val
			}
		}
	}

	return result, nil
}

// UpdateDynamicItem updates an existing item in a dynamic data table
func (d *DynamicHandlers) UpdateDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, itemID string, data map[string]interface{}) error {
	// Get tenant schema
	userTenantID, err := d.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := d.utils.GetTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := fmt.Sprintf(`"%s".data_%s`, tenantSchema, tableName)

	// Check if table exists
	exists, err := d.utils.TableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = d.db.Exec("SELECT set_user_context($1)", userID)
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
		if field != "id" && field != "created_at" && field != "created_by" {
			setParts = append(setParts, fmt.Sprintf(`"%s" = $%d`, field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s, updated_at = CURRENT_TIMESTAMP, updated_by = $%d WHERE id = $%d",
		dataTableName, strings.Join(setParts, ", "), argIndex, argIndex+1)
	args = append(args, userID, itemID)

	// Execute update
	result, err := d.db.Exec(query, args...)
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

// DeleteDynamicItem deletes an item from a dynamic data table
func (d *DynamicHandlers) DeleteDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, itemID string) error {
	// Get tenant schema
	userTenantID, err := d.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := d.utils.GetTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := fmt.Sprintf(`"%s".data_%s`, tenantSchema, tableName)

	// Check if table exists
	exists, err := d.utils.TableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = d.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}

	// Execute delete
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", dataTableName)
	result, err := d.db.Exec(query, itemID)
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
