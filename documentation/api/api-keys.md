# 🔑 API Key Management Testing Guide

## Overview
API keys provide programmatic access to the Basin API with the same permissions as the user they belong to.

## ✅ **Recent Fixes Applied**
1. ✅ Admin can now access `/items/roles` and `/items/permissions`
2. ✅ `/items/api_keys` GET endpoint no longer throws 500 error
3. ✅ API key authentication now works properly
4. ✅ All issues resolved after running `sqlc generate`
5. ✅ **UPDATE and DELETE operations now work** (no more demo mode!)
6. ✅ Complete CRUD operations for API keys, collections, fields, users

## 🧪 **Testing API Key Management**

### **1. Login as Admin**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'
```
Save the JWT token from the response.

### **2. Test Admin Access to System Tables (Previously Fixed Issues)**

**Test Roles Access (was 403, now works):**
```bash
curl -X GET http://localhost:8080/items/roles \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Test Permissions Access (was 403, now works):**
```bash
curl -X GET http://localhost:8080/items/permissions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### **3. View Current API Keys (was 500 error, now works)**
```bash
curl -X GET http://localhost:8080/items/api_keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Expected Response:**
```json
{
  "data": [
    {
      "id": "5835dec3-e289-45b1-bb16-d6372a880041",
      "user_id": "3c474327-8a8f-4b4b-b25f-25b775b9a76e",
      "name": "KEY1",
      "is_active": true,
      "expires_at": "2025-12-31T23:59:59Z",
      "last_used_at": null,
      "created_at": "2025-08-06T03:20:43.385548Z",
      "updated_at": "2025-08-06T03:20:43.385548Z"
    }
  ],
  "meta": {
    "table": "api_keys",
    "count": 1,
    "type": "schema"
  }
}
```

### **4. Create a New API Key**
```bash
curl -X POST http://localhost:8080/items/api_keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Admin API Key",
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

**Response will include:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "user-uuid", 
    "name": "My Admin API Key",
    "api_key": "basin_1234567890abcdef...",
    "is_active": true,
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**⚠️ IMPORTANT:** The `api_key` field is only returned during creation! Save it securely.

### **5. Test API Key Authentication (was 401, now works)**

Use the API key from the previous step:
```bash
curl -X GET http://localhost:8080/items/collections \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"
```

**Test Multiple Endpoints with API Key:**
```bash
# Test collections
curl -X GET http://localhost:8080/items/collections \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"

# Test fields  
curl -X GET http://localhost:8080/items/fields \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"

# Test users (admin should see all)
curl -X GET http://localhost:8080/items/users \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"

# Test roles (should now work)
curl -X GET http://localhost:8080/items/roles \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"

# Test permissions (should now work)
curl -X GET http://localhost:8080/items/permissions \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"

# Test API keys (should show only keys owned by this user)
curl -X GET http://localhost:8080/items/api_keys \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE"
```

### **6. Test API Key Data Operations**

**Create a new collection using API key:**
```bash
curl -X POST http://localhost:8080/items/collections \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api_test_posts",
    "display_name": "API Test Posts",
    "description": "Created via API key"
  }'
```

**Create fields for the collection:**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer basin_YOUR_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "COLLECTION_ID_FROM_ABOVE",
    "name": "title",
    "display_name": "Title",
    "type": "text",
    "required": true,
    "sort_order": 1
  }'
```

### **7. Create API Key for Another User (Admin Only)**
```bash
curl -X POST http://localhost:8080/items/api_keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "target-user-uuid",
    "name": "Manager API Key",
    "expires_at": "2025-06-30T23:59:59Z"
  }'
```

### **8. Update API Key (Now Actually Works!)**

**Deactivate an API key:**
```bash
curl -X PUT http://localhost:8080/items/api_keys/API_KEY_UUID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false
  }'
```

**Update name and expiration:**
```bash
curl -X PUT http://localhost:8080/items/api_keys/API_KEY_UUID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated API Key Name",
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

### **9. Delete API Key (Now Actually Works!)**
```bash
curl -X DELETE http://localhost:8080/items/api_keys/API_KEY_UUID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Expected Response:**
```json
{
  "meta": {
    "table": "api_keys",
    "id": "API_KEY_UUID"
  }
}
```

