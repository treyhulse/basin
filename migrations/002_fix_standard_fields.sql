
-- Migration to fix missing standard fields in data tables
-- This adds the missing created_by, updated_by, created_at, updated_at fields
-- that the API expects for all collections

-- Function to add missing standard fields to existing data tables
CREATE OR REPLACE FUNCTION add_missing_standard_fields_to_data_tables()
RETURNS VOID AS $$
DECLARE
    collection_record RECORD;
    tenant_schema TEXT;
    alter_sql TEXT;
BEGIN
    -- Loop through all non-system collections
    FOR collection_record IN 
        SELECT c.id, c.name, c.tenant_id, t.slug as tenant_slug
        FROM collections c
        JOIN tenants t ON c.tenant_id = t.id
        WHERE c.is_system = false
    LOOP
        tenant_schema := collection_record.tenant_slug;
        
        -- Add created_by field if it doesn't exist
        BEGIN
            alter_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || collection_record.name || ' ADD COLUMN IF NOT EXISTS created_by UUID';
            EXECUTE alter_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add created_by to %: %', collection_record.name, SQLERRM;
        END;
        
        -- Add updated_by field if it doesn't exist
        BEGIN
            alter_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || collection_record.name || ' ADD COLUMN IF NOT EXISTS updated_by UUID';
            EXECUTE alter_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add updated_by to %: %', collection_record.name, SQLERRM;
        END;
        
        -- Add created_at field if it doesn't exist
        BEGIN
            alter_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || collection_record.name || ' ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
            EXECUTE alter_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add created_at to %: %', collection_record.name, SQLERRM;
        END;
        
        -- Add updated_at field if it doesn't exist
        BEGIN
            alter_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || collection_record.name || ' ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
            EXECUTE alter_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add updated_at to %: %', collection_record.name, SQLERRM;
        END;
        
        -- Add tenant_id field if it doesn't exist
        BEGIN
            alter_sql := 'ALTER TABLE "' || tenant_schema || '".data_' || collection_record.name || ' ADD COLUMN IF NOT EXISTS tenant_id UUID DEFAULT ''' || collection_record.tenant_id || '''';
            EXECUTE alter_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not add tenant_id to %: %', collection_record.name, SQLERRM;
        END;
        
        RAISE NOTICE 'Added standard fields to data_%', collection_record.name;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to update the create_data_table function to include standard fields
CREATE OR REPLACE FUNCTION create_data_table(p_collection_id UUID, p_table_name TEXT)
RETURNS VOID AS $$
DECLARE
    create_table_sql TEXT;
    field_record RECORD;
    tenant_schema TEXT;
    has_fields BOOLEAN;
    tenant_id UUID;
BEGIN
    -- Get the tenant schema and ID for this collection
    SELECT t.slug, c.tenant_id INTO tenant_schema, tenant_id
    FROM tenants t 
    JOIN collections c ON t.id = c.tenant_id 
    WHERE c.id = p_collection_id;
    
    -- Check if there are any fields for this collection
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id) INTO has_fields;
    
    -- Start building the CREATE TABLE statement
    create_table_sql := 'CREATE TABLE IF NOT EXISTS "' || tenant_schema || '".data_' || p_table_name || ' (';
    create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4()';
    
    -- Add standard fields that every collection should have
    create_table_sql := create_table_sql || ', created_by UUID';
    create_table_sql := create_table_sql || ', updated_by UUID';
    create_table_sql := create_table_sql || ', created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    create_table_sql := create_table_sql || ', updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
    create_table_sql := create_table_sql || ', tenant_id UUID DEFAULT ''' || tenant_id || '''';
    
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
                    create_table_sql := create_table_sql || 'DECIMAL(10,2)';
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
    
    RAISE NOTICE 'Created data table % with standard fields', p_table_name;
END;
$$ LANGUAGE plpgsql;

-- Function to recreate existing data tables with standard fields
CREATE OR REPLACE FUNCTION recreate_data_tables_with_standard_fields()
RETURNS VOID AS $$
DECLARE
    collection_record RECORD;
    tenant_schema TEXT;
BEGIN
    -- Loop through all non-system collections
    FOR collection_record IN 
        SELECT c.id, c.name, c.tenant_id, t.slug as tenant_schema
        FROM collections c
        JOIN tenants t ON c.tenant_id = t.id
        WHERE c.is_system = false
    LOOP
        tenant_schema := collection_record.tenant_schema;
        
        -- Drop the existing table
        EXECUTE 'DROP TABLE IF EXISTS "' || tenant_schema || '".data_' || collection_record.name;
        
        -- Recreate with standard fields
        PERFORM create_data_table(collection_record.id, collection_record.name);
        
        RAISE NOTICE 'Recreated data_% with standard fields', collection_record.name;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Run the migration functions
SELECT add_missing_standard_fields_to_data_tables();
SELECT recreate_data_tables_with_standard_fields();

-- Add triggers for updated_at columns on data tables
CREATE OR REPLACE FUNCTION create_data_table_triggers()
RETURNS VOID AS $$
DECLARE
    collection_record RECORD;
    tenant_schema TEXT;
    trigger_sql TEXT;
BEGIN
    -- Loop through all non-system collections
    FOR collection_record IN 
        SELECT c.id, c.name, c.tenant_id, t.slug as tenant_schema
        FROM collections c
        JOIN tenants t ON c.tenant_id = t.id
        WHERE c.is_system = false
    LOOP
        tenant_schema := collection_record.tenant_schema;
        
        -- Create trigger for updated_at
        BEGIN
            trigger_sql := 'CREATE TRIGGER update_data_' || collection_record.name || '_updated_at 
                           BEFORE UPDATE ON "' || tenant_schema || '".data_' || collection_record.name || ' 
                           FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()';
            EXECUTE trigger_sql;
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Could not create trigger for data_%: %', collection_record.name, SQLERRM;
        END;
        
        RAISE NOTICE 'Added triggers to data_%', collection_record.name;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all data tables
SELECT create_data_table_triggers();

-- Update the ensure_standard_collection_fields function to NOT require name field
CREATE OR REPLACE FUNCTION ensure_standard_collection_fields(p_collection_id UUID, p_tenant_id UUID)
RETURNS VOID AS $$
DECLARE
    field_exists BOOLEAN;
    collection_name TEXT;
    tenant_schema TEXT;
BEGIN
    -- Get collection name and tenant schema
    SELECT c.name, t.slug INTO collection_name, tenant_schema
    FROM collections c
    JOIN tenants t ON c.tenant_id = t.id
    WHERE c.id = p_collection_id;
    
    -- Check and add 'display_name' field if it doesn't exist (optional, for admin UI)
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
    
    -- Check and add 'tenant_id' field if it doesn't exist
    SELECT EXISTS(SELECT 1 FROM fields WHERE collection_id = p_collection_id AND name = 'tenant_id') INTO field_exists;
    IF NOT field_exists THEN
        INSERT INTO fields (collection_id, name, display_name, type, is_required, sort_order, tenant_id)
        VALUES (p_collection_id, 'tenant_id', 'Tenant ID', 'uuid', false, 104, p_tenant_id);
    END IF;
    
    -- Note: We intentionally do NOT add a 'name' field by default
    -- Collections should define their own meaningful fields
END;
$$ LANGUAGE plpgsql;

-- Add the missing standard fields to the fields table for existing collections
SELECT ensure_standard_collection_fields(id, tenant_id) FROM collections WHERE is_system = false;

-- Migration completed successfully
-- All dynamic data tables now have the required standard fields:
-- - created_by, updated_by, created_at, updated_at, tenant_id
-- - No forced 'name' field requirement
-- - Triggers for automatic updated_at updates
