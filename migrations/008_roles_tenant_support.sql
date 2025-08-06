-- Migration 008: Add tenant support to roles table
-- This migration adds tenant_id column to roles table for multi-tenant role isolation

-- Add tenant_id column to roles table
ALTER TABLE roles ADD COLUMN tenant_id UUID REFERENCES tenants(id);

-- Create index for tenant isolation
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);

-- Update existing roles to belong to default tenant
UPDATE roles SET tenant_id = (SELECT id FROM tenants WHERE slug = 'default');

-- Make tenant_id NOT NULL after updating existing records
ALTER TABLE roles ALTER COLUMN tenant_id SET NOT NULL;

-- Update unique constraint to include tenant_id for role name uniqueness per tenant
-- First drop the existing unique constraint
ALTER TABLE roles DROP CONSTRAINT roles_name_key;

-- Add new unique constraint that includes tenant_id
ALTER TABLE roles ADD CONSTRAINT roles_name_tenant_unique UNIQUE (name, tenant_id);

-- Update existing permissions to ensure role-tenant consistency
-- This ensures permissions reference roles within the same tenant
UPDATE permissions 
SET tenant_id = (
    SELECT r.tenant_id 
    FROM roles r 
    WHERE r.id = permissions.role_id
) 
WHERE tenant_id IS NULL OR tenant_id != (
    SELECT r.tenant_id 
    FROM roles r 
    WHERE r.id = permissions.role_id
);

-- Add a constraint to ensure permissions and roles are in the same tenant
-- This prevents cross-tenant permission assignments
ALTER TABLE permissions 
ADD CONSTRAINT check_role_tenant_consistency 
CHECK (
    tenant_id = (SELECT tenant_id FROM roles WHERE id = role_id)
);

-- Create additional roles for each tenant if needed
-- This creates basic roles for the default tenant that already exist
-- Future tenants will need their own role setup

-- Add comment for documentation
COMMENT ON COLUMN roles.tenant_id IS 'Tenant isolation for roles - each tenant has its own set of roles';
COMMENT ON CONSTRAINT roles_name_tenant_unique ON roles IS 'Role names must be unique within each tenant';
COMMENT ON CONSTRAINT check_role_tenant_consistency ON permissions IS 'Ensures permissions and roles belong to the same tenant';