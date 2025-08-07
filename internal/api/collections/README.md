# Collections API

This directory contains documentation for the Collections API functionality within the Basin API system.

## Overview

The Collections API provides specialized handling for dynamic collections created from the `collections` and `fields` schema tables. This enables a Directus-style dynamic API where users can define their own data structures and interact with them through REST endpoints.

## Architecture

### CollectionsHandler

The `CollectionsHandler` (located in `../collections_handler.go`) provides specialized operations for user-created collections:

- **Schema Validation**: Validates incoming data against collection/field definitions
- **Field Type Validation & Conversion**: Ensures data types match field definitions and converts values appropriately
- **Dynamic Table Management**: Handles operations on tenant-specific dynamic tables
- **Enhanced Error Handling**: Provides collection-specific error messages and context

### Key Components

#### CollectionField
Represents a field definition from the `fields` table:
```go
type CollectionField struct {
    ID           uuid.UUID              `json:"id"`
    CollectionID uuid.UUID              `json:"collection_id"`
    Name         string                 `json:"name"`
    Type         string                 `json:"type"`
    Required     bool                   `json:"required"`
    Default      interface{}            `json:"default"`
    Validation   map[string]interface{} `json:"validation"`
    Options      map[string]interface{} `json:"options"`
}
```

#### Collection
Represents a collection definition from the `collections` table:
```go
type Collection struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    TenantID    uuid.UUID `json:"tenant_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## API Endpoints

Collections are accessed through the main `/items/:table` endpoints, where `:table` is the collection name. The system automatically routes requests to the appropriate handler based on whether the table is a user-created collection.

### Supported Operations

- **GET** `/items/:collection` - List all items in a collection
- **GET** `/items/:collection/:id` - Get a specific item by ID
- **POST** `/items/:collection` - Create a new item in a collection
- **PUT** `/items/:collection/:id` - Update an existing item
- **DELETE** `/items/:collection/:id` - Delete an item by ID

## Features

### Schema Validation
The CollectionsHandler validates incoming data against the collection's field definitions:

- **Required Fields**: Ensures all required fields are present
- **Field Types**: Validates data types match field definitions
- **Validation Rules**: Applies custom validation rules (min/max length, numeric ranges, patterns)
- **Default Values**: Applies default values for missing optional fields

### Type Conversion
Automatically converts field values to appropriate types:

- **String to Number**: Converts string representations to integers/floats
- **JSON Parsing**: Handles JSON string parsing for complex types
- **Boolean Conversion**: Converts various boolean representations
- **Date/Time**: Handles timestamp conversions

### Tenant Isolation
All collections are tenant-specific, ensuring data isolation between different organizations.

### RBAC Integration
Collections respect the same Role-Based Access Control (RBAC) system as other API endpoints:

- **Field-Level Permissions**: Users only see fields they have permission to access
- **Row-Level Security**: Users only see data they're authorized to view
- **Operation Permissions**: Create, read, update, delete permissions are enforced

## Complete CRUD Example

This example demonstrates the full lifecycle of creating, reading, updating, and deleting a collection item. We'll use a "blog_posts" collection as an example.

### Prerequisites
- You have a valid JWT token from logging in as admin
- The `blog_posts` collection has been created with appropriate fields
- Admin user has full permissions for the collection

### Authentication
First, get your authentication token:

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password"
  }'

# Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "ca7a49fa-6b4f-433b-ac4f-df67c301757d",
    "email": "admin@example.com",
    "first_name": "Admin",
    "last_name": "User"
  }
}
```

### 1. CREATE - Create a New Blog Post

```bash
# Create a new blog post
curl -X POST http://localhost:8080/items/blog_posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Getting Started with Basin API",
    "content": "This is a comprehensive guide to using the Basin API for dynamic collections...",
    "author": "Admin User",
    "status": "published",
    "tags": ["api", "tutorial", "basin"],
    "published_at": "2025-08-07T10:00:00Z"
  }'

