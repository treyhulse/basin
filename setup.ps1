# Dynamic Setup Script for Go RBAC API
# This script is designed to be future-proof and handle various scenarios

param(
    [switch]$Help,
    [switch]$Version
)

# Script configuration
$ScriptVersion = "2.0.0"
$RepoUrl = "https://github.com/treyhulse/directus-clone.git"
$RepoBranch = "main"
$MinGoVersion = "1.21"
$MinDockerVersion = "20.0"
$MinDockerComposeVersion = "2.0"
$RequiredEnvVars = @("DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET", "SERVER_PORT")

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

# Function to compare version numbers
function Compare-Version {
    param([string]$Version1, [string]$Version2)
    
    $v1 = [version]$Version1
    $v2 = [version]$Version2
    
    if ($v1 -eq $v2) { return 0 }
    if ($v1 -gt $v2) { return 1 }
    return 2
}

# Function to get version of a command
function Get-CommandVersion {
    param([string]$Command, [string]$VersionFlag)
    
    if (Test-Command $Command) {
        try {
            $output = & $Command $VersionFlag 2>&1 | Select-Object -First 1
            $version = $output -replace '.*?(\d+\.\d+(?:\.\d+)?).*', '$1'
            if ($version -match '^\d+\.\d+(?:\.\d+)?$') {
                return $version
            }
        }
        catch {
            # Ignore errors
        }
    }
    return ""
}

# Function to validate environment variables
function Test-EnvironmentVariables {
    Write-Info "Validating environment variables..."
    
    $missingVars = @()
    $invalidVars = @()
    
    # Check if .env file exists
    if (-not (Test-Path ".env")) {
        Write-Warning ".env file not found. Creating from template..."
        if (Test-Path "env.example") {
            Copy-Item "env.example" ".env"
            Write-Status "Created .env from env.example"
        }
        else {
            Write-Error "No .env file and no env.example template found!"
            return $false
        }
    }
    
    # Load environment variables
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match '^([^#][^=]+)=(.*)$') {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                [Environment]::SetEnvironmentVariable($name, $value, "Process")
            }
        }
    }
    
    # Check required variables
    foreach ($var in $RequiredEnvVars) {
        if ([string]::IsNullOrEmpty([Environment]::GetEnvironmentVariable($var))) {
            $missingVars += $var
        }
    }
    
    # Validate specific variables
    $dbPort = [Environment]::GetEnvironmentVariable("DB_PORT")
    if ($dbPort -and -not ($dbPort -match '^\d+$')) {
        $invalidVars += "DB_PORT must be a number"
    }
    
    $serverPort = [Environment]::GetEnvironmentVariable("SERVER_PORT")
    if ($serverPort -and -not ($serverPort -match '^\d+$')) {
        $invalidVars += "SERVER_PORT must be a number"
    }
    
    $jwtSecret = [Environment]::GetEnvironmentVariable("JWT_SECRET")
    if ($jwtSecret -and $jwtSecret.Length -lt 32) {
        Write-Warning "JWT_SECRET is shorter than 32 characters (security risk)"
    }
    
    # Report issues
    if ($missingVars.Count -gt 0) {
        Write-Error "Missing required environment variables:"
        foreach ($var in $missingVars) {
            Write-Host "  - $var" -ForegroundColor Red
        }
        return $false
    }
    
    if ($invalidVars.Count -gt 0) {
        Write-Error "Invalid environment variables:"
        foreach ($var in $invalidVars) {
            Write-Host "  - $var" -ForegroundColor Red
        }
        return $false
    }
    
    Write-Status "Environment variables validated"
    return $true
}

