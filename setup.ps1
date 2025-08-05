# Comprehensive Setup Script for Go RBAC API (PowerShell)
# Usage: .\setup.ps1

param(
    [switch]$Help,
    [switch]$Version,
    [switch]$SkipEnvCheck,
    [switch]$SkipMigrations,
    [switch]$SkipBuild
)

# Script configuration
$ScriptVersion = "2.0.0"

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor Blue
}

function Write-Header {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Cyan
}

# Function to check if a command exists
function Test-Command {
    param([string]$CommandName)
    return [bool](Get-Command -Name $CommandName -ErrorAction SilentlyContinue)
}

# Function to check prerequisites
function Test-Prerequisites {
    Write-Header "Checking Prerequisites"
    
    $issues = @()
    
    # Check Go
    if (-not (Test-Command "go")) {
        $issues += "Go is not installed. Please install Go 1.21+ from https://golang.org/dl/"
    }
    else {
        Write-Status "Go ✓"
    }
    
    # Check Docker
    if (-not (Test-Command "docker")) {
        $issues += "Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/"
    }
    else {
        Write-Status "Docker ✓"
    }
    
    # Check Docker Compose
    if (-not (Test-Command "docker-compose") -and -not (docker compose version 2>$null)) {
        $issues += "Docker Compose is not installed"
    }
    else {
        Write-Status "Docker Compose ✓"
    }
    
    # Check Git
    if (-not (Test-Command "git")) {
        $issues += "Git is not installed. Please install Git from https://git-scm.com/"
    }
    else {
        Write-Status "Git ✓"
    }
    
    # Report issues
    if ($issues.Count -gt 0) {
        Write-Error "Prerequisite issues found:"
        foreach ($issue in $issues) {
            Write-Host "  - $issue" -ForegroundColor Red
        }
        Write-Host ""
        Write-Info "Please install missing prerequisites and try again."
        return $false
    }
    
    Write-Status "All prerequisites met"
    return $true
}

# Function to setup environment variables
function Setup-Environment {
    Write-Header "Setting Up Environment Variables"
    
    $envFiles = @(".env.local", ".env")
    $envFile = $null
    
    # Check for existing env files
    foreach ($file in $envFiles) {
        if (Test-Path $file) {
            $envFile = $file
            Write-Info "Found existing environment file: $file"
            break
        }
    }
    
    # If no env file exists, create one
    if (-not $envFile) {
        Write-Info "No environment file found. Creating .env from template..."
        
        if (Test-Path "env.example") {
            Copy-Item "env.example" ".env"
            $envFile = ".env"
            Write-Status "Created .env from env.example"
        }
        else {
            Write-Warning "No env.example found. Creating basic .env file..."
            
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
            $envFile = ".env"
            Write-Status "Created basic .env file"
        }
    }
    
    # Load environment variables
    if (Test-Path $envFile) {
        Get-Content $envFile | ForEach-Object {
            if ($_ -match '^([^#][^=]+)=(.*)$') {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                [Environment]::SetEnvironmentVariable($name, $value, "Process")
            }
        }
        Write-Status "Environment variables loaded from $envFile"
    }
    
    # Validate required environment variables
    $requiredVars = @(
        "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
        "JWT_SECRET", "SERVER_PORT", "ADMIN_EMAIL", "ADMIN_PASSWORD"
    )
    
    $missingVars = @()
    foreach ($var in $requiredVars) {
        $value = [Environment]::GetEnvironmentVariable($var, "Process")
        if (-not $value) {
            $missingVars += $var
        }
    }
    
    if ($missingVars.Count -gt 0) {
        Write-Error "Missing required environment variables:"
        foreach ($var in $missingVars) {
            Write-Host "  - $var" -ForegroundColor Red
        }
        Write-Host ""
        Write-Info "Please add these variables to your $envFile file and try again."
        return $false
    }
    
    Write-Status "Environment variables validated"
    return $true
}

