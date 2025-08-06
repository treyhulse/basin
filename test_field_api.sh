#!/bin/bash

# Test Field Creation API
echo "🧪 Testing Field Creation for blog_posts Collection..."

# Step 1: Login and get token
echo "📝 Step 1: Login..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "Token: ${TOKEN:0:20}..."

# Step 2: Create a field
echo "📝 Step 2: Creating 'title' field..."
FIELD_RESPONSE=$(curl -s -X POST http://localhost:8080/items/fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "collection_id": "5c059eeb-cd5d-4fb7-a35a-c9c485b2e024",
    "name": "title",
    "display_name": "Title",
    "type": "text",
    "is_required": true,
    "is_unique": false,
    "sort_order": 1
  }')

echo "Field creation response:"
echo $FIELD_RESPONSE

echo "✅ Field creation test completed!"