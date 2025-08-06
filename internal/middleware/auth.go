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
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
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
				// Store user information in context
				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				c.Set("auth_type", "jwt")
				c.Next()
				return
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

			// Get user information
			user, err := db.Queries.GetUserByID(c.Request.Context(), apiKey.UserID)
			if err == nil {
				c.Set("user_id", user.ID)
				c.Set("email", user.Email)
				c.Set("auth_type", "api_key")
				c.Next()
				return
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

// hashAPIKey creates a SHA-256 hash of the API key for secure storage
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}
