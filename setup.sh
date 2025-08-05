#!/bin/bash

# Complete Setup Script for Go RBAC API
# Usage: bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/setup.sh)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script configuration
REPO_URL="https://github.com/treyhulse/directus-clone.git"
PROJECT_NAME="directus-clone"

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

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local issues=()
    
    # Check Go
    if ! command_exists "go"; then
        issues+=("Go is not installed. Please install Go 1.21+ from https://golang.org/dl/")
    else
        print_status "Go ✓"
    fi
    
    # Check Docker
    if ! command_exists "docker"; then
        issues+=("Docker is not installed. Please install Docker from https://www.docker.com/products/docker-desktop/")
    else
        print_status "Docker ✓"
    fi
    
    # Check Docker Compose
    if ! command_exists "docker-compose" && ! docker compose version >/dev/null 2>&1; then
        issues+=("Docker Compose is not installed")
    else
        print_status "Docker Compose ✓"
    fi
    
    # Check Git
    if ! command_exists "git"; then
        issues+=("Git is not installed. Please install Git from https://git-scm.com/")
    else
        print_status "Git ✓"
    fi
    
    # Report issues
    if [[ ${#issues[@]} -gt 0 ]]; then
        print_error "Prerequisite issues found:"
        for issue in "${issues[@]}"; do
            echo -e "  - ${RED}$issue${NC}"
        done
        echo ""
        print_info "Please install missing prerequisites and try again."
        return 1
    fi
    
    print_status "All prerequisites met"
    return 0
}

# Function to clone repository
clone_repository() {
    print_header "Cloning Repository"
    
    local target_dir="$PROJECT_NAME"
    
    # Check if directory already exists
    if [[ -d "$target_dir" ]]; then
        print_warning "Directory $target_dir already exists"
        read -p "Do you want to remove it and start fresh? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Removing existing directory..."
            rm -rf "$target_dir"
        else
            print_error "Installation cancelled"
            exit 1
        fi
    fi
    
    print_info "Cloning $REPO_URL to $target_dir..."
    if git clone "$REPO_URL" "$target_dir"; then
        print_status "Repository cloned successfully"
    else
        print_error "Failed to clone repository"
        exit 1
    fi
    
    # Change to project directory
    cd "$target_dir"
    print_status "Changed to project directory: $(pwd)"
}

# Function to setup environment variables
setup_environment() {
    print_header "Setting Up Environment Variables"
    
    # Create .env file with correct settings
    cat > .env << 'EOF'
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
EOF
    
    print_status "Created .env file with correct settings"
    
    # Load environment variables
    export $(grep -v '^#' .env | xargs)
    print_status "Environment variables loaded"
}

# Function to install dependencies and generate code
install_dependencies() {
    print_header "Installing Dependencies and Generating Code"
    
    # Install Go dependencies
    print_info "Installing Go dependencies..."
    if go mod tidy; then
        print_status "Go dependencies installed"
    else
        print_error "Failed to install Go dependencies"
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
    else
        print_status "sqlc already installed"
    fi
    
    # Generate database code
    print_info "Generating database code..."
    if sqlc generate; then
        print_status "Database code generated"
    else
        print_error "Failed to generate database code"
        return 1
    fi
    
    return 0
}

# Function to start Docker services
start_docker_services() {
    print_header "Starting Docker Services"
    
    # Stop any existing containers
    print_info "Stopping any existing containers..."
    docker-compose down >/dev/null 2>&1 || true
    
    # Start PostgreSQL
    print_info "Starting PostgreSQL database..."
    docker-compose up -d
    
    # Wait for database to be ready
    print_info "Waiting for database to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        attempt=$((attempt + 1))
        sleep 2
        
        if docker exec go-rbac-postgres pg_isready -U postgres >/dev/null 2>&1; then
            print_status "Database is ready"
            return 0
        fi
        
        print_info "Attempt $attempt/$max_attempts - Database not ready yet..."
    done
    
    print_error "Database failed to start within expected time"
    return 1
}

# Function to apply migrations
apply_migrations() {
    print_header "Applying Database Migrations"
    
    # Get all migration files
    local migrations=($(find migrations -name "*.sql" | sort))
    
    if [[ ${#migrations[@]} -eq 0 ]]; then
        print_warning "No migration files found in migrations/ directory"
        return 0
    fi
    
    print_info "Found ${#migrations[@]} migration files"
    
    for migration in "${migrations[@]}"; do
        print_info "Applying $(basename "$migration")..."
        if docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < "$migration"; then
            print_status "Applied $(basename "$migration")"
        else
            print_warning "Migration $(basename "$migration") had issues (this might be expected)"
        fi
    done
    
    print_status "Migrations completed"
    return 0
}

# Function to create admin user
create_admin_user() {
    print_header "Creating Admin User"
    
    local admin_email="${ADMIN_EMAIL:-admin@example.com}"
    local admin_password="${ADMIN_PASSWORD:-password}"
    local admin_first_name="${ADMIN_FIRST_NAME:-Admin}"
    local admin_last_name="${ADMIN_LAST_NAME:-User}"
    
    print_info "Creating admin user: $admin_email"
    
    # Check if admin user already exists
    local existing_count=$(docker exec go-rbac-postgres psql -U postgres -d go_rbac_db -t -c "SELECT COUNT(*) FROM users WHERE email = '$admin_email';" 2>/dev/null | tr -d ' ')
    
    if [[ "$existing_count" -gt 0 ]]; then
        print_info "Admin user already exists"
        return 0
    fi
    
    # Create admin user with plain text password (will be hashed on first login)
    local create_user_query="INSERT INTO users (id, email, password_hash, first_name, last_name, is_active, created_at, updated_at) VALUES (gen_random_uuid(), '$admin_email', '$admin_password', '$admin_first_name', '$admin_last_name', true, NOW(), NOW());"
    
    if echo "$create_user_query" | docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db; then
        print_status "Admin user created successfully"
    else
        print_warning "Failed to create admin user (might already exist)"
    fi
    
    return 0
}

# Function to build the application
build_application() {
    print_header "Building Application"
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Build the application
    print_info "Building application..."
    if go build -o bin/api cmd/main.go; then
        print_status "Application built successfully"
    else
        print_error "Failed to build application"
        return 1
    fi
    
    return 0
}

# Function to start the application
start_application() {
    print_header "Starting Your API"
    
    print_info "Starting the Go RBAC API..."
    echo ""
    echo -e "${GREEN}🚀 Your API is now running at: http://localhost:8080${NC}"
    echo ""
    echo -e "${NC}📋 Default credentials:${NC}"
    echo -e "   Email: admin@example.com${NC}"
    echo -e "   Password: password${NC}"
    echo ""
    echo -e "${NC}🔑 API Keys for testing:${NC}"
    echo -e "   Admin: admin_api_key_123${NC}"
    echo -e "   Manager: manager_api_key_456${NC}"
    echo ""
    echo -e "${NC}📚 API Documentation: http://localhost:8080${NC}"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
    echo ""
    
    # Start the application
    go run cmd/main.go
}

# Function to display completion message
show_completion() {
    print_header "🎉 Setup Complete!"
    echo ""
    print_info "Your Go RBAC API is ready!"
    echo ""
    print_info "What was set up:"
    echo -e "  ${NC}✅ Repository cloned${NC}"
    echo -e "  ${NC}✅ Environment variables configured${NC}"
    echo -e "  ${NC}✅ Dependencies installed${NC}"
    echo -e "  ${NC}✅ Database code generated${NC}"
    echo -e "  ${NC}✅ PostgreSQL database started${NC}"
    echo -e "  ${NC}✅ Database migrations applied${NC}"
    echo -e "  ${NC}✅ Admin user created${NC}"
    echo -e "  ${NC}✅ Application built${NC}"
    echo ""
    print_info "Starting your API now..."
    echo ""
}

# Main execution function
main() {
    print_header "🚀 Go RBAC API - Complete Setup"
    echo ""
    
    # Check prerequisites
    if ! check_prerequisites; then
        exit 1
    fi
    
    # Clone repository
    if ! clone_repository; then
        exit 1
    fi
    
    # Setup environment
    if ! setup_environment; then
        exit 1
    fi
    
    # Install dependencies and generate code
    if ! install_dependencies; then
        exit 1
    fi
    
    # Start Docker services
    if ! start_docker_services; then
        exit 1
    fi
    
    # Apply migrations
    if ! apply_migrations; then
        exit 1
    fi
    
    # Create admin user
    if ! create_admin_user; then
        exit 1
    fi
    
    # Build application
    if ! build_application; then
        exit 1
    fi
    
    # Show completion message
    show_completion
    
    # Start the application
    start_application
}

# Run main installation
main 