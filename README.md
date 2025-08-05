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

## 📋 Prerequisites

- Go 1.21+
- Docker & Docker Compose
- sqlc (optional, for code generation)

## 🚀 **Getting Started - Choose Your Path**

### **🎯 Quick Start (Recommended)**

**Want to get up and running in 2 minutes?**

1. **Open a terminal/command prompt**
2. **Run one command:**
   - **Windows:** `powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install-simple.ps1 -UseBasicParsing | iex }"`
   - **Mac/Linux:** `bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.sh)`
3. **Wait for setup to complete**
4. **Run:** `go run cmd/main.go`
5. **Visit:** `http://localhost:8080`

**That's it!** Your API is ready with admin user, database, and all permissions set up.

---

## 📋 **Detailed Setup Options**

### **Option 1: One-Command Installation (Easiest)**

**Start from any empty directory with a single command:**

#### **Windows (PowerShell):**
```powershell
# Simple version (recommended)
powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install-simple.ps1 -UseBasicParsing | iex }"
```

#### **Unix/Linux/macOS:**
```bash
bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.sh)
```

**What this does automatically:**
1. ✅ **Checks prerequisites** (Go 1.21+, Docker, Docker Compose)
2. 🔄 **Clones the repository** (creates `directus-clone` directory)
3. 🔍 **Sets up environment variables** (creates .env file)
4. 🐘 **Starts PostgreSQL database** (with health checks)
5. 🗄️ **Applies all database migrations**
6. 👤 **Creates admin user** (from environment variables)
7. 📦 **Installs Go dependencies**
8. 🔧 **Generates database code**
9. 🔨 **Builds the application**
10. 📋 **Shows you how to start the API**

**Perfect for:** New projects, demos, testing, learning

---

### **Option 2: Local Setup (If you already have the code)**

**If you've already cloned the repository:**

#### **Windows:**
```powershell
# Method 1: PowerShell script
.\setup.ps1

# Method 2: Double-click (easiest)
setup.bat

# Method 3: With options
.\setup.ps1 -Help
.\setup.ps1 -SkipEnvCheck  # Skip environment setup
.\setup.ps1 -SkipMigrations  # Skip database migrations
.\setup.ps1 -SkipBuild  # Skip building the app
```

#### **Unix/Linux/macOS:**
```bash
# Method 1: Make executable and run
chmod +x setup.sh
./setup.sh

# Method 2: Run directly with bash
bash setup.sh

# Method 3: With options
./setup.sh --help
./setup.sh --skip-migrations
```

**What this does:**
1. ✅ **Checks prerequisites** (Go 1.21+, Docker, Docker Compose, Git)
2. 🔍 **Sets up environment variables** (creates/validates .env file)
3. 🐘 **Starts PostgreSQL database** (with health checks)
4. 🗄️ **Applies all database migrations**
5. 👤 **Creates admin user** (from environment variables)
6. 📦 **Installs Go dependencies**
7. 🔧 **Generates database code**
8. 🔨 **Builds the application**
9. 📋 **Shows you how to start the API**

---

### **Option 3: Manual Setup (Advanced)**

**If you prefer to do everything manually:**

```bash
# 1. Clone the repository
git clone https://github.com/treyhulse/directus-clone.git
cd directus-clone

# 2. Set up environment variables
cp env.example .env
# Edit .env with your preferred settings

# 3. Start the database
docker-compose up -d

# 4. Wait for database to be ready (about 15 seconds)
# Then apply migrations
docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < migrations/001_init.sql
docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < migrations/002_api_keys.sql
docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < migrations/003_admin_permissions.sql

# 5. Install Go dependencies
go mod tidy

# 6. Install sqlc (if not already installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 7. Generate database code
sqlc generate

# 8. Build the application
go build -o bin/api cmd/main.go

# 9. Run the application
go run cmd/main.go
```

---

## 🎯 **After Setup - Start Your API**

**Once setup is complete, start your API:**

```bash
# Option 1: Run directly
go run cmd/main.go

# Option 2: Run the built binary
./bin/api  # Linux/macOS
.\bin\api.exe  # Windows
```

**Your API will be available at:** `http://localhost:8080`

**Default admin credentials:**
- **Email:** `admin@example.com`
- **Password:** `password`

**API Keys for testing:**
- **Admin:** `admin_api_key_123`
- **Manager:** `manager_api_key_456`

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

#### **"Go is not installed"**
- **Solution:** Install Go 1.21+ from https://golang.org/dl/
- **Verify:** Run `go version` in terminal

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

### **Option C: Install Make on Windows (Optional)**

If you prefer to use the make commands on Windows, you can install `make`:

- **Using Chocolatey**: `choco install make`
- **Using Scoop**: `scoop install make`
- **Using WSL**: Install Ubuntu and use `sudo apt-get install make`

Then follow Option A above.

The API will be available at `http://localhost:8080`

## 📚 API Documentation

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
GET /items/:table
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

## 🧪 Testing the API

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

#### Make Commands (Linux/macOS or Windows with make installed)
- `make help` - Show available commands
- `make dev` - Start development server
- `make build` - Build the application
- `make run` - Run the built application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make deps` - Download dependencies
- `make sqlc` - Generate database code
- `make docker-up` - Start PostgreSQL
- `make docker-down` - Stop PostgreSQL
- `make setup` - Complete development setup

#### Direct Commands (Windows without make)
- `go mod tidy` - Download dependencies
- `go run cmd/main.go` - Start development server
- `go build -o app.exe cmd/main.go` - Build the application
- `./app.exe` - Run the built application
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

- [ ] Add more comprehensive tests
- [ ] Implement real-time subscriptions
- [ ] Add GraphQL support
- [ ] Implement file upload functionality
- [ ] Add audit logging
- [ ] Create admin dashboard
- [ ] Add rate limiting
- [ ] Implement caching layer 