package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	sqlc "go-rbac-api/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthProvider provides centralized authentication context and session management
type AuthProvider struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	TenantID    uuid.UUID `json:"tenant_id"`
	TenantSlug  string    `json:"tenant_slug"`
	IsAdmin     bool      `json:"is_admin"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	SessionID   string    `json:"session_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantSlug string    `json:"tenant_slug"`
	SessionID  string    `json:"session_id"`
	jwt.RegisteredClaims
}

// Session represents a tenant-scoped authentication session
type Session struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// GenerateTokenWithTenant creates a JWT token that includes user and tenant information
func GenerateTokenWithTenant(user sqlc.User, tenant sqlc.Tenant, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(cfg.JWTExpiry)
	sessionID := uuid.New().String()

	claims := &Claims{
		UserID:     user.ID,
		Email:      user.Email,
		TenantID:   tenant.ID,
		TenantSlug: tenant.Slug,
		SessionID:  sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// GenerateToken creates a JWT token without tenant context (for system-wide operations)
func GenerateToken(user sqlc.User, cfg *config.Config) (string, error) {
	expirationTime := time.Now().Add(cfg.JWTExpiry)
	sessionID := uuid.New().String()

	claims := &Claims{
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// AuthMiddleware creates a middleware that validates JWT tokens and provides auth context
func AuthMiddleware(cfg *config.Config, db *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// Parse and validate JWT token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			// Get user roles and permissions
			userRoles, err := db.Queries.GetUserRoles(c.Request.Context(), claims.UserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
				c.Abort()
				return
			}

			// Check if user is admin
			isAdmin := false
			roles := make([]string, 0, len(userRoles))
			for _, role := range userRoles {
				roles = append(roles, role.Name)
				if role.Name == "admin" {
					isAdmin = true
				}
			}

			// Get user permissions if tenant context exists
			var permissions []string
			if claims.TenantID != uuid.Nil {
				userPermissions, err := db.Queries.GetPermissionsByUserAndTenant(c.Request.Context(), sqlc.GetPermissionsByUserAndTenantParams{
					UserID:   claims.UserID,
					TenantID: uuid.NullUUID{UUID: claims.TenantID, Valid: true},
				})
				if err == nil {
					permissions = make([]string, 0, len(userPermissions))
					for _, perm := range userPermissions {
						permissions = append(permissions, fmt.Sprintf("%s:%s", perm.TableName, perm.Action))
					}
				}
			}

			// Create auth provider
			authProvider := &AuthProvider{
				UserID:      claims.UserID,
				Email:       claims.Email,
				TenantID:    claims.TenantID,
				TenantSlug:  claims.TenantSlug,
				IsAdmin:     isAdmin,
				Roles:       roles,
				Permissions: permissions,
				SessionID:   claims.SessionID,
				ExpiresAt:   time.Unix(int64(claims.ExpiresAt.Unix()), 0),
			}

			// Store auth provider in context
			c.Set("auth", authProvider)
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("tenant_id", claims.TenantID)
			c.Set("tenant_slug", claims.TenantSlug)
			c.Set("is_admin", isAdmin)
			c.Set("auth_type", "jwt")

			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
	}
}

// GetAuthProvider retrieves the auth provider from the context
func GetAuthProvider(c *gin.Context) (*AuthProvider, bool) {
	auth, exists := c.Get("auth")
	if !exists {
		return nil, false
	}

	if provider, ok := auth.(*AuthProvider); ok {
		return provider, true
	}

	return nil, false
}

// GetUserID retrieves the user ID from the context
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

// RequireTenant creates a middleware that requires a tenant context
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := GetTenantID(c)
		if !exists || tenantID == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant context required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission creates a middleware that requires a specific permission
func RequirePermission(tableName, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, exists := GetAuthProvider(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Admin bypass
		if auth.IsAdmin {
			c.Next()
			return
		}

		// Check for specific permission
		requiredPermission := fmt.Sprintf("%s:%s", tableName, action)
		hasPermission := false
		for _, permission := range auth.Permissions {
			if permission == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, exists := GetAuthProvider(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Admin bypass
		if auth.IsAdmin {
			c.Next()
			return
		}

		// Check for specific role
		hasRole := false
		for _, role := range auth.Roles {
			if role == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}
