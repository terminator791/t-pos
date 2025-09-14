#!/bin/bash

# Comprehensive Sync Critical Debugging Script
# Tests sync operations with detailed database validation and error analysis

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

echo -e "${BLUE}=== SYNC CRITICAL DEBUGGING TESTS ===${NC}"
echo "This script will:"
echo "1. Verify database state and entity relationships"
echo "2. Test sync operations with problematic data"
echo "3. Provide detailed error analysis and debugging info"
echo "4. Suggest fixes for any identified issues"
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

# Function to test authentication and get user details
test_authentication() {
    local email=$1
    local password=$2
    
    echo -e "${BLUE}Testing authentication for $email${NC}"
    
    local auth_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Authentication request failed${NC}"
        return 1
    fi
    
    local status=$(echo "$auth_response" | jq -r '.status // "unknown"')
    if [[ "$status" != "success" ]]; then
        echo -e "${RED}❌ Authentication failed: $status${NC}"
        echo "$auth_response" | jq '.'
        return 1
    fi
    
    local token=$(echo "$auth_response" | jq -r '.data.token')
    local user_id=$(echo "$auth_response" | jq -r '.data.user.id')
    local role=$(echo "$auth_response" | jq -r '.data.user.role')
    local license_id=$(echo "$auth_response" | jq -r '.data.user.license_id')
    local shop_id=$(echo "$auth_response" | jq -r '.data.user.shop_id // "null"')
    
    echo -e "${GREEN}✅ Authentication successful${NC}"
    echo -e "${YELLOW}User Details:${NC}"
    echo "  - User ID: $user_id"
    echo "  - Role: $role"
    echo "  - License ID: $license_id"
    echo "  - Shop ID: $shop_id"
    
    # Return token for use in other requests
    echo "$token"
}

# Function to create test sync data that matches the problematic scenario
create_problematic_sync_data() {
    cat << 'EOF'
{
    "last_sync_timestamp": null,
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

# Function to test sync operation with detailed analysis
test_sync_with_analysis() {
    local token=$1
    local user_type=$2
    local sync_data=$3
    local test_description=$4
    
    echo -e "${BLUE}=== Testing: $test_description ===${NC}"
    
    echo -e "${YELLOW}Sending sync request...${NC}"
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
    
    # Analyze the response in detail
    echo -e "${YELLOW}=== DETAILED RESPONSE ANALYSIS ===${NC}"
    
    # Check for filtering warnings
    local errors=$(echo "$sync_response" | jq -r '.errors[]? // empty')
    if [[ -n "$errors" ]]; then
        echo -e "${YELLOW}📋 Filtering/Error Analysis:${NC}"
        echo "$sync_response" | jq -r '.errors[]? | "  - Entity: \(.entity_type), Code: \(.error_code), Message: \(.message)"'
        
        # Check specifically for the problematic filtering
        local stock_errors=$(echo "$sync_response" | jq -r '.errors[]? | select(.entity_type == "stock_histories") | .message')
        local transaction_errors=$(echo "$sync_response" | jq -r '.errors[]? | select(.entity_type == "transaction_products") | .message')
        
        if [[ -n "$stock_errors" ]]; then
            echo -e "${RED}🚨 STOCK HISTORIES FILTERING DETECTED:${NC}"
            echo "  $stock_errors"
        fi
        
        if [[ -n "$transaction_errors" ]]; then
            echo -e "${RED}🚨 TRANSACTION PRODUCTS FILTERING DETECTED:${NC}"
            echo "  $transaction_errors"
        fi
    else
        echo -e "${GREEN}✅ No filtering errors detected${NC}"
    fi
    
    # Show full response for debugging
    echo -e "${YELLOW}=== FULL RESPONSE ===${NC}"
    echo "$sync_response" | jq '.'
    
    return 0
}

# Main test execution
main() {
    echo -e "${BLUE}Starting Comprehensive Sync Debugging...${NC}"
    
    # Wait for server to be ready
    if ! wait_for_server; then
        exit 1
    fi
    
    # Test 1: Owner2 with Shop2 access (should work)
    echo -e "\n${BLUE}=== TEST 1: Owner2 with Shop2 Data (Should Work) ===${NC}"
    OWNER2_TOKEN=$(test_authentication "owner2@example.com" "password123")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Owner2 authentication failed${NC}"
        exit 1
    fi
    
    SYNC_DATA=$(create_problematic_sync_data)
    test_sync_with_analysis "$OWNER2_TOKEN" "owner2" "$SYNC_DATA" "Owner2 syncing Shop2 data - should succeed"
    
    # Test 2: Owner1 with Shop2 data (should filter)
    echo -e "\n${BLUE}=== TEST 2: Owner1 with Shop2 Data (Should Filter) ===${NC}"
    OWNER1_TOKEN=$(test_authentication "owner1@example.com" "password123")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Owner1 authentication failed${NC}"
        exit 1
    fi
    
    test_sync_with_analysis "$OWNER1_TOKEN" "owner1" "$SYNC_DATA" "Owner1 syncing Shop2 data - should filter correctly"
    
    # Test 3: Cashier2 with Shop2 data (should work)
    echo -e "\n${BLUE}=== TEST 3: Cashier2 with Shop2 Data (Should Work) ===${NC}"
    CASHIER2_TOKEN=$(test_authentication "cashier2@example.com" "password123")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Cashier2 authentication failed${NC}"
        exit 1
    fi
    
    test_sync_with_analysis "$CASHIER2_TOKEN" "cashier2" "$SYNC_DATA" "Cashier2 syncing Shop2 data - should succeed"
    
    # Test 4: Cashier1 with Shop2 data (should filter)
    echo -e "\n${BLUE}=== TEST 4: Cashier1 with Shop2 Data (Should Filter) ===${NC}"
    CASHIER1_TOKEN=$(test_authentication "cashier1@example.com" "password123")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Cashier1 authentication failed${NC}"
        exit 1
    fi
    
    test_sync_with_analysis "$CASHIER1_TOKEN" "cashier1" "$SYNC_DATA" "Cashier1 syncing Shop2 data - should filter correctly"
    
    echo -e "\n${GREEN}=== DEBUGGING TESTS COMPLETED ===${NC}"
    echo -e "${YELLOW}📋 Analysis Summary:${NC}"
    echo -e "${YELLOW}  - Tests 1 & 3 should succeed (Owner2 and Cashier2 have access to Shop2)${NC}"
    echo -e "${YELLOW}  - Tests 2 & 4 should filter properly (Owner1 and Cashier1 don't have access to Shop2)${NC}"
    echo -e "${YELLOW}  - If any unexpected filtering occurs, this indicates a bug that needs fixing${NC}"
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ jq is required but not installed. Please install jq first.${NC}"
    exit 1
fi

# Run main test
main "$@"