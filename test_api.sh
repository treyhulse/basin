#!/bin/bash

# Test script for Go RBAC API
# Make sure the server is running on localhost:8080

echo "🚀 Testing Go RBAC API"
echo "======================"

# Base URL
BASE_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Test 1: Health check
echo -e "\n${YELLOW}1. Testing health endpoint...${NC}"
curl -s "$BASE_URL/health" > /dev/null
print_status $? "Health check"

# Test 2: API documentation
echo -e "\n${YELLOW}2. Testing API documentation...${NC}"
curl -s "$BASE_URL/" > /dev/null
print_status $? "API documentation"

# Test 3: Login
echo -e "\n${YELLOW}3. Testing login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}')

if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    print_status 0 "Login successful"
    echo "   Token: ${TOKEN:0:20}..."
else
    print_status 1 "Login failed"
    echo "   Response: $LOGIN_RESPONSE"
    exit 1
fi

# Test 4: Get current user
echo -e "\n${YELLOW}4. Testing get current user...${NC}"
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/me" \
  -H "Authorization: Bearer $TOKEN")

if echo "$ME_RESPONSE" | grep -q "admin@example.com"; then
    print_status 0 "Get current user successful"
else
    print_status 1 "Get current user failed"
    echo "   Response: $ME_RESPONSE"
fi

# Test 5: Get products
echo -e "\n${YELLOW}5. Testing get products...${NC}"
PRODUCTS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/products" \
  -H "Authorization: Bearer $TOKEN")

if echo "$PRODUCTS_RESPONSE" | grep -q "data"; then
    print_status 0 "Get products successful"
    PRODUCT_COUNT=$(echo "$PRODUCTS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $PRODUCT_COUNT products"
else
    print_status 1 "Get products failed"
    echo "   Response: $PRODUCTS_RESPONSE"
fi

# Test 6: Get customers
echo -e "\n${YELLOW}6. Testing get customers...${NC}"
CUSTOMERS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/customers" \
  -H "Authorization: Bearer $TOKEN")

if echo "$CUSTOMERS_RESPONSE" | grep -q "data"; then
    print_status 0 "Get customers successful"
    CUSTOMER_COUNT=$(echo "$CUSTOMERS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $CUSTOMER_COUNT customers"
else
    print_status 1 "Get customers failed"
    echo "   Response: $CUSTOMERS_RESPONSE"
fi

# Test 7: Get orders
echo -e "\n${YELLOW}7. Testing get orders...${NC}"
ORDERS_RESPONSE=$(curl -s -X GET "$BASE_URL/items/orders" \
  -H "Authorization: Bearer $TOKEN")

if echo "$ORDERS_RESPONSE" | grep -q "data"; then
    print_status 0 "Get orders successful"
    ORDER_COUNT=$(echo "$ORDERS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    echo "   Found $ORDER_COUNT orders"
else
    print_status 1 "Get orders failed"
    echo "   Response: $ORDERS_RESPONSE"
fi

# Test 8: Create a test product (demo mode)
echo -e "\n${YELLOW}8. Testing create product (demo mode)...${NC}"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/items/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "description": "A test product created via API",
    "price": 29.99,
    "category": "Electronics",
    "stock_quantity": 100
  }')

if echo "$CREATE_RESPONSE" | grep -q "demo mode"; then
    print_status 0 "Create product (demo mode) successful"
else
    print_status 1 "Create product failed"
    echo "   Response: $CREATE_RESPONSE"
fi

echo -e "\n${GREEN}🎉 API testing completed!${NC}"
echo -e "\n${YELLOW}Note: Create/Update/Delete operations are in demo mode and don't actually modify the database.${NC}"
echo -e "${YELLOW}To see the full API documentation, visit: $BASE_URL${NC}" 