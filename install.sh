#!/bin/bash

# Remote Installation Script for Go RBAC API
# Usage: bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.sh)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_VERSION="1.0.0"
REPO_URL="https://github.com/treyhulse/directus-clone.git"
REPO_BRANCH="main"
PROJECT_NAME="directus-clone"
MIN_GO_VERSION="1.21"
MIN_DOCKER_VERSION="20.0"
MIN_DOCKER_COMPOSE_VERSION="2.0"

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

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    local issues=()
    
    # Check Go
    if ! command_exists "go"; then
        issues+=("Go is not installed. Please install Go 1.21+ from https://golang.org/dl/")
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
        issues+=("Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/")
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
    if git clone -b "$REPO_BRANCH" "$REPO_URL" "$target_dir"; then
        print_status "Repository cloned successfully"
    else
        print_error "Failed to clone repository"
        exit 1
    fi
    
    # Change to project directory
    cd "$target_dir"
    print_status "Changed to project directory: $(pwd)"
}

# Function to run the setup script
run_setup() {
    print_header "Running Setup"
    
    if [[ -f "setup.sh" ]]; then
        print_info "Found setup.sh, running it..."
        chmod +x setup.sh
        ./setup.sh
    else
        print_warning "setup.sh not found, running manual setup..."
        
        # Install sqlc if not present
        if ! command_exists "sqlc"; then
            print_info "Installing sqlc..."
            go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
        fi
        
        # Start database
        print_info "Starting PostgreSQL database..."
        docker-compose up -d
        
        # Wait for database
        print_info "Waiting for database to be ready..."
        sleep 15
        
        # Apply migrations
        print_info "Applying migrations..."
        for migration in migrations/*.sql; do
            if [[ -f "$migration" ]]; then
                print_info "Applying $migration..."
                docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < "$migration" || true
            fi
        done
        
        # Install dependencies and build
        print_info "Installing Go dependencies..."
        go mod tidy
        
        print_info "Generating database code..."
        sqlc generate
        
        print_info "Building application..."
        go build -o bin/api cmd/main.go
        
        print_status "Setup completed"
    fi
}

# Function to display completion message
show_completion() {
    print_header "🎉 Installation Complete!"
    echo ""
    print_info "Your Go RBAC API is ready!"
    echo ""
    print_info "Next steps:"
    echo "  1. cd $PROJECT_NAME"
    echo "  2. go run cmd/main.go"
    echo ""
    print_info "The API will be available at:"
    echo "  http://localhost:8080"
    echo ""
    print_info "Default credentials:"
    echo "  Email: admin@example.com"
    echo "  Password: password"
    echo ""
    print_info "API Keys (for testing):"
    echo "  Admin: admin_api_key_123"
    echo "  Manager: manager_api_key_456"
    echo ""
    print_status "Happy coding! 🚀"
}

# Main execution
main() {
    print_header "🚀 Go RBAC API Installer v$SCRIPT_VERSION"
    echo ""
    
    # Check prerequisites
    if ! check_prerequisites; then
        exit 1
    fi
    
    # Clone repository
    if ! clone_repository; then
        exit 1
    fi
    
    # Run setup
    if ! run_setup; then
        exit 1
    fi
    
    # Show completion message
    show_completion
}

# Handle script arguments
case "${1:-}" in
    --version)
        echo "Installer v$SCRIPT_VERSION"
        exit 0
        ;;
    --help)
        echo "Usage: bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.sh)"
        echo ""
        echo "This script will:"
        echo "  - Check all prerequisites"
        echo "  - Clone the repository"
        echo "  - Run the setup script"
        echo "  - Build the application"
        echo ""
        echo "Options:"
        echo "  --version    Show version information"
        echo "  --help       Show this help message"
        exit 0
        ;;
    "")
        # No arguments, run main installation
        main
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac 