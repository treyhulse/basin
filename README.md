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

## 🛠️ Quick Start

### 🚀 One-Command Installation (Recommended)

**From anywhere, with a single command:**

**Unix/Linux/macOS:**
```bash
bash <(curl -sL https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.sh)
```

**Windows (PowerShell):**
```powershell
powershell -ExecutionPolicy Bypass -Command "& { iwr https://raw.githubusercontent.com/treyhulse/directus-clone/main/install.ps1 -UseBasicParsing | iex }"
```

This will:
- ✅ **Check all prerequisites** (Go 1.21+, Docker 20.0+, Docker Compose 2.0+)
- 🔄 **Clone the repository** (creates `directus-clone` directory)
- 🔍 **Validate environment variables** (checks .env file and validates values)
- 🐘 **Start a fresh PostgreSQL database** (with health checks)
- 🗄️ **Apply all migrations dynamically** (finds and runs all .sql files in migrations/)
- 📦 **Install Go dependencies** (go mod tidy)
- 🔧 **Generate database code** (sqlc generate)
- 🔨 **Build the application** (go build)
- 📋 **Display all credentials and endpoints**

**Perfect for:**
- 🆕 **New projects** - Start from scratch in any directory
- 🔄 **Quick demos** - Get up and running in minutes
- 🧪 **Testing** - Fresh environment every time
- 📚 **Learning** - No complex setup required

### 🛠️ Local Setup (Alternative)

**Windows:**
```powershell
# PowerShell
.\setup.ps1

# Or simply double-click:
setup.bat

# With options:
.\setup.ps1 -Help
.\setup.ps1 -Version
```

**Unix/Linux/macOS:**
```bash
# Make executable and run
chmod +x setup.sh
./setup.sh

# Or run directly with bash
bash setup.sh

# With options:
./setup.sh --help
./setup.sh --version
```

This will:
- ✅ **Check all prerequisites** (Go 1.21+, Docker 20.0+, Docker Compose 2.0+)
- 🔄 **Clone/update repository** (automatically pulls latest changes)
- 🔍 **Validate environment variables** (checks .env file and validates values)
- 🐘 **Start a fresh PostgreSQL database** (with health checks)
- 🗄️ **Apply all migrations dynamically** (finds and runs all .sql files in migrations/)
- 📦 **Install Go dependencies** (go mod tidy)
- 🔧 **Generate database code** (sqlc generate)
- 🔨 **Build the application** (go build)
- 📋 **Display all credentials and endpoints**

**Dynamic Features:**
- 🆕 **Future-proof**: Automatically handles new migration files
- 🔧 **Environment validation**: Checks for required env vars and validates formats
- 📊 **Version checking**: Ensures minimum versions of Go, Docker, etc.
- 🔄 **Repository management**: Clones or updates from git automatically
- 🛡️ **Error handling**: Comprehensive error checking and reporting

### Manual Setup
```bash
# 1. Clone and Setup
git clone <repository-url>
cd directus-clone

# 2. Install Dependencies
go mod tidy

# 3. Start Database
docker-compose up -d

# 4. Install sqlc (if not already installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 5. Generate Database Code
sqlc generate

# 6. Run the Application
go run cmd/main.go
```

### Option C: Install Make on Windows (Optional)

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