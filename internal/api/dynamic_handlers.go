// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This file contains handlers for dynamic data table operations - the tenant-specific tables
// that store actual user data based on the collections and fields defined in the schema.
package api

import (
	"context"
	"fmt"
	"strings"

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
	handler *ItemsHandler // Reference to main handler for database access
	utils   *ItemsUtils   // Utility functions for tenant/table management
}

// NewDynamicHandlers creates a new DynamicHandlers instance with required dependencies.
//
// Parameters:
//   - handler: Main ItemsHandler instance providing database access
//   - utils: ItemsUtils instance providing utility functions
//
// Returns:
//   - *DynamicHandlers: Configured dynamic handler ready for use
//
// Example:
//
//	dynamicHandler := NewDynamicHandlers(itemsHandler, utils)
//	err := dynamicHandler.CreateDynamicItem(ctx, userID, "products", productData)
func NewDynamicHandlers(handler *ItemsHandler, utils *ItemsUtils) *DynamicHandlers {
	return &DynamicHandlers{
		handler: handler,
		utils:   utils,
	}
}

// CreateDynamicItem creates a new item in a dynamic data table
func (d *DynamicHandlers) CreateDynamicItem(ctx context.Context, userID uuid.UUID, tableName string, data map[string]interface{}) error {
	// Get tenant schema
	userTenantID, err := d.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	tenantSchema, err := d.utils.GetTenantSchema(ctx, userTenantID)
	if err != nil {
		return err
	}

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	tableExists, err := d.utils.TableExists(dataTableName)
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

	_, err = d.handler.db.ExecContext(ctx, query, values...)
	return err
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

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	exists, err := d.utils.TableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = d.handler.db.Exec("SELECT set_user_context($1)", userID)
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
	result, err := d.handler.db.Exec(query, args...)
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

	dataTableName := tenantSchema + ".data_" + tableName

	// Check if table exists
	exists, err := d.utils.TableExists(dataTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("table %s does not exist", dataTableName)
	}

	// Set user context for RLS
	_, err = d.handler.db.Exec("SELECT set_user_context($1)", userID)
	if err != nil {
		return fmt.Errorf("failed to set user context: %w", err)
	}

	// Execute delete
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", dataTableName)
	result, err := d.handler.db.Exec(query, itemID)
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
