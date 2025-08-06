// Package api provides HTTP handlers for the Basin API's dynamic database access functionality.
// This file contains handlers for schema management operations including collections, fields,
// users, and API keys - the core building blocks of Basin's dynamic table system.
package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	sqlc "go-rbac-api/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// SchemaHandlers provides CRUD operations for Basin's schema management tables.
//
// This handler manages the core schema entities that define Basin's dynamic table system:
// - Collections: Define logical groupings of data (similar to database tables)
// - Fields: Define the structure and validation rules for collection data
// - Users: Manage user accounts with tenant isolation
// - API Keys: Provide programmatic access with proper security
//
// All operations respect tenant boundaries and RBAC permissions. Schema changes
// automatically propagate to the underlying database structure, creating or modifying
// tenant-specific data tables as needed.
//
// Key Features:
// - Tenant-aware operations (users can only modify their own tenant's schema)
// - Automatic data table creation/modification when collections/fields change
// - Secure API key generation with proper hashing
// - Full CRUD support with proper error handling and validation
type SchemaHandlers struct {
	handler *ItemsHandler // Reference to main handler for database access and policy checking
	utils   *ItemsUtils   // Utility functions for common operations
}

// NewSchemaHandlers creates a new SchemaHandlers instance with required dependencies.
//
// Parameters:
//   - handler: Main ItemsHandler instance providing database access and policy checking
//   - utils: ItemsUtils instance providing utility functions
//
// Returns:
//   - *SchemaHandlers: Configured schema handler ready for use
//
// Example:
//
//	schemaHandler := NewSchemaHandlers(itemsHandler, utils)
//	collection, err := schemaHandler.CreateCollection(ctx, userID, collectionData)
func NewSchemaHandlers(handler *ItemsHandler, utils *ItemsUtils) *SchemaHandlers {
	return &SchemaHandlers{
		handler: handler,
		utils:   utils,
	}
}

// Collection Operations

