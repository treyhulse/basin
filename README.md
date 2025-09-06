# 🚀 Basin API - Dynamic Multi-Tenant REST API with RBAC

A **production-ready** Go-based REST API that provides Directus-style functionality with comprehensive Role-Based Access Control (RBAC), multi-tenancy, and dynamic schema management. This is a fully functional, enterprise-grade API that can dynamically create and manage database tables, collections, and fields through the same REST endpoints.

## 🌟 **Key Features**

- **🔑 Multi-Tenant Authentication** - JWT-based auth with tenant isolation
- **🔐 Advanced RBAC System** - Role-based access control with field-level and row-level security
- **📦 Dynamic Schema Management** - Create collections and fields via REST API
- **🏢 Multi-Tenancy** - Complete tenant isolation with user-tenant relationships
- **📊 Generic CRUD API** - `GET /items/:table` endpoints for any data table
- **🛡️ Comprehensive Security** - Field-level permissions, row-level filtering, and input validation
- **🐘 PostgreSQL** - Robust database with automatic table creation
- **⚡ Type-Safe DB Access** - Generated with sqlc for compile-time safety
- **🐳 Docker Ready** - One-command development setup
- **📚 OpenAPI/Swagger** - Auto-generated API documentation

## 🆕 **NEW: Self-Referential Schema Management**

The Basin API can manage its own schema through the same REST endpoints:

- **Collections** define data models (like database schemas)
- **Fields** define columns and validation rules
- **Dynamic Tables** are automatically created with `data_` prefix
- **Same RBAC System** applies to schema management
- **Triggers** handle automatic table creation/deletion

## 🚀 **Quick Start (2 Commands)**

