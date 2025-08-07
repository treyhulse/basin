# Go RBAC API - Directus-style API with Role-Based Access Control

A Go-based REST API that provides Directus-style functionality with comprehensive Role-Based Access Control (RBAC) including row-level and field-level security policies.

## 🚀 Features

- **🔑 JWT Authentication** - Secure token-based authentication
- **🔐 RBAC System** - Role-based access control with granular permissions
- **📦 Dynamic API** - Generic `GET /items/:table` endpoints with automatic filtering
- **🛡️ Field-Level Security** - Control which fields users can access
- **📊 Row-Level Security** - Filter data based on user permissions
- **🐘 PostgreSQL** - Robust database with UUID support
- **⚡ Type-Safe DB Access** - Generated with sqlc for compile-time safety
- **🐳 Docker Ready** - Easy development setup with Docker Compose

## 📋 **Prerequisites - Required Before Setup**

**⚠️ These must be installed and running BEFORE you start, or the setup will fail:**

### **Required Software:**

1. **🐳 Docker Desktop** - [Download here](https://www.docker.com/products/docker-desktop/)
   - **Windows:** Install Docker Desktop and make sure it's running (check system tray)
   - **macOS:** Install Docker Desktop and start it
   - **Linux:** Install Docker Engine and Docker Compose

2. **🐹 Go 1.21+** - [Download here](https://golang.org/dl/)
   - Verify installation: `go version`

3. **📦 Git** - [Download here](https://git-scm.com/)
   - Verify installation: `git --version`

### **Before Running Setup:**

- ✅ **Docker Desktop is running** (check system tray on Windows)
- ✅ **Go is installed** (`go version` works)
- ✅ **Git is installed** (`git --version` works)
- ✅ **Port 5432 is available** (stop any local PostgreSQL if running)

**If any of these are missing, the setup script will fail and tell you what's missing.**

### **Quick Verification (Optional)**
Run these commands to verify everything is ready:
```bash
# Check Go
go version

# Check Docker
docker --version
docker ps

# Check Git
git --version
```

## 🚀 **Getting Started - Super Simple Setup**

### **🎯 Two Commands to Get Everything Running**

**Want to get up and running in 2 minutes?**

#### **Step 1: Clone the Repository**
```bash
git clone https://github.com/treyhulse/directus-clone.git
cd directus-clone
```

#### **Step 2: Run the Setup Command**
```bash
make setup
```

**That's it!** Your API will automatically start at http://localhost:8080

#### **Step 3: Start the Application (after setup)**
```bash
make start
```

---

## 📋 **What the Setup Command Does**

The `make setup` command automatically handles everything:

1. ✅ **Checks prerequisites** (Go, Docker, Docker Compose, Git)
2. 🔍 **Sets up environment variables** (creates .env file)
3. 📦 **Installs Go dependencies** (`go mod tidy`)
4. 🔧 **Installs sqlc** (database code generator)
5. 🐘 **Starts PostgreSQL database** (with health checks)
6. 🗄️ **Applies all database migrations**
7. 🔨 **Generates database code** (`sqlc generate`)
8. 🏗️ **Builds the application**

**Perfect for:** New projects, demos, testing, learning - **zero additional steps needed!**

### **Daily Development Commands**
After initial setup, use these commands for daily development:

```bash
make start    # Start the application (cold start)
make dev      # Start development server  
make stop     # Stop the application
make restart  # Restart everything
```

---

### **Installing Make on Windows**

If you're on Windows and don't have Make installed, you can install it using Scoop:

```powershell
# Install Scoop first (if not already installed)
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod get.scoop.sh | Invoke-Expression

# Install Make
scoop install make
```

**Alternative package managers:**
- **Chocolatey:** `choco install make`
- **WSL:** `sudo apt-get install make`

---

### **Manual Setup (Advanced)**

**If you prefer to do everything manually:**

```bash
# 1. Clone the repository
git clone https://github.com/treyhulse/directus-clone.git
cd directus-clone

# 2. Set up environment variables
cp env.example .env
# Edit .env with your preferred settings

# 3. Start the database
make docker-up

# 4. Apply migrations
make migrate

# 5. Install Go dependencies
make deps

# 6. Generate database code
make generate

# 7. Build the application
make build

# 8. Run the application
make dev
```

**Or use individual Make commands:**
- `make help` - Show all available commands
- `make docker-up` - Start database
- `make migrate` - Apply migrations  
- `make deps` - Install dependencies
- `make generate` - Generate database code
- `make build` - Build application
- `make dev` - Start development server

---

## 🎯 **After Setup - Start Your API**

**Once setup is complete, start your API:**

```bash
# Recommended: Use Make commands
make start    # Cold start (stops containers, starts fresh)
make dev      # Development server (just runs the app)

# Alternative: Run directly  
go run cmd/main.go

# Alternative: Run the built binary
./bin/api     # Linux/macOS
.\bin\api.exe # Windows
```

**Your API will be available at:** `http://localhost:8080`

**Default admin credentials:**
- **Email:** `admin@example.com`
- **Password:** `password`

**API Keys for testing:**
- **Admin:** `admin_api_key_123`
- **Manager:** `manager_api_key_456`

---

## 🐳 **Docker Management - Simple Start/Stop**

**Need to quickly start/stop your database?**

### **Using Make Commands (Recommended):**
```bash
make stop         # Stop everything
make docker-up    # Start database only
make docker-down  # Stop database only
make docker-logs  # View database logs
make restart      # Stop and restart everything
```

### **Direct Docker Commands:**
```bash
# Stop the database (keeps data intact)
docker-compose down

# Start the database back up
docker-compose up -d

# Restart in one command
docker-compose restart

# See what's running
docker-compose ps
# or
docker ps

# Complete reset (⚠️ DELETES ALL DATA!)
docker-compose down -v
# Then start fresh
docker-compose up -d
```

**💡 Pro Tip:** The database data persists between stops/starts, so you won't lose your data when using `make stop` and `make start`.

---

## 🔧 **Environment Configuration**

The setup scripts will create a `.env` file with these settings:

```bash
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
```

**You can customize these values in your `.env` file before running the setup.**

---

## 🔧 **Troubleshooting Common Issues**

### **Setup Issues**

#### **"Docker is not installed"**
- **Solution:** Install Docker Desktop from https://www.docker.com/products/docker-desktop/
- **Windows:** Make sure Docker Desktop is running (check system tray)
- **Verify:** Run `docker --version` in terminal

#### **"Docker Desktop is not running"**
- **Solution:** 
  1. Start Docker Desktop application
  2. Wait for it to fully load (check system tray icon)
  3. Verify with `docker ps` command

#### **"Go is not installed"**
- **Solution:** Install Go 1.21+ from https://golang.org/dl/
- **Verify:** Run `go version` in terminal

#### **"Git is not installed"**
- **Solution:** Install Git from https://git-scm.com/
- **Verify:** Run `git --version` in terminal

#### **"docker-compose is not recognized"**
- **Solution:** 
  1. Make sure Docker Desktop is installed and running
  2. Try using `docker compose up -d` (with a space instead of hyphen)
  3. Restart your terminal after installing Docker Desktop

#### **Database connection errors**
- **Solution:** 
  1. Make sure PostgreSQL is running: `docker-compose ps`
  2. Check if port 5432 is available (stop any local PostgreSQL)
  3. Restart Docker Desktop if needed

#### **"sqlc is not recognized"**
- **Solution:** The setup script will install it automatically, or run: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### **Runtime Issues**

#### **"403 Insufficient permissions"**
- **Solution:** Make sure you're using the correct admin credentials:
  - Email: `admin@example.com`
  - Password: `password`

#### **"Connection refused" on localhost:8080**
- **Solution:** 
  1. Make sure the API is running: `go run cmd/main.go`
  2. Check if port 8080 is available
  3. Try a different port in your `.env` file

#### **Docker API errors (500 Internal Server Error)**
- **Solution:** 
  1. Restart Docker Desktop completely
  2. Run `docker system prune -a` to clean up
  3. Reset Docker Desktop to factory defaults if needed

---

### **Quick Commands Reference**

```bash
# First time setup
make setup

# Daily development
make start     # Cold start everything
make dev       # Just run the app (database should be running)
make stop      # Stop everything
make restart   # Restart everything

# Database management
make migrate       # Apply new migrations
make docker-up     # Start database only
make docker-down   # Stop database only
make docker-logs   # View database logs

# Development tasks
make build     # Build the application
make test      # Run tests
make clean     # Clean build artifacts
make deps      # Update dependencies
make generate  # Regenerate database code
```

The API will be available at `http://localhost:8080`

## 📚 API Documentation

### OpenAPI

- Swagger UI: `http://localhost:8080/swagger/index.html`
- Swagger JSON: `http://localhost:8080/swagger/doc.json`
- Generate a typed client for frontend using `openapi-typescript` or `orval`

### Authentication

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "password"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "admin@example.com",
    "first_name": "Admin",
    "last_name": "User",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### API Key Authentication
You can also authenticate using API keys for programmatic access:

```http
GET /items/products
Authorization: Bearer admin_api_key_123
```

**Available API Keys:**
- `admin_api_key_123` - Full admin access
- `manager_api_key_456` - Manager-level access

#### Get Current User
```http
GET /auth/me
Authorization: Bearer <token>
```

### Items API

All items endpoints require authentication. Include the JWT token in the Authorization header.

#### List Items
```http
GET /items/:table?limit=50&offset=0&sort=created_at&order=desc
Authorization: Bearer <token>
```

#### Get Single Item
```http
GET /items/:table/:id
Authorization: Bearer <token>
```

#### Create Item
```http
POST /items/:table
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "New Product",
  "description": "Product description",
  "price": 99.99
}
```

#### Update Item
```http
PUT /items/:table/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Product",
  "price": 89.99
}
```

#### Delete Item
```http
DELETE /items/:table/:id
Authorization: Bearer <token>
```

## 🗄️ Database Schema

### Core Tables

- **users** - User accounts with authentication
- **roles** - Available roles in the system
- **user_roles** - Many-to-many relationship between users and roles
- **permissions** - RBAC permissions with field-level access control

### Sample Tables

- **products** - Product catalog
- **customers** - Customer information
- **orders** - Order records
- **order_items** - Order line items

## 🔐 RBAC System

### Roles

1. **admin** - Full system access
2. **manager** - Can manage products and view orders
3. **sales** - Can view products and create orders
4. **customer** - Can view products and own orders

### Permission Structure

Each permission defines:
- **role_id** - Which role this applies to
- **table_name** - Which table this applies to
- **action** - What action is allowed (create, read, update, delete)
- **field_filter** - Row-level filtering (JSONB)
- **allowed_fields** - Field-level access control (array of field names)

### Example Permissions

```sql
-- Admin can do everything on products
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
VALUES (admin_role_id, 'products', 'read', ARRAY['*']);

-- Sales can only see certain product fields
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
VALUES (sales_role_id, 'products', 'read', ARRAY['id', 'name', 'description', 'price', 'category']);
```

## 🧪 Testing

### Automated Testing

The project includes comprehensive integration tests that test the real API against a running server and database:

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run with coverage report
make test-coverage

# Run integration tests specifically
make test-integration

# Run integration tests with automatic database setup
make test-integration-full
```

**Integration tests cover:**
- ✅ Real authentication flows (login, JWT validation)
- ✅ All API endpoints with actual database operations
- ✅ Role-based access control with real permissions
- ✅ Error handling and edge cases
- ✅ End-to-end functionality

### Manual API Testing

### 1. Login as Admin

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'
```

### 2. Get Products (with token from login)

```bash
curl -X GET http://localhost:8080/items/products \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 3. Create a Product

```bash
curl -X POST http://localhost:8080/items/products \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "A test product",
    "price": 29.99,
    "category": "Electronics",
    "stock_quantity": 100
  }'
```

## 🛠️ Development

### Project Structure

```
go-rbac-api/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── auth.go          # Authentication handlers
│   │   └── items.go         # Dynamic items API
│   ├── config/
│   │   └── env.go           # Configuration management
│   ├── db/
│   │   ├── postgres.go      # Database connection
│   │   ├── query.sql        # SQL queries for sqlc
│   │   └── sqlc/            # Generated database code
│   ├── middleware/
│   │   └── auth.go          # JWT authentication middleware
│   ├── models/
│   │   └── user.go          # User model and auth helpers
│   └── rbac/
│       └── policies.go      # RBAC policy checker
├── migrations/
│   └── 001_init.sql         # Database schema and seed data
├── docker-compose.yml       # PostgreSQL setup
├── sqlc.yaml               # sqlc configuration
├── go.mod                  # Go module file
└── Makefile                # Development commands
```

### Available Commands

#### Make Commands (Recommended)
- `make help` - Show all available commands
- `make setup` - Complete initial setup (first time only)
- `make start` - Cold start the application
- `make dev` - Start development server
- `make stop` - Stop the application
- `make restart` - Stop and restart everything
- `make build` - Build the application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make deps` - Download dependencies
- `make generate` - Generate database code
- `make migrate` - Apply database migrations
- `make docker-up` - Start PostgreSQL
- `make docker-down` - Stop PostgreSQL
- `make docker-logs` - Show Docker logs

#### Direct Commands (Alternative)
- `go mod tidy` - Download dependencies
- `go run cmd/main.go` - Start development server
- `go build -o bin/api cmd/main.go` - Build the application
- `./bin/api` - Run the built application
- `go test ./...` - Run tests
- `sqlc generate` - Generate database code
- `docker-compose up -d` - Start PostgreSQL
- `docker-compose down` - Stop PostgreSQL

### Environment Variables

Copy `env.example` to `.env` and configure:

```bash
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
```

## 🔧 Configuration

### Database

The application uses PostgreSQL with the following default settings:
- Host: localhost
- Port: 5432
- Database: go_rbac_db
- User: postgres
- Password: postgres

### JWT

- Secret: Configured via `JWT_SECRET` environment variable
- Expiry: Configurable via `JWT_EXPIRY` (default: 24h)

## 🚀 Deployment

### Build for Production

```bash
make build
```

### Docker Deployment

```bash
# Build the application
docker build -t go-rbac-api .

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e JWT_SECRET=your-jwt-secret \
  go-rbac-api
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the API documentation at `http://localhost:8080/` when running

## 🔧 Troubleshooting

### Common Issues

#### "make is not recognized"
- **Solution**: Use the direct commands in the Quick Start section above, or install make using Chocolatey (`choco install make`) or Scoop (`scoop install make`)

#### "docker-compose is not recognized"
- **Solution**: 
  1. Make sure Docker Desktop is installed and running
  2. Try using `docker compose up -d` (with a space instead of hyphen)
  3. Restart your terminal after installing Docker Desktop

#### Docker API errors (500 Internal Server Error)
- **Solution**: 
  1. Restart Docker Desktop completely
  2. Make sure Docker Desktop is fully started (check system tray icon)
  3. Try running `docker system prune -a` to clean up Docker cache
  4. If the issue persists, try switching Docker Desktop to Windows containers and back to Linux containers
  5. As a last resort, reset Docker Desktop to factory defaults

#### "sqlc is not recognized"
- **Solution**: Install sqlc using `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

#### Database connection errors
- **Solution**: Make sure PostgreSQL is running with `docker-compose ps` or `docker compose ps`

## 🔄 Roadmap

- [x] Add comprehensive integration tests
- [ ] Implement real-time subscriptions
- [ ] Add GraphQL support
- [ ] Implement file upload functionality
- [ ] Add audit logging
- [ ] Create admin dashboard
- [ ] Add rate limiting
- [ ] Implement caching layer 