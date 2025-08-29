package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	"go-rbac-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	TenantID   uuid.UUID `json:"tenant_id,omitempty"`
	TenantSlug string    `json:"tenant_slug,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(cfg.JWTExpiry)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// GenerateTokenWithTenant creates a JWT token that includes tenant context
func GenerateTokenWithTenant(user models.User, tenantID uuid.UUID, tenantSlug string, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(cfg.JWTExpiry)

	claims := &Claims{
		UserID:     user.ID,
		Email:      user.Email,
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func AuthMiddleware(cfg *config.Config, db *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// First, try to validate as JWT token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*Claims); ok {
				// Get user roles to check if admin
				userRoles, err := db.Queries.GetUserRoles(c.Request.Context(), claims.UserID)
				if err == nil {
					isAdmin := false
					for _, role := range userRoles {
						if role.Name == "admin" {
							isAdmin = true
							break
						}
					}

					// Store user information in context
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
					c.Set("auth_type", "jwt")
					c.Set("is_admin", isAdmin)

					// Store tenant information if present
					if claims.TenantID != uuid.Nil {
						c.Set("tenant_id", claims.TenantID)
						c.Set("tenant_slug", claims.TenantSlug)
					}

					c.Next()
					return
				}
			}
		}

		// If JWT validation failed, try API key authentication
		// Hash the API key to match what's stored in the database
		keyHash := hashAPIKey(tokenString)
		apiKey, err := db.Queries.GetAPIKeyByHash(c.Request.Context(), keyHash)
		if err == nil && apiKey.IsActive.Bool {
			// Check if API key is expired
			if apiKey.ExpiresAt.Valid && apiKey.ExpiresAt.Time.Before(time.Now()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
				c.Abort()
				return
			}

			// Update last used timestamp
			db.Queries.UpdateAPIKeyLastUsed(c.Request.Context(), apiKey.ID)

			// Get user information and roles
			user, err := db.Queries.GetUserByID(c.Request.Context(), apiKey.UserID)
			if err == nil {
				userRoles, err := db.Queries.GetUserRoles(c.Request.Context(), user.ID)
				if err == nil {
					isAdmin := false
					for _, role := range userRoles {
						if role.Name == "admin" {
							isAdmin = true
							break
						}
					}

					c.Set("user_id", user.ID)
					c.Set("email", user.Email)
					c.Set("auth_type", "api_key")
					c.Set("is_admin", isAdmin)
					c.Next()
					return
				}
			}
		}

		// If both JWT and API key authentication failed
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token or API key"})
		c.Abort()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}

	if id, ok := userID.(uuid.UUID); ok {
		return id, true
	}

	return uuid.Nil, false
}

func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}

	if e, ok := email.(string); ok {
		return e, true
	}

	return "", false
}

// GetTenantID retrieves the tenant ID from the context
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, false
	}

	if id, ok := tenantID.(uuid.UUID); ok {
		return id, true
	}

	return uuid.Nil, false
}

// GetTenantSlug retrieves the tenant slug from the context
func GetTenantSlug(c *gin.Context) (string, bool) {
	tenantSlug, exists := c.Get("tenant_slug")
	if !exists {
		return "", false
	}

	if slug, ok := tenantSlug.(string); ok {
		return slug, true
	}

	return "", false
}

// IsAdmin checks if the current user has admin role
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}

	if admin, ok := isAdmin.(bool); ok {
		return admin
	}

	return false
}

// hashAPIKey creates a SHA-256 hash of the API key for secure storage
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