## 🔐 **Permission Inheritance**

API keys inherit the exact same permissions as their owner:

### **Admin API Key**
- ✅ Can access all tables (collections, fields, users, roles, permissions, api_keys)
- ✅ Can create/read/update/delete any data
- ✅ Can manage other users' API keys

### **Manager API Key** 
- ✅ Can access collections, fields (read-only)
- ✅ Can manage products, view orders
- ✅ Can manage their own API keys only
- ❌ Cannot access users, roles, permissions tables

### **Sales API Key**
- ✅ Can view products (limited fields)
- ✅ Can create orders
- ✅ Can manage their own API keys only
- ❌ Cannot access admin tables

## 🛡️ **Security Features**

### **1. Secure Generation**
- API keys are 64 characters long with `basin_` prefix
- Generated using crypto/rand for cryptographic security
- Stored as SHA-256 hashes in database

### **2. Access Control**
- Users can only see/manage their own API keys (unless admin)
- Field-level permissions control what data is returned
- Row-level filtering ensures data isolation

### **3. Expiration & Monitoring**
- API keys can have expiration dates
- Last used timestamp is tracked
- Keys can be deactivated without deletion

### **4. Audit Trail**
- Creation and modification timestamps
- User association for accountability
- Activity tracking via last_used_at

## 🚀 **Programmatic Usage Examples**

### **Node.js Example**
```javascript
const apiKey = 'basin_1234567890abcdef...';
const baseURL = 'http://localhost:8080';

// Fetch collections
const response = await fetch(`${baseURL}/items/collections`, {
  headers: {
    'Authorization': `Bearer ${apiKey}`,
    'Content-Type': 'application/json'
  }
});

const collections = await response.json();
console.log(collections);
```

### **Python Example**
```python
import requests

api_key = 'basin_1234567890abcdef...'
base_url = 'http://localhost:8080'

headers = {
    'Authorization': f'Bearer {api_key}',
    'Content-Type': 'application/json'
}

# Create a new blog post
response = requests.post(f'{base_url}/items/blog_posts', 
    headers=headers,
    json={
        'title': 'API Created Post',
        'content': 'This post was created via API key!'
    }
)

print(response.json())
```

### **cURL Scripts**
```bash
#!/bin/bash
API_KEY="basin_1234567890abcdef..."
BASE_URL="http://localhost:8080"

# Function to make authenticated requests
api_request() {
    curl -H "Authorization: Bearer $API_KEY" \
         -H "Content-Type: application/json" \
         "$BASE_URL/$1" "${@:2}"
}

# Usage examples
api_request "items/collections"
api_request "items/blog_posts" -X POST -d '{"title":"Test","content":"API test"}'
```

## 🔧 **Management Best Practices**

### **1. Naming Convention**
- Use descriptive names: "Production API", "Mobile App Key", "Analytics Service"
- Include purpose and environment: "Staging Dashboard Key"

### **2. Expiration Policy**
- Set reasonable expiration dates (1 year max)
- Rotate keys regularly for security
- Use shorter expiration for high-privilege keys

### **3. Access Principle**
- Create keys with minimum required permissions
- Use separate keys for different services/purposes
- Monitor usage via last_used_at timestamps

### **4. Security**
- Never log API keys in plaintext
- Store securely in environment variables
- Revoke unused or compromised keys immediately
- Regular audit of active keys

This API key system provides secure, scalable programmatic access while maintaining the same RBAC security model as the web interface!

---

## 🔍 **What Was Fixed Under the Hood**

### **1. Admin Permissions Fix**
- Added missing `read` permissions for `roles` and `permissions` tables
- Admin now has full `*` access to all system tables
- Fixed via direct database permission inserts

### **2. API Keys GET Endpoint Fix**  
- Fixed tenant filtering logic for `api_keys` table
- API keys table doesn't have `tenant_id`, so we filter by `user_id` instead
- Users can only see their own API keys (unless admin)
- Updated `internal/api/items.go` filtering logic

