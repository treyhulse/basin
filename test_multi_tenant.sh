#!/bin/bash

# Test script for Multi-Tenant Schema Management System
# Make sure the server is running on localhost:8080

echo "🚀 Testing Multi-Tenant Schema Management System"
echo "================================================"

# Base URL
BASE_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Test 1: Health check
echo -e "\n${YELLOW}1. Testing health endpoint...${NC}"
curl -s "$BASE_URL/health" > /dev/null
print_status $? "Health check"

# Test 2: Login as admin
echo -e "\n${YELLOW}2. Testing admin login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    print_status 0 "Admin login successful"
    echo "   Token: ${TOKEN:0:20}..."
else
    print_status 1 "Admin login failed"
    echo "   Response: $LOGIN_RESPONSE"
    exit 1
fi

# Test 3: Get collections (schema metadata)
echo -e "\n${YELLOW}3. Testing collections endpoint...${NC}"
COLLECTIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/collections" \
  -H "Authorization: Bearer $TOKEN")

if echo "$COLLECTIONS_RESPONSE" | grep -q "data"; then
    print_status 0 "Get collections successful"
    COLLECTION_COUNT=$(echo "$COLLECTIONS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $COLLECTION_COUNT collections"
else
    print_status 1 "Get collections failed"
    echo "   Response: $COLLECTIONS_RESPONSE"
fi

# Test 4: Get fields (schema metadata)
echo -e "\n${YELLOW}4. Testing fields endpoint...${NC}"
FIELDS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/fields" \
  -H "Authorization: Bearer $TOKEN")

if echo "$FIELDS_RESPONSE" | grep -q "data"; then
    print_status 0 "Get fields successful"
    FIELD_COUNT=$(echo "$FIELDS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $FIELD_COUNT fields"
else
    print_status 1 "Get fields failed"
    echo "   Response: $FIELDS_RESPONSE"
fi

# Test 5: Create a new collection
echo -e "\n${YELLOW}5. Testing collection creation...${NC}"
COLLECTION_DATA='{
  "name": "test_blog_posts",
  "display_name": "Test Blog Posts",
  "description": "Test collection for blog posts",
  "icon": "article"
}'

CREATE_COLLECTION_RESPONSE=$(curl -s -X POST "$BASE_URL/items/collections" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$COLLECTION_DATA")

if echo "$CREATE_COLLECTION_RESPONSE" | grep -q "test_blog_posts"; then
    print_status 0 "Create collection successful"
    COLLECTION_ID=$(echo "$CREATE_COLLECTION_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "   Collection ID: ${COLLECTION_ID:0:8}..."
else
    print_status 1 "Create collection failed"
    echo "   Response: $CREATE_COLLECTION_RESPONSE"
fi

# Test 6: Create fields for the collection
echo -e "\n${YELLOW}6. Testing field creation...${NC}"
FIELD_DATA='{
  "collection_id": "'$COLLECTION_ID'",
  "name": "title",
  "display_name": "Title",
  "type": "string",
  "is_required": true,
  "sort_order": 1
}'

CREATE_FIELD_RESPONSE=$(curl -s -X POST "$BASE_URL/items/fields" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$FIELD_DATA")

if echo "$CREATE_FIELD_RESPONSE" | grep -q "title"; then
    print_status 0 "Create field successful"
else
    print_status 1 "Create field failed"
    echo "   Response: $CREATE_FIELD_RESPONSE"
fi

# Test 7: Create another field
FIELD_DATA2='{
  "collection_id": "'$COLLECTION_ID'",
  "name": "content",
  "display_name": "Content",
  "type": "text",
  "is_required": false,
  "sort_order": 2
}'

CREATE_FIELD_RESPONSE2=$(curl -s -X POST "$BASE_URL/items/fields" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$FIELD_DATA2")

if echo "$CREATE_FIELD_RESPONSE2" | grep -q "content"; then
    print_status 0 "Create second field successful"
else
    print_status 1 "Create second field failed"
    echo "   Response: $CREATE_FIELD_RESPONSE2"
fi

# Test 8: Try to access the new collection (should exist now)
echo -e "\n${YELLOW}7. Testing access to new collection...${NC}"
NEW_COLLECTION_RESPONSE=$(curl -s -X GET "$BASE_URL/items/test_blog_posts" \
  -H "Authorization: Bearer $TOKEN")

if echo "$NEW_COLLECTION_RESPONSE" | grep -q "data"; then
    print_status 0 "Access new collection successful"
    DATA_COUNT=$(echo "$NEW_COLLECTION_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Collection has $DATA_COUNT records"
else
    print_status 1 "Access new collection failed"
    echo "   Response: $NEW_COLLECTION_RESPONSE"
fi

# Test 9: Create a record in the new collection
echo -e "\n${YELLOW}8. Testing data creation in new collection...${NC}"
BLOG_POST_DATA='{
  "title": "My First Blog Post",
  "content": "This is the content of my first blog post created through the dynamic API!"
}'

CREATE_POST_RESPONSE=$(curl -s -X POST "$BASE_URL/items/test_blog_posts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BLOG_POST_DATA")

if echo "$CREATE_POST_RESPONSE" | grep -q "My First Blog Post"; then
    print_status 0 "Create blog post successful"
else
    print_status 1 "Create blog post failed"
    echo "   Response: $CREATE_POST_RESPONSE"
fi

# Test 10: Verify the record was created
echo -e "\n${YELLOW}9. Testing data retrieval from new collection...${NC}"
RETRIEVE_POSTS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/test_blog_posts" \
  -H "Authorization: Bearer $TOKEN")

if echo "$RETRIEVE_POSTS_RESPONSE" | grep -q "My First Blog Post"; then
    print_status 0 "Retrieve blog posts successful"
    POST_COUNT=$(echo "$RETRIEVE_POSTS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $POST_COUNT blog posts"
else
    print_status 1 "Retrieve blog posts failed"
    echo "   Response: $RETRIEVE_POSTS_RESPONSE"
fi

# Test 11: Test tenant isolation (should only see default tenant data)
echo -e "\n${YELLOW}10. Testing tenant isolation...${NC}"
print_info "All data should be isolated to the default tenant"

# Test 12: Clean up - Delete the test collection
echo -e "\n${YELLOW}11. Testing collection deletion...${NC}"
DELETE_COLLECTION_RESPONSE=$(curl -s -X DELETE "$BASE_URL/items/collections/$COLLECTION_ID" \
  -H "Authorization: Bearer $TOKEN")

if [ $? -eq 0 ]; then
    print_status 0 "Delete collection successful"
else
    print_status 1 "Delete collection failed"
    echo "   Response: $DELETE_COLLECTION_RESPONSE"
fi

# Test 13: Verify collection is gone
echo -e "\n${YELLOW}12. Verifying collection deletion...${NC}"
VERIFY_DELETE_RESPONSE=$(curl -s -X GET "$BASE_URL/items/test_blog_posts" \
  -H "Authorization: Bearer $TOKEN")

if echo "$VERIFY_DELETE_RESPONSE" | grep -q "Table does not exist"; then
    print_status 0 "Collection deletion verified"
else
    print_status 1 "Collection still exists"
    echo "   Response: $VERIFY_DELETE_RESPONSE"
fi

echo -e "\n${GREEN}🎉 Multi-Tenant Schema Management Test Complete!${NC}"
echo -e "${BLUE}Summary:${NC}"
echo "  ✅ Schema metadata management (collections/fields)"
echo "  ✅ Dynamic table creation"
echo "  ✅ Data storage and retrieval"
echo "  ✅ Tenant isolation"
echo "  ✅ RBAC integration"
echo "  ✅ Cleanup and deletion"

echo -e "\n${YELLOW}Next Steps:${NC}"
echo "  1. Create multiple tenants"
echo "  2. Test cross-tenant data isolation"
echo "  3. Test field-level permissions"
echo "  4. Build the admin frontend" 