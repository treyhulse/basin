.PHONY: help setup start stop restart dev build test test-verbose test-coverage clean deps generate sqlc migrate docker-up docker-down docker-logs docs

# Default target
help:
	@echo Basin API - Available Commands:
	@echo.
	@echo Setup ^& Start:
	@echo   setup     - Complete initial setup (env, deps, docker, migrations, build)
	@echo   start     - Cold start the application (docker, sqlc, build, run)
	@echo   restart   - Stop and restart the application
	@echo   dev       - Start development server with hot reload
	@echo.
	@echo Database Management:
	@echo   migrate   - Apply database migrations (manual use only)
	@echo   docker-up   - Start PostgreSQL with Docker
	@echo   docker-down - Stop PostgreSQL
	@echo   docker-logs - Show Docker logs
	@echo.
	@echo Build ^& Development:
	@echo   build     - Build the application
	@echo   test      - Run all tests (unit + integration)
	@echo   test-verbose - Run tests with detailed output  
	@echo   test-coverage - Run tests with coverage report
	@echo   clean     - Clean build artifacts
	@echo   deps      - Download dependencies
	@echo   generate  - Generate sqlc code
	@echo   docs      - Generate Swagger docs
	@echo   sqlc      - Run sqlc generate
	@echo.
	@echo Examples:
	@echo   make setup   # First time setup
	@echo   make start   # Start the application
	@echo   make restart # Restart everything

# Complete initial setup
setup:
	@echo Basin API - Complete Setup
	@echo.
	@echo Checking prerequisites...
	@powershell -Command "if (-not (Get-Command 'go' -ErrorAction SilentlyContinue)) { Write-Host 'ERROR: Go is not installed' -ForegroundColor Red; exit 1 }"
	@powershell -Command "if (-not (Get-Command 'docker' -ErrorAction SilentlyContinue)) { Write-Host 'ERROR: Docker is not installed' -ForegroundColor Red; exit 1 }"
	@echo SUCCESS: Prerequisites check passed
	@echo.
	@echo Setting up environment...
	@if not exist .env (copy env.example .env)
	@echo SUCCESS: Environment configured
	@echo.
	@echo Installing dependencies...
	@go mod tidy
	@echo SUCCESS: Dependencies installed
	@echo.
	@echo Installing sqlc...
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo SUCCESS: sqlc installed
	@echo.
	@echo Starting database...
	@docker-compose down 2>nul
	@docker-compose up -d
	@echo Waiting for database to be ready...
	@powershell -Command "$$attempt = 0; do { $$attempt++; Start-Sleep -Seconds 2; $$result = docker exec go-rbac-postgres pg_isready -U postgres 2>$$null; if ($$LASTEXITCODE -eq 0) { Write-Host 'SUCCESS: Database ready' -ForegroundColor Green; break } } while ($$attempt -lt 30)"
	@echo.
	@echo Applying migrations...
	@for %%f in (migrations\*.sql) do @echo   Applying %%~nxf... && type "%%f" | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
	@echo SUCCESS: Migrations applied
	@echo.
	@echo Generating database code...
	@sqlc generate
	@echo SUCCESS: Database code generated
	@echo.
	@echo Building application...
	@if not exist bin mkdir bin
	@go build -o bin/api cmd/main.go
	@echo SUCCESS: Application built
	@echo.
	@echo Setup complete! Run 'make start' to start the server.

# Cold start the application
start:
	@echo Basin API - Cold Start
	@echo.
	@echo Stopping any existing containers...
	@docker-compose down 2>nul
	@echo.
	@echo Starting database...
	@docker-compose up -d
	@echo Waiting for database to be ready...
	@powershell -Command "$$attempt = 0; do { $$attempt++; Start-Sleep -Seconds 2; $$result = docker exec go-rbac-postgres pg_isready -U postgres 2>$$null; if ($$LASTEXITCODE -eq 0) { Write-Host 'SUCCESS: Database ready' -ForegroundColor Green; break } } while ($$attempt -lt 30)"
	@echo.
	@echo Generating database code...
	@sqlc generate
	@echo SUCCESS: Database code generated
	@echo.
	@echo Generating Swagger docs...
	@swag init -g cmd/main.go -o docs 2>nul || echo "(tip) Install swag with: go install github.com/swaggo/swag/cmd/swag@latest"
	@echo Building application...
	@if not exist bin mkdir bin
	@go build -o bin/api cmd/main.go
	@echo SUCCESS: Application built
	@echo.
	@echo Starting Basin API server...
	@echo Server will be available at: http://localhost:8080
	@echo Health check: http://localhost:8080/health
	@echo.
	@echo Default admin credentials:
	@echo   Email: admin@example.com
	@echo   Password: password
	@echo.
	@echo Press Ctrl+C to stop the server
	@echo ================================
	@go run cmd/main.go

