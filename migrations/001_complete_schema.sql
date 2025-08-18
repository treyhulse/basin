-- Complete Basin API Schema Migration
-- This single migration file creates the complete database schema
-- Includes: users, roles, permissions, collections, fields, tenants, and API keys

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tenant schemas for data tables
CREATE SCHEMA IF NOT EXISTS main;

-- Create tenants table first (referenced by other tables)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255),
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    tenant_id UUID REFERENCES tenants(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id UUID REFERENCES tenants(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user_roles junction table
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Create permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete'
    field_filter JSONB, -- {"field": "value"} for row-level filtering
    allowed_fields TEXT[], -- array of allowed fields for field-level access
    tenant_id UUID REFERENCES tenants(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, table_name, action)
);

-- Create collections table
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    icon VARCHAR(100),
    is_system BOOLEAN DEFAULT false,
    tenant_id UUID REFERENCES tenants(id),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create fields table
CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID REFERENCES collections(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    type VARCHAR(50) NOT NULL, -- 'string', 'text', 'integer', 'decimal', 'boolean', 'datetime', 'json', 'uuid', 'relation'
    is_primary BOOLEAN DEFAULT false,
    is_required BOOLEAN DEFAULT false,
    is_unique BOOLEAN DEFAULT false,
    default_value TEXT,
    validation_rules JSONB,
    sort_order INTEGER DEFAULT 0,
    relation_config JSONB, -- {"related_collection": "table_name", "relation_type": "one_to_many"}
    tenant_id UUID REFERENCES tenants(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(collection_id, name)
);

-- Create API keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create sample tables for demonstration
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);
CREATE INDEX idx_collections_tenant_id ON collections(tenant_id);
CREATE INDEX idx_fields_tenant_id ON fields(tenant_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_roles_name_tenant_unique ON roles(name, tenant_id);

-- Create default tenant
INSERT INTO tenants (id, name, slug, domain) VALUES 
    (uuid_generate_v4(), 'Main Tenant', 'main', 'localhost');

-- Insert default roles
INSERT INTO roles (name, description, tenant_id) VALUES
    ('admin', 'Full system access', (SELECT id FROM tenants WHERE slug = 'main')),
    ('manager', 'Can manage products and view orders', (SELECT id FROM tenants WHERE slug = 'main')),
    ('sales', 'Can view products and create orders', (SELECT id FROM tenants WHERE slug = 'main')),
    ('customer', 'Can view products and own orders', (SELECT id FROM tenants WHERE slug = 'main'));

-- Insert default users (password: password)
INSERT INTO users (email, password_hash, first_name, last_name, tenant_id) VALUES
    ('admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', (SELECT id FROM tenants WHERE slug = 'main')),
    ('manager@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Manager', 'User', (SELECT id FROM tenants WHERE slug = 'main'));

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM users u, roles r 
WHERE u.email = 'admin@example.com' AND r.name = 'admin';

INSERT INTO user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM users u, roles r 
WHERE u.email = 'manager@example.com' AND r.name = 'manager';

-- Insert sample permissions for admin role
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin permissions (all tables, all actions, all fields)
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    -- Admin can manage system tables
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'users', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'roles', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'permissions', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'tenants', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main'));

-- Insert manager permissions
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    
    -- Manager can manage their own API keys
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'create', ARRAY['name', 'expires_at'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'read', ARRAY['id', 'name', 'is_active', 'expires_at', 'last_used_at', 'created_at'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'update', ARRAY['name', 'is_active'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main'));

-- Insert sales permissions
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    
    -- Sales can manage their own API keys
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'create', ARRAY['name', 'expires_at'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'read', ARRAY['id', 'name', 'is_active', 'expires_at', 'last_used_at', 'created_at'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'update', ARRAY['name', 'is_active'], (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM roles WHERE name = 'sales'), 'api_keys', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'main'));

-- Insert customer permissions
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    ((SELECT id FROM roles WHERE name = 'customer'), 'customers', 'read', ARRAY['id', 'first_name', 'last_name', 'email'], (SELECT id FROM tenants WHERE slug = 'main'));

-- Insert sample data
INSERT INTO customers (first_name, last_name, email, phone, address) VALUES
    ('John', 'Doe', 'john.doe@example.com', '+1234567890', '123 Main St, City, State'),
    ('Jane', 'Smith', 'jane.smith@example.com', '+0987654321', '456 Oak Ave, Town, State');

-- Insert sample API keys for testing
INSERT INTO api_keys (user_id, name, key_hash, expires_at) VALUES
    ((SELECT id FROM users WHERE email = 'admin@example.com'), 'Admin API Key', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NULL),
    ((SELECT id FROM users WHERE email = 'manager@example.com'), 'Manager API Key', '$2a$10$TKh8H1.PfQx37YgCzwiKb.KjNyWgaHb9cbcoQgdIVFlYg7B77UdFm', NULL);

-- Create sample collections for demonstration
INSERT INTO collections (name, display_name, description, icon, is_system, tenant_id, created_by) VALUES
    ('blog_posts', 'Blog Posts', 'Blog post articles with title, content, and author', '📝', false, (SELECT id FROM tenants WHERE slug = 'main'), (SELECT id FROM users WHERE email = 'admin@example.com')),
    ('customers', 'Customers', 'Customer information with contact details', '👥', false, (SELECT id FROM tenants WHERE slug = 'main'), (SELECT id FROM users WHERE email = 'admin@example.com'));

-- Create sample fields for blog_posts collection
INSERT INTO fields (collection_id, name, display_name, type, is_primary, is_required, is_unique, sort_order, tenant_id) VALUES
    ((SELECT id FROM collections WHERE name = 'blog_posts'), 'title', 'Title', 'text', true, true, false, 1, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'blog_posts'), 'content', 'Content', 'text', false, true, false, 2, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'blog_posts'), 'author', 'Author', 'text', false, true, false, 3, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'blog_posts'), 'published_at', 'Published At', 'datetime', false, false, false, 4, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'blog_posts'), 'tags', 'Tags', 'json', false, false, false, 5, (SELECT id FROM tenants WHERE slug = 'main'));

-- Create sample fields for customers collection
INSERT INTO fields (collection_id, name, display_name, type, is_primary, is_required, is_unique, sort_order, tenant_id) VALUES
    ((SELECT id FROM collections WHERE name = 'customers'), 'first_name', 'First Name', 'text', false, true, false, 1, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'customers'), 'last_name', 'Last Name', 'text', false, true, false, 2, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'customers'), 'email', 'Email', 'text', true, true, true, 3, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'customers'), 'phone', 'Phone', 'text', false, false, false, 4, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'customers'), 'address', 'Address', 'text', false, false, false, 5, (SELECT id FROM tenants WHERE slug = 'main')),
    ((SELECT id FROM collections WHERE name = 'customers'), 'date_of_birth', 'Date of Birth', 'datetime', false, false, false, 6, (SELECT id FROM tenants WHERE slug = 'main'));

