-- Migration 015: Fix collection triggers to properly call create_data_table function
-- This migration fixes the triggers to pass the correct parameters to the create_data_table function

-- Drop the existing triggers
DROP TRIGGER IF EXISTS trigger_create_data_table ON collections;
DROP TRIGGER IF EXISTS trigger_drop_data_table ON collections;

-- Create a new trigger function that calls create_data_table with parameters
CREATE OR REPLACE FUNCTION trigger_create_data_table()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create data tables for non-system collections
    IF NEW.is_system = false THEN
        -- Call the create_data_table function with the collection ID and name
        PERFORM create_data_table(NEW.id, NEW.name);
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a new trigger function that calls drop_data_table with parameters
CREATE OR REPLACE FUNCTION trigger_drop_data_table()
RETURNS TRIGGER AS $$
DECLARE
    tenant_schema TEXT;
BEGIN
    -- Only drop data tables for non-system collections
    IF OLD.is_system = false THEN
        -- Get the tenant schema for this collection
        SELECT t.slug INTO tenant_schema 
        FROM tenants t 
        WHERE t.id = OLD.tenant_id;
        
        -- Drop the data table
        EXECUTE 'DROP TABLE IF EXISTS "' || tenant_schema || '".data_' || OLD.name || ' CASCADE';
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Recreate the triggers
CREATE TRIGGER trigger_create_data_table
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION trigger_create_data_table();

CREATE TRIGGER trigger_drop_data_table
    AFTER DELETE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION trigger_drop_data_table();
