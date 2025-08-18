-- Migration 006: Add Admin Permissions for System Tables
-- This migration adds missing admin permissions for users, roles, and permissions tables

-- Add admin permissions for users table
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage users
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING;

-- Add admin permissions for roles table
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage roles
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING;

-- Add admin permissions for permissions table
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage permissions
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING;

-- Add admin permissions for tenants table (in case we want to manage tenants through API)
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage tenants
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING;

-- Add admin permissions for API keys table
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage API keys
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING;

-- Add permissions for regular users to manage their own API keys
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id, field_filter) VALUES
    -- Manager can manage their own API keys
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'create', ARRAY['name', 'expires_at'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'read', ARRAY['id', 'name', 'is_active', 'expires_at', 'last_used_at', 'created_at'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'update', ARRAY['name', 'is_active'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    
    -- Sales can manage their own API keys
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'create', ARRAY['name', 'expires_at'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'read', ARRAY['id', 'name', 'is_active', 'expires_at', 'last_used_at', 'created_at'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'update', ARRAY['name', 'is_active'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}'),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'), '{"user_id": "current_user_id"}')
ON CONFLICT (role_id, table_name, action) DO NOTHING;