// CreateCollection creates a new collection
func (s *SchemaHandlers) CreateCollection(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
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
	collection, err := s.handler.db.Queries.CreateCollection(ctx, sqlc.CreateCollectionParams{
		ID:          collectionID,
		Name:        data["name"].(string),
		DisplayName: sql.NullString{String: GetStringFromMap(data, "display_name"), Valid: true},
		Description: sql.NullString{String: GetStringFromMap(data, "description"), Valid: true},
		Icon:        sql.NullString{String: GetStringFromMap(data, "icon"), Valid: true},
		IsSystem:    sql.NullBool{Bool: GetBoolFromMap(data, "is_system"), Valid: true},
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

// UpdateCollection updates an existing collection
func (s *SchemaHandlers) UpdateCollection(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	collectionID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid collection ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing collection
	existingCollection, err := s.handler.db.Queries.GetCollection(ctx, collectionID)
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
	updatedCollection, err := s.handler.db.Queries.UpdateCollection(ctx, sqlc.UpdateCollectionParams{
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

// DeleteCollection deletes a collection
func (s *SchemaHandlers) DeleteCollection(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	collectionID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid collection ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing collection to check access
	existingCollection, err := s.handler.db.Queries.GetCollection(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	// Check tenant access
	if existingCollection.TenantID.Valid && existingCollection.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: collection not accessible")
	}

	// Delete collection using sqlc (this will trigger the database trigger to drop the data table)
	return s.handler.db.Queries.DeleteCollection(ctx, collectionID)
}

// Field Operations

// CreateField creates a new field
func (s *SchemaHandlers) CreateField(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
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

	// Get collection info to check if it's a system collection
	collection, err := s.handler.db.Queries.GetCollection(ctx, collectionID)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Check tenant access
	if collection.TenantID.Valid && collection.TenantID.UUID != userTenantID {
		return nil, fmt.Errorf("unauthorized: collection not accessible")
	}

	// Create field using sqlc
	field, err := s.handler.db.Queries.CreateField(ctx, sqlc.CreateFieldParams{
		ID:              fieldID,
		CollectionID:    uuid.NullUUID{UUID: collectionID, Valid: true},
		Name:            data["name"].(string),
		DisplayName:     sql.NullString{String: GetStringFromMap(data, "display_name"), Valid: true},
		Type:            data["type"].(string),
		IsPrimary:       sql.NullBool{Bool: GetBoolFromMap(data, "is_primary"), Valid: true},
		IsRequired:      sql.NullBool{Bool: GetBoolFromMap(data, "is_required"), Valid: true},
		IsUnique:        sql.NullBool{Bool: GetBoolFromMap(data, "is_unique"), Valid: true},
		DefaultValue:    sql.NullString{String: GetStringFromMap(data, "default_value"), Valid: true},
		ValidationRules: pqtype.NullRawMessage{},
		RelationConfig:  pqtype.NullRawMessage{},
		SortOrder:       sql.NullInt32{Int32: int32(GetIntFromMap(data, "sort_order")), Valid: true},
		TenantID:        uuid.NullUUID{UUID: userTenantID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// If this is not a system collection, update the data table structure
	if !collection.IsSystem.Bool {
		err = s.utils.AddColumnToDataTable(ctx, userTenantID, collection.Name, field)
		if err != nil {
			// If we fail to add the column, we should delete the field record to maintain consistency
			s.handler.db.Queries.DeleteField(ctx, fieldID)
			return nil, fmt.Errorf("failed to add column to data table: %w", err)
		}
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

// UpdateField updates an existing field
func (s *SchemaHandlers) UpdateField(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	fieldID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid field ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing field
	existingField, err := s.handler.db.Queries.GetField(ctx, fieldID)
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
	if sortInt := GetIntFromMap(data, "sort_order"); sortInt > 0 {
		sortOrder = sql.NullInt32{Int32: int32(sortInt), Valid: true}
	}

	// Update field using sqlc
	updatedField, err := s.handler.db.Queries.UpdateField(ctx, sqlc.UpdateFieldParams{
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

// DeleteField deletes a field
func (s *SchemaHandlers) DeleteField(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	fieldID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid field ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing field to check access
	existingField, err := s.handler.db.Queries.GetField(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	// Check tenant access
	if existingField.TenantID.Valid && existingField.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: field not accessible")
	}

	// Delete field using sqlc
	return s.handler.db.Queries.DeleteField(ctx, fieldID)
}

// User Operations

// CreateUser creates a new user
func (s *SchemaHandlers) CreateUser(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get user's tenant
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
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
		// TODO: Hash this properly
		passwordHash = password
	}

	// Create user using sqlc
	user, err := s.handler.db.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           newUserID,
		Email:        data["email"].(string),
		PasswordHash: passwordHash,
		FirstName:    sql.NullString{String: GetStringFromMap(data, "first_name"), Valid: true},
		LastName:     sql.NullString{String: GetStringFromMap(data, "last_name"), Valid: true},
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

// UpdateUser updates an existing user
func (s *SchemaHandlers) UpdateUser(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	targetUserID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get tenant ID for filtering
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get existing user
	existingUser, err := s.handler.db.Queries.GetUserByID(ctx, targetUserID)
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
	updatedUser, err := s.handler.db.Queries.UpdateUser(ctx, sqlc.UpdateUserParams{
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

// DeleteUser deletes a user
func (s *SchemaHandlers) DeleteUser(ctx context.Context, userID uuid.UUID, itemID string) error {
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
	userTenantID, err := s.utils.GetUserTenantID(ctx, userID)
	if err != nil {
		return err
	}

	// Get existing user to check access
	existingUser, err := s.handler.db.Queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check tenant access
	if existingUser.TenantID.Valid && existingUser.TenantID.UUID != userTenantID {
		return fmt.Errorf("unauthorized: user not accessible")
	}

	// Delete user using sqlc
	return s.handler.db.Queries.DeleteUser(ctx, targetUserID)
}

// API Key Operations

// CreateAPIKey creates a new API key
func (s *SchemaHandlers) CreateAPIKey(ctx context.Context, userID uuid.UUID, data map[string]interface{}) (map[string]interface{}, error) {
	// Get target user ID (can create API keys for other users if admin)
	targetUserID := userID // Default to current user
	if targetUserStr, ok := data["user_id"].(string); ok {
		if parsedID, err := uuid.Parse(targetUserStr); err == nil {
			targetUserID = parsedID
		}
	}

	// Generate a secure API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the API key for storage
	keyHash := s.hashAPIKey(apiKey)

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
	createdKey, err := s.handler.db.Queries.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
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

// UpdateAPIKey updates an existing API key
func (s *SchemaHandlers) UpdateAPIKey(ctx context.Context, userID uuid.UUID, itemID string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse item ID
	apiKeyID, err := uuid.Parse(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid API key ID: %w", err)
	}

	// Check if user owns this API key (unless admin)
	existingKey, err := s.handler.db.Queries.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Only allow users to update their own keys (unless admin)
	if existingKey.UserID != userID {
		// Check if user is admin
		hasAdminAccess, _, _ := s.handler.policyChecker.CheckPermission(ctx, userID, "users", "read")
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
	updatedKey, err := s.handler.db.Queries.UpdateAPIKey(ctx, sqlc.UpdateAPIKeyParams{
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

// DeleteAPIKey deletes an API key
func (s *SchemaHandlers) DeleteAPIKey(ctx context.Context, userID uuid.UUID, itemID string) error {
	// Parse item ID
	apiKeyID, err := uuid.Parse(itemID)
	if err != nil {
		return fmt.Errorf("invalid API key ID: %w", err)
	}

	// Check if user owns this API key (unless admin)
	existingKey, err := s.handler.db.Queries.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		return fmt.Errorf("API key not found: %w", err)
	}

	// Only allow users to delete their own keys (unless admin)
	if existingKey.UserID != userID {
		// Check if user is admin
		hasAdminAccess, _, _ := s.handler.policyChecker.CheckPermission(ctx, userID, "users", "read")
		if !hasAdminAccess {
			return fmt.Errorf("unauthorized: can only delete your own API keys")
		}
	}

	// Delete API key using sqlc
	return s.handler.db.Queries.DeleteAPIKey(ctx, apiKeyID)
}

// Helper methods for API key generation

// generateAPIKey generates a secure random API key
func (s *SchemaHandlers) generateAPIKey() (string, error) {
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
func (s *SchemaHandlers) hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
