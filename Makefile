.PHONY: help dev build run test clean deps generate sqlc

# Default target
help:
	@echo "Available commands:"
	@echo "  dev       - Start development server with hot reload"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  deps      - Download dependencies"
	@echo "  generate  - Generate sqlc code"
	@echo "  sqlc      - Run sqlc generate"
	@echo "  docker-up - Start PostgreSQL with Docker"
	@echo "  docker-down - Stop PostgreSQL"

# Development server
dev:
	@echo "Starting development server..."
	@go run cmd/main.go

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/go-rbac-api cmd/main.go

# Run the application
run: build
	@echo "Running application..."
	@./bin/go-rbac-api

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Generate sqlc code
generate: sqlc

# Run sqlc generate
sqlc:
	@echo "Generating sqlc code..."
	@sqlc generate

# Start PostgreSQL with Docker
docker-up:
	@echo "Starting PostgreSQL with Docker..."
	@docker-compose up -d

# Stop PostgreSQL
docker-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down

# Setup development environment
setup: deps docker-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Setup complete! Run 'make dev' to start the server."

# Install sqlc (if not already installed)
install-sqlc:
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest 