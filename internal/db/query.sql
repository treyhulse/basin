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