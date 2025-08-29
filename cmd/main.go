package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-rbac-api/internal/api"
	"go-rbac-api/internal/config"
	"go-rbac-api/internal/db"
	"go-rbac-api/internal/middleware"

	_ "go-rbac-api/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title        Basin API
// @version      1.0.0
// @description  Directus-style API with Role-Based Access Control (RBAC). A powerful, generic API that provides CRUD operations for any database table with comprehensive security, multi-tenancy, and dynamic schema management.
// @BasePath     /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
// @description  JWT Bearer token for user authentication
// @securityDefinitions.apikey ApiKeyAuth
// @in          header
// @name        Authorization
// @description  API key for programmatic access (format: Bearer YOUR_API_KEY)
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.ServerMode)

	// Initialize database
	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize handlers
	authHandler := api.NewAuthHandler(database, cfg)
	itemsHandler := api.NewItemsHandler(database)
	tenantHandler := api.NewTenantHandler(database, cfg)

	// Setup router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check endpoint
	// @Summary      Health Check
	// @Tags         system
	// @Produce      json
	// @Success      200 {object} models.HealthResponse
	// @Router       /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().UTC(),
		})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/signup", authHandler.SignUp)
		auth.GET("/me", middleware.AuthMiddleware(cfg, database), authHandler.Me)

		// User management (protected routes)
		users := auth.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg, database))
		{
			users.PUT("/:id", authHandler.UpdateUser)
			users.DELETE("/:id", authHandler.DeleteUser)
		}
	}

	// Items routes (protected) - Dynamic table access
	items := router.Group("/items")
	items.Use(middleware.AuthMiddleware(cfg, database))
	{
		items.GET("/:table", itemsHandler.GetItems)
		items.GET("/:table/:id", itemsHandler.GetItem)
		items.POST("/:table", itemsHandler.CreateItem)
		items.PUT("/:table/:id", itemsHandler.UpdateItem)
		items.DELETE("/:table/:id", itemsHandler.DeleteItem)
	}

	// Tenant routes (protected)
	tenant := router.Group("/tenants")
	tenant.Use(middleware.AuthMiddleware(cfg, database))
	{
		tenant.POST("/", tenantHandler.CreateTenant)
		tenant.GET("/", tenantHandler.GetTenants)
		tenant.GET("/:id", tenantHandler.GetTenant)
		tenant.PUT("/:id", tenantHandler.UpdateTenant)
		tenant.DELETE("/:id", tenantHandler.DeleteTenant)

		// User-tenant management
		tenant.POST("/:id/users", tenantHandler.AddUserToTenant)
		tenant.DELETE("/:id/users/:user_id", tenantHandler.RemoveUserFromTenant)
		tenant.POST("/:id/join", tenantHandler.JoinTenant) // New route for users to join tenants
	}

	// API documentation
	// @Summary      API Information
	// @Tags         system
	// @Produce      json
	// @Success      200 {object} models.APIInfoResponse
	// @Router       / [get]
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Dynamic auto-generated REST API with Role-Based Access Control on Postgres",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health": "/health",
				"auth": gin.H{
					"login": "POST /auth/login",
					"me":    "GET /auth/me",
				},
				"items": gin.H{
					"list":   "GET /items/:table",
					"get":    "GET /items/:table/:id",
					"create": "POST /items/:table",
					"update": "PUT /items/:table/:id",
					"delete": "DELETE /items/:table/:id",
				},
			},
			"sample_tables": []string{"blog_posts", "customers"},
			"default_admin": gin.H{
				"email":    "admin@example.com",
				"password": "password",
			},
		})
	})

	// Swagger UI and JSON (auto-generated)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
