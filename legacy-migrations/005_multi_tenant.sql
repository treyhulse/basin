-- Migration 005: Multi-Tenant and Enhanced User Isolation
-- This migration adds tenant support and enhanced user isolation

-- Create tenants table
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

-- Add tenant_id to existing tables
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE collections ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE fields ADD COLUMN tenant_id UUID REFERENCES tenants(id);

-- Create indexes for tenant isolation
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_collections_tenant_id ON collections(tenant_id);
CREATE INDEX idx_fields_tenant_id ON fields(tenant_id);

-- Add user_id to collections for ownership
ALTER TABLE collections ADD COLUMN created_by UUID REFERENCES users(id);
ALTER TABLE collections ADD COLUMN updated_by UUID REFERENCES users(id);

-- Enhanced permissions with tenant support
ALTER TABLE permissions ADD COLUMN tenant_id UUID REFERENCES tenants(id);
CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);

-- Create default tenant
INSERT INTO tenants (id, name, slug, domain) VALUES 
    (uuid_generate_v4(), 'Default Tenant', 'default', 'localhost');

-- Update existing records to belong to default tenant
UPDATE users SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE collections SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE fields SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE permissions SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default');

-- Add ownership tracking to data tables
-- This will be applied to new data tables via triggers

-- Enhanced trigger function for data table creation
CREATE OR REPLACE FUNCTION create_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
    create_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
BEGIN
    -- Only create data tables for non-system collections
    IF NEW.is_system = false THEN
        -- Get tenant schema name
        SELECT slug INTO tenant_schema FROM tenants WHERE id = NEW.tenant_id;
        
        -- Create tenant schema if it doesn't exist
        EXECUTE 'CREATE SCHEMA IF NOT EXISTS ' || tenant_schema;
        
        data_table_name := tenant_schema || '.data_' || NEW.name;
        
        -- Start building the CREATE TABLE statement
        create_table_sql := 'CREATE TABLE ' || data_table_name || ' (';
        create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),';
        create_table_sql := create_table_sql || 'created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
        create_table_sql := create_table_sql || 'updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
        create_table_sql := create_table_sql || 'created_by UUID REFERENCES public.users(id),';
        create_table_sql := create_table_sql || 'updated_by UUID REFERENCES public.users(id)';
        
        -- Add fields for this collection
        FOR field_record IN 
            SELECT name, type, is_required, is_unique, default_value, relation_config
            FROM fields 
            WHERE collection_id = NEW.id 
            ORDER BY sort_order
        LOOP
            create_table_sql := create_table_sql || ',';
            
            -- Add column name
            create_table_sql := create_table_sql || '"' || field_record.name || '" ';
            
            -- Add data type
            CASE field_record.type
                WHEN 'string' THEN create_table_sql := create_table_sql || 'VARCHAR(255)';
                WHEN 'text' THEN create_table_sql := create_table_sql || 'TEXT';
                WHEN 'integer' THEN create_table_sql := create_table_sql || 'INTEGER';
                WHEN 'decimal' THEN create_table_sql := create_table_sql || 'DECIMAL(10,2)';
                WHEN 'boolean' THEN create_table_sql := create_table_sql || 'BOOLEAN';
                WHEN 'datetime' THEN create_table_sql := create_table_sql || 'TIMESTAMP WITH TIME ZONE';
                WHEN 'json' THEN create_table_sql := create_table_sql || 'JSONB';
                WHEN 'uuid' THEN create_table_sql := create_table_sql || 'UUID';
                WHEN 'relation' THEN 
                    -- Handle relation fields with tenant schema
                    IF field_record.relation_config IS NOT NULL AND 
                       field_record.relation_config ? 'related_collection' THEN
                        create_table_sql := create_table_sql || 'UUID REFERENCES ' || tenant_schema || '.data_' || 
                                          (field_record.relation_config->>'related_collection') || '(id)';
                    ELSE
                        create_table_sql := create_table_sql || 'UUID';
                    END IF;
                ELSE create_table_sql := create_table_sql || 'TEXT';
            END CASE;
            
            -- Add constraints
            IF field_record.is_required THEN
                create_table_sql := create_table_sql || ' NOT NULL';
            END IF;
            
            IF field_record.is_unique THEN
                create_table_sql := create_table_sql || ' UNIQUE';
            END IF;
            
            IF field_record.default_value IS NOT NULL AND field_record.default_value != '' THEN
                create_table_sql := create_table_sql || ' DEFAULT ' || field_record.default_value;
            END IF;
        END LOOP;
        
        create_table_sql := create_table_sql || ')';
        
        -- Execute the CREATE TABLE statement
        EXECUTE create_table_sql;
        
        -- Add RLS (Row Level Security) policies
        EXECUTE 'ALTER TABLE ' || data_table_name || ' ENABLE ROW LEVEL SECURITY';
        
        -- Create policy for tenant isolation
        EXECUTE 'CREATE POLICY tenant_isolation ON ' || data_table_name || 
                ' FOR ALL USING (EXISTS (SELECT 1 FROM public.users WHERE id = current_setting(''app.current_user_id'')::uuid AND tenant_id = (SELECT tenant_id FROM public.collections WHERE name = ''' || NEW.name || ''')))';
        
        -- Create policy for user ownership (optional)
        EXECUTE 'CREATE POLICY user_ownership ON ' || data_table_name || 
                ' FOR ALL USING (created_by = current_setting(''app.current_user_id'')::uuid)';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Enhanced drop function
CREATE OR REPLACE FUNCTION drop_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
    tenant_schema TEXT;
BEGIN
    -- Only drop data tables for non-system collections
    IF OLD.is_system = false THEN
        -- Get tenant schema
        SELECT slug INTO tenant_schema FROM tenants WHERE id = OLD.tenant_id;
        data_table_name := tenant_schema || '.data_' || OLD.name;
        EXECUTE 'DROP TABLE IF EXISTS ' || data_table_name || ' CASCADE';
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Add tenant management functions
CREATE OR REPLACE FUNCTION create_tenant(tenant_name TEXT, tenant_slug TEXT, tenant_domain TEXT DEFAULT NULL)
RETURNS UUID AS $$
DECLARE
    new_tenant_id UUID;
BEGIN
    INSERT INTO tenants (id, name, slug, domain) 
    VALUES (uuid_generate_v4(), tenant_name, tenant_slug, tenant_domain)
    RETURNING id INTO new_tenant_id;
    
    -- Create tenant schema
    EXECUTE 'CREATE SCHEMA IF NOT EXISTS ' || tenant_slug;
    
    RETURN new_tenant_id;
END;
$$ LANGUAGE plpgsql;

-- Function to set current user context for RLS
CREATE OR REPLACE FUNCTION set_user_context(user_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_user_id', user_id::text, false);
END;
$$ LANGUAGE plpgsql;

-- Add tenant-based permissions (with conflict handling)
INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) VALUES
    -- Admin can manage their own tenant's collections
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    
    -- Admin can manage their own tenant's fields
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'create', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'read', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'update', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default')),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'delete', ARRAY['*'], (SELECT id FROM tenants WHERE slug = 'default'))
ON CONFLICT (role_id, table_name, action) DO NOTHING; 