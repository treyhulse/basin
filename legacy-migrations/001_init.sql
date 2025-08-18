-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user_roles junction table
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Create permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    table_name VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete'
    field_filter JSONB, -- {"field": "value"} for row-level filtering
    allowed_fields TEXT[], -- array of allowed fields for field-level access
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, table_name, action)
);

-- Create sample tables for demonstration
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    category VARCHAR(100),
    stock_quantity INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID REFERENCES customers(id),
    order_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('manager', 'Can manage products and view orders'),
    ('sales', 'Can view products and create orders'),
    ('customer', 'Can view products and own orders');

-- Insert default users (password: admin123 for admin, password for manager)
INSERT INTO users (email, password_hash, first_name, last_name) VALUES
    ('admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User'),
    ('manager@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Manager', 'User');

-- Assign roles to users
INSERT INTO user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM users u, roles r 
WHERE u.email = 'admin@example.com' AND r.name = 'admin';

INSERT INTO user_roles (user_id, role_id) 
SELECT u.id, r.id 
FROM users u, roles r 
WHERE u.email = 'manager@example.com' AND r.name = 'manager';

-- Insert sample permissions
INSERT INTO permissions (role_id, table_name, action, allowed_fields) VALUES
    -- Admin permissions (all tables, all actions, all fields)
    ((SELECT id FROM roles WHERE name = 'admin'), 'products', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'products', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'products', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'products', 'delete', ARRAY['*']),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'customers', 'delete', ARRAY['*']),
    
    ((SELECT id FROM roles WHERE name = 'admin'), 'orders', 'create', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'orders', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'orders', 'update', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'admin'), 'orders', 'delete', ARRAY['*']),
    
    -- Manager permissions
    ((SELECT id FROM roles WHERE name = 'manager'), 'products', 'create', ARRAY['name', 'description', 'price', 'category', 'stock_quantity']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'products', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'products', 'update', ARRAY['name', 'description', 'price', 'category', 'stock_quantity']),
    
    ((SELECT id FROM roles WHERE name = 'manager'), 'orders', 'read', ARRAY['*']),
    ((SELECT id FROM roles WHERE name = 'manager'), 'orders', 'update', ARRAY['status', 'notes']),
    
    -- Sales permissions
    ((SELECT id FROM roles WHERE name = 'sales'), 'products', 'read', ARRAY['id', 'name', 'description', 'price', 'category']),
    ((SELECT id FROM roles WHERE name = 'sales'), 'orders', 'create', ARRAY['customer_id', 'total_amount', 'notes']),
    ((SELECT id FROM roles WHERE name = 'sales'), 'orders', 'read', ARRAY['*']),
    
    -- Customer permissions (limited to own orders)
    ((SELECT id FROM roles WHERE name = 'customer'), 'products', 'read', ARRAY['id', 'name', 'description', 'price', 'category']);

-- Insert sample data
INSERT INTO products (name, description, price, category, stock_quantity) VALUES
    ('Laptop Pro', 'High-performance laptop for professionals', 1299.99, 'Electronics', 50),
    ('Wireless Mouse', 'Ergonomic wireless mouse', 29.99, 'Accessories', 100),
    ('Office Chair', 'Comfortable office chair with lumbar support', 199.99, 'Furniture', 25);

INSERT INTO customers (first_name, last_name, email, phone, address) VALUES
    ('John', 'Doe', 'john.doe@example.com', '+1234567890', '123 Main St, City, State'),
    ('Jane', 'Smith', 'jane.smith@example.com', '+0987654321', '456 Oak Ave, Town, State');

INSERT INTO orders (customer_id, status, total_amount, notes) VALUES
    ((SELECT id FROM customers WHERE email = 'john.doe@example.com'), 'completed', 1329.98, 'Express delivery requested'),
    ((SELECT id FROM customers WHERE email = 'jane.smith@example.com'), 'pending', 29.99, 'Standard shipping');

INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
    ((SELECT id FROM orders WHERE notes = 'Express delivery requested'), (SELECT id FROM products WHERE name = 'Laptop Pro'), 1, 1299.99),
    ((SELECT id FROM orders WHERE notes = 'Express delivery requested'), (SELECT id FROM products WHERE name = 'Wireless Mouse'), 1, 29.99),
    ((SELECT id FROM orders WHERE notes = 'Standard shipping'), (SELECT id FROM products WHERE name = 'Wireless Mouse'), 1, 29.99); 