# Expected Response (201 Created):
{
  "data": {
    "title": "Getting Started with Basin API",
    "content": "This is a comprehensive guide to using the Basin API for dynamic collections...",
    "author": "Admin User",
    "status": "published",
    "tags": ["api", "tutorial", "basin"],
    "published_at": "2025-08-07T10:00:00Z",
    "created_by": "ca7a49fa-6b4f-433b-ac4f-df67c301757d",
    "updated_by": "ca7a49fa-6b4f-433b-ac4f-df67c301757d",
    "created_at": "2025-08-07T13:45:30.123456Z",
    "updated_at": "2025-08-07T13:45:30.123456Z"
  },
  "meta": {
    "table": "blog_posts",
    "type": "collection"
  }
}
```

### 2. READ - Get All Blog Posts

```bash
# Get all blog posts
curl -X GET http://localhost:8080/items/blog_posts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected Response (200 OK):
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Getting Started with Basin API",
      "content": "This is a comprehensive guide to using the Basin API for dynamic collections...",
      "author": "Admin User",
      "status": "published",
      "tags": ["api", "tutorial", "basin"],
      "published_at": "2025-08-07T10:00:00Z",
      "created_at": "2025-08-07T13:45:30.123456Z",
      "updated_at": "2025-08-07T13:45:30.123456Z"
    }
  ],
  "meta": {
    "table": "blog_posts",
    "count": 1,
    "type": "collection",
    "collection": "blog_posts"
  }
}
```

### 3. READ - Get a Specific Blog Post

```bash
# Get a specific blog post by ID
curl -X GET http://localhost:8080/items/blog_posts/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected Response (200 OK):
{
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Getting Started with Basin API",
    "content": "This is a comprehensive guide to using the Basin API for dynamic collections...",
    "author": "Admin User",
    "status": "published",
    "tags": ["api", "tutorial", "basin"],
    "published_at": "2025-08-07T10:00:00Z",
    "created_at": "2025-08-07T13:45:30.123456Z",
    "updated_at": "2025-08-07T13:45:30.123456Z"
  },
  "meta": {
    "table": "blog_posts",
    "id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

### 4. UPDATE - Update the Blog Post

```bash
# Update the blog post
curl -X PUT http://localhost:8080/items/blog_posts/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Getting Started with Basin API - Updated",
    "content": "This is an updated comprehensive guide to using the Basin API for dynamic collections...",
    "status": "published",
    "tags": ["api", "tutorial", "basin", "updated"]
  }'

# Expected Response (200 OK):
{
  "data": {
    "title": "Getting Started with Basin API - Updated",
    "content": "This is an updated comprehensive guide to using the Basin API for dynamic collections...",
    "status": "published",
    "tags": ["api", "tutorial", "basin", "updated"]
  },
  "meta": {
    "table": "blog_posts",
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "collection"
  }
}
```

### 5. DELETE - Delete the Blog Post

```bash
# Delete the blog post
curl -X DELETE http://localhost:8080/items/blog_posts/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected Response (200 OK):
{
  "meta": {
    "table": "blog_posts",
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "type": "collection"
  }
}
```

### 6. Verify Deletion

```bash
# Verify the blog post is deleted
curl -X GET http://localhost:8080/items/blog_posts \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected Response (200 OK):
{
  "data": [],
  "meta": {
    "table": "blog_posts",
    "count": 0,
    "type": "collection",
    "collection": "blog_posts"
  }
}
```

## Error Handling Examples

### Validation Error (400 Bad Request)
```bash
# Try to create with missing required field
curl -X POST http://localhost:8080/items/blog_posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "content": "This post has no title"
  }'

# Expected Response (400 Bad Request):
{
  "error": "Failed to create collection item: validation failed: required field 'title' is missing"
}
```

### Permission Error (403 Forbidden)
```bash
# Try to access without proper permissions
curl -X GET http://localhost:8080/items/blog_posts \
  -H "Authorization: Bearer INVALID_TOKEN"

# Expected Response (403 Forbidden):
{
  "error": "Insufficient permissions"
}
```

### Not Found Error (404 Not Found)
```bash
# Try to get a non-existent item
curl -X GET http://localhost:8080/items/blog_posts/non-existent-id \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Expected Response (404 Not Found):
{
  "error": "Item not found"
}
```

## Field Type Validation

The CollectionsHandler automatically validates and converts field types:

```bash
# Create with various data types
curl -X POST http://localhost:8080/items/blog_posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Type Validation Example",
    "content": "Testing various data types",
    "view_count": "150",           // String converted to integer
    "rating": 4.5,                 // Float value
    "is_featured": "true",         // String converted to boolean
    "metadata": "{\"key\": \"value\"}", // JSON string parsed
    "publish_date": "2025-08-07"   // Date string parsed
  }'
```

This example demonstrates the complete CRUD lifecycle with proper error handling and data validation.

## Error Handling

The CollectionsHandler provides detailed error messages for validation failures:

- **Missing Required Fields**: Clear indication of which required fields are missing
- **Type Mismatches**: Specific error messages for type validation failures
- **Validation Rule Violations**: Detailed feedback on validation rule failures
- **Collection Not Found**: Proper error handling for non-existent collections

## Testing

The CollectionsHandler includes comprehensive tests covering:

- **Field Type Validation**: Tests for various data types and conversions
- **Validation Rules**: Tests for min/max length, numeric ranges, patterns
- **Error Handling**: Tests for various error conditions
- **Integration**: Tests with the main API flow

Run tests with:
```bash
go test ./internal/api -v -run TestCollectionsHandler
```

## Integration with Main API

The CollectionsHandler is integrated into the main `ItemsHandler` through delegation:

1. **Request Routing**: The main handler determines if a table is a user-created collection
2. **Delegation**: Collection operations are delegated to the CollectionsHandler
3. **Validation**: The CollectionsHandler performs schema validation and type conversion
4. **Execution**: Validated operations are delegated to the DynamicHandlers for actual database operations
5. **Response**: Results are returned through the main handler's response formatting

This architecture provides a clean separation of concerns while maintaining the unified API interface.
