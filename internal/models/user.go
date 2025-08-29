package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	TenantSlug string `json:"tenant_slug,omitempty"` // Optional: specify tenant for login
}

type LoginResponse struct {
	Token      string    `json:"token"`
	User       User      `json:"user"`
	TenantID   uuid.UUID `json:"tenant_id,omitempty"`
	TenantSlug string    `json:"tenant_slug,omitempty"`
}

type SignUpRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	TenantSlug string `json:"tenant_slug,omitempty"` // Optional: create/join tenant during signup
}

type SignUpResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}

type SwitchTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" binding:"required"`
}

type SwitchTenantResponse struct {
	Token      string    `json:"token"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantSlug string    `json:"tenant_slug"`
	Message    string    `json:"message"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
