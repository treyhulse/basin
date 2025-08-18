-- Fix the create_data_table function to properly handle the "default" schema name
-- The issue is that "default" is a reserved keyword in PostgreSQL and needs to be quoted

CREATE OR REPLACE FUNCTION create_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
    tenant_schema TEXT;
    create_table_sql TEXT;
    field_record RECORD;
BEGIN
    -- Only create data tables for non-system collections
    IF NEW.is_system = false THEN
        -- Get tenant schema name
        SELECT slug INTO tenant_schema FROM tenants WHERE id = NEW.tenant_id;
        
        -- Create tenant schema if it doesn't exist (properly quoted)
        EXECUTE 'CREATE SCHEMA IF NOT EXISTS "' || tenant_schema || '"';
        
        -- Use quoted schema name in table reference
        data_table_name := '"' || tenant_schema || '".data_' || NEW.name;
        
        -- Start building the CREATE TABLE statement
        create_table_sql := 'CREATE TABLE ' || data_table_name || ' (';
        create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),';
        create_table_sql := create_table_sql || 'created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
        create_table_sql := create_table_sql || 'updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
        create_table_sql := create_table_sql || 'created_by UUID REFERENCES public.users(id),';
        create_table_sql := create_table_sql || 'updated_by UUID REFERENCES public.users(id)';
        
        -- Add fields for this collection
        FOR field_record IN 
            SELECT * FROM fields WHERE collection_id = NEW.id ORDER BY sort_order
        LOOP
            create_table_sql := create_table_sql || ',' || field_record.name || ' ';
            
            -- Map field types to PostgreSQL types
            CASE field_record.type
                WHEN 'text' THEN
                    create_table_sql := create_table_sql || 'TEXT';
                WHEN 'number' THEN
                    create_table_sql := create_table_sql || 'NUMERIC';
                WHEN 'boolean' THEN
                    create_table_sql := create_table_sql || 'BOOLEAN';
                WHEN 'date' THEN
                    create_table_sql := create_table_sql || 'DATE';
                WHEN 'datetime' THEN
                    create_table_sql := create_table_sql || 'TIMESTAMP WITH TIME ZONE';
                ELSE
                    create_table_sql := create_table_sql || 'TEXT';
            END CASE;
            
            -- Add NOT NULL constraint if required
            IF field_record.is_required = true THEN
                create_table_sql := create_table_sql || ' NOT NULL';
            END IF;
        END LOOP;
        
        -- Close the CREATE TABLE statement
        create_table_sql := create_table_sql || ')';
        
        -- Execute the CREATE TABLE statement
        EXECUTE create_table_sql;
        
        -- Enable Row Level Security
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

-- Also fix the drop_data_table function
CREATE OR REPLACE FUNCTION drop_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
    tenant_schema TEXT;
BEGIN
    -- Only drop data tables for non-system collections
    IF OLD.is_system = false THEN
        -- Get tenant schema name
        SELECT slug INTO tenant_schema FROM tenants WHERE id = OLD.tenant_id;
        
        -- Use quoted schema name in table reference
        data_table_name := '"' || tenant_schema || '".data_' || OLD.name;
        
        -- Drop the data table if it exists
        EXECUTE 'DROP TABLE IF EXISTS ' || data_table_name || ' CASCADE';
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Fix the create_tenant function as well
CREATE OR REPLACE FUNCTION create_tenant(tenant_name TEXT, tenant_slug TEXT)
RETURNS UUID AS $$
DECLARE
    new_tenant_id UUID;
BEGIN
    -- Generate new tenant ID
    new_tenant_id := uuid_generate_v4();
    
    -- Insert the tenant
    INSERT INTO tenants (id, name, slug) VALUES (new_tenant_id, tenant_name, tenant_slug);
    
    -- Create the tenant schema (properly quoted)
    EXECUTE 'CREATE SCHEMA IF NOT EXISTS "' || tenant_slug || '"';
    
    RETURN new_tenant_id;
END;
$$ LANGUAGE plpgsql;