#!/bin/bash

# Dynamic Setup Script for Go RBAC API
# This script is designed to be future-proof and handle various scenarios

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_VERSION="2.0.0"
REPO_URL="https://github.com/treyhulse/directus-clone.git"
REPO_BRANCH="main"
MIN_GO_VERSION="1.21"
MIN_DOCKER_VERSION="20.0"
MIN_DOCKER_COMPOSE_VERSION="2.0"
REQUIRED_ENV_VARS=("DB_HOST" "DB_PORT" "DB_USER" "DB_PASSWORD" "DB_NAME" "JWT_SECRET" "SERVER_PORT")

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_header() {
    echo -e "${CYAN}$1${NC}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to compare version numbers
version_compare() {
    if [[ $1 == $2 ]]; then
        return 0
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 1
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 2
        fi
    done
    return 0
}

# Function to get version of a command
get_version() {
    local cmd=$1
    local version_flag=$2
    
    if command_exists "$cmd"; then
        local version_output
        version_output=$($cmd $version_flag 2>&1 | head -n 1)
        echo "$version_output" | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n 1
    else
        echo ""
    fi
}

# Function to validate environment variables
validate_env_vars() {
    print_info "Validating environment variables..."
    
    local missing_vars=()
    local invalid_vars=()
    
    # Check if .env file exists
    if [[ ! -f ".env" ]]; then
        print_warning ".env file not found. Creating from template..."
        if [[ -f "env.example" ]]; then
            cp env.example .env
            print_status "Created .env from env.example"
        else
            print_error "No .env file and no env.example template found!"
            return 1
        fi
    fi
    
    # Load environment variables
    if [[ -f ".env" ]]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Check required variables
    for var in "${REQUIRED_ENV_VARS[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    # Validate specific variables
    if [[ -n "$DB_PORT" ]] && ! [[ "$DB_PORT" =~ ^[0-9]+$ ]]; then
        invalid_vars+=("DB_PORT must be a number")
    fi
    
    if [[ -n "$SERVER_PORT" ]] && ! [[ "$SERVER_PORT" =~ ^[0-9]+$ ]]; then
        invalid_vars+=("SERVER_PORT must be a number")
    fi
    
    if [[ -n "$JWT_SECRET" ]] && [[ ${#JWT_SECRET} -lt 32 ]]; then
        print_warning "JWT_SECRET is shorter than 32 characters (security risk)"
    fi
    
    # Report issues
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    if [[ ${#invalid_vars[@]} -gt 0 ]]; then
        print_error "Invalid environment variables:"
        for var in "${invalid_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    print_status "Environment variables validated"
    return 0
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local issues=()
    
    # Check Go
    if ! command_exists "go"; then
        issues+=("Go is not installed")
    else
        local go_version
        go_version=$(get_version "go" "version")
        if [[ -n "$go_version" ]]; then
            version_compare "$go_version" "$MIN_GO_VERSION"
            if [[ $? -eq 2 ]]; then
                issues+=("Go version $go_version is older than required $MIN_GO_VERSION")
            else
                print_status "Go $go_version ✓"
            fi
        else
            issues+=("Could not determine Go version")
        fi
    fi
    
    # Check Docker
    if ! command_exists "docker"; then
        issues+=("Docker is not installed")
    else
        local docker_version
        docker_version=$(get_version "docker" "version")
        if [[ -n "$docker_version" ]]; then
            version_compare "$docker_version" "$MIN_DOCKER_VERSION"
            if [[ $? -eq 2 ]]; then
                issues+=("Docker version $docker_version is older than required $MIN_DOCKER_VERSION")
            else
                print_status "Docker $docker_version ✓"
            fi
        else
            issues+=("Could not determine Docker version")
        fi
    fi
    
    # Check Docker Compose
    if ! command_exists "docker-compose" && ! docker compose version >/dev/null 2>&1; then
        issues+=("Docker Compose is not installed")
    else
        local compose_version
        if command_exists "docker-compose"; then
            compose_version=$(get_version "docker-compose" "version")
        else
            compose_version=$(get_version "docker" "compose version")
        fi
        if [[ -n "$compose_version" ]]; then
            version_compare "$compose_version" "$MIN_DOCKER_COMPOSE_VERSION"
            if [[ $? -eq 2 ]]; then
                issues+=("Docker Compose version $compose_version is older than required $MIN_DOCKER_COMPOSE_VERSION")
            else
                print_status "Docker Compose $compose_version ✓"
            fi
        else
            issues+=("Could not determine Docker Compose version")
        fi
    fi
    
    # Check sqlc
    if ! command_exists "sqlc"; then
        print_warning "sqlc not found, will install during setup"
    else
        local sqlc_version
        sqlc_version=$(get_version "sqlc" "version")
        if [[ -n "$sqlc_version" ]]; then
            print_status "sqlc $sqlc_version ✓"
        else
            print_warning "Could not determine sqlc version"
        fi
    fi
    
    # Report issues
    if [[ ${#issues[@]} -gt 0 ]]; then
        print_error "Prerequisite issues found:"
        for issue in "${issues[@]}"; do
            echo "  - $issue"
        done
        echo ""
        print_info "Please install missing prerequisites and try again."
        return 1
    fi
    
    print_status "All prerequisites met"
    return 0
}

# Function to clone or update repository
setup_repository() {
    print_header "Setting up Repository"
    
    if [[ ! -d ".git" ]]; then
        print_info "Not a git repository. Cloning from $REPO_URL..."
        if git clone -b "$REPO_BRANCH" "$REPO_URL" .; then
            print_status "Repository cloned successfully"
        else
            print_error "Failed to clone repository"
            return 1
        fi
    else
        print_info "Git repository found. Checking for updates..."
        git fetch origin
        local current_branch
        current_branch=$(git branch --show-current)
        if [[ "$current_branch" != "$REPO_BRANCH" ]]; then
            print_warning "Current branch is $current_branch, switching to $REPO_BRANCH"
            git checkout "$REPO_BRANCH"
        fi
        git pull origin "$REPO_BRANCH"
        print_status "Repository updated"
    fi
}

# Function to find all migration files
find_migrations() {
    local migrations=()
    if [[ -d "migrations" ]]; then
        while IFS= read -r -d '' file; do
            migrations+=("$file")
        done < <(find migrations -name "*.sql" -type f -print0 | sort -z)
    fi
    echo "${migrations[@]}"
}

# Function to apply migrations
apply_migrations() {
    print_header "Applying Database Migrations"
    
    local migrations
    readarray -t migrations < <(find_migrations)
    
    if [[ ${#migrations[@]} -eq 0 ]]; then
        print_warning "No migration files found"
        return 0
    fi
    
    print_info "Found ${#migrations[@]} migration files"
    
    for migration in "${migrations[@]}"; do
        print_info "Applying $migration..."
        if docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < "$migration"; then
            print_status "Applied $migration"
        else
            print_warning "Migration $migration had issues (this might be expected for duplicate entries)"
        fi
    done
}

# Function to check database health
check_database_health() {
    print_header "Checking Database Health"
    
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        attempt=$((attempt + 1))
        print_info "Checking database connection (attempt $attempt/$max_attempts)..."
        
        if docker exec go-rbac-postgres pg_isready -U postgres >/dev/null 2>&1; then
            print_status "Database is ready!"
            return 0
        fi
        
        if [[ $attempt -eq $max_attempts ]]; then
            print_error "Database failed to start after $max_attempts attempts"
            return 1
        fi
        
        sleep 2
    done
}

# Function to build application
build_application() {
    print_header "Building Application"
    
    # Install Go dependencies
    print_info "Installing Go dependencies..."
    if go mod tidy; then
        print_status "Dependencies installed"
    else
        print_error "Failed to install dependencies"
        return 1
    fi
    
    # Install sqlc if not present
    if ! command_exists "sqlc"; then
        print_info "Installing sqlc..."
        if go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; then
            print_status "sqlc installed"
        else
            print_error "Failed to install sqlc"
            return 1
        fi
    fi
    
    # Generate database code
    print_info "Generating database code..."
    if sqlc generate; then
        print_status "Database code generated"
    else
        print_error "Failed to generate database code"
        return 1
    fi
    
    # Build the application
    print_info "Building application..."
    if go build -o bin/api cmd/main.go; then
        print_status "Application built successfully"
    else
        print_error "Build failed"
        return 1
    fi
}

# Function to display summary
display_summary() {
    print_header "Setup Complete!"
    echo "=================="
    echo ""
    
    print_info "Database Status:"
    docker-compose ps
    echo ""
    
    print_info "Default Admin Credentials:"
    echo "   Email: admin@example.com"
    echo "   Password: password"
    echo ""
    
    print_info "API Keys (for testing):"
    echo "   Admin API Key: admin_api_key_123"
    echo "   Manager API Key: manager_api_key_456"
    echo ""
    
    print_info "To start the API server:"
    echo "   go run cmd/main.go"
    echo "   or"
    echo "   ./bin/api"
    echo ""
    
    print_info "API will be available at:"
    echo "   http://localhost:${SERVER_PORT:-8080}"
    echo ""
    
    print_info "Available endpoints:"
    echo "   POST /auth/login - Login with email/password"
    echo "   GET  /auth/me - Get current user info"
    echo "   GET  /items/:table - Generic data access (products, customers, orders, etc.)"
    echo "   GET  /health - Health check"
    echo ""
    
    print_info "Useful commands:"
    echo "   docker-compose down - Stop database"
    echo "   docker-compose up -d - Start database"
    echo "   docker exec -it go-rbac-postgres psql -U postgres -d go_rbac_db - Connect to database"
    echo ""
    
    print_status "You're all set! Run 'go run cmd/main.go' to start the API server."
}

# Main execution
main() {
    print_header "🚀 Dynamic Setup Script for Go RBAC API v$SCRIPT_VERSION"
    echo ""
    
    # Check prerequisites
    if ! check_prerequisites; then
        exit 1
    fi
    
    # Setup repository
    if ! setup_repository; then
        exit 1
    fi
    
    # Validate environment
    if ! validate_env_vars; then
        exit 1
    fi
    
    # Stop and remove any existing containers
    print_header "Cleaning up existing containers"
    docker-compose down -v 2>/dev/null || true
    docker system prune -f 2>/dev/null || true
    
    # Start fresh database
    print_header "Starting PostgreSQL database"
    if docker-compose up -d; then
        print_status "Database started"
    else
        print_error "Failed to start database"
        exit 1
    fi
    
    # Wait for database to be ready
    print_info "Waiting for database to be ready..."
    sleep 10
    
    # Check database health
    if ! check_database_health; then
        exit 1
    fi
    
    # Apply migrations
    if ! apply_migrations; then
        print_warning "Some migrations had issues, but continuing..."
    fi
    
    # Build application
    if ! build_application; then
        exit 1
    fi
    
    # Display summary
    display_summary
}

# Handle script arguments
case "${1:-}" in
    --version)
        echo "Setup Script v$SCRIPT_VERSION"
        exit 0
        ;;
    --help)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --version    Show version information"
        echo "  --help       Show this help message"
        echo ""
        echo "This script will:"
        echo "  - Check all prerequisites"
        echo "  - Clone/update the repository"
        echo "  - Validate environment variables"
        echo "  - Start a fresh PostgreSQL database"
        echo "  - Apply all migrations dynamically"
        echo "  - Install Go dependencies"
        echo "  - Generate database code"
        echo "  - Build the application"
        exit 0
        ;;
    "")
        # No arguments, run main setup
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac 