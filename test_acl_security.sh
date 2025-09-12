#!/bin/bash
# ACL Security Validation Script
# Tests domain-specific filtering and authorization bypass prevention

# Configuration
BASE_URL="http://localhost:8080/api/v1"
TOKEN_SUPER_ADMIN=""
TOKEN_ADMIN=""
TOKEN_OWNER1=""
TOKEN_OWNER2=""
TOKEN_CASHIER1=""
TOKEN_CASHIER2=""

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== T-POS ACL Security Validation Test ===${NC}"
echo ""

# Function to make authenticated API request
make_request() {
    local method=$1
    local endpoint=$2
    local token=$3
    local expected_status=$4
    local description=$5
    
    echo -e "${YELLOW}Testing: $description${NC}"
    echo "  → $method $endpoint"
    
    if [ -z "$token" ]; then
        echo -e "  ${RED}✗ SKIP - Token not available${NC}"
        return
    fi
    
    response=$(curl -s -o /dev/null -w "%{http_code}" \
        -X "$method" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        "$BASE_URL$endpoint")
    
    if [ "$response" = "$expected_status" ]; then
        echo -e "  ${GREEN}✓ SUCCESS - Status: $response (Expected: $expected_status)${NC}"
    else
        echo -e "  ${RED}✗ FAILED - Status: $response (Expected: $expected_status)${NC}"
    fi
    echo ""
}

# Function to get authentication tokens
get_tokens() {
    echo -e "${BLUE}Getting authentication tokens...${NC}"
    
    # Login as super admin (if exists)
    # TOKEN_SUPER_ADMIN=$(curl -s -X POST "$BASE_URL/auth/owner/login" \
    #     -H "Content-Type: application/json" \
    #     -d '{"email":"superadmin@example.com","password":"password123"}' | \
    #     jq -r '.data.token // empty')
    
    # For testing, we'll use sample tokens - replace with actual login
    echo -e "  ${YELLOW}Note: Please provide actual authentication tokens${NC}"
    echo -e "  ${YELLOW}This script template shows the testing structure${NC}"
    echo ""
}

# Function to test list endpoints with domain filtering
test_list_endpoints() {
    echo -e "${BLUE}=== Testing List Endpoints Domain Filtering ===${NC}"
    echo ""
    
    # Test Categories List
    echo -e "${YELLOW}Categories List Access:${NC}"
    make_request "GET" "/categories" "$TOKEN_SUPER_ADMIN" "200" "Super Admin can list all categories"
    make_request "GET" "/categories" "$TOKEN_ADMIN" "200" "Admin can list all categories"
    make_request "GET" "/categories" "$TOKEN_OWNER1" "200" "Owner1 can list their license categories"
    make_request "GET" "/categories" "$TOKEN_OWNER2" "200" "Owner2 can list their license categories"
    make_request "GET" "/categories" "$TOKEN_CASHIER1" "200" "Cashier1 can list their shop categories"
    make_request "GET" "/categories" "$TOKEN_CASHIER2" "200" "Cashier2 can list their shop categories"
    
    # Test Products List
    echo -e "${YELLOW}Products List Access:${NC}"
    make_request "GET" "/products" "$TOKEN_SUPER_ADMIN" "200" "Super Admin can list all products"
    make_request "GET" "/products" "$TOKEN_ADMIN" "200" "Admin can list all products"
    make_request "GET" "/products" "$TOKEN_OWNER1" "200" "Owner1 can list their license products"
    make_request "GET" "/products" "$TOKEN_OWNER2" "200" "Owner2 can list their license products"
    make_request "GET" "/products" "$TOKEN_CASHIER1" "200" "Cashier1 can list their shop products"
    make_request "GET" "/products" "$TOKEN_CASHIER2" "200" "Cashier2 can list their shop products"
    
    # Test Shops List
    echo -e "${YELLOW}Shops List Access:${NC}"
    make_request "GET" "/shops" "$TOKEN_SUPER_ADMIN" "200" "Super Admin can list all shops"
    make_request "GET" "/shops" "$TOKEN_ADMIN" "200" "Admin can list all shops"
    make_request "GET" "/shops" "$TOKEN_OWNER1" "200" "Owner1 can list their license shops"
    make_request "GET" "/shops" "$TOKEN_OWNER2" "200" "Owner2 can list their license shops"
    make_request "GET" "/shops" "$TOKEN_CASHIER1" "200" "Cashier1 can list their assigned shop"
    make_request "GET" "/shops" "$TOKEN_CASHIER2" "200" "Cashier2 can list their assigned shop"
}

# Function to test cross-tenant access attempts
test_cross_tenant_access() {
    echo -e "${BLUE}=== Testing Cross-Tenant Access Prevention ===${NC}"
    echo ""
    
    # Shop-specific endpoint tests
    local SHOP1_ID="11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
    local SHOP2_ID="22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
    
    echo -e "${YELLOW}Shop-Specific Endpoint Access:${NC}"
    # Cashier1 should access Shop1 but not Shop2
    make_request "GET" "/transactions/shop/$SHOP1_ID" "$TOKEN_CASHIER1" "200" "Cashier1 can access Shop1 transactions"
    make_request "GET" "/transactions/shop/$SHOP2_ID" "$TOKEN_CASHIER1" "403" "Cashier1 CANNOT access Shop2 transactions"
    
    # Cashier2 should access Shop2 but not Shop1
    make_request "GET" "/transactions/shop/$SHOP2_ID" "$TOKEN_CASHIER2" "200" "Cashier2 can access Shop2 transactions"
    make_request "GET" "/transactions/shop/$SHOP1_ID" "$TOKEN_CASHIER2" "403" "Cashier2 CANNOT access Shop1 transactions"
    
    # Similar tests for other shop-specific endpoints
    make_request "GET" "/products/low-stock?shop_id=$SHOP1_ID" "$TOKEN_CASHIER1" "200" "Cashier1 can get Shop1 low stock"
    make_request "GET" "/products/low-stock?shop_id=$SHOP2_ID" "$TOKEN_CASHIER1" "403" "Cashier1 CANNOT get Shop2 low stock"
    
    make_request "GET" "/categories?shop_id=$SHOP2_ID" "$TOKEN_CASHIER2" "200" "Cashier2 can get Shop2 categories"
    make_request "GET" "/categories?shop_id=$SHOP1_ID" "$TOKEN_CASHIER2" "403" "Cashier2 CANNOT get Shop1 categories"
}

