package models

import "time"

// API Response Models for Swagger Documentation

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string    `json:"status" example:"ok"`
	Time   time.Time `json:"time" example:"2024-01-01T00:00:00Z"`
}

// APIInfoResponse represents the root endpoint response
type APIInfoResponse struct {
	Message    string                 `json:"message" example:"Go RBAC API - Directus-style API with Role-Based Access Control"`
	Version    string                 `json:"version" example:"1.0.0"`
	Endpoints  map[string]interface{} `json:"endpoints"`
	SampleData []string               `json:"sample_tables" example:"customers,blog_posts,collections,fields"`
	AdminInfo  map[string]string      `json:"default_admin"`
}

// ItemsListResponse represents a paginated list of items
type ItemsListResponse struct {
	Data []map[string]interface{} `json:"data"`
	Meta ItemsListMeta            `json:"meta"`
}

// ItemsListMeta represents metadata for item list responses
type ItemsListMeta struct {
	Table  string `json:"table" example:"customers"`
	Count  int    `json:"count" example:"25"`
	Total  int    `json:"total" example:"100"`
	Limit  int    `json:"limit" example:"25"`
	Offset int    `json:"offset" example:"0"`
	Type   string `json:"type" example:"data"`
}

// ItemResponse represents a single item response
type ItemResponse struct {
	Data map[string]interface{} `json:"data"`
	Meta ItemMeta               `json:"meta"`
}

// ItemMeta represents metadata for single item responses
type ItemMeta struct {
	Table string `json:"table" example:"customers"`
	ID    string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// CreateItemResponse represents a create item response
type CreateItemResponse struct {
	Data map[string]interface{} `json:"data"`
	Meta CreateItemMeta         `json:"meta"`
}

// CreateItemMeta represents metadata for create item responses
type CreateItemMeta struct {
	Table   string `json:"table" example:"customers"`
	Message string `json:"message" example:"Item created successfully"`
}

// UpdateItemResponse represents an update item response
type UpdateItemResponse struct {
	Data map[string]interface{} `json:"data"`
	Meta UpdateItemMeta         `json:"meta"`
}

// UpdateItemMeta represents metadata for update item responses
type UpdateItemMeta struct {
	Table string `json:"table" example:"customers"`
	ID    string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// DeleteItemResponse represents a delete item response
type DeleteItemResponse struct {
	Meta DeleteItemMeta `json:"meta"`
}

// DeleteItemMeta represents metadata for delete item responses
type DeleteItemMeta struct {
	Table   string `json:"table" example:"customers"`
	ID      string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Message string `json:"message" example:"Item deleted successfully"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid table name"`
	Details string `json:"details,omitempty" example:"Table 'invalid_table' does not exist or is not accessible"`
	Code    string `json:"code,omitempty" example:"INVALID_TABLE"`
}
