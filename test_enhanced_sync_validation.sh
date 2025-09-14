#!/bin/bash

# Enhanced Sync Validation Testing Script
# Tests sync operations with valid data references and enhanced error messaging

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_HOST="localhost"
SERVER_PORT="8080"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"
MAX_RETRIES=10
RETRY_DELAY=3

echo -e "${BLUE}=== ENHANCED SYNC VALIDATION TESTING ===${NC}"
echo "Testing sync operations with:"
echo "1. Valid data references using seeded database entities"
echo "2. Enhanced error messaging for missing vs inaccessible entities"
echo "3. Role-based access control validation"
echo "4. Comprehensive filtering logic verification"
echo

# Function to wait for server
wait_for_server() {
    echo -e "${YELLOW}Waiting for server to be ready...${NC}"
    for i in $(seq 1 $MAX_RETRIES); do
        if curl -s "$BASE_URL/api/v1/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Server is ready${NC}"
            return 0
        fi
        echo "Attempt $i/$MAX_RETRIES - Server not ready, waiting $RETRY_DELAY seconds..."
        sleep $RETRY_DELAY
    done
    echo -e "${RED}❌ Server is not responding after $((MAX_RETRIES * RETRY_DELAY)) seconds${NC}"
    return 1
}

# Function to test authentication
test_authentication() {
    local username=$1
    local password=$2
    local expected_role=$3
    
    echo -e "${BLUE}Testing authentication for $username (expected role: $expected_role)...${NC}"
    
    local auth_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Authentication request failed for $username${NC}"
        return 1
    fi
    
    local token=$(echo "$auth_response" | jq -r '.data.token // empty')
    if [[ -z "$token" || "$token" == "null" ]]; then
        echo -e "${RED}❌ Failed to get token for $username${NC}"
        echo "Response: $auth_response"
        return 1
    fi
    
    echo -e "${GREEN}✅ Authentication successful for $username${NC}"
    echo "$token"
}

# Function to create sync data with VALID references (using correct seeded IDs)
create_owner_sync_data_valid() {
    cat << 'EOF'
{
    "carts": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440001",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
            "user_id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
            "quantity": 2,
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z"
        }
    ],
    "products": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440002",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "name": "Test Product Owner Valid",
            "sale": 25000.0,
            "buy": 18000.0,
            "stock": 40,
            "is_schedule": false,
            "is_have_stock": true,
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z"
        }
    ],
    "stock_histories": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440004",
            "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
            "stock": 100,
            "last_stock": 50,
            "stocked_at": "2024-01-15T10:00:00Z",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z"
        }
    ],
    "transactions": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440020",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "cashier_id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
            "total_price": 50000.0,
            "status": "completed",
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        }
    ],
    "transaction_products": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440021",
            "transaction_id": "550e8400-e29b-41d4-a716-446655440020",
            "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
            "quantity": 2,
            "unit_price": 25000.0,
            "total_price": 50000.0,
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        }
    ]
}
EOF
}

# Function to create sync data with INVALID references (to test error messaging)
create_sync_data_with_missing_references() {
    cat << 'EOF'
{
    "stock_histories": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440030",
            "product_id": "99999999-9999-9999-9999-999999999999",
            "stock": 100,
            "last_stock": 50,
            "stocked_at": "2024-01-15T10:00:00Z",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z"
        }
    ],
    "transaction_products": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440031",
            "transaction_id": "99999999-9999-9999-9999-999999999999",
            "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
            "quantity": 1,
            "unit_price": 25000.0,
            "total_price": 25000.0,
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        }
    ]
}
EOF
}