# Restart the application
restart: stop start

# Stop the application
stop:
	@echo Stopping Basin API...
	@docker-compose down
	@echo SUCCESS: Application stopped

# Development server with hot reload
dev:
	@echo Starting development server...
	@echo Generating Swagger docs...
	@swag init -g cmd/main.go -o docs 2>nul || echo "(tip) Install swag with: go install github.com/swaggo/swag/cmd/swag@latest"
	@go run cmd/main.go

# Build the application
build:
	@echo Building application...
	@if not exist bin mkdir bin
	@go build -o bin/api cmd/main.go
	@echo SUCCESS: Application built

# Run tests (clean output)
test:
	@echo Running all tests (unit + integration)...
	@echo.
	@echo Starting database if not running...
	@docker-compose up -d
	@echo Waiting for database to be ready...
	@powershell -Command "$$attempt = 0; do { $$attempt++; Start-Sleep -Seconds 2; $$result = docker exec go-rbac-postgres pg_isready -U postgres 2>$$null; if ($$LASTEXITCODE -eq 0) { Write-Host 'SUCCESS: Database ready' -ForegroundColor Green; break } } while ($$attempt -lt 30)"
	@echo.
	@echo Running all tests...
	@go test ./...
	@echo.
	@echo Tests completed successfully!

# Run tests with verbose output
test-verbose:
	@echo Running tests with verbose output...
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@echo Running tests with coverage...
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@del coverage.out
	@echo.
	@echo Coverage report generated: coverage.html



# Clean build artifacts
clean:
	@echo Cleaning build artifacts...
	@if exist bin rmdir /s /q bin
	@go clean
	@echo SUCCESS: Cleaned

# Download dependencies
deps:
	@echo Downloading dependencies...
	@go mod download
	@go mod tidy
	@echo SUCCESS: Dependencies updated

# Generate sqlc code
generate: sqlc

# Run sqlc generate
sqlc:
	@echo Generating sqlc code...
	@sqlc generate
	@echo SUCCESS: Database code generated

# Generate swagger docs
docs:
	@echo Generating Swagger docs...
	@swag init -g cmd/main.go -o docs
	@echo SUCCESS: Swagger docs generated at docs/

# Apply database migrations (only if needed)
migrate:
	@echo Applying database migrations...
	@echo Checking if database is running...
	@docker exec go-rbac-postgres pg_isready -U postgres >nul 2>&1 || (echo Database not running, starting it... && docker-compose up -d)
	@echo Waiting for database to be ready...
	@powershell -Command "$$attempt = 0; do { $$attempt++; Start-Sleep -Seconds 2; $$result = docker exec go-rbac-postgres pg_isready -U postgres 2>$$null; if ($$LASTEXITCODE -eq 0) { Write-Host 'SUCCESS: Database ready' -ForegroundColor Green; break } } while ($$attempt -lt 30)"
	@echo.
	@for %%f in (migrations\*.sql) do @echo   Applying %%~nxf... && type "%%f" | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
	@echo SUCCESS: Migrations applied

# Start PostgreSQL with Docker
docker-up:
	@echo Starting PostgreSQL with Docker...
	@docker-compose up -d
	@echo SUCCESS: Database started

# Stop PostgreSQL
docker-down:
	@echo Stopping PostgreSQL...
	@docker-compose down
	@echo SUCCESS: Database stopped

# Show Docker logs
docker-logs:
	@echo Showing Docker logs...
	@docker-compose logs -f

# Install sqlc (if not already installed)
install-sqlc:
	@echo Installing sqlc...
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo SUCCESS: sqlc installed 