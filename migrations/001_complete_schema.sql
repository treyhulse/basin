-- Complete Basin API Schema Migration
-- This migration creates the complete database schema with multi-schema architecture
-- Core tables in public schema, data tables in 'data' schema with tenant-specific naming
-- Data tables named: collectionName-data-tenantId (e.g., customers-data-tenant123)

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create data schema for tenant-specific data tables
CREATE SCHEMA IF NOT EXISTS data;

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

-- Create user_tenants junction table for many-to-many user-tenant relationships
CREATE TABLE user_tenants (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, tenant_id)
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

-- Create collections table with tenant isolation and data table reference
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL, -- Display name (e.g., "Customers", "Products")
    slug VARCHAR(255) NOT NULL, -- URL-friendly name (e.g., "customers", "products")
    data_table_name VARCHAR(255) NOT NULL, -- Actual table name in data schema (e.g., "customers-data-tenant123")
    display_name VARCHAR(255),
    description TEXT,
    icon VARCHAR(100),
    is_system BOOLEAN DEFAULT false,
    tenant_id UUID REFERENCES tenants(id),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Ensure collection names are unique within each tenant
    UNIQUE(tenant_id, slug),
    -- Ensure data table names are globally unique
    UNIQUE(data_table_name)
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
    relation_config JSONB, -- {"related_collection": "collection_name", "field": "field_name"}
    sort_order INTEGER DEFAULT 0,
    tenant_id UUID REFERENCES tenants(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create API keys table
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    permissions JSONB NOT NULL, -- {"tables": ["table1", "table2"], "actions": ["read", "write"]}
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    tenant_id UUID REFERENCES tenants(id),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);
CREATE INDEX idx_collections_tenant_id ON collections(tenant_id);
CREATE INDEX idx_collections_slug ON collections(tenant_id, slug);
CREATE INDEX idx_fields_tenant_id ON fields(tenant_id);
CREATE INDEX idx_fields_collection_id ON fields(collection_id);
CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);

-- Create function to generate unique data table name
CREATE OR REPLACE FUNCTION generate_data_table_name(p_collection_slug TEXT, p_tenant_id UUID)
RETURNS TEXT AS $$
BEGIN
    -- Format: collectionSlug-data-tenantId (e.g., customers-data-6e68062f-c4c6-42df-9e01-e2d1081664f4)
    RETURN p_collection_slug || '-data-' || p_tenant_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to create data tables in data schema with tenant-specific naming
CREATE OR REPLACE FUNCTION create_data_table(p_collection_id UUID, p_collection_slug TEXT, p_tenant_id UUID)
RETURNS TEXT AS $$
DECLARE
    create_table_sql TEXT;
    field_record RECORD;
    has_fields BOOLEAN;
    data_table_name TEXT;
