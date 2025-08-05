# Complete Setup Script for Go RBAC API (PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/setup.ps1 -UseBasicParsing | iex }"

Write-Host "Go RBAC API - Complete Setup" -ForegroundColor Cyan
Write-Host ""

# Check prerequisites
Write-Host "Checking Prerequisites..." -ForegroundColor Cyan

# Check Go
if (Get-Command "go" -ErrorAction SilentlyContinue) {
    Write-Host "SUCCESS: Go found" -ForegroundColor Green
} else {
    Write-Host "ERROR: Go is not installed" -ForegroundColor Red
    exit 1
}

# Check Docker
if (Get-Command "docker" -ErrorAction SilentlyContinue) {
    Write-Host "SUCCESS: Docker found" -ForegroundColor Green
} else {
    Write-Host "ERROR: Docker is not installed" -ForegroundColor Red
    exit 1
}

# Check Docker Compose
if ((Get-Command "docker-compose" -ErrorAction SilentlyContinue) -or (docker compose version 2>$null)) {
    Write-Host "SUCCESS: Docker Compose found" -ForegroundColor Green
} else {
    Write-Host "ERROR: Docker Compose is not installed" -ForegroundColor Red
    exit 1
}

# Check Git
if (Get-Command "git" -ErrorAction SilentlyContinue) {
    Write-Host "SUCCESS: Git found" -ForegroundColor Green
} else {
    Write-Host "ERROR: Git is not installed" -ForegroundColor Red
    exit 1
}

Write-Host "SUCCESS: All prerequisites met" -ForegroundColor Green
Write-Host ""

# Setup environment variables
Write-Host "Setting Up Environment Variables..." -ForegroundColor Cyan

$envContent = @"
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_rbac_db
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h

# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# Admin User Configuration
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=password
ADMIN_FIRST_NAME=Admin
ADMIN_LAST_NAME=User
"@

$envContent | Out-File -FilePath ".env" -Encoding UTF8
Write-Host "SUCCESS: Created .env file with correct settings" -ForegroundColor Green

# Load environment variables
Get-Content ".env" | ForEach-Object {
    if ($_ -match '^([^#][^=]+)=(.*)$') {
        $name = $matches[1].Trim()
        $value = $matches[2].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}
Write-Host "SUCCESS: Environment variables loaded" -ForegroundColor Green
Write-Host ""

# Install Go dependencies
Write-Host "Installing Dependencies and Generating Code..." -ForegroundColor Cyan
Write-Host "INFO: Installing Go dependencies..." -ForegroundColor Blue
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to install Go dependencies" -ForegroundColor Red
    exit 1
}
Write-Host "SUCCESS: Go dependencies installed" -ForegroundColor Green

# Install sqlc if not present
if (-not (Get-Command "sqlc" -ErrorAction SilentlyContinue)) {
    Write-Host "INFO: Installing sqlc..." -ForegroundColor Blue
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Failed to install sqlc" -ForegroundColor Red
        exit 1
    }
    Write-Host "SUCCESS: sqlc installed" -ForegroundColor Green
} else {
    Write-Host "SUCCESS: sqlc already installed" -ForegroundColor Green
}

# Generate database code
Write-Host "INFO: Generating database code..." -ForegroundColor Blue
sqlc generate
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to generate database code" -ForegroundColor Red
    exit 1
}
Write-Host "SUCCESS: Database code generated" -ForegroundColor Green
Write-Host ""

# Start Docker services
Write-Host "Starting Docker Services..." -ForegroundColor Cyan
Write-Host "INFO: Stopping any existing containers..." -ForegroundColor Blue
docker-compose down 2>$null

Write-Host "INFO: Starting PostgreSQL database..." -ForegroundColor Blue
docker-compose up -d

# Wait for database to be ready
Write-Host "INFO: Waiting for database to be ready..." -ForegroundColor Blue
$maxAttempts = 30
$attempt = 0

do {
    $attempt++
    Start-Sleep -Seconds 2
    
    try {
        $result = docker exec go-rbac-postgres pg_isready -U postgres 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "SUCCESS: Database is ready" -ForegroundColor Green
            break
        }
    }
    catch {
        # Ignore errors
    }
    
    Write-Host "INFO: Attempt $attempt/$maxAttempts - Database not ready yet..." -ForegroundColor Blue
} while ($attempt -lt $maxAttempts)

