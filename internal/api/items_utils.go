// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This file contains utility functions and helpers used across the items API handlers.
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"

	"github.com/google/uuid"
)

// ItemsUtils provides utility functions for database operations, data conversion,
// and tenant management used across the Basin API handlers.
//
// This struct encapsulates common operations like:
// - Converting SQL rows to Go maps with proper type handling
// - Managing tenant-specific database schemas and access
// - Validating table existence across different schemas
// - Handling UUID and data type conversions safely
//
// All methods are designed to be thread-safe and can be used concurrently
// across multiple HTTP requests.
type ItemsUtils struct {
	db *db.DB // Database connection pool for executing queries
}

// NewItemsUtils creates a new ItemsUtils instance with the provided database connection.
//
// Parameters:
//   - db: Database connection pool that will be used for all utility operations
//
// Returns:
//   - *ItemsUtils: Configured utility instance ready for use
//
// Example:
//
//	utils := NewItemsUtils(dbConnection)
//	results := utils.ScanRowsToMaps(rows)
func NewItemsUtils(db *db.DB) *ItemsUtils {
	return &ItemsUtils{db: db}
}

// ScanRowsToMaps converts SQL result rows into a slice of string-keyed maps for JSON serialization.
//
// This method handles the complex task of converting database rows with unknown column types
// into Go maps that can be easily serialized to JSON. It properly handles:
// - JSON/JSONB columns (unmarshals to native Go types)
// - NULL values (converts to Go nil)
// - Binary data (converts to strings when not valid JSON)
// - All standard SQL types
//
// Parameters:
//   - rows: Active SQL rows result set (must not be closed)
//
// Returns:
//   - []map[string]interface{}: Slice of maps where each map represents one database row.
//     Keys are column names, values are the converted column values.
//     Returns empty slice if no rows or on column scanning errors.
//
// Example:
//
//	rows, _ := db.Query("SELECT id, name, metadata FROM users")
//	defer rows.Close()
//	results := utils.ScanRowsToMaps(rows)
//	// results[0] = {"id": "123e4567-...", "name": "John", "metadata": {"age": 30}}
func (u *ItemsUtils) ScanRowsToMaps(rows *sql.Rows) []map[string]interface{} {
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

// TableExists checks whether a specified table exists in the database.
//
// This method supports both simple table names and schema-qualified table names.
// It uses the PostgreSQL information_schema to safely check table existence
// without risking SQL injection attacks.
//
// Parameters:
//   - tableName: Table name to check. Can be "table_name" or "schema.table_name"
//
// Returns:
//   - bool: true if table exists, false otherwise
//   - error: Database error if the query fails
//
// Examples:
//
//	exists, err := utils.TableExists("users")           // Check in default schema
//	exists, err := utils.TableExists("tenant1.data_products") // Check in specific schema
func (u *ItemsUtils) TableExists(tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = split_part($1, '.', 1)
			AND table_name = split_part($1, '.', 2)
		)
	`
	var exists bool
	err := u.db.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

// GetUserTenantID retrieves the tenant ID associated with a specific user.
//
// In Basin's multi-tenant architecture, each user belongs to exactly one tenant.
// This method is essential for enforcing tenant isolation and ensuring users
// can only access data within their own tenant's scope.
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - userID: UUID of the user whose tenant ID should be retrieved
//
// Returns:
//   - uuid.UUID: The tenant ID that the user belongs to
//   - error: Database error or user not found error
//
// Example:
//
//	tenantID, err := utils.GetUserTenantID(ctx, userUUID)
//	if err != nil {
//	    return fmt.Errorf("user not found or no tenant assigned: %w", err)
//	}
func (u *ItemsUtils) GetUserTenantID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT tenant_id FROM users WHERE id = $1`
	var tenantID uuid.UUID
	err := u.db.QueryRowContext(ctx, query, userID).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, err
	}
	return tenantID, nil
}

// GetTenantSchema retrieves the database schema name for a specific tenant.
//
// Each tenant in Basin has its own PostgreSQL schema to isolate their data tables.
// This method maps tenant UUIDs to their corresponding schema names (usually based on tenant slug).
// If the tenant is not found, it falls back to the "default" schema to maintain system stability.
//
// Parameters:
//   - ctx: Request context for cancellation and timeout handling
//   - tenantID: UUID of the tenant whose schema name should be retrieved
//
// Returns:
//   - string: Schema name for the tenant (e.g., "tenant_abc", "default")
//   - error: Database error (note: returns "default" schema name even on error as fallback)
//
// Example:
//
//	schema, err := utils.GetTenantSchema(ctx, tenantUUID)
//	tableName := fmt.Sprintf("%s.data_products", schema) // "tenant_abc.data_products"
func (u *ItemsUtils) GetTenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `SELECT slug FROM tenants WHERE id = $1`
	var schema string
	err := u.db.QueryRowContext(ctx, query, tenantID).Scan(&schema)
	if err != nil {
		return "default", err // Fallback to default schema
	}
	return schema, nil
}