-- Create trigger function to update updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_collections_updated_at BEFORE UPDATE ON collections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_fields_updated_at BEFORE UPDATE ON fields FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to set user context for RLS (Row Level Security)
CREATE OR REPLACE FUNCTION set_user_context(user_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_user_id', user_id::text, false);
END;
$$ LANGUAGE plpgsql;

-- Create function to create data tables for collections
CREATE OR REPLACE FUNCTION create_data_table(p_collection_id UUID, p_table_name TEXT)
RETURNS VOID AS $$
DECLARE
    create_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
    has_fields BOOLEAN;
BEGIN
    -- Get the tenant schema for this collection
    SELECT t.slug INTO tenant_schema 
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = p_collection_id;
    
    -- Check if there are any fields for this collection
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id) INTO has_fields;
    
    -- Start building the CREATE TABLE statement
    create_table_sql := 'CREATE TABLE IF NOT EXISTS "' || tenant_schema || '".data_' || p_table_name || ' (';
    create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4()';
    
    -- Add fields from the fields table (if any exist)
    IF has_fields THEN
        FOR field_record IN 
            SELECT f.name, f.type, f.is_required, f.is_unique, f.default_value, f.relation_config
            FROM fields f
            WHERE f.collection_id = p_collection_id 
            ORDER BY f.sort_order, f.name
        LOOP
            create_table_sql := create_table_sql || ',';
            
            -- Add column name
            create_table_sql := create_table_sql || '"' || field_record.name || '" ';
            
            -- Map field types to PostgreSQL types
            CASE field_record.type
                WHEN 'string', 'text' THEN
                    create_table_sql := create_table_sql || 'TEXT';
                WHEN 'integer', 'int' THEN
                    create_table_sql := create_table_sql || 'INTEGER';
                WHEN 'float', 'decimal' THEN
                    create_table_sql := create_table_sql || 'DECIMAL';
                WHEN 'boolean', 'bool' THEN
                    create_table_sql := create_table_sql || 'BOOLEAN';
                WHEN 'json', 'object' THEN
                    create_table_sql := create_table_sql || 'JSONB';
                WHEN 'date', 'datetime' THEN
                    create_table_sql := create_table_sql || 'TIMESTAMP WITH TIME ZONE';
                WHEN 'uuid' THEN
                    create_table_sql := create_table_sql || 'UUID';
                ELSE
                    create_table_sql := create_table_sql || 'TEXT';
            END CASE;
            
            -- Add NOT NULL constraint for required fields
            IF field_record.is_required THEN
                create_table_sql := create_table_sql || ' NOT NULL';
            END IF;
            
            -- Add default value if specified
            IF field_record.default_value IS NOT NULL AND field_record.default_value != '' THEN
                create_table_sql := create_table_sql || ' DEFAULT ' || field_record.default_value;
            END IF;
        END LOOP;
    END IF;
    
    -- Close the statement
    create_table_sql := create_table_sql || ')';
    
    -- Execute the CREATE TABLE statement
    EXECUTE create_table_sql;
