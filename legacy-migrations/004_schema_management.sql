-- Migration 004: Schema Management Tables
-- This migration adds the collections and fields tables that will be managed
-- through the existing dynamic /items/:table API endpoints

-- Create collections table
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    description TEXT,
    icon VARCHAR(50),
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create fields table
CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    collection_id UUID REFERENCES collections(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    type VARCHAR(50) NOT NULL, -- 'string', 'text', 'integer', 'decimal', 'boolean', 'datetime', 'json', 'uuid', 'relation'
    is_primary BOOLEAN DEFAULT false,
    is_required BOOLEAN DEFAULT false,
    is_unique BOOLEAN DEFAULT false,
    default_value TEXT,
    validation_rules JSONB,
    relation_config JSONB, -- For foreign key relationships
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(collection_id, name)
);

-- Create indexes for better performance
CREATE INDEX idx_collections_name ON collections(name);
CREATE INDEX idx_fields_collection_id ON fields(collection_id);
CREATE INDEX idx_fields_sort_order ON fields(collection_id, sort_order);

-- Insert system collections (these will be managed through the API)
INSERT INTO collections (id, name, display_name, description, icon, is_system) VALUES
    (uuid_generate_v4(), 'collections', 'Collections', 'System collection for managing data models', 'database', true),
    (uuid_generate_v4(), 'fields', 'Fields', 'System collection for managing field definitions', 'list', true),
    (uuid_generate_v4(), 'users', 'Users', 'System collection for user management', 'users', true),
    (uuid_generate_v4(), 'roles', 'Roles', 'System collection for role management', 'shield', true),
    (uuid_generate_v4(), 'permissions', 'Permissions', 'System collection for permission management', 'lock', true);

-- Insert fields for collections table
INSERT INTO fields (id, collection_id, name, display_name, type, is_primary, is_required, sort_order) VALUES
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'id', 'ID', 'uuid', true, true, 1),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'name', 'Name', 'string', false, true, 2),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'display_name', 'Display Name', 'string', false, false, 3),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'description', 'Description', 'text', false, false, 4),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'icon', 'Icon', 'string', false, false, 5),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'is_system', 'Is System', 'boolean', false, false, 6),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'created_at', 'Created At', 'datetime', false, false, 7),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'collections'), 'updated_at', 'Updated At', 'datetime', false, false, 8);

-- Insert fields for fields table
INSERT INTO fields (id, collection_id, name, display_name, type, is_primary, is_required, sort_order) VALUES
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'id', 'ID', 'uuid', true, true, 1),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'collection_id', 'Collection ID', 'uuid', false, true, 2),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'name', 'Name', 'string', false, true, 3),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'display_name', 'Display Name', 'string', false, false, 4),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'type', 'Type', 'string', false, true, 5),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'is_primary', 'Is Primary', 'boolean', false, false, 6),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'is_required', 'Is Required', 'boolean', false, false, 7),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'is_unique', 'Is Unique', 'boolean', false, false, 8),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'default_value', 'Default Value', 'text', false, false, 9),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'validation_rules', 'Validation Rules', 'json', false, false, 10),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'relation_config', 'Relation Config', 'json', false, false, 11),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'sort_order', 'Sort Order', 'integer', false, false, 12),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'created_at', 'Created At', 'datetime', false, false, 13),
    ((SELECT uuid_generate_v4()), (SELECT id FROM collections WHERE name = 'fields'), 'updated_at', 'Updated At', 'datetime', false, false, 14);

-- Add admin permissions for schema management
INSERT INTO permissions (role_id, table_name, action, allowed_fields) VALUES
    -- Admin can manage collections
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'collections', 'delete', ARRAY['*']),
    
    -- Admin can manage fields
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'fields', 'delete', ARRAY['*']);

-- Create a function to automatically create data tables when collections are created
CREATE OR REPLACE FUNCTION create_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
    create_table_sql TEXT;
    field_record RECORD;
BEGIN
    -- Only create data tables for non-system collections
    IF NEW.is_system = false THEN
        data_table_name := 'data_' || NEW.name;
        
        -- Start building the CREATE TABLE statement
        create_table_sql := 'CREATE TABLE ' || data_table_name || ' (';
        create_table_sql := create_table_sql || 'id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),';
        create_table_sql := create_table_sql || 'created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),';
        create_table_sql := create_table_sql || 'updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()';
        
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
                    -- Handle relation fields
                    IF field_record.relation_config IS NOT NULL AND 
                       field_record.relation_config ? 'related_collection' THEN
                        create_table_sql := create_table_sql || 'UUID REFERENCES data_' || 
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
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically create data tables
CREATE TRIGGER trigger_create_data_table
    AFTER INSERT ON collections
    FOR EACH ROW
    EXECUTE FUNCTION create_data_table();

-- Create a function to drop data tables when collections are deleted
CREATE OR REPLACE FUNCTION drop_data_table()
RETURNS TRIGGER AS $$
DECLARE
    data_table_name TEXT;
BEGIN
    -- Only drop data tables for non-system collections
    IF OLD.is_system = false THEN
        data_table_name := 'data_' || OLD.name;
        EXECUTE 'DROP TABLE IF EXISTS ' || data_table_name || ' CASCADE';
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically drop data tables
CREATE TRIGGER trigger_drop_data_table
    AFTER DELETE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION drop_data_table(); 