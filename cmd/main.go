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
// @version      1.0
// @description  Directus-style API with Role-Based Access Control (RBAC)
// @BasePath     /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
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
		auth.GET("/me", middleware.AuthMiddleware(cfg, database), authHandler.Me)
	}

	// Items routes (protected)
	items := router.Group("/items")
	items.Use(middleware.AuthMiddleware(cfg, database))
	{
		items.GET("/:table", itemsHandler.GetItems)
		items.GET("/:table/:id", itemsHandler.GetItem)
		items.POST("/:table", itemsHandler.CreateItem)
		items.PUT("/:table/:id", itemsHandler.UpdateItem)
		items.DELETE("/:table/:id", itemsHandler.DeleteItem)
	}

	// API documentation
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Go RBAC API - Directus-style API with Role-Based Access Control",
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
			"sample_tables": []string{"products", "customers", "orders"},
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