# Function to create cashier sync data with cross-shop references
create_cashier_sync_data_with_cross_shop() {
    cat << 'EOF'
{
    "carts": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440040",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "product_id": "11111111-2222-bbbb-bbbb-aaaaaaaaaaaa",
            "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
            "quantity": 1,
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        },
        {
            "id": "550e8400-e29b-41d4-a716-446655440041", 
            "shop_id": "11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
            "product_id": "11111111-1111-aaaa-aaaa-aaaaaaaaaaaa",
            "user_id": "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
            "quantity": 3,
            "created_at": "2024-01-15T11:00:00Z",
            "updated_at": "2024-01-15T11:00:00Z"
        }
    ],
    "stock_histories": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440042",
            "product_id": "11111111-1111-aaaa-aaaa-aaaaaaaaaaaa",
            "stock": 75,
            "last_stock": 50,
            "stocked_at": "2024-01-15T10:00:00Z",
            "created_at": "2024-01-15T10:00:00Z",
            "updated_at": "2024-01-15T10:00:00Z"
        }
    ]
}
EOF
}

# Function to test sync operation
test_sync_operation() {
    local token=$1
    local user_type=$2
    local sync_data=$3
    local test_description=$4
    
    echo -e "${BLUE}Testing sync operation: $test_description${NC}"
    
    local sync_response=$(curl -s -X POST "$BASE_URL/api/v1/sync" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$sync_data")
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Sync request failed for $user_type${NC}"
        return 1
    fi
    
    local status=$(echo "$sync_response" | jq -r '.status // "unknown"')
    
    echo -e "${YELLOW}Response Status: $status${NC}"
    echo -e "${YELLOW}Response:${NC}"
    echo "$sync_response" | jq '.'
    
    # Check for enhanced error messaging
    local errors=$(echo "$sync_response" | jq -r '.errors[]? // empty')
    if [[ -n "$errors" ]]; then
        echo -e "${YELLOW}Enhanced Error Analysis:${NC}"
        echo "$sync_response" | jq -r '.errors[]? | "- \(.entity_type): \(.message)"'
    fi
    
    return 0
}

# Main test execution
main() {
    echo -e "${BLUE}Starting Enhanced Sync Validation Tests...${NC}"
    
    # Wait for server to be ready
    if ! wait_for_server; then
        exit 1
    fi
    
    # Test 1: Owner business with valid data
    echo -e "\n${BLUE}=== TEST 1: Owner Business - Valid Data References ===${NC}"
    OWNER_TOKEN=$(test_authentication "owner1@example.com" "password123" "owner_business")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Owner authentication failed${NC}"
        exit 1
    fi
    
    OWNER_SYNC_DATA=$(create_owner_sync_data_valid)
    test_sync_operation "$OWNER_TOKEN" "owner_business" "$OWNER_SYNC_DATA" "Owner with valid product/transaction references"
    
    # Test 2: Sync with missing references (to test enhanced error messaging)
    echo -e "\n${BLUE}=== TEST 2: Owner Business - Missing References (Error Messaging Test) ===${NC}"
    MISSING_REF_DATA=$(create_sync_data_with_missing_references)
    test_sync_operation "$OWNER_TOKEN" "owner_business" "$MISSING_REF_DATA" "Owner with missing product/transaction references"
    
    # Test 3: Cashier with cross-shop data (to test inaccessible shop filtering)
    echo -e "\n${BLUE}=== TEST 3: Cashier - Cross-Shop References (Access Control Test) ===${NC}"
    CASHIER_TOKEN=$(test_authentication "cashier1@example.com" "password123" "cashier")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Cashier authentication failed${NC}"
        exit 1
    fi
    
    CROSS_SHOP_DATA=$(create_cashier_sync_data_with_cross_shop)
    test_sync_operation "$CASHIER_TOKEN" "cashier" "$CROSS_SHOP_DATA" "Cashier with cross-shop references"
    
    echo -e "\n${GREEN}=== Enhanced Sync Validation Tests Completed ===${NC}"
    echo -e "${GREEN}✅ All test scenarios executed successfully${NC}"
    echo -e "${YELLOW}📋 Review the enhanced error messages to verify:${NC}"
    echo -e "${YELLOW}   - Missing entity references are properly identified${NC}"
    echo -e "${YELLOW}   - Inaccessible shop access is clearly distinguished${NC}"
    echo -e "${YELLOW}   - Role-based filtering works correctly${NC}"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required but not installed. Please install jq first.${NC}"
    exit 1
fi

# Run main test
main "$@"