// addColumnToDataTable adds a column to a data table when a field is created
func (u *ItemsUtils) AddColumnToDataTable(ctx context.Context, tenantID uuid.UUID, collectionName string, field sqlc.Field) error {
	// Get tenant schema
	tenantSchema, err := u.GetTenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}

	// For table existence check, use unquoted schema name
	unquotedTableName := tenantSchema + ".data_" + collectionName
	// For ALTER TABLE, use quoted schema name
	quotedTableName := "\"" + tenantSchema + "\".data_" + collectionName

	// Check if table exists
	tableExists, err := u.TableExists(unquotedTableName)
	if err != nil {
		return err
	}

	if !tableExists {
		return fmt.Errorf("data table %s does not exist", unquotedTableName)
	}

	// Build ALTER TABLE statement
	var columnType string
	switch field.Type {
	case "text":
		columnType = "TEXT"
	case "number":
		columnType = "NUMERIC"
	case "boolean":
		columnType = "BOOLEAN"
	case "date":
		columnType = "DATE"
	case "datetime":
		columnType = "TIMESTAMP WITH TIME ZONE"
	default:
		columnType = "TEXT"
	}

	// Build the ALTER TABLE query
	alterQuery := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN "%s" %s`, quotedTableName, field.Name, columnType)

	// Add NOT NULL constraint if required
	if field.IsRequired.Bool {
		alterQuery += " NOT NULL"
	}

	// Add default value if provided
	if field.DefaultValue.Valid && field.DefaultValue.String != "" {
		alterQuery += fmt.Sprintf(" DEFAULT '%s'", field.DefaultValue.String)
	}

	// Execute the ALTER TABLE statement
	_, err = u.db.ExecContext(ctx, alterQuery)
	if err != nil {
		return fmt.Errorf("failed to add column to data table: %w", err)
	}

	return nil
}

// Helper functions to safely extract values from map with type conversion and nil safety.
// These functions are used when processing JSON request bodies that have been unmarshaled
// into map[string]interface{} structures, providing safe type assertions with fallback values.

// GetStringFromMap safely extracts a string value from a map with proper type checking.
//
// This function handles the common scenario where JSON data has been unmarshaled into a
// map[string]interface{} and you need to safely extract string values without panicking
// on type assertion failures or missing keys.
//
// Parameters:
//   - data: Map containing the data (typically from JSON unmarshaling)
//   - key: Key to look up in the map
//
// Returns:
//   - string: The string value if found and valid, empty string otherwise
//
// Example:
//
//	data := map[string]interface{}{"name": "John", "age": 30}
//	name := GetStringFromMap(data, "name")     // Returns "John"
//	missing := GetStringFromMap(data, "email") // Returns ""
func GetStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetBoolFromMap safely extracts a boolean value from a map with proper type checking.
//
// This function provides safe extraction of boolean values from JSON-unmarshaled maps,
// handling cases where the key doesn't exist or the value is not a boolean type.
//
// Parameters:
//   - data: Map containing the data (typically from JSON unmarshaling)
//   - key: Key to look up in the map
//
// Returns:
//   - bool: The boolean value if found and valid, false otherwise
//
// Example:
//
//	data := map[string]interface{}{"active": true, "count": 5}
//	active := GetBoolFromMap(data, "active")   // Returns true
//	missing := GetBoolFromMap(data, "deleted") // Returns false
func GetBoolFromMap(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetIntFromMap safely extracts an integer value from a map with type conversion support.
//
// This function handles both integer and float64 types (since JSON numbers are often
// unmarshaled as float64 in Go) and safely converts them to integers. It provides
// fallback behavior for missing keys or invalid types.
//
// Parameters:
//   - data: Map containing the data (typically from JSON unmarshaling)
//   - key: Key to look up in the map
//
// Returns:
//   - int: The integer value if found and valid, 0 otherwise
//
// Example:
//
//	data := map[string]interface{}{"count": 42, "price": 29.99, "name": "test"}
//	count := GetIntFromMap(data, "count")  // Returns 42
//	price := GetIntFromMap(data, "price")  // Returns 29 (truncated from 29.99)
//	missing := GetIntFromMap(data, "age")  // Returns 0
func GetIntFromMap(data map[string]interface{}, key string) int {
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

// Contains checks if a slice contains a specific string with wildcard support.
//
// This function is primarily used for checking if a field name is allowed in RBAC
// field filtering. It supports the special "*" wildcard that grants access to all fields.
// This is essential for admin users or when full field access is granted.
//
// Parameters:
//   - slice: Slice of strings to search through (typically allowed field names)
//   - item: String to search for (typically a field name)
//
// Returns:
//   - bool: true if the item is found in the slice OR if slice contains "*", false otherwise
//
// Example:
//
//	allowed := []string{"id", "name", "email"}
//	Contains(allowed, "name")    // Returns true
//	Contains(allowed, "secret")  // Returns false
//
//	adminAllowed := []string{"*"}
//	Contains(adminAllowed, "any_field") // Returns true (wildcard match)
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item || s == "*" {
			return true
		}
	}
	return false
}
