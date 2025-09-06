# Complete Tenant Initialization Test

This document demonstrates how to test the new complete tenant initialization flow.

## What the Enhanced CreateTenant Does

When you create a tenant now, it automatically:

1. **Creates the tenant** in the database
2. **Creates default roles**: admin, manager, editor, viewer
3. **Adds the creator as admin** to the new tenant
4. **Sets up permissions** for all system tables
5. **Creates default collections**: customers, products, orders
6. **Adds appropriate fields** to each collection
7. **Sets up the database schema** (via existing triggers)

## Test the Complete Flow

### 1. Start the Server
```bash
go run cmd/main.go
```

### 2. Login to Get a Token
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password"
  }'
```

### 3. Create a New Tenant (Full Initialization)
```bash
curl -X POST http://localhost:8080/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "My New Company",
    "slug": "my-new-company",
    "domain": "mynewcompany.com"
  }'
```

### 4. Verify the Tenant was Created
```bash
curl -X GET http://localhost:8080/tenants \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. Check Available Collections
```bash
curl -X GET http://localhost:8080/items/collections \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 6. Check Available Roles
```bash
curl -X GET http://localhost:8080/items/roles \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 7. Switch to the New Tenant
```bash
curl -X POST http://localhost:8080/auth/switch-tenant \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "tenant_id": "NEW_TENANT_ID_FROM_STEP_3"
  }'
```

### 8. Verify Auth Context in New Tenant
```bash
curl -X GET http://localhost:8080/auth/context \
  -H "Authorization: Bearer NEW_TOKEN_FROM_STEP_7"
```

## Expected Results

After creating a tenant, you should see:

- **Tenant created** with the provided name and slug
- **4 default roles**: admin, manager, editor, viewer
- **Creator is admin** of the new tenant
- **3 default collections**: customers, products, orders
- **Appropriate fields** for each collection
- **Full permissions** set up for system tables
- **Database tables** automatically created (via triggers)

## Database Verification

You can also verify the initialization by checking the database directly:

```sql
-- Check tenant was created
SELECT * FROM tenants WHERE slug = 'my-new-company';

-- Check roles were created
SELECT * FROM roles WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'my-new-company');

-- Check user-tenant relationship
SELECT * FROM user_tenants WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'my-new-company');

-- Check collections were created
SELECT * FROM collections WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'my-new-company');

-- Check fields were created
SELECT f.* FROM fields f 
JOIN collections c ON f.collection_id = c.id 
WHERE c.tenant_id = (SELECT id FROM tenants WHERE slug = 'my-new-company');

-- Check permissions were created
SELECT p.* FROM permissions p 
JOIN roles r ON p.role_id = r.id 
WHERE r.tenant_id = (SELECT id FROM tenants WHERE slug = 'my-new-company');
```

## What Makes This Special

This implementation provides:

✅ **Complete tenant setup** in one API call
✅ **Proper RBAC structure** with roles and permissions
✅ **Default data collections** ready to use
✅ **Creator as admin** with full access
✅ **Database schema** automatically created
✅ **Transaction safety** (all-or-nothing initialization)
✅ **Ready to use immediately** after creation

The tenant is now fully functional and ready for users to start working with immediately!