# Function to start Docker services
function Start-DockerServices {
    Write-Header "Starting Docker Services"
    
    # Stop any existing containers
    Write-Info "Stopping any existing containers..."
    docker-compose down 2>$null
    
    # Start PostgreSQL
    Write-Info "Starting PostgreSQL database..."
    docker-compose up -d
    
    # Wait for database to be ready
    Write-Info "Waiting for database to be ready..."
    $maxAttempts = 30
    $attempt = 0
    
    do {
        $attempt++
        Start-Sleep -Seconds 2
        
        try {
            $result = docker exec go-rbac-postgres pg_isready -U postgres 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Status "Database is ready"
                return $true
            }
        }
        catch {
            # Ignore errors
        }
        
        Write-Info "Attempt $attempt/$maxAttempts - Database not ready yet..."
    } while ($attempt -lt $maxAttempts)
    
    Write-Error "Database failed to start within expected time"
    return $false
}

# Function to apply migrations
function Apply-Migrations {
    Write-Header "Applying Database Migrations"
    
    # Get all migration files
    $migrations = Get-ChildItem -Path "migrations" -Filter "*.sql" | Sort-Object Name
    
    if ($migrations.Count -eq 0) {
        Write-Warning "No migration files found in migrations/ directory"
        return $true
    }
    
    Write-Info "Found $($migrations.Count) migration files"
    
    foreach ($migration in $migrations) {
        Write-Info "Applying $($migration.Name)..."
        try {
            Get-Content $migration.FullName | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
            if ($LASTEXITCODE -eq 0) {
                Write-Status "Applied $($migration.Name)"
            }
            else {
                Write-Warning "Migration $($migration.Name) had issues (this might be expected)"
            }
        }
        catch {
            Write-Warning "Migration $($migration.Name) had issues (this might be expected)"
        }
    }
    
    Write-Status "Migrations completed"
    return $true
}

# Function to create admin user
function Create-AdminUser {
    Write-Header "Creating Admin User"
    
    $adminEmail = [Environment]::GetEnvironmentVariable("ADMIN_EMAIL", "Process")
    $adminPassword = [Environment]::GetEnvironmentVariable("ADMIN_PASSWORD", "Process")
    $adminFirstName = [Environment]::GetEnvironmentVariable("ADMIN_FIRST_NAME", "Process")
    $adminLastName = [Environment]::GetEnvironmentVariable("ADMIN_LAST_NAME", "Process")
    
    if (-not $adminEmail -or -not $adminPassword) {
        Write-Warning "Admin email or password not set in environment variables"
        return $true
    }
    
    Write-Info "Creating admin user: $adminEmail"
    
    # Check if admin user already exists
    $checkQuery = "SELECT COUNT(*) FROM users WHERE email = '$adminEmail';"
    $existingCount = docker exec go-rbac-postgres psql -U postgres -d go_rbac_db -t -c $checkQuery 2>$null
    
    if ($existingCount -match '\d+' -and [int]$matches[0] -gt 0) {
        Write-Info "Admin user already exists"
        return $true
    }
    
    # Hash the password using the Go utility
    $hashedPassword = ""
    if (Test-Path "scripts/hash_password.go") {
        try {
            $hashedPassword = go run scripts/hash_password.go $adminPassword 2>$null
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "Failed to hash password, using plain text (will be hashed on first login)"
                $hashedPassword = $adminPassword
            }
        }
        catch {
            Write-Warning "Failed to hash password, using plain text (will be hashed on first login)"
            $hashedPassword = $adminPassword
        }
    }
    else {
        Write-Warning "Password hashing utility not found, using plain text (will be hashed on first login)"
        $hashedPassword = $adminPassword
    }
    
    # Create admin user
    $createUserQuery = @"
INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    '$adminEmail',
    '$hashedPassword',
    '$adminFirstName',
    '$adminLastName',
    true,
    NOW(),
    NOW()
);
"@
    
    try {
        $createUserQuery | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
        Write-Status "Admin user created successfully"
    }
    catch {
        Write-Warning "Failed to create admin user (might already exist)"
    }
    
    return $true
}

