// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This file contains the CollectionsHandler for managing dynamic collections created from
// the collections and fields schema tables.
//
// CollectionsHandler provides specialized operations for user-created collections:
// - Schema validation against collections/fields definitions
// - Field type validation and conversion
// - Dynamic table creation and management
// - Collection-specific business logic and constraints
// - Enhanced error handling for collection operations
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-rbac-api/internal/db"

	"github.com/google/uuid"
)

// CollectionField represents a field definition from the fields table
type CollectionField struct {
	ID           uuid.UUID              `json:"id"`
	CollectionID uuid.UUID              `json:"collection_id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	IsRequired   bool                   `json:"is_required"`
	Default      interface{}            `json:"default"`
	Validation   map[string]interface{} `json:"validation"`
	Options      map[string]interface{} `json:"options"`
}

// Collection represents a collection definition from the collections table
type Collection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TenantID    uuid.UUID `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CollectionsHandler provides specialized operations for dynamic collections.
//
// This handler manages user-created collections that are defined in the collections
// and fields tables. It provides enhanced functionality beyond the basic DynamicHandlers:
//
// - Schema validation against collection/field definitions
// - Field type validation and conversion
// - Dynamic table creation when collections are created
// - Collection-specific business logic and constraints
// - Enhanced error handling with collection context
//
// Key Features:
// - Validates data against field definitions (types, required fields, etc.)
// - Handles field type conversions (string to int, JSON parsing, etc.)
// - Manages collection lifecycle (creation, updates, deletion)
// - Provides collection-specific error messages
// - Integrates with tenant isolation and RBAC
type CollectionsHandler struct {
	db              *db.DB           // Database connection for queries
	utils           *ItemsUtils      // Utility functions
	dynamicHandlers *DynamicHandlers // Basic dynamic operations
}

// NewCollectionsHandler creates a new CollectionsHandler with required dependencies.
//
// Parameters:
//   - db: Database connection pool
//   - utils: ItemsUtils instance for utility functions
//   - dynamicHandlers: DynamicHandlers instance for basic operations
//
// Returns:
//   - *CollectionsHandler: Configured handler ready for use
func NewCollectionsHandler(db *db.DB, utils *ItemsUtils, dynamicHandlers *DynamicHandlers) *CollectionsHandler {
	return &CollectionsHandler{
		db:              db,
		utils:           utils,
		dynamicHandlers: dynamicHandlers,
	}
}

// GetCollection retrieves a collection definition by name
func (ch *CollectionsHandler) GetCollection(ctx context.Context, tenantID uuid.UUID, collectionName string) (*Collection, error) {
	query := `
		SELECT id, name, description, tenant_id, created_at, updated_at
		FROM collections 
		WHERE name = $1 AND tenant_id = $2
	`

	var collection Collection
	err := ch.db.QueryRowContext(ctx, query, collectionName, tenantID).Scan(
		&collection.ID,
		&collection.Name,
		&collection.Description,
		&collection.TenantID,
		&collection.CreatedAt,
		&collection.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	return &collection, nil
}

// GetCollectionFields retrieves all fields for a collection
func (ch *CollectionsHandler) GetCollectionFields(ctx context.Context, collectionID uuid.UUID) ([]CollectionField, error) {
	query := `
		SELECT id, collection_id, name, type, is_required, default_value, validation_rules, relation_config
		FROM fields 
		WHERE collection_id = $1
		ORDER BY name
	`

	rows, err := ch.db.QueryContext(ctx, query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fields: %w", err)
	}
	defer rows.Close()

	var fields []CollectionField
	for rows.Next() {
		var field CollectionField
		var defaultVal, validation, options []byte

		err := rows.Scan(
			&field.ID,
			&field.CollectionID,
			&field.Name,
			&field.Type,
			&field.IsRequired,
			&defaultVal,
			&validation,
			&options,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field: %w", err)
		}

		// Parse JSON fields
		if len(defaultVal) > 0 {
			json.Unmarshal(defaultVal, &field.Default)
		}
		if len(validation) > 0 {
			json.Unmarshal(validation, &field.Validation)
		}
		if len(options) > 0 {
			json.Unmarshal(options, &field.Options)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// ValidateCollectionData validates data against collection field definitions
func (ch *CollectionsHandler) ValidateCollectionData(ctx context.Context, tenantID uuid.UUID, collectionName string, data map[string]interface{}) error {
	// Get collection definition
	collection, err := ch.GetCollection(ctx, tenantID, collectionName)
	if err != nil {
		return fmt.Errorf("collection validation failed: %w", err)
	}

	// Get field definitions
	fields, err := ch.GetCollectionFields(ctx, collection.ID)
	if err != nil {
		return fmt.Errorf("field validation failed: %w", err)
	}

	// Create field map for quick lookup
	fieldMap := make(map[string]CollectionField)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	// Validate each provided field
	for fieldName, value := range data {
		field, exists := fieldMap[fieldName]
		if !exists {
			return fmt.Errorf("field '%s' is not defined in collection '%s'", fieldName, collectionName)
		}

		// Validate required fields
		if field.IsRequired && (value == nil || value == "") {
			return fmt.Errorf("field '%s' is required", fieldName)
		}

		// Skip validation for nil/empty values (unless required)
		if value == nil || value == "" {
			continue
		}

		// Validate field type
		if err := ch.validateFieldType(field, value); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", fieldName, err)
		}

		// Apply field-specific validation rules
		if err := ch.applyFieldValidation(field, value); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", fieldName, err)
		}
	}

	// Check for missing required fields
	for _, field := range fields {
		if field.IsRequired {
			if _, provided := data[field.Name]; !provided {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	return nil
}

// validateFieldType validates that a value matches the expected field type
func (ch *CollectionsHandler) validateFieldType(field CollectionField, value interface{}) error {
	switch field.Type {
	case "string", "text":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

	case "integer", "int":
		switch v := value.(type) {
		case int, int32, int64, float64:
			// Accept various numeric types
			return nil
		case string:
			// Try to parse string as number
			if _, err := fmt.Sscanf(v, "%d", new(int)); err != nil {
				return fmt.Errorf("cannot convert string '%s' to integer", v)
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}

	case "float", "decimal":
		switch v := value.(type) {
		case float32, float64, int, int32, int64:
			// Accept various numeric types
			return nil
		case string:
			// Try to parse string as float
			if _, err := fmt.Sscanf(v, "%f", new(float64)); err != nil {
				return fmt.Errorf("cannot convert string '%s' to float", v)
			}
		default:
			return fmt.Errorf("expected float, got %T", value)
		}

	case "boolean", "bool":
		switch value.(type) {
		case bool:
			return nil
		case string:
			// Accept string representations
			str := value.(string)
			if str != "true" && str != "false" && str != "1" && str != "0" {
				return fmt.Errorf("cannot convert string '%s' to boolean", str)
			}
		default:
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "json", "object":
		// JSON fields can be maps, slices, or JSON strings
		switch value.(type) {
		case map[string]interface{}, []interface{}:
			return nil
		case string:
			// Try to parse as JSON
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(value.(string)), &jsonVal); err != nil {
				return fmt.Errorf("invalid JSON string: %w", err)
			}
		default:
			return fmt.Errorf("expected JSON object or string, got %T", value)
		}

	case "date", "datetime":
		switch v := value.(type) {
		case time.Time:
			return nil
		case string:
			// Try to parse common date formats
			formats := []string{
				"2006-01-02",
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05.000Z",
				time.RFC3339,
			}
			for _, format := range formats {
				if _, err := time.Parse(format, v); err == nil {
					return nil
				}
			}
			return fmt.Errorf("cannot parse date string '%s'", v)
		default:
			return fmt.Errorf("expected date/time, got %T", value)
		}

	default:
		// Unknown field type - accept any value
		return nil
	}

	return nil
}

// applyFieldValidation applies field-specific validation rules
func (ch *CollectionsHandler) applyFieldValidation(field CollectionField, value interface{}) error {
	if field.Validation == nil {
		return nil
	}

	// Apply length validation for strings
	if field.Type == "string" || field.Type == "text" {
		if str, ok := value.(string); ok {
			if minLength, exists := field.Validation["min_length"]; exists {
				if min, ok := minLength.(float64); ok && len(str) < int(min) {
					return fmt.Errorf("minimum length is %d characters", int(min))
				}
			}
			if maxLength, exists := field.Validation["max_length"]; exists {
				if max, ok := maxLength.(float64); ok && len(str) > int(max) {
					return fmt.Errorf("maximum length is %d characters", int(max))
				}
			}
		}
	}

	// Apply range validation for numbers
	if field.Type == "integer" || field.Type == "float" {
		var num float64
		switch v := value.(type) {
		case int:
			num = float64(v)
		case int32:
			num = float64(v)
		case int64:
			num = float64(v)
		case float32:
			num = float64(v)
		case float64:
			num = v
		case string:
			if _, err := fmt.Sscanf(v, "%f", &num); err != nil {
				return fmt.Errorf("cannot parse number: %w", err)
			}
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

		if min, exists := field.Validation["min"]; exists {
			if minVal, ok := min.(float64); ok && num < minVal {
				return fmt.Errorf("minimum value is %f", minVal)
			}
		}
		if max, exists := field.Validation["max"]; exists {
			if maxVal, ok := max.(float64); ok && num > maxVal {
				return fmt.Errorf("maximum value is %f", maxVal)
			}
		}
	}

	// Apply pattern validation for strings
	if field.Type == "string" {
		if pattern, exists := field.Validation["pattern"]; exists {
			if regexPattern, ok := pattern.(string); ok {
				// Basic pattern matching (you might want to use regexp package for more complex patterns)
				if str, ok := value.(string); ok && strings.Contains(regexPattern, "@") && !strings.Contains(str, "@") {
					return fmt.Errorf("value must match pattern: %s", regexPattern)
				}
			}
		}
	}

	return nil
}

// ConvertFieldValues converts field values to appropriate types based on field definitions
func (ch *CollectionsHandler) ConvertFieldValues(ctx context.Context, tenantID uuid.UUID, collectionName string, data map[string]interface{}) (map[string]interface{}, error) {
	// Get collection and field definitions
	collection, err := ch.GetCollection(ctx, tenantID, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	fields, err := ch.GetCollectionFields(ctx, collection.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// Create field map for quick lookup
	fieldMap := make(map[string]CollectionField)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	converted := make(map[string]interface{})

	// Convert each field value
	for fieldName, value := range data {
		field, exists := fieldMap[fieldName]
		if !exists {
			// Skip unknown fields
			continue
		}

		// Convert value based on field type
		convertedValue, err := ch.convertFieldValue(field, value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field '%s': %w", fieldName, err)
		}

		converted[fieldName] = convertedValue
	}

	// Add default values for missing fields
	for _, field := range fields {
		if _, exists := converted[field.Name]; !exists && field.Default != nil {
			converted[field.Name] = field.Default
		}
	}

	return converted, nil
}

// convertFieldValue converts a single field value to the appropriate type
func (ch *CollectionsHandler) convertFieldValue(field CollectionField, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch field.Type {
	case "string", "text":
		return fmt.Sprintf("%v", value), nil

	case "integer", "int":
		switch v := value.(type) {
		case int:
			return v, nil
		case int32:
			return int(v), nil
		case int64:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			var result int
			if _, err := fmt.Sscanf(v, "%d", &result); err != nil {
				return nil, fmt.Errorf("cannot convert '%s' to integer", v)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to integer", value)
		}

	case "float", "decimal":
		switch v := value.(type) {
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int32:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			var result float64
			if _, err := fmt.Sscanf(v, "%f", &result); err != nil {
				return nil, fmt.Errorf("cannot convert '%s' to float", v)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to float", value)
		}

	case "boolean", "bool":
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			switch strings.ToLower(v) {
			case "true", "1", "yes", "on":
				return true, nil
			case "false", "0", "no", "off":
				return false, nil
			default:
				return nil, fmt.Errorf("cannot convert '%s' to boolean", v)
			}
		case int, int32, int64:
			return value != 0, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to boolean", value)
		}

	case "json", "object":
		switch v := value.(type) {
		case map[string]interface{}, []interface{}:
			return v, nil
		case string:
			var result interface{}
			if err := json.Unmarshal([]byte(v), &result); err != nil {
				return nil, fmt.Errorf("invalid JSON: %w", err)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to JSON", value)
		}

	case "date", "datetime":
		switch v := value.(type) {
		case time.Time:
			return v, nil
		case string:
			formats := []string{
				"2006-01-02",
				"2006-01-02T15:04:05Z",
				"2006-01-02T15:04:05.000Z",
				time.RFC3339,
			}
			for _, format := range formats {
				if parsed, err := time.Parse(format, v); err == nil {
					return parsed, nil
				}
			}
			return nil, fmt.Errorf("cannot parse date '%s'", v)
		default:
			return nil, fmt.Errorf("cannot convert %T to date", value)
		}

	default:
		// Unknown type - return as-is
		return value, nil
	}
}

// CreateCollectionItem creates a new item in a collection with full validation
func (ch *CollectionsHandler) CreateCollectionItem(ctx context.Context, userID uuid.UUID, collectionName string, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := ch.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenant: %w", err)
	}

	// Validate data against collection schema
	if err := ch.ValidateCollectionData(ctx, userTenantID, collectionName, data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert field values to appropriate types
	convertedData, err := ch.ConvertFieldValues(ctx, userTenantID, collectionName, data)
	if err != nil {
		return nil, fmt.Errorf("field conversion failed: %w", err)
	}

	// Create the item using dynamic handlers
	err = ch.dynamicHandlers.CreateDynamicItem(ctx, userID, collectionName, convertedData)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return convertedData, nil
}

// GetCollectionItem retrieves a specific item from a collection
func (ch *CollectionsHandler) GetCollectionItem(ctx context.Context, userID uuid.UUID, collectionName string, itemID string) (map[string]interface{}, error) {
	// Get the item using dynamic handlers
	item, err := ch.dynamicHandlers.GetDynamicItem(ctx, userID, collectionName, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return item, nil
}

// UpdateCollectionItem updates an item in a collection with full validation
func (ch *CollectionsHandler) UpdateCollectionItem(ctx context.Context, userID uuid.UUID, collectionName string, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := ch.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenant: %w", err)
	}

	// Validate data against collection schema
	if err := ch.ValidateCollectionData(ctx, userTenantID, collectionName, data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert field values to appropriate types
	convertedData, err := ch.ConvertFieldValues(ctx, userTenantID, collectionName, data)
	if err != nil {
		return nil, fmt.Errorf("field conversion failed: %w", err)
	}

	// Update the item using dynamic handlers
	err = ch.dynamicHandlers.UpdateDynamicItem(ctx, userID, collectionName, itemID, convertedData)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return convertedData, nil
}

// DeleteCollectionItem deletes an item from a collection
func (ch *CollectionsHandler) DeleteCollectionItem(ctx context.Context, userID uuid.UUID, collectionName string, itemID string) error {
	// Delete the item using dynamic handlers
	err := ch.dynamicHandlers.DeleteDynamicItem(ctx, userID, collectionName, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}