### **Prerequisites**
- **🐳 Docker Desktop** - [Download here](https://www.docker.com/products/docker-desktop/)
- **🐹 Go 1.21+** - [Download here](https://golang.org/dl/)
- **📦 Git** - [Download here](https://git-scm.com/)

### **Setup Commands**
```bash
# 1. Clone the repository
git clone <your-repo-url>
cd basin

# 2. Run the setup command
make setup
```

**That's it!** Your API will automatically start at http://localhost:8080

### **Daily Development Commands**
```bash
make start    # Cold start everything
make dev      # Start development server
make stop     # Stop everything
make restart  # Restart everything
```

---

## 🏗️ **Architecture Overview**

### **Core Components**
- **Multi-Tenant System** - Complete tenant isolation with user-tenant relationships
- **Dynamic Schema Engine** - Collections and fields managed via REST API
- **RBAC Engine** - Comprehensive permission system with field-level control
- **Generic CRUD Handler** - Single endpoint handles any table with automatic filtering

### **Data Flow**
1. **Collections** define data structure (e.g., "blog_posts")
2. **Fields** define columns and validation (e.g., "title", "content", "author")
3. **Dynamic Tables** are automatically created (`data_blog_posts`)
4. **RBAC Policies** control access at field and row levels
5. **Generic API** provides CRUD operations for any collection

---

## 🔌 **API Endpoints**

### **Authentication**
- `POST /auth/login` - User login with tenant context
- `POST /auth/signup` - User registration
- `GET /auth/me` - Get current user info
- `POST /auth/switch-tenant` - Switch between user's tenants
- `GET /auth/context` - Get current auth context
- `GET /auth/tenants` - Get user's accessible tenants

### **Dynamic CRUD Operations**
- `GET /items/:table` - List items with RBAC filtering, pagination, and sorting
- `GET /items/:table/:id` - Get single item
- `POST /items/:table` - Create new item
- `PUT /items/:table/:id` - Update item
- `DELETE /items/:table/:id` - Delete item

### **Schema Management (Same Endpoints!)**
- `GET /items/collections` - List all collections
- `POST /items/collections` - Create new collection
- `PUT /items/collections/:id` - Update collection
- `DELETE /items/collections/:id` - Delete collection

- `GET /items/fields` - List all fields
- `POST /items/fields` - Create new field
- `PUT /items/fields/:id` - Update field
- `DELETE /items/fields/:id` - Delete field

### **Tenant Management**
- `POST /tenants` - Create new tenant
- `GET /tenants` - List all tenants
- `GET /tenants/:id` - Get tenant details
- `PUT /tenants/:id` - Update tenant
- `DELETE /tenants/:id` - Delete tenant
- `POST /tenants/:id/users` - Add user to tenant
- `DELETE /tenants/:id/users/:user_id` - Remove user from tenant
- `POST /tenants/:id/join` - Join existing tenant

### **System**
- `GET /health` - Health check
- `GET /` - API information
- `GET /swagger/*` - OpenAPI/Swagger documentation

---

## 🗄️ **Database Schema**

### **Core Tables**
- **`tenants`** - Multi-tenant organization system
- **`users`** - User accounts with authentication
- **`roles`** - Role definitions per tenant
- **`user_roles`** - User-role assignments
- **`user_tenants`** - Many-to-many user-tenant relationships
- **`permissions`** - RBAC policies with field-level access control

### **Schema Management Tables**
- **`collections`** - Data model definitions
- **`fields`** - Field definitions for collections

### **Dynamic Data Tables**
- **`data_[collection_name]`** - Automatically created for each collection

### **Sample Business Tables (Created Per Tenant)**
- **`customers`** - Customer information
- **`products`** - Product catalog  
- **`orders`** - Order management

---

## 🔐 **RBAC System**

### **Permission Structure**
```sql
permissions (
    role_id UUID,           -- Which role this applies to
    table_name VARCHAR(100), -- Which table this applies to
    action VARCHAR(50),      -- 'create', 'read', 'update', 'delete'
    field_filter JSONB,      -- Row-level filtering {"field": "value"}
    allowed_fields TEXT[],   -- Field-level access control
    tenant_id UUID           -- Tenant isolation
)
```

### **Security Features**
- **Field-Level Security** - Control which columns users can see
- **Row-Level Security** - Filter records based on user context
- **Action-Based Permissions** - CRUD operation granularity
- **Role Inheritance** - Users can have multiple roles
- **Tenant Isolation** - Complete data separation between tenants

---

## 🔄 **Dynamic Schema Management Example**

### **1. Create a Collection**
```bash
POST /items/collections
{
  "name": "blog_posts",
  "display_name": "Blog Posts",
  "description": "Blog post management",
  "icon": "article"
}
```

### **2. Add Fields to the Collection**
```bash
POST /items/fields
{
  "collection_id": "collection-uuid",
  "name": "title",
  "display_name": "Title",
  "type": "string",
  "is_required": true
}

POST /items/fields
{
  "collection_id": "collection-uuid",
  "name": "content",
  "display_name": "Content",
  "type": "text",
  "is_required": true
}
```

### **3. System Automatically Creates**
- `data_blog_posts` table with proper columns
- Triggers for automatic updates
- Indexes for performance

### **4. Use the New Collection**
```bash
# Create a blog post
POST /items/blog_posts
{
  "title": "My First Post",
  "content": "Hello, world!"
}

# List all blog posts
GET /items/blog_posts
```

---

## 🚀 **Getting Started**

### **Option 1: Automated Setup (Recommended)**
```bash
# Clone and setup
git clone <your-repo-url>
cd basin
make setup
```

### **Option 2: Manual Setup**
```bash
# 1. Clone the repository
git clone <your-repo-url>
cd basin

# 2. Set up environment variables
cp env.example .env
# Edit .env with your preferred settings

# 3. Install dependencies
make deps

# 4. Start the database
make docker-up

# 5. Apply migrations
make migrate

# 6. Generate database code
make generate

# 7. Build the application
make build

# 8. Start the server
make dev
```

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

---

## 🧪 **Testing the API**

### **1. Login as Admin**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password"
  }'
```

### **2. Create a Tenant**
```bash
curl -X POST http://localhost:8080/tenants \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "slug": "my-company"
  }'
```

### **3. Create a Collection**
```bash
curl -X POST http://items/collections \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "display_name": "Products",
    "description": "Product catalog"
  }'
```

### **4. Add Fields to Collection**
```bash
curl -X POST http://items/fields \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "COLLECTION_UUID",
    "name": "name",
    "display_name": "Product Name",
    "type": "string",
    "is_required": true
  }'
