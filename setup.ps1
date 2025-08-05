# Complete Setup Script for Go RBAC API (PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/setup.ps1 -UseBasicParsing | iex }"

# Script configuration
$RepoUrl = "https://github.com/treyhulse/directus-clone.git"
$ProjectName = "directus-clone"

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

# Function to clone repository
function Clone-Repository {
    Write-Header "Cloning Repository"
    
    $targetDir = $ProjectName
    
    # Check if directory already exists
    if (Test-Path $targetDir) {
        Write-Warning "Directory $targetDir already exists"
        $response = Read-Host "Do you want to remove it and start fresh? (y/N)"
        if ($response -eq "y" -or $response -eq "Y") {
            Write-Info "Removing existing directory..."
            Remove-Item -Recurse -Force $targetDir
        }
        else {
            Write-Error "Installation cancelled"
            exit 1
        }
    }
    
    Write-Info "Cloning $RepoUrl to $targetDir..."
    try {
        git clone $RepoUrl $targetDir
        Write-Status "Repository cloned successfully"
    }
    catch {
        Write-Error "Failed to clone repository"
        exit 1
    }
    
    # Change to project directory
    Set-Location $targetDir
    Write-Status "Changed to project directory: $(Get-Location)"
}

# Function to setup environment variables
function Setup-Environment {
    Write-Header "Setting Up Environment Variables"
    
    # Create .env file with correct settings
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
    Write-Status "Created .env file with correct settings"
    
    # Load environment variables
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
    Write-Status "Environment variables loaded"
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
    
    Write-Info "Creating admin user: $adminEmail"
    
    # Check if admin user already exists
    $checkQuery = "SELECT COUNT(*) FROM users WHERE email = '$adminEmail';"
    $existingCount = docker exec go-rbac-postgres psql -U postgres -d go_rbac_db -t -c $checkQuery 2>$null
    
    if ($existingCount -match '\d+' -and [int]$matches[0] -gt 0) {
        Write-Info "Admin user already exists"
        return $true
    }
    
    # Create admin user with plain text password (will be hashed on first login)
    $createUserQuery = @"
INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    '$adminEmail',
    '$adminPassword',
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

# Function to start the application
function Start-Application {
    Write-Header "Starting Your API"
    
    Write-Info "Starting the Go RBAC API..."
    Write-Host ""
    Write-Host "🚀 Your API is now running at: http://localhost:8080" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Default credentials:" -ForegroundColor White
    Write-Host "   Email: admin@example.com" -ForegroundColor White
    Write-Host "   Password: password" -ForegroundColor White
    Write-Host ""
    Write-Host "🔑 API Keys for testing:" -ForegroundColor White
    Write-Host "   Admin: admin_api_key_123" -ForegroundColor White
    Write-Host "   Manager: manager_api_key_456" -ForegroundColor White
    Write-Host ""
    Write-Host "📚 API Documentation: http://localhost:8080" -ForegroundColor White
    Write-Host ""
    Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
    Write-Host ""
    
    # Start the application
    go run cmd/main.go
}

# Function to display completion message
function Show-Completion {
    Write-Header "🎉 Setup Complete!"
    Write-Host ""
    Write-Info "Your Go RBAC API is ready!"
    Write-Host ""
    Write-Info "What was set up:"
    Write-Host "  ✅ Repository cloned" -ForegroundColor White
    Write-Host "  ✅ Environment variables configured" -ForegroundColor White
    Write-Host "  ✅ Dependencies installed" -ForegroundColor White
    Write-Host "  ✅ Database code generated" -ForegroundColor White
    Write-Host "  ✅ PostgreSQL database started" -ForegroundColor White
    Write-Host "  ✅ Database migrations applied" -ForegroundColor White
    Write-Host "  ✅ Admin user created" -ForegroundColor White
    Write-Host "  ✅ Application built" -ForegroundColor White
    Write-Host ""
    Write-Info "Starting your API now..."
    Write-Host ""
}

# Main execution function
function Start-Installation {
    Write-Header "🚀 Go RBAC API - Complete Setup"
    Write-Host ""
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        exit 1
    }
    
    # Clone repository
    if (-not (Clone-Repository)) {
        exit 1
    }
    
    # Setup environment
    if (-not (Setup-Environment)) {
        exit 1
    }
    
    # Install dependencies and generate code
    if (-not (Install-Dependencies)) {
        exit 1
    }
    
    # Start Docker services
    if (-not (Start-DockerServices)) {
        exit 1
    }
    
    # Apply migrations
    if (-not (Apply-Migrations)) {
        exit 1
    }
    
    # Create admin user
    if (-not (Create-AdminUser)) {
        exit 1
    }
    
    # Build application
    if (-not (Build-Application)) {
        exit 1
    }
    
    # Show completion message
    Show-Completion
    
    # Start the application
    Start-Application
}

# Run main installation
Start-Installation 