END;
$$ LANGUAGE plpgsql;

-- Create function to drop data tables for collections
CREATE OR REPLACE FUNCTION drop_data_table(collection_id UUID, table_name TEXT)
RETURNS VOID AS $$
DECLARE
    tenant_schema TEXT;
BEGIN
    -- Get the tenant schema for this collection
    SELECT t.slug INTO tenant_schema 
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = collection_id;
    
    -- Drop the table if it exists
    EXECUTE 'DROP TABLE IF EXISTS "' || tenant_schema || '".data_' || table_name;
END;
$$ LANGUAGE plpgsql;

-- Create function to ensure standard fields for collections
CREATE OR REPLACE FUNCTION ensure_standard_collection_fields(p_collection_id UUID, p_tenant_id UUID)
RETURNS VOID AS $$
DECLARE
    field_exists BOOLEAN;
BEGIN
    -- Check and add 'name' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'name') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, is_unique, sort_order, tenant_id)
        VALUES (p_collection_id, 'name', 'Name', 'text', true, false, 1, p_tenant_id);
    END IF;

    -- Check and add 'display_name' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'display_name') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, is_unique, sort_order, tenant_id)
        VALUES (p_collection_id, 'display_name', 'Display Name', 'text', false, false, 2, p_tenant_id);
    END IF;

    -- Check and add 'created_at' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'created_at') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, sort_order, tenant_id)
        VALUES (p_collection_id, 'created_at', 'Created At', 'datetime', false, 100, p_tenant_id);
    END IF;

    -- Check and add 'created_by' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'created_by') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, sort_order, tenant_id)
        VALUES (p_collection_id, 'created_by', 'Created By', 'uuid', false, 101, p_tenant_id);
    END IF;

    -- Check and add 'updated_at' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'updated_at') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, sort_order, tenant_id)
        VALUES (p_collection_id, 'updated_at', 'Updated At', 'datetime', false, 102, p_tenant_id);
    END IF;

    -- Check and add 'updated_by' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'updated_by') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, sort_order, tenant_id)
        VALUES (p_collection_id, 'updated_by', 'Updated By', 'uuid', false, 103, p_tenant_id);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to add standard fields to new collections
CREATE OR REPLACE FUNCTION add_standard_fields_to_new_collection()
RETURNS TRIGGER AS $$
BEGIN
    -- Only add standard fields to non-system collections
    IF NEW.is_system = false THEN
        -- First add standard fields to the fields table
        PERFORM ensure_standard_collection_fields(NEW.id, NEW.tenant_id);
        -- Then create the data table (which will now include the standard fields)
        PERFORM create_data_table(NEW.id, NEW.name);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create function to add permissions for new collections
CREATE OR REPLACE FUNCTION admin_collection_permissions_trigger()
RETURNS TRIGGER AS $$
DECLARE
    admin_role_id UUID;
    creator_role_id UUID;
