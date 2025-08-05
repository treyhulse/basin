-- Migration: Add API Keys table
-- This allows users to generate API keys for programmatic access

-- Create API keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);

-- Add trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_api_keys_updated_at 
    BEFORE UPDATE ON api_keys 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert some sample API keys for testing
-- Note: These are bcrypt hashes of the actual keys
-- Key: "admin_api_key_123" -> Hash: $2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
-- Key: "manager_api_key_456" -> Hash: $2a$10$TKh8H1.PfQx37YgCzwiKb.KjNyWgaHb9cbcoQgdIVFlYg7B77UdFm
INSERT INTO api_keys (user_id, name, key_hash, expires_at) VALUES
    ((SELECT id FROM users WHERE email = 'admin@example.com'), 'Admin API Key', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NULL),
    ((SELECT id FROM users WHERE email = 'manager@example.com'), 'Manager API Key', '$2a$10$TKh8H1.PfQx37YgCzwiKb.KjNyWgaHb9cbcoQgdIVFlYg7B77UdFm', NULL);

-- Add permissions for API key management (admin can manage all, users can manage their own)
INSERT INTO permissions (role_id, table_name, action, allowed_fields) VALUES
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'api_keys', 'delete', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'read', ARRAY['id', 'name', 'is_active', 'expires_at', 'last_used_at', 'created_at']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'create', ARRAY['name', 'expires_at']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'update', ARRAY['name', 'is_active', 'expires_at']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'api_keys', 'delete', ARRAY['*']); 