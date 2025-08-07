-- Migration 009: Add admin permissions for dynamic collections
-- This migration ensures the admin user has full access to all dynamically created collections
-- The admin user should have open access to everything (*) as requested

-- Add admin permissions for all dynamically created collections
-- This uses a wildcard approach to ensure admin has access to any collection created via the collections table

-- First, let's add permissions for any existing collections that might have been created
-- We'll use a function to dynamically add permissions for collections

CREATE OR REPLACE FUNCTION ensure_admin_collection_permissions()
RETURNS VOID AS $$
DECLARE
    collection_record RECORD;
    admin_role_id UUID;
    default_tenant_id UUID;
BEGIN
    -- Get admin role ID and default tenant ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = (SELECT id FROM tenants WHERE slug = 'default');
    SELECT id INTO default_tenant_id FROM tenants WHERE slug = 'default';
    
    -- Loop through all non-system collections and ensure admin has permissions
    FOR collection_record IN 
        SELECT name FROM collections 
        WHERE is_system = false AND tenant_id = default_tenant_id
    LOOP
        -- Add permissions for this collection if they don't exist
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'create', ARRAY['*'], default_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'read', ARRAY['*'], default_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'update', ARRAY['*'], default_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
        
        INSERT INTO permissions (role_id, table_name, action, allowed_fields, tenant_id) 
        VALUES 
            (admin_role_id, collection_record.name, 'delete', ARRAY['*'], default_tenant_id)
        ON CONFLICT (role_id, table_name, action) DO NOTHING;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Execute the function to add permissions for existing collections
SELECT ensure_admin_collection_permissions();

-- Create a trigger function to automatically add admin permissions when new collections are created
CREATE OR REPLACE FUNCTION add_admin_collection_permissions()
RETURNS TRIGGER AS $$
DECLARE
    admin_role_id UUID;
BEGIN
    -- Only add permissions for non-system collections
    IF NEW.is_system = false THEN
        -- Get admin role ID for the same tenant
        SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = NEW.tenant_id;
        
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

-- Create trigger to automatically add admin permissions when collections are created
DROP TRIGGER IF EXISTS trigger_add_admin_collection_permissions ON collections;
CREATE TRIGGER trigger_add_admin_collection_permissions
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION add_admin_collection_permissions();

-- Also create a trigger to remove admin permissions when collections are deleted
CREATE OR REPLACE FUNCTION remove_admin_collection_permissions()
RETURNS TRIGGER AS $$
DECLARE
    admin_role_id UUID;
BEGIN
    -- Only remove permissions for non-system collections
    IF OLD.is_system = false THEN
        -- Get admin role ID for the same tenant
        SELECT id INTO admin_role_id FROM roles WHERE name = 'admin' AND tenant_id = OLD.tenant_id;
        
        -- Remove permissions for the deleted collection
        DELETE FROM permissions 
        WHERE role_id = admin_role_id 
        AND table_name = OLD.name 
        AND tenant_id = OLD.tenant_id;
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically remove admin permissions when collections are deleted
DROP TRIGGER IF EXISTS trigger_remove_admin_collection_permissions ON collections;
CREATE TRIGGER trigger_remove_admin_collection_permissions
    AFTER DELETE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION remove_admin_collection_permissions();

-- Add comment for documentation
COMMENT ON FUNCTION ensure_admin_collection_permissions() IS 'Ensures admin user has permissions for all existing collections';
COMMENT ON FUNCTION add_admin_collection_permissions() IS 'Automatically adds admin permissions when new collections are created';
COMMENT ON FUNCTION remove_admin_collection_permissions() IS 'Automatically removes admin permissions when collections are deleted';
