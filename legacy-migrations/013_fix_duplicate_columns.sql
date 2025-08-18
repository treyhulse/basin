-- Migration 013: Fix duplicate columns in create_data_table function
-- This migration fixes the create_data_table function to avoid duplicate columns

DROP FUNCTION IF EXISTS create_data_table(UUID, TEXT);

CREATE OR REPLACE FUNCTION create_data_table(p_collection_id UUID, p_table_name TEXT)
RETURNS VOID AS $$
DECLARE
    create_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
    has_created_at BOOLEAN := false;
    has_updated_at BOOLEAN := false;
BEGIN
    -- Get the tenant schema for this collection
    SELECT t.slug INTO tenant_schema 
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = p_collection_id;
    
    -- Check if created_at and updated_at fields already exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'created_at') INTO has_created_at;
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'updated_at') INTO has_updated_at;
    
    -- Start building the CREATE TABLE statement
    create_table_sql := 'CREATE TABLE IF NOT EXISTS "' || tenant_schema || '".data_' || p_table_name || ' (';
    create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),';
    
    -- Only add created_at if it doesn't exist as a field
    IF NOT has_created_at THEN
        create_table_sql := create_table_sql || 'created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
    END IF;
    
    -- Only add updated_at if it doesn't exist as a field
    IF NOT has_updated_at THEN
        create_table_sql := create_table_sql || 'updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
    END IF;
    
    -- Add fields from the fields table
    FOR field_record IN 
        SELECT name, type, is_required, is_unique, default_value, relation_config
        FROM fields 
        WHERE collection_id = p_collection_id 
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

-- Now create the missing main.data_blog_posts table
DO $$
DECLARE
    blog_posts_id UUID;
BEGIN
    -- Get the blog_posts collection ID
    SELECT id INTO blog_posts_id FROM collections WHERE name = 'blog_posts' LIMIT 1;
    
    IF blog_posts_id IS NOT NULL THEN
        PERFORM create_data_table(blog_posts_id, 'blog_posts');
    END IF;
END $$;
