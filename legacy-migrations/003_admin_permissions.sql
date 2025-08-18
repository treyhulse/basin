-- Migration: Add all missing permissions for admin role
-- This ensures admin has full access to all tables

-- Add missing permissions for order_items
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'order_items', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'order_items' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'order_items', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'order_items' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'order_items', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'order_items' 
    AND action = 'delete'
);

-- Add missing permissions for products (if any)
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'products', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'products' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'products', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'products' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'products', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'products' 
    AND action = 'delete'
);

-- Add missing permissions for customers (if any)
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'customers', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'customers' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'customers', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'customers' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'customers', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'customers' 
    AND action = 'delete'
);

-- Add missing permissions for orders (if any)
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'orders', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'orders' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'orders', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'orders' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'orders', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'orders' 
    AND action = 'delete'
);

-- Add missing permissions for system tables
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'users', 'read', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'users' 
    AND action = 'read'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'users', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'users' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'users', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'users' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'users', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'users' 
    AND action = 'delete'
);

-- Add missing permissions for roles table
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'roles', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'roles' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'roles', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'roles' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'roles', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'roles' 
    AND action = 'delete'
);

-- Add missing permissions for user_roles table
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'user_roles', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'user_roles' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'user_roles', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'user_roles' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'user_roles', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'user_roles' 
    AND action = 'delete'
);

-- Add missing permissions for permissions table
INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'create', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'permissions' 
    AND action = 'create'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'update', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'permissions' 
    AND action = 'update'
);

INSERT INTO permissions (role_id, table_name, action, allowed_fields) 
SELECT (SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'delete', ARRAY['*']
WHERE NOT EXISTS (
    SELECT 1 FROM permissions 
    WHERE role_id = (SELECT id FROM roles WHERE name = 'admin') 
    AND table_name = 'permissions' 
    AND action = 'delete'
); 