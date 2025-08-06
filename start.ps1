# Basin API Start Script
# This script sets up and starts the Basin API server

Write-Host "Starting Basin API..." -ForegroundColor Green

# Step 1: Start Docker containers
Write-Host "Starting Docker containers..." -ForegroundColor Yellow
docker-compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to start Docker containers" -ForegroundColor Red
    exit 1
}

# Step 2: Wait for PostgreSQL to be ready
Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
do {
    $attempt++
    Start-Sleep -Seconds 2
    $result = docker exec go-rbac-postgres pg_isready -U postgres 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "PostgreSQL is ready!" -ForegroundColor Green
        break
    }
    Write-Host "Attempt $attempt/$maxAttempts - PostgreSQL not ready yet..." -ForegroundColor Gray
} while ($attempt -lt $maxAttempts)

if ($attempt -eq $maxAttempts) {
    Write-Host "PostgreSQL failed to start within timeout" -ForegroundColor Red
    exit 1
}

# Step 3: Run database migrations
Write-Host "Running database migrations..." -ForegroundColor Yellow
$migrations = @(
    "001_init.sql",
    "002_api_keys.sql", 
    "003_admin_permissions.sql",
    "004_schema_management.sql",
    "005_multi_tenant.sql",
    "006_admin_system_permissions.sql",
    "007_fix_default_schema.sql"
)

foreach ($migration in $migrations) {
    $migrationPath = "migrations/$migration"
    if (Test-Path $migrationPath) {
        Write-Host "  Applying $migration..." -ForegroundColor Cyan
        Get-Content $migrationPath | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Failed to apply migration: $migration" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "Migration file not found: $migration" -ForegroundColor Yellow
    }
}

# Step 4: Generate SQLC code
Write-Host "Generating SQLC code..." -ForegroundColor Yellow
sqlc generate
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to generate SQLC code" -ForegroundColor Red
    exit 1
}

# Step 5: Start the API server
Write-Host "Starting Basin API server..." -ForegroundColor Green
Write-Host "Server will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Health check: http://localhost:8080/health" -ForegroundColor Cyan
Write-Host "" 
Write-Host "Default admin credentials:" -ForegroundColor Yellow
Write-Host "   Email: admin@example.com" -ForegroundColor White
Write-Host "   Password: password" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Gray
Write-Host "================================" -ForegroundColor Gray

go run cmd/main.go