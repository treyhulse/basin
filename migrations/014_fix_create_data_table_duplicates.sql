-- Migration 014: Fix create_data_table function to handle duplicate columns more robustly
-- This migration makes the create_data_table function more robust by checking for existing columns

DROP FUNCTION IF EXISTS create_data_table(UUID, TEXT);

CREATE OR REPLACE FUNCTION create_data_table(p_collection_id UUID, p_table_name TEXT)
RETURNS VOID AS $$
DECLARE
    create_table_sql TEXT;
    alter_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
    table_exists BOOLEAN;
    column_exists BOOLEAN;
BEGIN
    -- Get the tenant schema for this collection
    SELECT t.slug INTO tenant_schema 
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = p_collection_id;
    
    -- Check if the table already exists
    SELECT EXISTS(
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = tenant_schema 
        AND table_name = 'data_' || p_table_name
    ) INTO table_exists;
    
    IF NOT table_exists THEN
        -- Create the table with basic structure
        create_table_sql := 'CREATE TABLE "' || tenant_schema || '".data_' || p_table_name || ' (';
        create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4()';
        create_table_sql := create_table_sql || ')';
        
        EXECUTE create_table_sql;
    END IF;
    
    -- Now add columns one by one, checking if they exist first
    FOR field_record IN 
        SELECT name, type, is_required, is_unique, default_value, relation_config
        FROM fields 
        WHERE collection_id = p_collection_id 
        ORDER BY sort_order, name
    LOOP
        -- Check if column already exists
        SELECT EXISTS(
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = tenant_schema 
            AND table_name = 'data_' || p_table_name 
            AND column_name = field_record.name
        ) INTO column_exists;
        
        IF NOT column_exists THEN
            -- Build ALTER TABLE statement to add the column
            alter_table_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || p_table_name || ' ADD COLUMN "' || field_record.name || '" ';
            
            -- Map field types to PostgreSQL types
            CASE field_record.type
                WHEN 'string', 'text' THEN
                    alter_table_sql := alter_table_sql || 'TEXT';
                WHEN 'integer', 'int' THEN
                    alter_table_sql := alter_table_sql || 'INTEGER';
                WHEN 'float', 'decimal' THEN
                    alter_table_sql := alter_table_sql || 'DECIMAL';
                WHEN 'boolean', 'bool' THEN
                    alter_table_sql := alter_table_sql || 'BOOLEAN';
                WHEN 'json', 'object' THEN
                    alter_table_sql := alter_table_sql || 'JSONB';
                WHEN 'date', 'datetime' THEN
                    alter_table_sql := alter_table_sql || 'TIMESTAMP WITH TIME ZONE';
                WHEN 'uuid' THEN
                    alter_table_sql := alter_table_sql || 'UUID';
                ELSE
                    alter_table_sql := alter_table_sql || 'TEXT';
            END CASE;
            
            -- Add NOT NULL constraint for required fields
            IF field_record.is_required THEN
                alter_table_sql := alter_table_sql || ' NOT NULL';
            END IF;
            
            -- Add default value if specified
            IF field_record.default_value IS NOT NULL AND field_record.default_value != '' THEN
                alter_table_sql := alter_table_sql || ' DEFAULT ' || field_record.default_value;
            END IF;
            
            -- Execute the ALTER TABLE statement
            EXECUTE alter_table_sql;
        END IF;
    END LOOP;
    
    -- Add standard columns if they don't exist
    -- created_at
    SELECT EXISTS(
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = tenant_schema 
        AND table_name = 'data_' || p_table_name 
        AND column_name = 'created_at'
    ) INTO column_exists;
    
    IF NOT column_exists THEN
        EXECUTE 'ALTER TABLE "' || tenant_schema || '".data_' || p_table_name || ' ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    END IF;
    
    -- updated_at
    SELECT EXISTS(
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = tenant_schema 
        AND table_name = 'data_' || p_table_name 
        AND column_name = 'updated_at'
    ) INTO column_exists;
    
    IF NOT column_exists THEN
        EXECUTE 'ALTER TABLE "' || tenant_schema || '".data_' || p_table_name || ' ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    END IF;
    
    -- created_by
    SELECT EXISTS(
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = tenant_schema 
        AND table_name = 'data_' || p_table_name 
        AND column_name = 'created_by'
    ) INTO column_exists;
    
    IF NOT column_exists THEN
        EXECUTE 'ALTER TABLE "' || tenant_schema || '".data_' || p_table_name || ' ADD COLUMN created_by UUID';
    END IF;
    
    -- updated_by
    SELECT EXISTS(
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = tenant_schema 
        AND table_name = 'data_' || p_table_name 
        AND column_name = 'updated_by'
    ) INTO column_exists;
    
    IF NOT column_exists THEN
        EXECUTE 'ALTER TABLE "' || tenant_schema || '".data_' || p_table_name || ' ADD COLUMN updated_by UUID';
    END IF;
END;
$$ LANGUAGE plpgsql;
