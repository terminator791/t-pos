#!/bin/bash

# Test script for diagnosing critical sync issues
# This script will test both owner and cashier sync scenarios mentioned in the comment

BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== CRITICAL SYNC ISSUE TESTING ===${NC}"
echo "Testing the two critical errors mentioned:"
echo "1. Owner sync: SQL transaction abort error"
echo "2. Cashier sync: Validation error about inaccessible shop"
echo ""

# Function to login and get token
login() {
    local email=$1
    local password=$2
    echo -e "${YELLOW}Logging in as $email${NC}"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "$CONTENT_TYPE" \
        -d "{
            \"email\": \"$email\",
            \"password\": \"$password\"
        }")
    
    token=$(echo $response | jq -r '.data.token // empty')
    if [ -z "$token" ] || [ "$token" = "null" ]; then
        echo -e "${RED}❌ Login failed for $email${NC}"
        echo "Response: $response"
        return 1
    fi
    
    echo -e "${GREEN}✅ Login successful for $email${NC}"
    echo $token
}

# Function to test sync operation
test_sync() {
    local description=$1
    local token=$2
    local sync_data=$3
    
    echo -e "${YELLOW}Testing: $description${NC}"
    
    response=$(curl -s -X POST "$BASE_URL/api/v1/sync" \
        -H "$CONTENT_TYPE" \
        -H "Authorization: Bearer $token" \
        -d "$sync_data")
    
    echo "Response: $response"
    
    status=$(echo $response | jq -r '.status // empty')
    if [ "$status" = "success" ]; then
        echo -e "${GREEN}✅ $description - SUCCESS${NC}"
        return 0
    else
        echo -e "${RED}❌ $description - FAILED${NC}"
        echo "Error details: $(echo $response | jq -r '.errors // .message // "Unknown error"')"
        return 1
    fi
}

# Wait for server to be ready
echo "Waiting for server to be ready..."
for i in {1..30}; do
    if curl -s "$BASE_URL/api/v1/auth/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ Server is not responding after 30 seconds${NC}"
        exit 1
    fi
    sleep 1
done

echo ""
echo -e "${BLUE}=== TESTING OWNER BUSINESS SYNC ===${NC}"

# Login as owner business (should have access to license LIC-001-DEMO and its shops)
OWNER_TOKEN=$(login "owner1@example.com" "password123")
if [ -z "$OWNER_TOKEN" ]; then
    echo -e "${RED}❌ Failed to login as owner, skipping owner tests${NC}"
else
    # Test 1: Owner sync with empty data (minimal test)
    OWNER_SYNC_EMPTY='{
        "last_sync_timestamp": null,
        "carts": [],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 1: Owner sync with empty data ==="
    test_sync "Owner - Empty sync" "$OWNER_TOKEN" "$OWNER_SYNC_EMPTY"
    
    # Test 2: Owner sync with cart data from their license shops
    OWNER_SYNC_WITH_CART='{
        "last_sync_timestamp": null,
        "carts": [
            {
                "id": "12345678-1234-1234-1234-123456789012",
                "shop_id": "11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                "product_id": "11111111-1111-aaaa-aaaa-aaaaaaaaaaaa",
                "user_id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
                "quantity": 2,
                "created_at": "2024-01-01T10:00:00Z",
                "updated_at": "2024-01-01T10:00:00Z"
            }
        ],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 2: Owner sync with cart data ==="
    test_sync "Owner - Cart sync" "$OWNER_TOKEN" "$OWNER_SYNC_WITH_CART"
fi

echo ""
echo -e "${BLUE}=== TESTING CASHIER SYNC ===${NC}"

# Login as cashier (should only have access to their assigned shop)
CASHIER_TOKEN=$(login "cashier1@example.com" "password123")
if [ -z "$CASHIER_TOKEN" ]; then
    echo -e "${RED}❌ Failed to login as cashier, skipping cashier tests${NC}"
