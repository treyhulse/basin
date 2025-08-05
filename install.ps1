# Remote Installation Script for Go RBAC API (PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.ps1 -UseBasicParsing | iex }"

param(
    [switch]$Help,
    [switch]$Version
)

# Script configuration
$ScriptVersion = "1.0.0"
$RepoUrl = "https://github.com/treyhulse/directus-clone.git"
$RepoBranch = "main"
$ProjectName = "directus-clone"
$MinGoVersion = "1.21"
$MinDockerVersion = "20.0"
$MinDockerComposeVersion = "2.0"

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

# Function to check prerequisites
function Test-Prerequisites {
    Write-Header "Checking Prerequisites"
    
    $issues = @()
    
    # Check Go
    if (-not (Test-Command "go")) {
        $issues += "Go is not installed. Please install Go 1.21+ from https://golang.org/dl/"
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
        $issues += "Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/"
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
        git clone -b $RepoBranch $RepoUrl $targetDir
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

# Function to run the setup script
function Start-Setup {
    Write-Header "Running Setup"
    
    if (Test-Path "setup.ps1") {
        Write-Info "Found setup.ps1, running it..."
        & .\setup.ps1
    }
    else {
        Write-Warning "setup.ps1 not found, running manual setup..."
        
        # Install sqlc if not present
        if (-not (Test-Command "sqlc")) {
            Write-Info "Installing sqlc..."
            go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        }
        
        # Start database
        Write-Info "Starting PostgreSQL database..."
        docker-compose up -d
        
        # Wait for database
        Write-Info "Waiting for database to be ready..."
        Start-Sleep -Seconds 15
        
        # Apply migrations
        Write-Info "Applying migrations..."
        $migrations = Get-ChildItem -Path "migrations" -Filter "*.sql" -Recurse
        foreach ($migration in $migrations) {
            Write-Info "Applying $($migration.Name)..."
            try {
                Get-Content $migration.FullName | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
            }
            catch {
                Write-Warning "Migration $($migration.Name) had issues (this might be expected)"
            }
        }
        
        # Install dependencies and build
        Write-Info "Installing Go dependencies..."
        go mod tidy
        
        Write-Info "Generating database code..."
        sqlc generate
        
        Write-Info "Building application..."
        go build -o bin/api cmd/main.go
        
        Write-Status "Setup completed"
    }
}

# Function to display completion message
function Show-Completion {
    Write-Header "🎉 Installation Complete!"
    Write-Host ""
    Write-Info "Your Go RBAC API is ready!"
    Write-Host ""
    Write-Info "Next steps:"
    Write-Host "  1. cd $ProjectName" -ForegroundColor White
    Write-Host "  2. go run cmd/main.go" -ForegroundColor White
    Write-Host ""
    Write-Info "The API will be available at:"
    Write-Host "  http://localhost:8080" -ForegroundColor White
    Write-Host ""
    Write-Info "Default credentials:"
    Write-Host "  Email: admin@example.com" -ForegroundColor White
    Write-Host "  Password: password" -ForegroundColor White
    Write-Host ""
    Write-Info "API Keys (for testing):"
    Write-Host "  Admin: admin_api_key_123" -ForegroundColor White
    Write-Host "  Manager: manager_api_key_456" -ForegroundColor White
    Write-Host ""
    Write-Status "Happy coding! 🚀"
}

# Main execution function
function Start-Installation {
    Write-Header "🚀 Go RBAC API Installer v$ScriptVersion"
    Write-Host ""
    
    # Check prerequisites
    if (-not (Test-Prerequisites)) {
        exit 1
    }
    
    # Clone repository
    if (-not (Clone-Repository)) {
        exit 1
    }
    
    # Run setup
    if (-not (Start-Setup)) {
        exit 1
    }
    
    # Show completion message
    Show-Completion
}

# Handle script arguments
if ($Version) {
    Write-Host "Installer v$ScriptVersion"
    exit 0
}

if ($Help) {
    Write-Host "Usage: powershell -ExecutionPolicy Bypass -Command `"& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.ps1 -UseBasicParsing | iex }`""
    Write-Host ""
    Write-Host "This script will:"
    Write-Host "  - Check all prerequisites"
    Write-Host "  - Clone the repository"
    Write-Host "  - Run the setup script"
    Write-Host "  - Build the application"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Version    Show version information"
    Write-Host "  -Help       Show this help message"
    exit 0
}

# Run main installation
Start-Installation 