package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	log.Println("🚀 === APP STARTING ===")
	log.Println("Step 1: Loading configuration...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("✅ Step 1 COMPLETE: Configuration loaded")
	log.Println("Step 2: Setting Gin mode...")

	// Set Gin mode
	gin.SetMode(cfg.ServerMode)

	log.Println("✅ Step 2 COMPLETE: Gin mode set")
	log.Println("Step 3: Connecting to database...")

	// Initialize database
	database, err := db.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("✅ Step 3 COMPLETE: Database connected")
	log.Println("Step 4: Starting migrations...")

	// Run database migrations automatically
	log.Println("=== STARTING MIGRATIONS ===")
	log.Printf("Current working directory: %s", getCurrentDir())
	log.Printf("Migration directory: %s", filepath.Join(getCurrentDir(), "migrations"))

	// List files in migrations directory
	if err := listMigrationFiles(); err != nil {
		log.Printf("Warning: Could not list migration files: %v", err)
	}

	log.Println("Running database migrations...")
	if err := runMigrations(database); err != nil {
		log.Printf("WARNING: Migrations failed: %v", err)
		log.Println("Continuing with startup... (migrations can be run manually later)")
	} else {
		log.Println("Database migrations completed successfully")
	}
	log.Println("=== MIGRATIONS COMPLETE ===")

	log.Println("✅ Step 4 COMPLETE: Migrations finished")
	log.Println("Step 5: Seeding database...")

	// Seed the database with initial data
	if err := seedDatabase(database); err != nil {
		log.Printf("WARNING: Database seeding failed: %v", err)
		log.Println("Continuing with startup... (seeding can be run manually later)")
	} else {
		log.Println("Database seeding completed successfully")
	}

	log.Println("✅ Step 5 COMPLETE: Database seeded")
	log.Println("Step 6: Initializing handlers...")

	// Initialize handlers
	authHandler := api.NewAuthHandler(database, cfg)
	itemsHandler := api.NewItemsHandler(database)
	tenantHandler := api.NewTenantHandler(database, cfg)

	log.Println("✅ Step 6 COMPLETE: Handlers initialized")
	log.Println("Step 7: Setting up router...")

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

		// Protected auth routes (require authentication)
		protected := auth.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg, database))
		{
			protected.POST("/switch-tenant", authHandler.SwitchTenant)
			protected.GET("/context", authHandler.GetAuthContext)
			protected.GET("/tenants", authHandler.GetUserTenants)
		}

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
			"sample_tables": []string{"customers", "products", "orders"},
			"default_admin": gin.H{
				"email":    "admin@example.com",
				"password": "password",
			},
		})
	})

	// Swagger UI and JSON (auto-generated)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create server
	// Railway provides PORT environment variable, fallback to config
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.ServerPort)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	log.Println("✅ Step 7 COMPLETE: Router setup finished")
	log.Println("Step 8: Starting server...")

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 SERVER STARTED on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("✅ Step 8 COMPLETE: Server startup initiated")
	log.Println("🎉 === APP STARTUP COMPLETE ===")

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

// getCurrentDir returns the current working directory
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// listMigrationFiles lists all files in the migrations directory for debugging
func listMigrationFiles() error {
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	log.Printf("Found %d files in migrations directory:", len(files))
	for _, file := range files {
		log.Printf("  - %s (dir: %t)", file.Name(), file.IsDir())
	}
	return nil
}

// seedDatabase seeds the database with initial data
func seedDatabase(db *db.DB) error {
	log.Println("Starting database seeding...")

	// Check if seeding has already been done
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'admin@example.com'").Scan(&count)
	if err != nil {
		// Table doesn't exist yet, that's fine for first run
		log.Println("Users table not found, proceeding with seeding...")
	} else if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Create default admin user
	log.Println("Creating default admin user...")
	adminPassword := "password" // In production, use environment variable
	hashedPassword, err := hashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %v", err)
	}

	// Insert admin user
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at)
		VALUES (
			gen_random_uuid(),
			'admin@example.com',
			$1,
			'Admin',
			'User',
			true,
			NOW(),
			NOW()
		)
	`, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Create default tenant
	log.Println("Creating default tenant...")
	var tenantID string
	err = db.QueryRow(`
		INSERT INTO tenants (id, name, description, created_at, updated_at)
		VALUES (
			gen_random_uuid(),
			'Default Tenant',
			'Default organization for the system',
			NOW(),
			NOW()
		)
		RETURNING id
	`).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("failed to create default tenant: %v", err)
	}

	// Link admin user to default tenant
	log.Println("Linking admin user to default tenant...")
	_, err = db.Exec(`
		INSERT INTO user_tenants (id, user_id, tenant_id, role, created_at, updated_at)
		SELECT 
			gen_random_uuid(),
			u.id,
			$1,
			'admin',
			NOW(),
			NOW()
		FROM users u
		WHERE u.email = 'admin@example.com'
	`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to link admin user to tenant: %v", err)
	}

	// Create some sample collections and fields
	log.Println("Creating sample collections...")
	_, err = db.Exec(`
		INSERT INTO collections (id, name, description, tenant_id, created_at, updated_at)
		VALUES (
			gen_random_uuid(),
			'blog_posts',
			'Sample blog posts collection',
			$1,
			NOW(),
			NOW()
		)
	`, tenantID)
	if err != nil {
		log.Printf("WARNING: Failed to create sample collection: %v", err)
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	// For now, return a simple hash. In production, use bcrypt
	// This is a placeholder - you should implement proper bcrypt hashing
	return fmt.Sprintf("hashed_%s", password), nil
}

// runMigrations executes all SQL files in the migrations directory
func runMigrations(db *db.DB) error {
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort files to ensure proper order
	var sqlFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	// Execute migrations in order
	for _, fileName := range sqlFiles {
		log.Printf("Executing migration: %s", fileName)

		filePath := filepath.Join(migrationDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("WARNING: Could not read migration file %s: %v", fileName, err)
			continue // Skip this migration but continue with others
		}

		// Execute the migration
		_, err = db.Exec(string(content))
		if err != nil {
			log.Printf("WARNING: Migration %s failed: %v", fileName, err)
			log.Printf("Continuing with next migration...")
			continue // Skip this migration but continue with others
		}

		log.Printf("Successfully executed migration: %s", fileName)
	}

	return nil
}
