-- Migration 011: Fix reserved keyword issues
-- This migration changes all references from 'default' (reserved keyword) to 'main'
-- to prevent PostgreSQL syntax errors

-- First, let's update the tenant slug from 'default' to 'main'
UPDATE tenants SET slug = 'main' WHERE slug = 'default';

-- Update all references to the old 'default' tenant
UPDATE users SET tenant_id = (SELECT id FROM tenants WHERE slug = 'main') WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE collections SET tenant_id = (SELECT id FROM tenants WHERE slug = 'main') WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE fields SET tenant_id = (SELECT id FROM tenants WHERE slug = 'main') WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE permissions SET tenant_id = (SELECT id FROM tenants WHERE slug = 'main') WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
UPDATE roles SET tenant_id = (SELECT id FROM tenants WHERE slug = 'main') WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'default');

-- Drop and recreate the create_data_table function to use 'main' instead of 'default'
DROP FUNCTION IF EXISTS create_data_table(UUID, TEXT);

CREATE OR REPLACE FUNCTION create_data_table(collection_id UUID, table_name TEXT)
RETURNS VOID AS $$
DECLARE
    create_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
BEGIN
    -- Get the tenant schema for this collection
    SELECT t.slug INTO tenant_schema 
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = collection_id;
    
    -- Start building the CREATE TABLE statement
    create_table_sql := 'CREATE TABLE IF NOT EXISTS "' || tenant_schema || '".data_' || table_name || ' (';
    create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),';
    create_table_sql := create_table_sql || 'created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
    create_table_sql := create_table_sql || 'updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
    
    -- Add fields from the fields table
    FOR field_record IN 
        SELECT name, type, is_required, is_unique, default_value, relation_config
        FROM fields 
        WHERE collection_id = $1 
        ORDER BY sort_order, name
    LOOP
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
        
        create_table_sql := create_table_sql || ',';
    END LOOP;
    
    -- Remove the trailing comma and close the statement
    create_table_sql := rtrim(create_table_sql, ',') || ')';
    
    -- Execute the CREATE TABLE statement
    EXECUTE create_table_sql;
END;
$$ LANGUAGE plpgsql;

-- Drop and recreate the drop_data_table function
DROP FUNCTION IF EXISTS drop_data_table(UUID, TEXT);

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

-- Update the ensure_admin_collection_permissions function
DROP FUNCTION IF EXISTS ensure_admin_collection_permissions();

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

-- Update the ensure_standard_collection_fields function
DROP FUNCTION IF EXISTS ensure_standard_collection_fields(UUID, UUID);

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

-- Update the add_standard_fields_to_new_collection function
DROP FUNCTION IF EXISTS add_standard_fields_to_new_collection() CASCADE;

CREATE OR REPLACE FUNCTION add_standard_fields_to_new_collection()
RETURNS TRIGGER AS $$
BEGIN
    -- Only add standard fields to non-system collections
    IF NEW.is_system = false THEN
        PERFORM ensure_standard_collection_fields(NEW.id, NEW.tenant_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Update the admin_collection_permissions_trigger function
DROP FUNCTION IF EXISTS admin_collection_permissions_trigger() CASCADE;

CREATE OR REPLACE FUNCTION admin_collection_permissions_trigger()
RETURNS TRIGGER AS $$
DECLARE
    admin_role_id UUID;
BEGIN
    -- Get admin role ID for the tenant
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = NEW.tenant_id;
    
    IF admin_role_id IS NOT NULL THEN
        -- Add permissions for the new collection
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
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Update existing blog_posts collection to have standard fields
DO $$
DECLARE
    blog_posts_id UUID;
    main_tenant_id UUID;
BEGIN
    -- Get the blog_posts collection ID
    SELECT id INTO blog_posts_id FROM collections WHERE name = 'blog_posts' LIMIT 1;
    SELECT id INTO main_tenant_id FROM tenants WHERE slug = 'main' LIMIT 1;
    
    IF blog_posts_id IS NOT NULL AND main_tenant_id IS NOT NULL THEN
        PERFORM ensure_standard_collection_fields(blog_posts_id, main_tenant_id);
    END IF;
END $$;

-- Run the admin collection permissions function to update existing permissions
SELECT ensure_admin_collection_permissions();
