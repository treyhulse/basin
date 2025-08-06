#!/bin/bash

# Basin API Start Script
# This script sets up and starts the Basin API server

echo "🚀 Starting Basin API..."

# Step 1: Start Docker containers
echo "📦 Starting Docker containers..."
docker-compose up -d
if [ $? -ne 0 ]; then
    echo "❌ Failed to start Docker containers"
    exit 1
fi

# Step 2: Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    attempt=$((attempt + 1))
    sleep 2
    if docker exec go-rbac-postgres pg_isready -U postgres > /dev/null 2>&1; then
        echo "✅ PostgreSQL is ready!"
        break
    fi
    echo "⏳ Attempt $attempt/$max_attempts - PostgreSQL not ready yet..."
done

if [ $attempt -eq $max_attempts ]; then
    echo "❌ PostgreSQL failed to start within timeout"
    exit 1
fi

# Step 3: Run database migrations
echo "🗄️ Running database migrations..."
migrations=(
    "001_init.sql"
    "002_api_keys.sql" 
    "003_admin_permissions.sql"
    "004_schema_management.sql"
    "005_multi_tenant.sql"
    "006_admin_system_permissions.sql"
    "007_fix_default_schema.sql"
)

for migration in "${migrations[@]}"; do
    migration_path="migrations/$migration"
    if [ -f "$migration_path" ]; then
        echo "  📝 Applying $migration..."
        docker exec -i go-rbac-postgres psql -U postgres -d go_rbac_db < "$migration_path"
        if [ $? -ne 0 ]; then
            echo "❌ Failed to apply migration: $migration"
            exit 1
        fi
    else
        echo "⚠️ Migration file not found: $migration"
    fi
done

# Step 4: Generate SQLC code
echo "🔧 Generating SQLC code..."
sqlc generate
if [ $? -ne 0 ]; then
    echo "❌ Failed to generate SQLC code"
    exit 1
fi

# Step 5: Start the API server
echo "🌟 Starting Basin API server..."
echo "📍 Server will be available at: http://localhost:8080"
echo "📍 Health check: http://localhost:8080/health"
echo ""
echo "🔑 Default admin credentials:"
echo "   Email: admin@example.com"
echo "   Password: password"
echo ""
echo "Press Ctrl+C to stop the server"
echo "================================"

go run cmd/main.go