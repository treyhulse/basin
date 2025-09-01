# Environment variables
BUILD_ENV ?= production
DOCKER_IMAGE_NAME = basin-backend
DOCKER_TAG = latest

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Basin Backend - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment: BUILD_ENV=$(BUILD_ENV)"
	@echo "Docker Image: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"

# Docker build targets
.PHONY: docker-build
docker-build: ## Build Docker image for current environment
	@echo "Building Docker image for $(BUILD_ENV) environment..."
	docker build \
		--target $(BUILD_ENV) \
		--build-arg BUILD_ENV=$(BUILD_ENV) \
		-t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) \
		.

.PHONY: docker-build-prod
docker-build-prod: ## Build Docker image for production
	@$(MAKE) docker-build BUILD_ENV=production

.PHONY: docker-build-dev
docker-build-dev: ## Build Docker image for development
	@$(MAKE) docker-build BUILD_ENV=development

# Local development targets
.PHONY: local-dev
local-dev: ## Start local development environment with Docker Compose
	@echo "Starting local development environment..."
	docker-compose up --build

.PHONY: local-dev-detached
local-dev-detached: ## Start local development environment in background
	@echo "Starting local development environment in background..."
	docker-compose up --build -d

.PHONY: local-stop
local-stop: ## Stop local development environment
	@echo "Stopping local development environment..."
	docker-compose down

.PHONY: local-logs
local-logs: ## Show local development logs
	docker-compose logs -f

# Railway deployment targets
.PHONY: railway-deploy
railway-deploy: ## Deploy to Railway using Docker
	@echo "Deploying to Railway using Docker..."
	@echo ""
	@echo "Prerequisites:"
	@echo "  1. Install Railway CLI: npm install -g @railway/cli"
	@echo "  2. Login to Railway: railway login"
	@echo "  3. Link to project: railway link"
	@echo "  4. Ensure PostgreSQL service exists in Railway"
	@echo ""
	@echo "Building production Docker image..."
	@$(MAKE) docker-build-prod
	@echo "Deploying to Railway..."
	railway up

.PHONY: railway-logs
railway-logs: ## View Railway logs
	@echo "Fetching Railway logs..."
	railway logs

.PHONY: railway-status
railway-status: ## Check Railway deployment status
	@echo "Checking Railway status..."
	railway status

# Database targets
.PHONY: test-db-local
test-db-local: ## Test local database connection
	@echo "Testing local database connection..."
	@echo "Note: Make sure local-dev is running first"
	@go run test_connection.go.bak

.PHONY: test-db-railway
test-db-railway: ## Test Railway database connection
	@echo "Testing Railway database connection..."
	@echo "Note: You need DATABASE_URL set for this to work"
	@DEPLOYMENT_MODE=railway go run test_connection.go.bak

.PHONY: test-db-auto
test-db-auto: ## Test database connection with auto-detection
	@echo "Testing database connection with auto-detection..."
	@go run test_connection.go.bak

# Utility targets
.PHONY: clean
clean: ## Clean up Docker images and containers
	@echo "Cleaning up Docker resources..."
	docker system prune -f
	docker image prune -f

.PHONY: clean-all
clean-all: ## Clean up all Docker resources (including images)
	@echo "Cleaning up all Docker resources..."
	docker system prune -af

.PHONY: docker-shell
docker-shell: ## Open shell in running container
	@echo "Opening shell in running container..."
	docker-compose exec app sh

# Migration targets (now handled automatically by the app)
.PHONY: railway-migrate
railway-migrate: ## Migrations now run automatically during deployment!
	@echo "Migrations now run automatically during deployment!"
	@echo ""
	@echo "To run migrations:"
	@echo "  1. Deploy your app: make railway-deploy"
	@echo "  2. Migrations will run automatically on startup"
	@echo ""
	@echo "To check migration status, view Railway logs:"
	@echo "  make railway-logs" 