#!/bin/bash

# Test script to validate ACL security for CREATE/POST operations
# This script tests authorization bypass prevention for endpoints that accept shop_id in request body

set -e

echo "🔒 Testing ACL Security for CREATE/POST Operations"
echo "=================================================="

# Configuration
BASE_URL="http://localhost:8080/api/v1"
SHOP1_ID="11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"  # Cashier1's shop (Jakarta)
SHOP2_ID="22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"  # Cashier2's shop (Bandung)
LICENSE1_ID="11111111-1111-1111-1111-111111111111"  # LIC-001-DEMO
LICENSE2_ID="22222222-2222-2222-2222-222222222222"  # LIC-002-DEMO

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print test results
print_result() {
    local test_name="$1"
    local expected="$2" 
    local actual="$3"
    local response="$4"
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}✅ PASS${NC}: $test_name (Status: $actual)"
    else
        echo -e "${RED}❌ FAIL${NC}: $test_name (Expected: $expected, Got: $actual)"
        echo -e "${YELLOW}Response: $response${NC}"
    fi
}

# Function to login and get JWT token
login_user() {
    local email="$1"
    local password="$2"
    
    local response=$(curl -s -X POST "$BASE_URL/auth/owner/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")
    
    echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4
}

# Function to login cashier and get JWT token  
login_cashier() {
    local email="$1"
    local pin="$2"
    
    local response=$(curl -s -X POST "$BASE_URL/auth/cashier/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"pin\":\"$pin\"}")
    
    echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4
}

# Function to test API endpoint
test_endpoint() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local token="$4"
    local data="$5"
    local expected_status="$6"
    
    local response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" "$BASE_URL$endpoint" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "$data")
    
    local body=$(echo "$response" | sed -E 's/HTTPSTATUS:[0-9]{3}$//')
    local status=$(echo "$response" | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    print_result "$test_name" "$expected_status" "$status" "$body"
}

echo -e "${BLUE}Step 1: Starting backend server...${NC}"
# Note: In a real test, you would start the server here
# For this demo, assume server is already running

echo -e "${BLUE}Step 2: Getting authentication tokens...${NC}"

# Get tokens for different users
SUPER_ADMIN_TOKEN=$(login_user "superadmin@example.com" "password123")
OWNER1_TOKEN=$(login_user "owner1@example.com" "password123")  # License 1
OWNER2_TOKEN=$(login_user "owner2@example.com" "password123")  # License 2
CASHIER1_TOKEN=$(login_cashier "cashier1@example.com" "123456")  # Shop 1
CASHIER2_TOKEN=$(login_cashier "cashier2@example.com" "123456")  # Shop 2

if [ -z "$SUPER_ADMIN_TOKEN" ] || [ -z "$OWNER1_TOKEN" ] || [ -z "$OWNER2_TOKEN" ] || [ -z "$CASHIER1_TOKEN" ] || [ -z "$CASHIER2_TOKEN" ]; then
    echo -e "${RED}❌ Failed to get authentication tokens. Please ensure the server is running and seeded.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All authentication tokens obtained successfully${NC}"

echo -e "\n${BLUE}Step 3: Testing Product Creation Authorization${NC}"
echo "=============================================="

# Test 1: Cashier1 creating product for their own shop (should succeed)
test_endpoint \
    "Cashier1 → Create Product for Own Shop1" \
    "POST" \
    "/products" \
    "$CASHIER1_TOKEN" \
    "{\"name\":\"Test Product 1\",\"sale\":10000,\"buy\":5000,\"shop_id\":\"$SHOP1_ID\",\"stock_quantity\":10}" \
    "201"

# Test 2: Cashier1 trying to create product for Shop2 (should fail - 403)
test_endpoint \
    "Cashier1 → Create Product for Other Shop2" \
    "POST" \
    "/products" \
    "$CASHIER1_TOKEN" \
    "{\"name\":\"Test Product 2\",\"sale\":10000,\"buy\":5000,\"shop_id\":\"$SHOP2_ID\",\"stock_quantity\":10}" \
    "403"

# Test 3: Owner1 creating product for shop under their license (should succeed)
test_endpoint \
    "Owner1 → Create Product for Shop1 (Under License1)" \
    "POST" \
    "/products" \
    "$OWNER1_TOKEN" \
    "{\"name\":\"Test Product 3\",\"sale\":15000,\"buy\":8000,\"shop_id\":\"$SHOP1_ID\",\"stock_quantity\":20}" \
    "201"

# Test 4: Owner1 trying to create product for shop under different license (should fail - 403)
test_endpoint \
    "Owner1 → Create Product for Shop2 (Different License)" \
    "POST" \
    "/products" \
    "$OWNER1_TOKEN" \
    "{\"name\":\"Test Product 4\",\"sale\":15000,\"buy\":8000,\"shop_id\":\"$SHOP2_ID\",\"stock_quantity\":20}" \
    "403"

echo -e "\n${BLUE}Step 4: Testing Category Creation Authorization${NC}"
echo "=============================================="

# Test 5: Cashier2 creating category for their own shop (should succeed)
test_endpoint \
    "Cashier2 → Create Category for Own Shop2" \
    "POST" \
    "/categories" \
    "$CASHIER2_TOKEN" \
    "{\"name\":\"Test Category 1\",\"shop_id\":\"$SHOP2_ID\"}" \
    "201"

