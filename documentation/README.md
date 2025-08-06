# Basin API Documentation

Welcome to the Basin API documentation. Basin is a dynamic REST API that provides secure, role-based access to PostgreSQL databases with automatic CRUD operations.

## 📁 Documentation Structure

### API Documentation (`/api/`)
- **[API Keys](api/api-keys.md)** - Authentication, token management, and API key usage
- **[Field Creation](api/field-creation.md)** - Dynamic schema management and field operations  
- **[Query Collection Fields](api/query-collection-fields.md)** - Advanced querying and field filtering

### Testing Documentation (`/testing/`)
- **[Field API Test](testing/field-api-test.ps1)** - PowerShell script for testing field operations

## 🚀 Quick Start

1. **Setup the application:**
   ```bash
   make setup
   ```

2. **Start the API:**
   ```bash
   make start
   ```

3. **Run tests:**
   ```bash
   make test
   ```

## 📖 Core Concepts

### Dynamic API
Basin automatically generates REST endpoints for any table in your PostgreSQL database. No code generation or manual endpoint creation required.

### Role-Based Access Control (RBAC)
Built-in security system that controls:
- Which tables users can access
- Which fields are visible/editable
- Row-level data filtering
- Operation permissions (read/write/delete)

### Authentication
JWT-based authentication with:
- Secure login endpoints
- Token validation middleware
- User session management
- Role assignment

## 🔗 Related Documentation

- **[Main README](../README.md)** - Project overview and setup
- **[API Handlers](../internal/api/README.md)** - Detailed API endpoint documentation
- **[Project Overview](../Project.md)** - Technical architecture and design decisions

## 🛠️ Development

### Available Make Commands
```bash
make help          # Show all available commands
make setup         # Initial project setup
make start         # Start the application
make test          # Run tests (clean output)
make test-verbose  # Run tests with detailed output
make test-coverage # Run tests with coverage report
make migrate       # Apply database migrations
make build         # Build the application
make clean         # Clean build artifacts
```

### Project Structure
```
basin/
├── cmd/                    # Application entry point
├── internal/
│   ├── api/               # HTTP handlers (see API README)
│   ├── config/            # Configuration management
│   ├── db/                # Database connection and queries
│   ├── middleware/        # HTTP middleware
│   ├── rbac/              # Role-based access control
│   └── schema/            # Schema management
├── test/                  # Integration tests
├── documentation/         # This documentation
├── migrations/            # Database migrations
└── scripts/               # Utility scripts
```

## 📝 Contributing

When adding new features:
1. Add comprehensive tests
2. Update relevant documentation
3. Follow the existing code patterns
4. Ensure all tests pass: `make test`

## 📞 Support

For questions or issues:
1. Check the documentation in this directory
2. Review the API README at `internal/api/README.md`
3. Look at test files for usage examples
4. Check the main project README