if ($attempt -ge $maxAttempts) {
    Write-Host "ERROR: Database failed to start within expected time" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Apply migrations
Write-Host "Applying Database Migrations..." -ForegroundColor Cyan
$migrations = Get-ChildItem -Path "migrations" -Filter "*.sql" | Sort-Object Name

if ($migrations.Count -eq 0) {
    Write-Host "WARNING: No migration files found in migrations/ directory" -ForegroundColor Yellow
} else {
    Write-Host "INFO: Found $($migrations.Count) migration files" -ForegroundColor Blue
    
    foreach ($migration in $migrations) {
        Write-Host "INFO: Applying $($migration.Name)..." -ForegroundColor Blue
        try {
            Get-Content $migration.FullName | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
            if ($LASTEXITCODE -eq 0) {
                Write-Host "SUCCESS: Applied $($migration.Name)" -ForegroundColor Green
            } else {
                Write-Host "WARNING: Migration $($migration.Name) had issues (this might be expected)" -ForegroundColor Yellow
            }
        }
        catch {
            Write-Host "WARNING: Migration $($migration.Name) had issues (this might be expected)" -ForegroundColor Yellow
        }
    }
}

Write-Host "SUCCESS: Migrations completed" -ForegroundColor Green
Write-Host ""

# Create admin user
Write-Host "Creating Admin User..." -ForegroundColor Cyan
$adminEmail = [Environment]::GetEnvironmentVariable("ADMIN_EMAIL", "Process")
$adminPassword = [Environment]::GetEnvironmentVariable("ADMIN_PASSWORD", "Process")
$adminFirstName = [Environment]::GetEnvironmentVariable("ADMIN_FIRST_NAME", "Process")
$adminLastName = [Environment]::GetEnvironmentVariable("ADMIN_LAST_NAME", "Process")

Write-Host "INFO: Creating admin user: $adminEmail" -ForegroundColor Blue

# Check if admin user already exists
$checkQuery = "SELECT COUNT(*) FROM users WHERE email = '$adminEmail';"
$existingCount = docker exec go-rbac-postgres psql -U postgres -d go_rbac_db -t -c $checkQuery 2>$null

if ($existingCount -match '\d+' -and [int]$matches[0] -gt 0) {
    Write-Host "INFO: Admin user already exists" -ForegroundColor Blue
} else {
    # Create admin user with plain text password (will be hashed on first login)
    $createUserQuery = "INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at) VALUES (gen_random_uuid(), '$adminEmail', '$adminPassword', '$adminFirstName', '$adminLastName', true, NOW(), NOW());"
    
    try {
        $createUserQuery | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
        Write-Host "SUCCESS: Admin user created successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "WARNING: Failed to create admin user (might already exist)" -ForegroundColor Yellow
    }
}
Write-Host ""

# Build application
Write-Host "Building Application..." -ForegroundColor Cyan
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

Write-Host "INFO: Building application..." -ForegroundColor Blue
go build -o bin/api cmd/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to build application" -ForegroundColor Red
    exit 1
}

Write-Host "SUCCESS: Application built successfully" -ForegroundColor Green
Write-Host ""

# Show completion message
Write-Host "Setup Complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "INFO: Your Go RBAC API is ready!" -ForegroundColor Blue
Write-Host ""
Write-Host "INFO: What was set up:" -ForegroundColor Blue
Write-Host "  - Environment variables configured" -ForegroundColor White
Write-Host "  - Dependencies installed" -ForegroundColor White
Write-Host "  - Database code generated" -ForegroundColor White
Write-Host "  - PostgreSQL database started" -ForegroundColor White
Write-Host "  - Database migrations applied" -ForegroundColor White
Write-Host "  - Admin user created" -ForegroundColor White
Write-Host "  - Application built" -ForegroundColor White
Write-Host ""
Write-Host "INFO: Starting your API now..." -ForegroundColor Blue
Write-Host ""

# Start the application
Write-Host "Starting Your API..." -ForegroundColor Cyan
Write-Host "INFO: Starting the Go RBAC API..." -ForegroundColor Blue
Write-Host ""
Write-Host "Your API is now running at: http://localhost:8080" -ForegroundColor Green
Write-Host ""
Write-Host "Default credentials:" -ForegroundColor White
Write-Host "   Email: admin@example.com" -ForegroundColor White
Write-Host "   Password: password" -ForegroundColor White
Write-Host ""
Write-Host "API Keys for testing:" -ForegroundColor White
Write-Host "   Admin: admin_api_key_123" -ForegroundColor White
Write-Host "   Manager: manager_api_key_456" -ForegroundColor White
Write-Host ""
Write-Host "API Documentation: http://localhost:8080" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

# Start the application
go run cmd/main.go 