# Test 6: Cashier2 trying to create category for Shop1 (should fail - 403)
test_endpoint \
    "Cashier2 → Create Category for Other Shop1" \
    "POST" \
    "/categories" \
    "$CASHIER2_TOKEN" \
    "{\"name\":\"Test Category 2\",\"shop_id\":\"$SHOP1_ID\"}" \
    "403"

# Test 7: Owner2 creating category for shop under their license (should succeed)
test_endpoint \
    "Owner2 → Create Category for Shop2 (Under License2)" \
    "POST" \
    "/categories" \
    "$OWNER2_TOKEN" \
    "{\"name\":\"Test Category 3\",\"shop_id\":\"$SHOP2_ID\"}" \
    "201"

echo -e "\n${BLUE}Step 5: Testing Cart Operations Authorization${NC}"
echo "==========================================="

# Assuming we have some product IDs to use
PRODUCT1_ID="11111111-1111-aaaa-aaaa-aaaaaaaaaaaa"
PRODUCT5_ID="11111111-2222-bbbb-bbbb-aaaaaaaaaaaa"

# Test 8: Cashier1 adding product from their shop to cart (should succeed)
test_endpoint \
    "Cashier1 → Add Product from Own Shop1 to Cart" \
    "POST" \
    "/carts" \
    "$CASHIER1_TOKEN" \
    "{\"product_id\":\"$PRODUCT1_ID\",\"shop_id\":\"$SHOP1_ID\",\"quantity\":2}" \
    "201"

# Test 9: Cashier1 trying to add product from different shop to cart (should fail - 403)
test_endpoint \
    "Cashier1 → Add Product from Other Shop2 to Cart" \
    "POST" \
    "/carts" \
    "$CASHIER1_TOKEN" \
    "{\"product_id\":\"$PRODUCT5_ID\",\"shop_id\":\"$SHOP2_ID\",\"quantity\":1}" \
    "403"

echo -e "\n${BLUE}Step 6: Testing Transaction Creation Authorization${NC}"
echo "=============================================="

# Test 10: Cashier1 creating transaction (should use their shop automatically)
test_endpoint \
    "Cashier1 → Create Transaction (Auto Shop Assignment)" \
    "POST" \
    "/transactions" \
    "$CASHIER1_TOKEN" \
    "{\"customer_name\":\"Test Customer 1\",\"items\":[{\"product_id\":\"$PRODUCT1_ID\",\"quantity\":1}]}" \
    "201"

# Test 11: Owner1 creating transaction for shop under their license (should succeed)
test_endpoint \
    "Owner1 → Create Transaction for Shop1 (Under License1)" \
    "POST" \
    "/transactions" \
    "$OWNER1_TOKEN" \
    "{\"shop_id\":\"$SHOP1_ID\",\"customer_name\":\"Test Customer 2\",\"items\":[{\"product_id\":\"$PRODUCT1_ID\",\"quantity\":1}]}" \
    "201"

# Test 12: Owner1 trying to create transaction for shop under different license (should fail - 403)
test_endpoint \
    "Owner1 → Create Transaction for Shop2 (Different License)" \
    "POST" \
    "/transactions" \
    "$OWNER1_TOKEN" \
    "{\"shop_id\":\"$SHOP2_ID\",\"customer_name\":\"Test Customer 3\",\"items\":[{\"product_id\":\"$PRODUCT5_ID\",\"quantity\":1}]}" \
    "403"

echo -e "\n${BLUE}Step 7: Testing Shop Creation Authorization${NC}"
echo "========================================"

# Test 13: Owner1 creating shop under their license (should succeed)
test_endpoint \
    "Owner1 → Create Shop under License1" \
    "POST" \
    "/shops" \
    "$OWNER1_TOKEN" \
    "{\"name\":\"Test Shop License1\",\"license_id\":\"$LICENSE1_ID\",\"user_id\":\"cccccccc-cccc-cccc-cccc-cccccccccccc\",\"domain\":\"shop-test1\"}" \
    "201"

# Test 14: Owner1 trying to create shop under different license (should fail - 403)
test_endpoint \
    "Owner1 → Create Shop under License2 (Unauthorized)" \
    "POST" \
    "/shops" \
    "$OWNER1_TOKEN" \
    "{\"name\":\"Test Shop License2\",\"license_id\":\"$LICENSE2_ID\",\"user_id\":\"cccccccc-cccc-cccc-cccc-cccccccccccc\",\"domain\":\"shop-test2\"}" \
    "403"

# Test 15: Super Admin can create anything (should succeed)
test_endpoint \
    "Super Admin → Create Shop under Any License" \
    "POST" \
    "/shops" \
    "$SUPER_ADMIN_TOKEN" \
    "{\"name\":\"Super Admin Test Shop\",\"license_id\":\"$LICENSE2_ID\",\"user_id\":\"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa\",\"domain\":\"shop-super-test\"}" \
    "201"

echo -e "\n${BLUE}Step 8: Testing Authorization Summary${NC}"
echo "===================================="

echo -e "${GREEN}✅ Authorization Bypass Prevention Tests Completed${NC}"
echo -e "${YELLOW}Expected Results:${NC}"
echo "- Users can only create resources for shops they have access to"
echo "- Cashiers: Only their assigned shop"
echo "- Owner Business: Only shops under their license"
echo "- Super Admin/Admin: All shops (global access)"
echo "- All unauthorized attempts should return 403 Forbidden"

echo -e "\n${BLUE}🔒 ACL Security Validation Complete${NC}"
echo "=================================="
echo "If all tests show expected results, ACL security is properly implemented."
echo "Any FAIL results indicate authorization bypass vulnerabilities that need fixing."