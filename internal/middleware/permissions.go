package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks if the user has permission to access a specific table/action
// Admin users automatically bypass all permission checks
func PermissionMiddleware(tableName, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth provider to check admin status
		auth, exists := GetAuthProvider(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Admin users bypass all permission checks
		if auth.IsAdmin {
			c.Next()
			return
		}

		// For non-admin users, you can implement permission checking here
		// This would involve checking the permissions table against the user's roles
		// For now, we'll just allow access (you can implement the actual logic later)

		// Example permission check (commented out for now):
		// hasPermission := false
		// for _, permission := range auth.Permissions {
		//     if permission == fmt.Sprintf("%s:%s", tableName, action) {
		//         hasPermission = true
		//         break
		//     }
		// }
		// if !hasPermission {
		//     c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		//     c.Abort()
		//     return
		// }

		c.Next()
	}
}

// CollectionPermissionMiddleware checks permissions for collection operations
func CollectionPermissionMiddleware(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth provider to check admin status
		auth, exists := GetAuthProvider(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Admin users bypass all permission checks
		if auth.IsAdmin {
			c.Next()
			return
		}

		// For non-admin users, implement collection-specific permission checking
		// This would check if the user has access to the specific collection

		c.Next()
	}
}

// DataPermissionMiddleware checks permissions for data operations within collections
func DataPermissionMiddleware(collectionName, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get auth provider to check admin status
		auth, exists := GetAuthProvider(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Admin users bypass all permission checks
		if auth.IsAdmin {
			c.Next()
			return
		}

		// For non-admin users, implement data-specific permission checking
		// This would check if the user has access to the specific data within the collection

		c.Next()
	}
}