# Function to install dependencies and generate code
function Install-Dependencies {
    Write-Header "Installing Dependencies and Generating Code"
    
    # Install Go dependencies
    Write-Info "Installing Go dependencies..."
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to install Go dependencies"
        return $false
    }
    Write-Status "Go dependencies installed"
    
    # Install sqlc if not present
    if (-not (Test-Command "sqlc")) {
        Write-Info "Installing sqlc..."
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to install sqlc"
            return $false
        }
        Write-Status "sqlc installed"
    }
    else {
        Write-Status "sqlc already installed"
    }
    
    # Generate database code
    Write-Info "Generating database code..."
    sqlc generate
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to generate database code"
        return $false
    }
    Write-Status "Database code generated"
    
    return $true
}

# Function to build the application
function Build-Application {
    Write-Header "Building Application"
    
    # Create bin directory if it doesn't exist
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    # Build the application
    Write-Info "Building application..."
    go build -o bin/api cmd/main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build application"
        return $false
    }
    
    Write-Status "Application built successfully"
    return $true
}

# Function to display completion message
function Show-Completion {
    Write-Header "🎉 Setup Complete!"
    Write-Host ""
    Write-Info "Your Go RBAC API is ready!"
    Write-Host ""
    Write-Info "Next steps:"
    Write-Host "  1. go run cmd/main.go" -ForegroundColor White
    Write-Host "  2. Or run the built binary: ./bin/api" -ForegroundColor White
    Write-Host ""
    Write-Info "The API will be available at:"
    Write-Host "  http://localhost:8080" -ForegroundColor White
    Write-Host ""
    Write-Info "Default credentials:"
    $adminEmail = [Environment]::GetEnvironmentVariable("ADMIN_EMAIL", "Process")
    $adminPassword = [Environment]::GetEnvironmentVariable("ADMIN_PASSWORD", "Process")
    Write-Host "  Email: $adminEmail" -ForegroundColor White
    Write-Host "  Password: $adminPassword" -ForegroundColor White
    Write-Host ""
    Write-Info "API Keys (for testing):"
    Write-Host "  Admin: admin_api_key_123" -ForegroundColor White
    Write-Host "  Manager: manager_api_key_456" -ForegroundColor White
    Write-Host ""
    Write-Info "Database:"
    Write-Host "  Host: localhost:5432" -ForegroundColor White
    Write-Host "  Database: go_rbac_db" -ForegroundColor White
    Write-Host "  User: postgres" -ForegroundColor White
    Write-Host ""
    Write-Status "Happy coding! 🚀"
}

# Main execution function
function Start-Setup {
    Write-Header "🚀 Go RBAC API Setup v$ScriptVersion"
    Write-Host ""
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        exit 1
    }
    
    # Setup environment
    if (-not $SkipEnvCheck -and -not (Setup-Environment)) {
        exit 1
    }
    
    # Start Docker services
    if (-not (Start-DockerServices)) {
        exit 1
    }
    
    # Apply migrations
    if (-not $SkipMigrations -and -not (Apply-Migrations)) {
        exit 1
    }
    
    # Create admin user
    if (-not (Create-AdminUser)) {
        exit 1
    }
    
    # Install dependencies and generate code
    if (-not (Install-Dependencies)) {
        exit 1
    }
    
    # Build application
    if (-not $SkipBuild -and -not (Build-Application)) {
        exit 1
    }
    
    # Show completion message
    Show-Completion
}

# Handle script arguments
if ($Version) {
    Write-Host "Setup v$ScriptVersion"
    exit 0
}

if ($Help) {
    Write-Host "Usage: .\setup.ps1 [options]"
    Write-Host ""
    Write-Host "This script will:"
    Write-Host "  - Check all prerequisites"
    Write-Host "  - Setup environment variables"
    Write-Host "  - Start Docker services"
    Write-Host "  - Apply database migrations"
    Write-Host "  - Create admin user"
    Write-Host "  - Install dependencies"
    Write-Host "  - Generate database code"
    Write-Host "  - Build the application"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -SkipEnvCheck    Skip environment variable setup"
    Write-Host "  -SkipMigrations  Skip database migrations"
    Write-Host "  -SkipBuild       Skip application build"
    Write-Host "  -Version         Show version information"
    Write-Host "  -Help            Show this help message"
    exit 0
}

# Run main setup
Start-Setup 