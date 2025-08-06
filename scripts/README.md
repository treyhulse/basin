# Basin API Scripts

This directory contains utility scripts for the Basin API project.

## Using the Makefile

The project uses a Makefile for all build and deployment tasks. All commands should be run from the project root.

```bash
# Show all available commands
make help

# Complete initial setup
make setup

# Cold start the application
make start

# Restart the application
make restart
```

## Installing Make on Windows

Make is already installed via Scoop. If you need to reinstall or install on another system:

### Option 1: Scoop (Recommended)
```powershell
# Install Scoop first
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod get.scoop.sh | Invoke-Expression

# Then install Make
scoop install make
```

### Option 2: Chocolatey
```powershell
choco install make
```

### Option 3: Windows Subsystem for Linux (WSL)
```bash
sudo apt-get install make
```

## Command Descriptions

### Setup Commands

- **`setup`** - Complete initial setup including:
  - Prerequisites check (Go, Docker)
  - Environment configuration
  - Dependency installation
  - sqlc installation
  - Database startup
  - Migration application
  - Code generation
  - Application build

- **`start`** - Cold start the application:
  - Stops any existing containers
  - Starts database
  - Applies migrations
  - Generates database code
  - Builds application
  - Starts the server

- **`restart`** - Stop and restart everything
- **`stop`** - Stop all containers

### Development Commands

- **`dev`** - Start development server with hot reload
- **`build`** - Build the application
- **`test`** - Run tests
- **`clean`** - Clean build artifacts
- **`deps`** - Update dependencies
- **`generate`** - Generate sqlc database code

### Docker Commands

- **`docker-up`** - Start PostgreSQL database
- **`docker-down`** - Stop PostgreSQL database
- **`docker-logs`** - Show Docker container logs

## Quick Start

1. **First time setup:**
   ```bash
   make setup
   ```

2. **Start the application:**
   ```bash
   make start
   ```

3. **For development:**
   ```bash
   make dev
   ```

## Utility Scripts

- `hash_password.go` - Utility for hashing passwords

## Troubleshooting

### Docker Issues
Make sure Docker Desktop is running before using any commands that start containers.

### Database Connection Issues
If the database fails to start, try:
```bash
make docker-down
make docker-up
``` 