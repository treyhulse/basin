# Simple Remote Installation Script for Go RBAC API (PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install-simple.ps1 -UseBasicParsing | iex }"

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
    Write-Header "🚀 Go RBAC API Installer"
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

# Run main installation
Start-Installation 