BEGIN
    -- Get admin role ID for the tenant
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = NEW.tenant_id;
    
    -- Get the creator's role (assuming they have admin role for now)
    SELECT ur.role_id INTO creator_role_id 
    FROM user_roles ur 
    WHERE ur.user_id = NEW.created_by 
    LIMIT 1;
    
    -- Add admin permissions for the new collection
    IF admin_role_id IS NOT NULL THEN
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, NEW.name, 'create', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, NEW.name, 'read', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, NEW.name, 'update', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id)
        VALUES 
            (admin_role_id, NEW.name, 'delete', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
    END IF;
    
    -- Add creator permissions (full access to their own collection)
    IF creator_role_id IS NOT NULL THEN
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (creator_role_id, NEW.name, 'create', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (creator_role_id, NEW.name, 'read', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (creator_role_id, NEW.name, 'update', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (creator_role_id, NEW.name, 'delete', ARRAY['*'], NEW.tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for collections
CREATE TRIGGER trigger_add_standard_fields 
    AFTER INSERT ON collections 
    FOR EACH ROW 
    EXECUTE FUNCTION add_standard_fields_to_new_collection();

CREATE TRIGGER trigger_add_admin_collection_permissions 
    AFTER INSERT ON collections 
    FOR EACH ROW 
    EXECUTE FUNCTION admin_collection_permissions_trigger();

-- Create function to ensure admin permissions for existing collections
CREATE OR REPLACE FUNCTION ensure_admin_collection_permissions()
RETURNS VOID AS $$
DECLARE
    collection_record RECORD;
    admin_role_id UUID;
    main_tenant_id UUID;
BEGIN
    -- Get admin role ID and main tenant ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = (SELECT id FROM tenants WHERE slug = 'main');
    SELECT id INTO main_tenant_id FROM tenants WHERE slug = 'main';
    
    -- Add permissions for all existing collections
    FOR collection_record IN 
        SELECT name FROM collections 
        WHERE is_system = false AND tenant_id = main_tenant_id
    LOOP
        -- Add permissions for this collection if they don't exist
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'create', ARRAY['*'], main_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'read', ARRAY['*'], main_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'update', ARRAY['*'], main_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'delete', ARRAY['*'], main_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Run the admin collection permissions function to update existing permissions
SELECT ensure_admin_collection_permissions();

-- Create function to ensure admin has access to ALL collections (including future ones)
CREATE OR REPLACE FUNCTION ensure_admin_unlimited_access()
RETURNS VOID AS $$
DECLARE
    admin_role_id UUID;
    main_tenant_id UUID;
    collection_record RECORD;
BEGIN
    -- Get admin role ID and main tenant ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = (SELECT id FROM tenants WHERE slug = 'main');
    SELECT id INTO main_tenant_id FROM tenants WHERE slug = 'main';
    
    -- Ensure admin has access to all system tables
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'tenants', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'tenants', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'tenants', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'tenants', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'users', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'users', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'users', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'users', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'roles', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'roles', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'roles', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'roles', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'permissions', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'permissions', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'permissions', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'permissions', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'collections', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'collections', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'collections', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'collections', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'fields', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'fields', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'fields', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'fields', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'api_keys', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'api_keys', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'api_keys', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'api_keys', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
    VALUES 
        (admin_role_id, 'customers', 'create', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'customers', 'read', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'customers', 'update', ARRAY['*'], main_tenant_id),
        (admin_role_id, 'customers', 'delete', ARRAY['*'], main_tenant_id)
    ON CONFLICT (role_id, table_name, action) DO NOTHING;
    
    -- Ensure admin has access to all existing collections
    FOR collection_record IN 
        SELECT name FROM collections 
        WHERE is_system = false AND tenant_id = main_tenant_id
    LOOP
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'create', ARRAY['*'], main_tenant_id),
            (admin_role_id, collection_record.name, 'read', ARRAY['*'], main_tenant_id),
            (admin_role_id, collection_record.name, 'update', ARRAY['*'], main_tenant_id),
            (admin_role_id, collection_record.name, 'delete', ARRAY['*'], main_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Run the admin unlimited access function
SELECT ensure_admin_unlimited_access();

-- Manually create data tables for initial collections (since triggers don't run on bulk inserts)
SELECT create_data_table(id, name) FROM collections WHERE name IN ('blog_posts', 'customers');

-- Add comments for documentation
COMMENT ON TABLE tenants IS 'Multi-tenant support - each tenant has isolated data';
COMMENT ON TABLE users IS 'User accounts with tenant isolation';
COMMENT ON TABLE roles IS 'Role definitions with tenant isolation';
COMMENT ON TABLE permissions IS 'Role-based permissions for table access';
COMMENT ON TABLE collections IS 'Dynamic collections that can be created by users';
COMMENT ON TABLE fields IS 'Field definitions for dynamic collections';
COMMENT ON TABLE api_keys IS 'API keys for programmatic access';
COMMENT ON FUNCTION set_user_context(UUID) IS 'Sets the current user context for RLS policies';
COMMENT ON FUNCTION create_data_table(UUID, TEXT) IS 'Creates data tables for collections';
COMMENT ON FUNCTION drop_data_table(UUID, TEXT) IS 'Drops data tables for collections';