# Function to test individual resource access
test_individual_resource_access() {
    echo -e "${BLUE}=== Testing Individual Resource Access ===${NC}"
    echo ""
    
    # Test with sample resource IDs (replace with actual IDs from seeded data)
    local PRODUCT1_ID="11111111-1111-aaaa-aaaa-aaaaaaaaaaaa"  # Product from Shop1
    local PRODUCT2_ID="11111111-2222-bbbb-bbbb-aaaaaaaaaaaa"  # Product from Shop2
    local CATEGORY1_ID="aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa" # Category from Shop1
    local CATEGORY2_ID="aaaaaaaa-2222-2222-2222-aaaaaaaaaaaa" # Category from Shop2
    
    echo -e "${YELLOW}Individual Product Access:${NC}"
    make_request "GET" "/products/$PRODUCT1_ID" "$TOKEN_CASHIER1" "200" "Cashier1 can access Shop1 product"
    make_request "GET" "/products/$PRODUCT2_ID" "$TOKEN_CASHIER1" "403" "Cashier1 CANNOT access Shop2 product"
    
    make_request "GET" "/products/$PRODUCT2_ID" "$TOKEN_CASHIER2" "200" "Cashier2 can access Shop2 product"
    make_request "GET" "/products/$PRODUCT1_ID" "$TOKEN_CASHIER2" "403" "Cashier2 CANNOT access Shop1 product"
    
    echo -e "${YELLOW}Individual Category Access:${NC}"
    make_request "GET" "/categories/$CATEGORY1_ID" "$TOKEN_CASHIER1" "200" "Cashier1 can access Shop1 category"
    make_request "GET" "/categories/$CATEGORY2_ID" "$TOKEN_CASHIER1" "403" "Cashier1 CANNOT access Shop2 category"
    
    make_request "GET" "/categories/$CATEGORY2_ID" "$TOKEN_CASHIER2" "200" "Cashier2 can access Shop2 category"
    make_request "GET" "/categories/$CATEGORY1_ID" "$TOKEN_CASHIER2" "403" "Cashier2 CANNOT access Shop1 category"
}

# Function to test search functionality
test_search_functionality() {
    echo -e "${BLUE}=== Testing Search Functionality ===${NC}"
    echo ""
    
    local SHOP1_ID="11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
    local SHOP2_ID="22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
    
    echo -e "${YELLOW}Product Search Access:${NC}"
    make_request "GET" "/products/search?q=test&shop_id=$SHOP1_ID" "$TOKEN_CASHIER1" "200" "Cashier1 can search in Shop1"
    make_request "GET" "/products/search?q=test&shop_id=$SHOP2_ID" "$TOKEN_CASHIER1" "403" "Cashier1 CANNOT search in Shop2"
    
    make_request "GET" "/products/search?q=test&shop_id=$SHOP2_ID" "$TOKEN_CASHIER2" "200" "Cashier2 can search in Shop2"
    make_request "GET" "/products/search?q=test&shop_id=$SHOP1_ID" "$TOKEN_CASHIER2" "403" "Cashier2 CANNOT search in Shop1"
}

# Main execution
main() {
    echo -e "${BLUE}Starting ACL Security Validation...${NC}"
    echo ""
    
    # Get authentication tokens
    get_tokens
    
    # Run test suites
    test_list_endpoints
    test_cross_tenant_access
    test_individual_resource_access
    test_search_functionality
    
    echo -e "${BLUE}=== ACL Security Validation Complete ===${NC}"
    echo ""
    echo -e "${YELLOW}Summary:${NC}"
    echo "- List endpoints should filter data by user's accessible shops/licenses"
    echo "- Cashiers should only access their assigned shop"
    echo "- Owner_business should only access shops under their license"
    echo "- Cross-tenant access should be blocked with 403 Forbidden"
    echo "- Individual resource access should be validated"
    echo ""
    echo -e "${YELLOW}To run actual tests:${NC}"
    echo "1. Start the T-POS backend server: make run-backend"
    echo "2. Seed the database: make seed"
    echo "3. Get actual authentication tokens by logging in"
    echo "4. Update the TOKEN variables in this script"
    echo "5. Run this script again"
}

# Check if server is running
check_server() {
    echo -e "${YELLOW}Checking if server is running...${NC}"
    if curl -s "$BASE_URL/../health" > /dev/null; then
        echo -e "${GREEN}✓ Server is running${NC}"
        return 0
    else
        echo -e "${RED}✗ Server is not running. Please start with: make run-backend${NC}"
        return 1
    fi
}

# Execute main function if server is available
if check_server; then
    main
fi