### **3. API Key Authentication Fix**
- Fixed middleware to hash incoming API keys before database lookup  
- API keys are stored as SHA-256 hashes but middleware was looking up raw values
- Added `hashAPIKey()` function to `internal/middleware/auth.go`
- Required running `sqlc generate` to update generated code

### **4. Security Improvements**
- API keys inherit exact same permissions as their owner
- Row-level filtering ensures users can only see their own keys
- Proper tenant isolation maintained for all operations
- Consistent security model across JWT and API key authentication

## 🎯 **Current Status**

After these fixes:
- ✅ Admin can access all system tables (`users`, `roles`, `permissions`, `collections`, `fields`, `api_keys`)
- ✅ API key creation and listing works without errors
- ✅ API keys can be used for authentication with same permissions as owner
- ✅ Multi-tenant isolation is maintained
- ✅ All authentication methods work consistently

**Key Lesson:** Always run `sqlc generate` after making database schema or query changes!

## 🚀 **Latest Update: Full CRUD Operations**

### **What Just Got Fixed:**
- ✅ **UPDATE operations** - No more "demo mode - not actually updated"
- ✅ **DELETE operations** - No more "demo mode - not actually deleted" 
- ✅ **Complete API key management** - Create, read, update, delete all work
- ✅ **All system tables** - collections, fields, users, roles, permissions, api_keys
- ✅ **Dynamic data tables** - Full CRUD on your custom collections
- ✅ **Security & permissions** - All operations respect RBAC and tenant isolation

### **New Features:**
- **Ownership validation** - Users can only update/delete their own API keys (unless admin)
- **Self-deletion prevention** - Users cannot delete their own user account
- **Tenant isolation** - All operations respect multi-tenant boundaries
- **Proper error handling** - Detailed error messages for debugging
- **Data validation** - Field validation and type checking

### **What Works Now:**
```bash
# Update any API key field
PUT /items/api_keys/:id - ✅ WORKS
DELETE /items/api_keys/:id - ✅ WORKS

# Update any collection
PUT /items/collections/:id - ✅ WORKS  
DELETE /items/collections/:id - ✅ WORKS (triggers data table deletion)

# Update any field  
PUT /items/fields/:id - ✅ WORKS
DELETE /items/fields/:id - ✅ WORKS

# Update any user
PUT /items/users/:id - ✅ WORKS
DELETE /items/users/:id - ✅ WORKS (with self-deletion protection)

# Update/delete data in your custom collections
PUT /items/your_collection/:id - ✅ WORKS
DELETE /items/your_collection/:id - ✅ WORKS
```

**Your API is now production-ready with full CRUD operations!** 🎉

## 🔧 **Latest Fix: Collection Creation Now Works**

### **Issue Resolved:**
The collection creation was failing with `syntax error at or near "default"` because PostgreSQL treats "default" as a reserved keyword. The database trigger function wasn't properly quoting schema names.

### **What Was Fixed:**
- ✅ **Schema Creation** - Properly quoted "default" schema name in PostgreSQL
- ✅ **Trigger Functions** - Updated `create_data_table()` and `drop_data_table()` functions
- ✅ **Tenant Management** - Fixed `create_tenant()` function for proper schema handling

### **Test Collection Creation:**
```bash
# Login first
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'

# Create a collection (should now work!)
curl -X POST http://localhost:8080/items/collections \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "blog_posts",
    "display_name": "Blog Posts", 
    "description": "Dynamic blog post collection",
    "icon": "article"
  }'
```

**Expected Response:**
```json
{
  "data": {
    "id": "uuid-here",
    "name": "blog_posts",
    "display_name": "Blog Posts",
    "description": "Dynamic blog post collection",
    "icon": "article",
    "is_system": false,
    "tenant_id": "tenant-uuid",
    "created_by": "user-uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  },
  "meta": {
    "table": "collections"
  }
}
```

### **What Happens Automatically:**
1. ✅ **Collection Created** - Record added to `collections` table
2. ✅ **Schema Created** - `"default"` schema created (properly quoted)
3. ✅ **Data Table Created** - `"default".data_blog_posts` table created automatically
4. ✅ **RLS Enabled** - Row-level security policies applied
5. ✅ **Permissions Set** - Tenant isolation and user ownership policies created

**Collection creation and field management are now fully operational!** 🚀