BEGIN
    -- Generate unique data table name
    data_table_name := generate_data_table_name(p_collection_slug, p_tenant_id);
    
    -- Check if there are any fields for this collection
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id) INTO has_fields;
    
    -- Start building the CREATE TABLE statement in data schema
    create_table_sql := 'CREATE TABLE IF NOT EXISTS data.' || data_table_name || ' (';
    create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4()';
    
    -- Add standard fields that every collection should have
    create_table_sql := create_table_sql || ', created_by UUID';
    create_table_sql := create_table_sql || ', updated_by UUID';
    create_table_sql := create_table_sql || ', created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    create_table_sql := create_table_sql || ', updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    create_table_sql := create_table_sql || ', tenant_id UUID DEFAULT ''' || p_tenant_id || '''';
    
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
                WHEN 'decimal', 'float' THEN
                    create_table_sql := create_table_sql || 'DECIMAL(10,2)';
                WHEN 'boolean', 'bool' THEN
                    create_table_sql := create_table_sql || 'BOOLEAN';
                WHEN 'datetime', 'timestamp' THEN
                    create_table_sql := create_table_sql || 'TIMESTAMP WITH TIME ZONE';
                WHEN 'json' THEN
                    create_table_sql := create_table_sql || 'JSONB';
                WHEN 'uuid' THEN
                    create_table_sql := create_table_sql || 'UUID';
                WHEN 'relation' THEN
                    -- Handle relation fields - reference the related table
                    IF field_record.relation_config IS NOT NULL THEN
                        create_table_sql := create_table_sql || 'UUID';
                    ELSE
                        create_table_sql := create_table_sql || 'UUID';
                    END IF;
                ELSE
                    create_table_sql := create_table_sql || 'TEXT';
            END CASE;
            
            -- Add constraints
            IF field_record.is_required THEN
                create_table_sql := create_table_sql || ' NOT NULL';
            END IF;
            
            IF field_record.is_unique THEN
                create_table_sql := create_table_sql || ' UNIQUE';
            END IF;
            
            IF field_record.default_value IS NOT NULL AND field_record.default_value != '' THEN
                create_table_sql := create_table_sql || ' DEFAULT ''' || field_record.default_value || '''';
            END IF;
        END LOOP;
    END IF;
    
    -- Close the CREATE TABLE statement
    create_table_sql := create_table_sql || ')';
    
    -- Execute the CREATE TABLE statement
    EXECUTE create_table_sql;
    
    -- Enable Row Level Security for tenant isolation
    EXECUTE 'ALTER TABLE data.' || data_table_name || ' ENABLE ROW LEVEL SECURITY';
    
    -- Create RLS policy for tenant isolation
    EXECUTE 'CREATE POLICY tenant_isolation_policy ON data.' || data_table_name || 
            ' FOR ALL TO PUBLIC USING (tenant_id = current_setting(''app.current_tenant_id'', true)::uuid)';
    
    -- Create indexes for better performance
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_' || data_table_name || '_tenant_id ON data.' || data_table_name || ' (tenant_id)';
    EXECUTE 'CREATE INDEX IF NOT EXISTS idx_' || data_table_name || '_created_at ON data.' || data_table_name || ' (created_at)';
    
    RAISE NOTICE 'Created data table: data.% with tenant_id isolation and RLS', data_table_name;
    
    -- Return the data table name
    RETURN data_table_name;
END;
$$ LANGUAGE plpgsql;

-- Create function to drop data tables
CREATE OR REPLACE FUNCTION drop_data_table(p_data_table_name TEXT)
RETURNS VOID AS $$
BEGIN
    -- Drop the data table from data schema
    EXECUTE 'DROP TABLE IF EXISTS data.' || p_data_table_name || ' CASCADE';
    RAISE NOTICE 'Dropped data table: data.%', p_data_table_name;
END;
$$ LANGUAGE plpgsql;

-- Create function to set user context for RLS
CREATE OR REPLACE FUNCTION set_user_context(p_user_id UUID, p_tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    -- Set the tenant_id in the session for RLS policies
    PERFORM set_config('app.current_tenant_id', p_tenant_id::text, true);
    PERFORM set_config('app.current_user_id', p_user_id::text, true);
END;
$$ LANGUAGE plpgsql;

-- Create trigger function to automatically create data tables when collections are created
CREATE OR REPLACE FUNCTION create_data_table_triggers()
RETURNS VOID AS $$
BEGIN
    -- Drop existing trigger if it exists
    DROP TRIGGER IF EXISTS trigger_create_data_table ON collections;
    
    -- Create new trigger for new collections
    CREATE TRIGGER trigger_create_data_table
        AFTER INSERT ON collections
        FOR EACH ROW
        EXECUTE FUNCTION create_collection_data_table();
    
    RAISE NOTICE 'Created data table trigger for multi-schema architecture';
END;
$$ LANGUAGE plpgsql;

-- Create trigger function that generates data table name and creates the table
CREATE OR REPLACE FUNCTION create_collection_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
BEGIN
    -- Generate the data table name
    data_table_name := generate_data_table_name(NEW.slug, NEW.tenant_id);
    
    -- Update the collection record with the data table name
    UPDATE collections 
    SET data_table_name = data_table_name 
    WHERE id = NEW.id;
    
    -- Create the actual data table
    PERFORM create_data_table(NEW.id, NEW.slug, NEW.tenant_id);
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger
SELECT create_data_table_triggers();

-- Insert default tenant
INSERT INTO tenants (id, name, slug, domain, is_active) 
VALUES ('6e68062f-c4c6-42df-9e01-e2d1081664f4', 'Main Tenant', 'main', 'localhost', true);

-- Insert default admin user
INSERT INTO users (id, email, password_hash, first_name, last_name, tenant_id, is_active)
VALUES ('38eae290-37b8-46a7-82ee-ae842d85c894', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', '6e68062f-c4c6-42df-9e01-e2d1081664f4', true);

-- Insert default roles for main tenant
INSERT INTO roles (id, name, description, tenant_id)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'admin', 'Administrator with full access', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('550e8400-e29b-41d4-a716-446655440002', 'user', 'Regular user with limited access', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('550e8400-e29b-41d4-a716-446655440003', 'viewer', 'Read-only access', '6e68062f-c4c6-42df-9e01-e2d1081664f4');

-- Add admin user to main tenant with admin role
INSERT INTO user_tenants (user_id, tenant_id, role_id, is_active)
VALUES ('38eae290-37b8-46a7-82ee-ae842d85c894', '6e68062f-c4c6-42df-9e01-e2d1081664f4', '550e8400-e29b-41d4-a716-446655440001', true);

-- Add admin role to user_roles table
INSERT INTO user_roles (user_id, role_id)
VALUES ('38eae290-37b8-46a7-82ee-ae842d85c894', '550e8400-e29b-41d4-a716-446655440001');

-- Insert default permissions for system tables
INSERT INTO permissions (id, role_id, table_name, action, tenant_id)
VALUES 
    -- Admin permissions for all system tables
    ('650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'users', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'users', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'users', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'users', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', 'roles', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', 'roles', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440001', 'roles', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440001', 'roles', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440001', 'permissions', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440001', 'permissions', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440001', 'permissions', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440001', 'permissions', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440013', '550e8400-e29b-41d4-a716-446655440001', 'collections', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440014', '550e8400-e29b-41d4-a716-446655440001', 'collections', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440015', '550e8400-e29b-41d4-a716-446655440001', 'collections', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440016', '550e8400-e29b-41d4-a716-446655440001', 'collections', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440017', '550e8400-e29b-41d4-a716-446655440001', 'fields', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440018', '550e8400-e29b-41d4-a716-446655440001', 'fields', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440019', '550e8400-e29b-41d4-a716-446655440001', 'fields', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440020', '550e8400-e29b-41d4-a716-446655440001', 'fields', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440021', '550e8400-e29b-41d4-a716-446655440001', 'api_keys', 'create', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440022', '550e8400-e29b-41d4-a716-446655440001', 'api_keys', 'read', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440023', '550e8400-e29b-41d4-a716-446655440001', 'api_keys', 'update', '6e68062f-c4c6-42df-9e01-e2d1081664f4'),
    ('650e8400-e29b-41d4-a716-446655440024', '550e8400-e29b-41d4-a716-446655440001', 'api_keys', 'delete', '6e68062f-c4c6-42df-9e01-e2d1081664f4');

-- Add comments
COMMENT ON TABLE tenants IS 'Tenant information and settings';
COMMENT ON TABLE users IS 'User accounts with tenant association';
COMMENT ON TABLE roles IS 'Roles within each tenant';
COMMENT ON TABLE permissions IS 'Permissions for roles on specific tables';
COMMENT ON TABLE collections IS 'Dynamic collections with tenant isolation and data table references';
COMMENT ON TABLE fields IS 'Field definitions for collections';
COMMENT ON TABLE api_keys IS 'API keys for programmatic access';

COMMENT ON FUNCTION generate_data_table_name(TEXT, UUID) IS 'Generates unique data table names: collectionSlug-data-tenantId';
COMMENT ON FUNCTION create_data_table(UUID, TEXT, UUID) IS 'Creates data tables in data schema with tenant-specific naming and RLS';
COMMENT ON FUNCTION drop_data_table(TEXT) IS 'Drops data tables from data schema';
COMMENT ON FUNCTION set_user_context(UUID, UUID) IS 'Sets user and tenant context for RLS policies';