else
    # Test 3: Cashier sync with empty data
    CASHIER_SYNC_EMPTY='{
        "last_sync_timestamp": null,
        "carts": [],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 3: Cashier sync with empty data ==="
    test_sync "Cashier - Empty sync" "$CASHIER_TOKEN" "$CASHIER_SYNC_EMPTY"
    
    # Test 4: Cashier sync with cart data from their own shop (should work)
    CASHIER_SYNC_OWN_SHOP='{
        "last_sync_timestamp": null,
        "carts": [
            {
                "id": "87654321-4321-4321-4321-210987654321",
                "shop_id": "11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                "product_id": "11111111-1111-aaaa-aaaa-aaaaaaaaaaaa",
                "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
                "quantity": 1,
                "created_at": "2024-01-01T11:00:00Z",
                "updated_at": "2024-01-01T11:00:00Z"
            }
        ],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 4: Cashier sync with own shop cart ==="
    test_sync "Cashier - Own shop cart sync" "$CASHIER_TOKEN" "$CASHIER_SYNC_OWN_SHOP"
    
    # Test 5: Cashier sync with cart data from inaccessible shop (should fail)
    CASHIER_SYNC_OTHER_SHOP='{
        "last_sync_timestamp": null,
        "carts": [
            {
                "id": "87654321-4321-4321-4321-210987654322",
                "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
                "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
                "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
                "quantity": 1,
                "created_at": "2024-01-01T11:00:00Z",
                "updated_at": "2024-01-01T11:00:00Z"
            }
        ],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 5: Cashier sync with inaccessible shop cart (should fail) ==="
    test_sync "Cashier - Inaccessible shop cart sync" "$CASHIER_TOKEN" "$CASHIER_SYNC_OTHER_SHOP"
fi

echo ""
echo -e "${BLUE}=== TESTING ADDITIONAL EDGE CASES ===${NC}"

if [ ! -z "$CASHIER_TOKEN" ]; then
    # Test 6: Cashier sync with mixed data (some accessible, some not)
    CASHIER_SYNC_MIXED='{
        "last_sync_timestamp": null,
        "carts": [
            {
                "id": "11111111-1111-1111-1111-111111111111",
                "shop_id": "11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                "product_id": "11111111-1111-aaaa-aaaa-aaaaaaaaaaaa",
                "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
                "quantity": 1,
                "created_at": "2024-01-01T11:00:00Z",
                "updated_at": "2024-01-01T11:00:00Z"
            },
            {
                "id": "22222222-2222-2222-2222-222222222222",
                "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
                "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
                "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
                "quantity": 1,
                "created_at": "2024-01-01T11:00:00Z",
                "updated_at": "2024-01-01T11:00:00Z"
            }
        ],
        "categories": [],
        "products": [],
        "transactions": [],
        "expenses": [],
        "payments": [],
        "receipts": [],
        "histories": [],
        "shops": [],
        "stock_histories": [],
        "transaction_products": [],
        "users": []
    }'
    
    echo ""
    echo "=== Test 6: Cashier sync with mixed accessible/inaccessible data ==="
    test_sync "Cashier - Mixed shop cart sync" "$CASHIER_TOKEN" "$CASHIER_SYNC_MIXED"
fi

echo ""
echo -e "${BLUE}=== SYNC TESTING COMPLETE ===${NC}"
echo "Review the results above to identify specific issues"
echo "Common issues to check:"
echo "1. SQL transaction abort errors (owner sync)"
echo "2. Shop access validation errors (cashier sync)"
echo "3. Database connection/timeout issues"
echo "4. Authorization policy mismatches"

# Test sync info endpoint for debugging
echo ""
echo -e "${BLUE}=== SYNC INFO DEBUGGING ===${NC}"

if [ ! -z "$OWNER_TOKEN" ]; then
    echo "Owner sync info:"
    curl -s -X GET "$BASE_URL/api/v1/sync/info" \
        -H "Authorization: Bearer $OWNER_TOKEN" | jq .
fi

if [ ! -z "$CASHIER_TOKEN" ]; then
    echo ""
    echo "Cashier sync info:"
    curl -s -X GET "$BASE_URL/api/v1/sync/info" \
        -H "Authorization: Bearer $CASHIER_TOKEN" | jq .
fi