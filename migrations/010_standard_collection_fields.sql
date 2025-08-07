-- Migration 010: Add standard core fields to all dynamic collections
-- This migration ensures all dynamic collections have standard fields for tracking and identification

-- Function to add standard fields to a collection if they don't exist
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

-- Add standard fields to existing collections
DO $$
DECLARE
    collection_record RECORD;
BEGIN
    FOR collection_record IN 
        SELECT id, tenant_id FROM collections WHERE is_system = false
    LOOP
        PERFORM ensure_standard_collection_fields(collection_record.id, collection_record.tenant_id);
    END LOOP;
END $$;

-- Create trigger to automatically add standard fields to new collections
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

-- Create trigger
DROP TRIGGER IF EXISTS trigger_add_standard_fields ON collections;
CREATE TRIGGER trigger_add_standard_fields
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION add_standard_fields_to_new_collection();

-- Update existing blog_posts collection to have standard fields
-- (This will be handled by the DO block above, but let's be explicit)
DO $$
DECLARE
    blog_posts_id UUID;
    default_tenant_id UUID;
BEGIN
    -- Get the blog_posts collection ID
    SELECT id INTO blog_posts_id FROM collections WHERE name = 'blog_posts' LIMIT 1;
    SELECT id INTO default_tenant_id FROM tenants WHERE slug = 'default' LIMIT 1;
    
    IF blog_posts_id IS NOT NULL AND default_tenant_id IS NOT NULL THEN
        PERFORM ensure_standard_collection_fields(blog_posts_id, default_tenant_id);
    END IF;
END $$;
