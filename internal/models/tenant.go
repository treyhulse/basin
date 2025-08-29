package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Domain    string    `json:"domain,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	Name   string `json:"name" binding:"required"`
	Slug   string `json:"slug" binding:"required"`
	Domain string `json:"domain,omitempty"`
}

type UpdateTenantRequest struct {
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type TenantResponse struct {
	Message string `json:"message"`
	Tenant  Tenant `json:"tenant"`
}

type DeleteTenantResponse struct {
	Message string `json:"message"`
}

// User-Tenant relationship
type UserTenant struct {
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	RoleID    uuid.UUID `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type AddUserToTenantRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	TenantID uuid.UUID `json:"tenant_id" binding:"required"`
	RoleID   uuid.UUID `json:"role_id" binding:"required"`
}

type RemoveUserFromTenantRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	TenantID uuid.UUID `json:"tenant_id" binding:"required"`
}

type UserTenantResponse struct {
	Message string `json:"message"`
}
