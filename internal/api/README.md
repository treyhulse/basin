# Basin API Handlers

This directory contains the HTTP handlers for the Basin API, providing dynamic database access with role-based access control.

## Overview

Basin is a dynamic API that allows you to interact with any PostgreSQL database table through REST endpoints, with built-in authentication, authorization, and field-level security.

## Files

### Core Handlers
- **`auth.go`** - Authentication endpoints (`/auth/login`, `/auth/me`)
- **`items.go`** - Dynamic CRUD operations for database tables

### Tests
- **`auth_test.go`** - Comprehensive authentication handler tests
- **`items_test.go`** - Dynamic API endpoint tests

## Authentication Endpoints

### POST /auth/login
Authenticates a user and returns a JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true
  }
}
```

### GET /auth/me
Returns current user information from JWT token.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "is_active": true,
  "roles": ["admin"]
}
```

## Dynamic Items API

The items API provides CRUD operations for any database table with automatic role-based filtering.

### GET /:table
Retrieve all items from a table.

**Example:** `GET /products`

**Query Parameters:**
- `limit` - Maximum number of records (default: 100)
- `offset` - Number of records to skip (default: 0)
- `order` - Sort order (e.g., `created_at DESC`)
- `filter` - JSON filter object

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Product Name",
      "price": 29.99,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "total": 1,
    "limit": 100,
    "offset": 0
  }
}
```

### GET /:table/:id
Retrieve a specific item by ID.

**Example:** `GET /products/123e4567-e89b-12d3-a456-426614174000`

### POST /:table
Create a new item in the table.

**Example:** `POST /products`
```json
{
  "name": "New Product",
  "price": 39.99,
  "category": "electronics"
}
```

### PUT /:table/:id
Update an existing item.

**Example:** `PUT /products/123e4567-e89b-12d3-a456-426614174000`
```json
{
  "name": "Updated Product",
  "price": 49.99
}
```

### DELETE /:table/:id
Delete an item by ID.

**Example:** `DELETE /products/123e4567-e89b-12d3-a456-426614174000`

## Security Features

### Role-Based Access Control (RBAC)
- **Admin** - Full access to all tables and fields
- **Manager** - Access to business tables, restricted sensitive fields
- **Sales** - Read/write access to orders, customers, limited product access
- **Customer** - Read-only access to own data only

### Field-Level Security
Sensitive fields are automatically filtered based on user roles:
- `cost`, `wholesale_price` - Admin/Manager only
- `ssn`, `tax_id` - Admin only
- `internal_notes` - Staff only

### Row-Level Security
Users can only access data they're authorized to see:
- Customers see only their own orders
- Sales reps see their assigned accounts
- Admins have unrestricted access

### Input Validation
- SQL injection prevention
- UUID format validation
- Table name whitelisting
- Content-type validation

## Error Handling

The API returns standard HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

Error responses include detailed messages:
```json
{
  "error": "Invalid table name",
  "details": "Table 'invalid_table' does not exist or is not accessible"
}
```

## Testing

Run the API tests:
```bash
# All tests
make test

# Verbose output
make test-verbose

# With coverage
make test-coverage
```

The test suite covers:
- Authentication flows (login, token validation)
- CRUD operations for all HTTP methods
- Input validation and sanitization
- Role-based access control
- Error handling scenarios
- Security headers and CORS

## Usage Examples

See the `/documentation/api/` directory for detailed examples:
- **API Keys** - Authentication and token usage
- **Field Creation** - Dynamic schema management
- **Query Examples** - Advanced filtering and sorting