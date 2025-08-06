package schema

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SchemaManager struct {
	db *sql.DB
}

type Collection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fields      []Field   `json:"fields,omitempty"`
}

type Field struct {
	ID              uuid.UUID              `json:"id"`
	CollectionID    uuid.UUID              `json:"collection_id"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Type            string                 `json:"type"`
	IsPrimary       bool                   `json:"is_primary"`
	IsRequired      bool                   `json:"is_required"`
	IsUnique        bool                   `json:"is_unique"`
	DefaultValue    string                 `json:"default_value"`
	ValidationRules map[string]interface{} `json:"validation_rules"`
	RelationConfig  map[string]interface{} `json:"relation_config"`
	SortOrder       int                    `json:"sort_order"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func NewSchemaManager(db *sql.DB) *SchemaManager {
	return &SchemaManager{db: db}
}

// CreateCollection creates a new collection and its corresponding data table
func (sm *SchemaManager) CreateCollection(ctx context.Context, collection Collection) error {
	// Start transaction
	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert collection record
	collectionQuery := `
		INSERT INTO collections (id, name, display_name, description, icon, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, collectionQuery,
		collection.ID, collection.Name, collection.DisplayName,
		collection.Description, collection.Icon, collection.IsSystem)
	if err != nil {
		return fmt.Errorf("failed to insert collection: %w", err)
	}

	// Create corresponding data table
	dataTableName := "data_" + collection.Name
	createTableQuery := sm.buildCreateTableQuery(dataTableName, collection.Fields)

	_, err = tx.ExecContext(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create data table: %w", err)
	}

	// Insert fields
	for _, field := range collection.Fields {
		fieldQuery := `
			INSERT INTO fields (id, collection_id, name, display_name, type, is_primary, 
				is_required, is_unique, default_value, validation_rules, relation_config, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err = tx.ExecContext(ctx, fieldQuery,
			field.ID, collection.ID, field.Name, field.DisplayName, field.Type,
			field.IsPrimary, field.IsRequired, field.IsUnique, field.DefaultValue,
			field.ValidationRules, field.RelationConfig, field.SortOrder)
		if err != nil {
			return fmt.Errorf("failed to insert field: %w", err)
		}
	}

	return tx.Commit()
}

// UpdateCollection updates a collection and its data table schema
func (sm *SchemaManager) UpdateCollection(ctx context.Context, collection Collection) error {
	// Start transaction
	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update collection record
	collectionQuery := `
		UPDATE collections 
		SET display_name = $2, description = $3, icon = $4, updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, collectionQuery,
		collection.ID, collection.DisplayName, collection.Description, collection.Icon)
	if err != nil {
		return fmt.Errorf("failed to update collection: %w", err)
	}

	// Handle field updates (simplified - in production you'd need more sophisticated schema migration)
	// For now, we'll just update the field records
	for _, field := range collection.Fields {
		fieldQuery := `
			UPDATE fields 
			SET display_name = $3, type = $4, is_primary = $5, is_required = $6, 
				is_unique = $7, default_value = $8, validation_rules = $9, 
				relation_config = $10, sort_order = $11, updated_at = NOW()
			WHERE id = $1 AND collection_id = $2
		`
		_, err = tx.ExecContext(ctx, fieldQuery,
			field.ID, collection.ID, field.DisplayName, field.Type,
			field.IsPrimary, field.IsRequired, field.IsUnique, field.DefaultValue,
			field.ValidationRules, field.RelationConfig, field.SortOrder)
		if err != nil {
			return fmt.Errorf("failed to update field: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteCollection deletes a collection and its data table
func (sm *SchemaManager) DeleteCollection(ctx context.Context, collectionID uuid.UUID) error {
	// Start transaction
	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get collection name
	var collectionName string
	err = tx.QueryRowContext(ctx, "SELECT name FROM collections WHERE id = $1", collectionID).Scan(&collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection name: %w", err)
	}

	// Drop data table
	dataTableName := "data_" + collectionName
	dropTableQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", dataTableName)
	_, err = tx.ExecContext(ctx, dropTableQuery)
	if err != nil {
		return fmt.Errorf("failed to drop data table: %w", err)
	}

	// Delete collection (fields will be deleted via CASCADE)
	_, err = tx.ExecContext(ctx, "DELETE FROM collections WHERE id = $1", collectionID)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return tx.Commit()
}

// GetCollection retrieves a collection with its fields
func (sm *SchemaManager) GetCollection(ctx context.Context, collectionID uuid.UUID) (*Collection, error) {
	// Get collection
	var collection Collection
	collectionQuery := `
		SELECT id, name, display_name, description, icon, is_system, created_at, updated_at
		FROM collections WHERE id = $1
	`
	err := sm.db.QueryRowContext(ctx, collectionQuery, collectionID).Scan(
		&collection.ID, &collection.Name, &collection.DisplayName, &collection.Description,
		&collection.Icon, &collection.IsSystem, &collection.CreatedAt, &collection.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Get fields
	fieldsQuery := `
		SELECT id, collection_id, name, display_name, type, is_primary, is_required, 
			is_unique, default_value, validation_rules, relation_config, sort_order, created_at, updated_at
		FROM fields WHERE collection_id = $1 ORDER BY sort_order
	`
	rows, err := sm.db.QueryContext(ctx, fieldsQuery, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field Field
		err := rows.Scan(
			&field.ID, &field.CollectionID, &field.Name, &field.DisplayName, &field.Type,
			&field.IsPrimary, &field.IsRequired, &field.IsUnique, &field.DefaultValue,
			&field.ValidationRules, &field.RelationConfig, &field.SortOrder,
			&field.CreatedAt, &field.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field: %w", err)
		}
		collection.Fields = append(collection.Fields, field)
	}

	return &collection, nil
}

// ListCollections retrieves all collections
func (sm *SchemaManager) ListCollections(ctx context.Context) ([]Collection, error) {
	query := `
		SELECT id, name, display_name, description, icon, is_system, created_at, updated_at
		FROM collections ORDER BY name
	`
	rows, err := sm.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var collection Collection
		err := rows.Scan(
			&collection.ID, &collection.Name, &collection.DisplayName, &collection.Description,
			&collection.Icon, &collection.IsSystem, &collection.CreatedAt, &collection.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan collection: %w", err)
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// buildCreateTableQuery builds the SQL to create a data table based on fields
func (sm *SchemaManager) buildCreateTableQuery(tableName string, fields []Field) string {
	var columns []string

	// Always add standard columns
	columns = append(columns, "id UUID PRIMARY KEY DEFAULT uuid_generate_v4()")
	columns = append(columns, "created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()")
	columns = append(columns, "updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()")

	// Add field columns
	for _, field := range fields {
		columnDef := sm.buildColumnDefinition(field)
		columns = append(columns, columnDef)
	}

	query := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", tableName, strings.Join(columns, ",\n  "))
	return query
}

// buildColumnDefinition builds the SQL definition for a field
func (sm *SchemaManager) buildColumnDefinition(field Field) string {
	var parts []string

	// Column name
	parts = append(parts, fmt.Sprintf(`"%s"`, field.Name))

	// Data type
	switch field.Type {
	case "string":
		parts = append(parts, "VARCHAR(255)")
	case "text":
		parts = append(parts, "TEXT")
	case "integer":
		parts = append(parts, "INTEGER")
	case "decimal":
		parts = append(parts, "DECIMAL(10,2)")
	case "boolean":
		parts = append(parts, "BOOLEAN")
	case "datetime":
		parts = append(parts, "TIMESTAMP WITH TIME ZONE")
	case "json":
		parts = append(parts, "JSONB")
	case "uuid":
		parts = append(parts, "UUID")
	case "relation":
		// Handle relation fields - would need to reference the related table
		if relConfig, ok := field.RelationConfig["related_collection"].(string); ok {
			parts = append(parts, fmt.Sprintf("UUID REFERENCES data_%s(id)", relConfig))
		} else {
			parts = append(parts, "UUID")
		}
	default:
		parts = append(parts, "TEXT")
	}

	// Constraints
	if field.IsRequired {
		parts = append(parts, "NOT NULL")
	}

	if field.IsUnique {
		parts = append(parts, "UNIQUE")
	}

	if field.DefaultValue != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", field.DefaultValue))
	}

	return strings.Join(parts, " ")
}
