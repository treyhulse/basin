package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sqlc "go-rbac-api/internal/db/sqlc"

	"github.com/google/uuid"
)

type PolicyChecker struct {
	db *sqlc.Queries
}

func NewPolicyChecker(db *sqlc.Queries) *PolicyChecker {
	return &PolicyChecker{db: db}
}

// CheckPermission checks if a user has permission to perform an action on a table
func (pc *PolicyChecker) CheckPermission(ctx context.Context, userID uuid.UUID, tableName, action string) (bool, []string, error) {
	// Get user roles
	roles, err := pc.db.GetUserRoles(ctx, userID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Get user's tenant ID
	user, err := pc.db.GetUserByID(ctx, userID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check permissions for each role with tenant isolation
	for _, role := range roles {
		// Check permissions for this role and tenant
		permissions, err := pc.db.GetPermissionsByRoleAndTenant(ctx, sqlc.GetPermissionsByRoleAndTenantParams{
			RoleID:   uuid.NullUUID{UUID: role.ID, Valid: true},
			TenantID: user.TenantID,
		})
		if err != nil {
			continue // Skip this role if there's an error
		}

		for _, permission := range permissions {
			// Check if permission matches table and action
			if permission.TableName == tableName && permission.Action == action {
				allowedFields := permission.AllowedFields
				if len(allowedFields) == 0 {
					allowedFields = []string{"*"} // Default to all fields
				}
				return true, allowedFields, nil
			}
		}
	}

	return false, nil, nil
}

// CheckPermissionWithTenant checks if a user has permission with explicit tenant context
func (pc *PolicyChecker) CheckPermissionWithTenant(ctx context.Context, userID, tenantID uuid.UUID, tableName, action string) (bool, []string, error) {
	// Get user roles
	roles, err := pc.db.GetUserRoles(ctx, userID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check permissions for each role with specific tenant
	for _, role := range roles {
		permissions, err := pc.db.GetPermissionsByRoleAndTenant(ctx, sqlc.GetPermissionsByRoleAndTenantParams{
			RoleID:   uuid.NullUUID{UUID: role.ID, Valid: true},
			TenantID: uuid.NullUUID{UUID: tenantID, Valid: true},
		})
		if err != nil {
			continue // Skip this role if there's an error
		}

		for _, permission := range permissions {
			// Check if permission matches table and action
			if permission.TableName == tableName && permission.Action == action {
				allowedFields := permission.AllowedFields
				if len(allowedFields) == 0 {
					allowedFields = []string{"*"} // Default to all fields
				}
				return true, allowedFields, nil
			}
		}
	}

	return false, nil, nil
}

// FilterFields filters the data based on allowed fields for the user
func (pc *PolicyChecker) FilterFields(data map[string]interface{}, allowedFields []string) map[string]interface{} {
	if len(allowedFields) == 0 {
		return data
	}

	// Check if all fields are allowed
	for _, field := range allowedFields {
		if field == "*" {
			return data // All fields allowed
		}
	}

	// Filter to only allowed fields
	filtered := make(map[string]interface{})
	for _, field := range allowedFields {
		if value, exists := data[field]; exists {
			filtered[field] = value
		}
	}

	return filtered
}

// FilterRecords applies row-level filtering based on field filters
func (pc *PolicyChecker) FilterRecords(records []map[string]interface{}, fieldFilter json.RawMessage) ([]map[string]interface{}, error) {
	if fieldFilter == nil {
		return records, nil
	}

	var filter map[string]interface{}
	if err := json.Unmarshal(fieldFilter, &filter); err != nil {
		return nil, fmt.Errorf("failed to unmarshal field filter: %w", err)
	}

	var filtered []map[string]interface{}
	for _, record := range records {
		match := true
		for key, value := range filter {
			if recordValue, exists := record[key]; !exists || recordValue != value {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, record)
		}
	}

	return filtered, nil
}

// ConvertToMap converts a struct to a map for filtering
func ConvertToMap(data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// ConvertFromMap converts a map back to JSON bytes
func ConvertFromMap(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// ValidateTableName ensures the table name is safe
func ValidateTableName(tableName string) bool {
	// Simple validation - only allow alphanumeric and underscores
	for _, char := range tableName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

// BuildSelectQuery builds a safe SELECT query with field filtering
func BuildSelectQuery(tableName string, allowedFields []string) string {
	if len(allowedFields) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", tableName)
	}

	// Check if all fields are allowed
	for _, field := range allowedFields {
		if field == "*" {
			return fmt.Sprintf("SELECT * FROM %s", tableName)
		}
	}

	// Build field list
	fields := make([]string, len(allowedFields))
	for i, field := range allowedFields {
		fields[i] = fmt.Sprintf(`"%s"`, field)
	}

	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), tableName)
}

// BuildSelectQueryWithTenant builds a safe SELECT query with tenant schema
func BuildSelectQueryWithTenant(tenantSchema, tableName string, allowedFields []string) string {
	fullTableName := fmt.Sprintf("%s.data_%s", tenantSchema, tableName)

	if len(allowedFields) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", fullTableName)
	}

	// Check if all fields are allowed
	for _, field := range allowedFields {
		if field == "*" {
			return fmt.Sprintf("SELECT * FROM %s", fullTableName)
		}
	}

	// Build field list
	fields := make([]string, len(allowedFields))
	for i, field := range allowedFields {
		fields[i] = fmt.Sprintf(`"%s"`, field)
	}

	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ", "), fullTableName)
}