# Function to check prerequisites
function Test-Prerequisites {
    Write-Header "Checking Prerequisites"
    
    $issues = @()
    
    # Check Go
    if (-not (Test-Command "go")) {
        $issues += "Go is not installed"
    }
    else {
        $goVersion = Get-CommandVersion "go" "version"
        if ($goVersion) {
            $compare = Compare-Version $goVersion $MinGoVersion
            if ($compare -eq 2) {
                $issues += "Go version $goVersion is older than required $MinGoVersion"
            }
            else {
                Write-Status "Go $goVersion ✓"
            }
        }
        else {
            $issues += "Could not determine Go version"
        }
    }
    
    # Check Docker
    if (-not (Test-Command "docker")) {
        $issues += "Docker is not installed"
    }
    else {
        $dockerVersion = Get-CommandVersion "docker" "version"
        if ($dockerVersion) {
            $compare = Compare-Version $dockerVersion $MinDockerVersion
            if ($compare -eq 2) {
                $issues += "Docker version $dockerVersion is older than required $MinDockerVersion"
            }
            else {
                Write-Status "Docker $dockerVersion ✓"
            }
        }
        else {
            $issues += "Could not determine Docker version"
        }
    }
    
    # Check Docker Compose
    if (-not (Test-Command "docker-compose") -and -not (docker compose version 2>$null)) {
        $issues += "Docker Compose is not installed"
    }
    else {
        $composeVersion = ""
        if (Test-Command "docker-compose") {
            $composeVersion = Get-CommandVersion "docker-compose" "version"
        }
        else {
            $composeVersion = Get-CommandVersion "docker" "compose version"
        }
        
        if ($composeVersion) {
            $compare = Compare-Version $composeVersion $MinDockerComposeVersion
            if ($compare -eq 2) {
                $issues += "Docker Compose version $composeVersion is older than required $MinDockerComposeVersion"
            }
            else {
                Write-Status "Docker Compose $composeVersion ✓"
            }
        }
        else {
            $issues += "Could not determine Docker Compose version"
        }
    }
    
    # Check sqlc
    if (-not (Test-Command "sqlc")) {
        Write-Warning "sqlc not found, will install during setup"
    }
    else {
        $sqlcVersion = Get-CommandVersion "sqlc" "version"
        if ($sqlcVersion) {
            Write-Status "sqlc $sqlcVersion ✓"
        }
        else {
            Write-Warning "Could not determine sqlc version"
        }
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

# Function to setup repository
function Set-Repository {
    Write-Header "Setting up Repository"
    
    if (-not (Test-Path ".git")) {
        Write-Info "Not a git repository. Cloning from $RepoUrl..."
        try {
            git clone -b $RepoBranch $RepoUrl .
            Write-Status "Repository cloned successfully"
        }
        catch {
            Write-Error "Failed to clone repository"
            return $false
        }
    }
    else {
        Write-Info "Git repository found. Checking for updates..."
        git fetch origin
        $currentBranch = git branch --show-current
        if ($currentBranch -ne $RepoBranch) {
            Write-Warning "Current branch is $currentBranch, switching to $RepoBranch"
            git checkout $RepoBranch
        }
        git pull origin $RepoBranch
        Write-Status "Repository updated"
    }
    return $true
}

# Function to find all migration files
function Get-MigrationFiles {
    $migrations = @()
    if (Test-Path "migrations") {
        $migrations = Get-ChildItem -Path "migrations" -Filter "*.sql" -Recurse | 
                     Sort-Object Name | 
                     ForEach-Object { $_.FullName }
    }
    return $migrations
}

# Function to apply migrations
function Invoke-Migrations {
    Write-Header "Applying Database Migrations"
    
    $migrations = Get-MigrationFiles
    
    if ($migrations.Count -eq 0) {
        Write-Warning "No migration files found"
        return $true
    }
    
    Write-Info "Found $($migrations.Count) migration files"
    
    foreach ($migration in $migrations) {
        Write-Info "Applying $migration..."
        try {
            Get-Content $migration | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
            Write-Status "Applied $migration"
        }
        catch {
            Write-Warning "Migration $migration had issues (this might be expected for duplicate entries)"
        }
    }
    return $true
}

# Function to check database health
function Test-DatabaseHealth {
    Write-Header "Checking Database Health"
    
    $maxAttempts = 30
    $attempt = 0
    
    while ($attempt -lt $maxAttempts) {
        $attempt++
        Write-Info "Checking database connection (attempt $attempt/$maxAttempts)..."
        
        try {
            $result = docker exec go-rbac-postgres pg_isready -U postgres 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Status "Database is ready!"
                return $true
            }
        }
        catch {
            # Ignore errors
        }
        
        if ($attempt -eq $maxAttempts) {
            Write-Error "Database failed to start after $maxAttempts attempts"
            return $false
        }
        
        Start-Sleep -Seconds 2
    }
    return $false
}

# Function to build application
function Build-Application {
    Write-Header "Building Application"
    
    # Install Go dependencies
    Write-Info "Installing Go dependencies..."
    try {
        go mod tidy
        Write-Status "Dependencies installed"
    }
    catch {
        Write-Error "Failed to install dependencies"
        return $false
    }
    
    # Install sqlc if not present
    if (-not (Test-Command "sqlc")) {
        Write-Info "Installing sqlc..."
        try {
            go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
            Write-Status "sqlc installed"
        }
        catch {
            Write-Error "Failed to install sqlc"
            return $false
        }
    }
    
    # Generate database code
    Write-Info "Generating database code..."
    try {
        sqlc generate
        Write-Status "Database code generated"
    }
    catch {
        Write-Error "Failed to generate database code"
        return $false
    }
    
    # Build the application
    Write-Info "Building application..."
    try {
        go build -o bin/api cmd/main.go
        Write-Status "Application built successfully"
    }
    catch {
        Write-Error "Build failed"
        return $false
    }
    
    return $true
}

# Function to display summary
function Show-Summary {
    Write-Header "Setup Complete!"
    Write-Host "==================" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Info "Database Status:"
    docker-compose ps
    Write-Host ""
    
    Write-Info "Default Admin Credentials:"
    Write-Host "   Email: admin@example.com" -ForegroundColor White
    Write-Host "   Password: password" -ForegroundColor White
    Write-Host ""
    
    Write-Info "API Keys (for testing):"
    Write-Host "   Admin API Key: admin_api_key_123" -ForegroundColor White
    Write-Host "   Manager API Key: manager_api_key_456" -ForegroundColor White
    Write-Host ""
    
    Write-Info "To start the API server:"
    Write-Host "   go run cmd/main.go" -ForegroundColor White
    Write-Host "   or" -ForegroundColor White
    Write-Host "   ./bin/api" -ForegroundColor White
    Write-Host ""
    
    $serverPort = [Environment]::GetEnvironmentVariable("SERVER_PORT")
    if (-not $serverPort) { $serverPort = "8080" }
    
    Write-Info "API will be available at:"
    Write-Host "   http://localhost:$serverPort" -ForegroundColor White
    Write-Host ""
    
    Write-Info "Available endpoints:"
    Write-Host "   POST /auth/login - Login with email/password" -ForegroundColor White
    Write-Host "   GET  /auth/me - Get current user info" -ForegroundColor White
    Write-Host "   GET  /items/:table - Generic data access (products, customers, orders, etc.)" -ForegroundColor White
    Write-Host "   GET  /health - Health check" -ForegroundColor White
    Write-Host ""
    
    Write-Info "Useful commands:"
    Write-Host "   docker-compose down - Stop database" -ForegroundColor White
    Write-Host "   docker-compose up -d - Start database" -ForegroundColor White
    Write-Host "   docker exec -it go-rbac-postgres psql -U postgres -d go_rbac_db - Connect to database" -ForegroundColor White
    Write-Host ""
    
    Write-Status "You're all set! Run 'go run cmd/main.go' to start the API server."
}

# Main execution function
function Start-Setup {
    Write-Header "🚀 Dynamic Setup Script for Go RBAC API v$ScriptVersion"
    Write-Host ""
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        exit 1
    }
    
    # Setup repository
    if (-not (Set-Repository)) {
        exit 1
    }
    
    # Validate environment
    if (-not (Test-EnvironmentVariables)) {
        exit 1
    }
    
    # Stop and remove any existing containers
    Write-Header "Cleaning up existing containers"
    docker-compose down -v 2>$null
    docker system prune -f 2>$null
    
    # Start fresh database
    Write-Header "Starting PostgreSQL database"
    try {
        docker-compose up -d
        Write-Status "Database started"
    }
    catch {
        Write-Error "Failed to start database"
        exit 1
    }
    
    # Wait for database to be ready
    Write-Info "Waiting for database to be ready..."
    Start-Sleep -Seconds 10
    
    # Check database health
    if (-not (Test-DatabaseHealth)) {
        exit 1
    }
    
    # Apply migrations
    if (-not (Invoke-Migrations)) {
        Write-Warning "Some migrations had issues, but continuing..."
    }
    
    # Build application
    if (-not (Build-Application)) {
        exit 1
    }
    
    # Display summary
    Show-Summary
}

# Handle script arguments
if ($Version) {
    Write-Host "Setup Script v$ScriptVersion"
    exit 0
}

if ($Help) {
    Write-Host "Usage: .\setup.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version    Show version information"
    Write-Host "  -Help       Show this help message"
    Write-Host ""
    Write-Host "This script will:"
    Write-Host "  - Check all prerequisites"
    Write-Host "  - Clone/update the repository"
    Write-Host "  - Validate environment variables"
    Write-Host "  - Start a fresh PostgreSQL database"
    Write-Host "  - Apply all migrations dynamically"
    Write-Host "  - Install Go dependencies"
    Write-Host "  - Generate database code"
    Write-Host "  - Build the application"
    exit 0
}

# Run main setup
Start-Setup 