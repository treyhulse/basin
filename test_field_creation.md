# 🏗️ Field Creation Guide

## 📋 **Overview**
Now that collection creation is working, let's create fields for our collections. Fields define the structure of your dynamic data tables.

## 🔑 **Step 1: Get Your Auth Token**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'
```

**Save the JWT token from the response!**

## 📊 **Step 2: Get Collection ID**
```bash
# List collections to get the ID
curl -X GET http://localhost:8080/items/collections \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**From our test, the blog_posts collection ID is:** `5c059eeb-cd5d-4fb7-a35a-c9c485b2e024`

## 🏗️ **Step 3: Create Fields for Blog Posts Collection**

### **Field 1: Title (Text, Required)**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "title",
    "display_name": "Title",
    "type": "text",
    "is_required": true,
    "is_unique": false,
    "sort_order": 1
  }'
```

### **Field 2: Content (Text, Optional)**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "content",
    "display_name": "Content",
    "type": "text",
    "is_required": false,
    "is_unique": false,
    "sort_order": 2
  }'
```

### **Field 3: Published Date (DateTime, Optional)**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "published_at",
    "display_name": "Published At",
    "type": "datetime",
    "is_required": false,
    "is_unique": false,
    "sort_order": 3
  }'
```

### **Field 4: Is Published (Boolean, Default false)**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "is_published",
    "display_name": "Is Published",
    "type": "boolean",
    "is_required": false,
    "is_unique": false,
    "default_value": "false",
    "sort_order": 4
  }'
```

### **Field 5: View Count (Number, Default 0)**
```bash
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "view_count",
    "display_name": "View Count",
    "type": "number",
    "is_required": false,
    "is_unique": false,
    "default_value": "0",
    "sort_order": 5
  }'
```

## 🔍 **Step 4: Verify Fields Were Created**
```bash
# List all fields for the collection
curl -X GET http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 📋 **Step 5: Check the Dynamic Table Structure**
After creating fields, the system should automatically update the `"default".data_blog_posts` table structure.

**Check the table structure:**
```sql
-- Connect to database and check table
docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db -c '\d "default".data_blog_posts'
```

**Expected table structure:**
```
Table "default.data_blog_posts"
    Column     |           Type           | Nullable |      Default       
---------------+--------------------------+----------+--------------------
 id            | uuid                     | not null | uuid_generate_v4()
 created_at    | timestamp with time zone |          | now()
 updated_at    | timestamp with time zone |          | now()
 created_by    | uuid                     |          |
 updated_by    | uuid                     |          |
 title         | text                     | not null |
 content       | text                     |          |
 published_at  | timestamp with time zone |          |
 is_published  | boolean                  |          |
 view_count    | numeric                  |          |
```

## 🎯 **Supported Field Types**

| Type | PostgreSQL Type | Description |
|------|----------------|-------------|
| `text` | TEXT | String/text data |
| `number` | NUMERIC | Numbers (integer/decimal) |
| `boolean` | BOOLEAN | True/false values |
| `date` | DATE | Date only |
| `datetime` | TIMESTAMP WITH TIME ZONE | Date and time |

## ⚙️ **Field Properties**

| Property | Type | Description |
|----------|------|-------------|
| `collection_id` | UUID | **Required** - ID of the collection |
| `name` | string | **Required** - Field name (database column) |
| `display_name` | string | Human-readable field name |
| `type` | string | **Required** - Field type (see above) |
| `is_primary` | boolean | Whether this is a primary key |
| `is_required` | boolean | Whether field is required (NOT NULL) |
| `is_unique` | boolean | Whether field values must be unique |
| `default_value` | string | Default value for the field |
| `sort_order` | number | Display order in forms/tables |

## 🔧 **Latest Fix: Field Creation Now Works**

### **Issue Resolved:**
Field creation was failing with `"data table does not exist"` error because the `tableExists` function couldn't handle quoted schema names properly.

### **What Was Fixed:**
- ✅ **Schema Name Handling** - Fixed `addColumnToDataTable` function to use unquoted schema names for existence checks
- ✅ **ALTER TABLE Queries** - Properly quoted schema names for ALTER TABLE statements
- ✅ **Table Detection** - `tableExists` function now correctly finds tables in the "default" schema

### **Test Field Creation Now:**
```bash
# 1. Login first
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'

# 2. Create a field (should work now!)
curl -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "title",
    "display_name": "Title",
    "type": "text",
    "is_required": true,
    "is_unique": false,
    "sort_order": 1
  }'
```

## 🚨 **Important Notes**

1. **Collection Must Exist First** - Create the collection before adding fields
2. **Field Names** - Use snake_case for database compatibility
3. **Required Fields** - Will add NOT NULL constraint to table
4. **Dynamic Updates** - The data table structure updates automatically
5. **Tenant Isolation** - Fields are isolated by tenant (multi-tenant safe)
6. **Schema Quoting** - System handles PostgreSQL reserved keywords automatically

## 🎉 **Next Steps**

After creating fields, you can:
1. ✅ **Insert data** into your dynamic collection: `POST /items/blog_posts`
2. ✅ **Query data** from your collection: `GET /items/blog_posts`
3. ✅ **Update records**: `PUT /items/blog_posts/:id`
4. ✅ **Delete records**: `DELETE /items/blog_posts/:id`

**Your Directus-style CMS is now ready for dynamic content management!** 🚀