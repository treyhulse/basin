-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: GetPermissionsByRole :many
SELECT * FROM permissions WHERE role_id = $1;

-- name: GetPermissionsByRoleAndTable :many
SELECT * FROM permissions WHERE role_id = $1 AND table_name = $2;

-- name: GetPermissionsByRoleAndAction :many
SELECT * FROM permissions WHERE role_id = $1 AND table_name = $2 AND action = $3;

-- name: GetProducts :many
SELECT * FROM products;

-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, description, price, category, stock_quantity) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateProduct :one
UPDATE products SET name = $2, description = $3, price = $4, category = $5, stock_quantity = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: GetCustomers :many
SELECT * FROM customers;

-- name: GetCustomer :one
SELECT * FROM customers WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (first_name, last_name, email, phone, address) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateCustomer :one
UPDATE customers SET first_name = $2, last_name = $3, email = $4, phone = $5, address = $6, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers WHERE id = $1;

-- name: GetOrders :many
SELECT * FROM orders;

-- name: GetOrder :one
SELECT * FROM orders WHERE id = $1;

-- name: CreateOrder :one
INSERT INTO orders (customer_id, total_amount, notes) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateOrder :one
UPDATE orders SET status = $2, notes = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1;

-- name: GetOrderItems :many
SELECT * FROM order_items;

-- name: GetOrderItem :one
SELECT * FROM order_items WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateOrderItem :one
UPDATE order_items SET order_id = $2, product_id = $3, quantity = $4, unit_price = $5 WHERE id = $1 RETURNING *;

-- name: DeleteOrderItem :exec
DELETE FROM order_items WHERE id = $1;

-- API Key Management Queries
-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND is_active = true;

-- name: GetAPIKeysByUser :many
SELECT * FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, name, key_hash, expires_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: UpdateAPIKey :one
UPDATE api_keys SET name = $2, is_active = $3, expires_at = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys WHERE id = $1;

-- name: GetAPIKeyByID :one
SELECT * FROM api_keys WHERE id = $1;

-- Schema Management Queries
-- name: GetCollections :many
SELECT * FROM collections ORDER BY name;

-- name: GetCollection :one
SELECT * FROM collections WHERE id = $1;

-- name: CreateCollection :one
INSERT INTO collections (id, name, display_name, description, icon, is_system, tenant_id, created_by) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;

-- name: UpdateCollection :one
UPDATE collections 
SET display_name = $2, description = $3, icon = $4, updated_at = CURRENT_TIMESTAMP, updated_by = $5
WHERE id = $1 RETURNING *;

-- name: DeleteCollection :exec
DELETE FROM collections WHERE id = $1;

-- name: GetFields :many
SELECT * FROM fields ORDER BY sort_order;

-- name: GetFieldsByCollection :many
SELECT * FROM fields WHERE collection_id = $1 ORDER BY sort_order;

-- name: GetField :one
SELECT * FROM fields WHERE id = $1;

-- name: CreateField :one
INSERT INTO fields (id, collection_id, name, display_name, type, is_primary, is_required, is_unique, default_value, validation_rules, relation_config, sort_order, tenant_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING *;

-- name: UpdateField :one
UPDATE fields 
SET display_name = $2, type = $3, is_primary = $4, is_required = $5, is_unique = $6, default_value = $7, validation_rules = $8, relation_config = $9, sort_order = $10, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields WHERE id = $1;

-- Tenant Management Queries
-- name: GetTenants :many
SELECT * FROM tenants ORDER BY name;

-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: CreateTenant :one
INSERT INTO tenants (id, name, slug, domain, settings) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants SET name = $2, slug = $3, domain = $4, settings = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: DeleteTenant :exec
DELETE FROM tenants WHERE id = $1;

-- Enhanced User Queries with Tenant Support
-- name: GetUsersByTenant :many
SELECT * FROM users WHERE tenant_id = $1 ORDER BY email;

-- name: GetUserWithTenant :one
SELECT u.*, t.name as tenant_name, t.slug as tenant_slug 
FROM users u 
JOIN tenants t ON u.tenant_id = t.id 
WHERE u.id = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, first_name, last_name, tenant_id) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: UpdateUser :one
UPDATE users 
SET email = $2, first_name = $3, last_name = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- Enhanced Permission Queries with Tenant Support
-- name: GetPermissionsByRoleAndTenant :many
SELECT * FROM permissions WHERE role_id = $1 AND tenant_id = $2;

-- name: GetPermissionsByUserAndTenant :many
SELECT p.* FROM permissions p
JOIN user_roles ur ON p.role_id = ur.role_id
WHERE ur.user_id = $1 AND p.tenant_id = $2;

-- name: CreatePermission :one
INSERT INTO permissions (id, role_id, table_name, action, field_filter, allowed_fields, tenant_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: UpdatePermission :one
UPDATE permissions 
SET field_filter = $2, allowed_fields = $3, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1 RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions WHERE id = $1; 