```

---

## 🛠️ **Development Commands**

### **Make Commands (Recommended)**
```bash
make help          # Show all available commands
make setup         # Complete initial setup (first time only)
make start         # Cold start the application
make dev           # Start development server
make stop          # Stop the application
make restart       # Stop and restart everything
make build         # Build the application
make test          # Run tests
make clean         # Clean build artifacts
make deps          # Update dependencies
make generate      # Regenerate database code
make migrate       # Apply database migrations
make docker-up     # Start PostgreSQL
make docker-down   # Stop PostgreSQL
make docker-logs   # Show Docker logs
make docs          # Generate Swagger documentation
```

### **Direct Commands (Alternative)**
```bash
go mod tidy                    # Download dependencies
go run cmd/main.go            # Start development server
go build -o bin/api cmd/main.go # Build the application
./bin/api                     # Run the built application
go test ./...                 # Run tests
sqlc generate                 # Generate database code
docker-compose up -d          # Start PostgreSQL
docker-compose down           # Stop PostgreSQL
```

---

## 📊 **API Features**

### **Query Parameters**
- **Pagination**: `limit`, `offset`, `page`, `per_page`
- **Sorting**: `sort`, `order` (asc/desc)
- **Filtering**: Field-based filtering via query parameters

### **Response Format**
```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20,
    "total_pages": 5
  }
}
```

### **Error Handling**
- **400** - Bad Request (validation errors)
- **401** - Unauthorized (authentication required)
- **403** - Forbidden (insufficient permissions)
- **404** - Not Found (resource doesn't exist)
- **409** - Conflict (duplicate resource)
- **500** - Internal Server Error

---

## 🔒 **Security Features**

### **Authentication**
- JWT tokens with configurable expiry
- Secure password hashing with bcrypt
- Token-based session management

### **Authorization**
- Role-based access control (RBAC)
- Field-level permissions
- Row-level security with JSONB filters
- Tenant isolation

### **Input Validation**
- SQL injection protection via sqlc
- Table name validation
- JSON payload validation
- UUID validation

---

## 🚀 **Deployment**

### **Production Build**
```bash
make build
```

### **Docker Deployment**
```bash
# Build the application
docker build -t basin-api .

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-db-password \
  -e JWT_SECRET=your-jwt-secret \
  basin-api
```

### **Environment Variables for Production**
```bash
DB_HOST=your-production-db-host
DB_PASSWORD=your-secure-password
JWT_SECRET=your-super-secure-jwt-secret
SERVER_MODE=release
```

---

## 🔄 **Roadmap & Future Enhancements**

### **Completed Features** ✅
- [x] Multi-tenant architecture
- [x] Dynamic schema management
- [x] Comprehensive RBAC system
- [x] Generic CRUD API
- [x] Field-level and row-level security
- [x] OpenAPI/Swagger documentation
- [x] Docker development environment
- [x] Comprehensive testing suite

### **Planned Enhancements** 🚧
- [ ] Real-time subscriptions (WebSocket)
- [ ] GraphQL support
- [ ] File upload functionality
- [ ] Audit logging system
- [ ] Rate limiting
- [ ] Caching layer
- [ ] Admin dashboard frontend
- [ ] API versioning
- [ ] Bulk operations
- [ ] Advanced query language

---

## 🤝 **Contributing**

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 **License**

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 **Support**

For support and questions:
- Create an issue in the repository
- Check the API documentation at `http://localhost:8080/swagger/` when running
- Review the comprehensive test suite for usage examples

---

## 🔧 **Troubleshooting**

### **Common Issues**

#### **"make is not recognized"**
- **Windows**: Install using Scoop (`scoop install make`) or Chocolatey (`choco install make`)
- **Alternative**: Use the direct commands listed above

#### **"docker-compose is not recognized"**
- Make sure Docker Desktop is installed and running
- Try using `docker compose up -d` (with a space instead of hyphen)
- Restart your terminal after installing Docker Desktop

#### **Database connection errors**
- Ensure PostgreSQL is running: `docker-compose ps`
- Check if port 5432 is available
- Restart Docker Desktop if needed

#### **"sqlc is not recognized"**
- The setup script installs it automatically
- Manual installation: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

---

**🎯 The Basin API is production-ready and provides enterprise-grade functionality for building scalable, secure, and flexible applications with dynamic schema management and comprehensive RBAC.** 