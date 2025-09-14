#!/bin/bash

# Test script to verify the sync filtering fix
# This tests the exact scenario from the problem statement

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== TESTING SYNC FILTERING FIX ===${NC}"
echo "Testing the exact scenario from the problem statement"
echo

# Configuration
SERVER_HOST="localhost"
SERVER_PORT="8080"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"

# Test data matching the problem statement exactly
TEST_SYNC_DATA=$(cat << 'EOF'
{
  "transactions": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440005",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "cashier_id": "ffffffff-ffff-ffff-ffff-ffffffffffff",
            "customer_name": "John Doe",
            "discount": 10.00,
            "discount_percentage": 5.0,
            "additional_cost": 0.00,
            "status": "completed",
            "total_price": 589.98,
            "profit_transaction": 289.98,
            "cashier_name": "Cashier 2",
            "change": 10.02,
            "amount": 600,
            "initial_payment_status": "paid",
            "created_at": "2026-01-15T12:00:00Z",
            "updated_at": "2026-01-15T12:05:00Z"
        }
    ],
    "transaction_products": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440006",
            "transaction_id": "550e8400-e29b-41d4-a716-446655440005",
            "product_id": "33333333-2222-bbbb-bbbb-cccccccccccc",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "quantity": 2,
            "unit_price": 25000.00,
            "total_price": 15000.00,
            "created_at": "2024-01-15T12:00:00Z",
            "updated_at": "2024-01-15T12:00:00Z"
        }
    ],
        "stock_histories": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440014",
            "shop_id": "22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
            "product_id": "22222222-2222-bbbb-bbbb-bbbbbbbbbbbb",
            "stock": 10,
            "last_stock": 50,
            "stocked_at": "2024-01-15T12:10:00Z",
            "created_at": "2024-01-15T12:10:00Z",
            "updated_at": "2024-01-15T12:10:00Z"
        }
    ]
}
EOF
)

# Function to wait for server
wait_for_server() {
    echo -e "${YELLOW}Waiting for server to be ready...${NC}"
    for i in {1..30}; do
        if curl -s "$BASE_URL/api/v1/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Server is ready${NC}"
            return 0
        fi
        echo "Attempt $i/30 - Server not ready, waiting 2 seconds..."
        sleep 2
    done
    echo -e "${RED}❌ Server is not responding after 60 seconds${NC}"
    return 1
}

# Function to authenticate and get token
authenticate() {
    local email=$1
    local password=$2
    
    echo -e "${BLUE}Authenticating as $email${NC}"
    
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
    
    echo -e "${GREEN}✅ Authentication successful${NC}"
    echo -e "${YELLOW}User Details:${NC}"
    echo "  - User ID: $user_id"
    echo "  - Role: $role"
    echo "  - License ID: $license_id"
    
    echo "$token"
}

# Function to test sync operation
test_sync() {
    local token=$1
    local description=$2
    
    echo -e "${BLUE}=== $description ===${NC}"
    
    local sync_response=$(curl -s -X POST "$BASE_URL/api/v1/sync" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$TEST_SYNC_DATA")
    
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Sync request failed${NC}"
        return 1
    fi
    
    local status=$(echo "$sync_response" | jq -r '.status // "unknown"')
    echo -e "${YELLOW}Sync Status: $status${NC}"
    
    # Check for the specific errors mentioned in the problem statement
    local stock_errors=$(echo "$sync_response" | jq -r '.data.errors[]? | select(.entity_type == "stock_histories") | .message')
    local transaction_errors=$(echo "$sync_response" | jq -r '.data.errors[]? | select(.entity_type == "transaction_products") | .message')
    
    echo -e "${YELLOW}=== ANALYSIS ===${NC}"
    
    if [[ -n "$stock_errors" ]]; then
        echo -e "${RED}❌ STOCK HISTORIES FILTERING DETECTED:${NC}"
        echo "  $stock_errors"
        echo -e "${RED}  This indicates the bug is still present!${NC}"
    else
        echo -e "${GREEN}✅ No stock history filtering errors${NC}"
    fi
    
    if [[ -n "$transaction_errors" ]]; then
        echo -e "${RED}❌ TRANSACTION PRODUCTS FILTERING DETECTED:${NC}"
        echo "  $transaction_errors"
        echo -e "${RED}  This indicates the bug is still present!${NC}"
    else
        echo -e "${GREEN}✅ No transaction product filtering errors${NC}"
    fi
    
    # Check if the referenced entities exist in the response
    local product_exists=$(echo "$sync_response" | jq -r '.data.products[]? | select(.id == "22222222-2222-bbbb-bbbb-bbbbbbbbbbbb") | .id')
    local transaction_exists=$(echo "$sync_response" | jq -r '.data.transactions[]? | select(.id == "550e8400-e29b-41d4-a716-446655440005") | .id')
    
    echo -e "${YELLOW}=== ENTITY VERIFICATION ===${NC}"
    
    if [[ -n "$product_exists" ]]; then
        echo -e "${GREEN}✅ Product 22222222-2222-bbbb-bbbb-bbbbbbbbbbbb exists in response${NC}"
    else
        echo -e "${YELLOW}⚠️  Product 22222222-2222-bbbb-bbbb-bbbbbbbbbbbb not in response${NC}"
    fi
    
    if [[ -n "$transaction_exists" ]]; then
        echo -e "${GREEN}✅ Transaction 550e8400-e29b-41d4-a716-446655440005 exists in response${NC}"
    else
        echo -e "${YELLOW}⚠️  Transaction 550e8400-e29b-41d4-a716-446655440005 not in response${NC}"
    fi
    
    # Show error details
    echo -e "${YELLOW}=== ERROR DETAILS ===${NC}"
    local errors=$(echo "$sync_response" | jq -r '.data.errors[]?')
    if [[ -n "$errors" ]]; then
        echo "$sync_response" | jq -r '.data.errors[]? | "  - Entity: \(.entity_type), Code: \(.error_code), Message: \(.message)"'
    else
        echo -e "${GREEN}✅ No errors reported${NC}"
    fi
    
    # Show partial response for verification
    echo -e "${YELLOW}=== PARTIAL RESPONSE ===${NC}"
    echo "$sync_response" | jq '{
        status: .status,
        message: .message,
        data: {
            sync_timestamp: .data.sync_timestamp,
            products: (.data.products // [] | length),
            transactions: (.data.transactions // [] | length),
            stock_histories: (.data.stock_histories // [] | length),
            transaction_products: (.data.transaction_products // [] | length),
            errors: (.data.errors // [] | length),
            stats: .data.stats
        }
    }'
    
    return 0
}

# Main test execution
main() {
    echo -e "${BLUE}Starting sync filtering fix verification...${NC}"
    
    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}❌ jq is required but not installed${NC}"
        exit 1
    fi
    
    # Wait for server to be ready
    if ! wait_for_server; then
        echo -e "${RED}❌ Server is not ready. Please start the backend server first.${NC}"
        echo "Run: cd backend && go run cmd/main.go"
        exit 1
    fi
    
    # Test with owner2 who should have access to the shop
    echo -e "\n${BLUE}=== TESTING WITH OWNER2 (Should have access to shop 22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb) ===${NC}"
    
    OWNER2_TOKEN=$(authenticate "owner2@example.com" "password123")
    if [[ $? -ne 0 ]]; then
        echo -e "${RED}❌ Owner2 authentication failed${NC}"
        echo "Note: This test requires the system to have seeded test data."
        echo "Make sure you've run the database migrations and seeding."
        exit 1
    fi
    
    test_sync "$OWNER2_TOKEN" "Testing sync with owner2 - should NOT have filtering errors"
    
    echo -e "\n${GREEN}=== TEST COMPLETED ===${NC}"
    echo -e "${YELLOW}Expected results:${NC}"
    echo -e "${YELLOW}  - No stock history filtering errors (product exists and is accessible)${NC}"
    echo -e "${YELLOW}  - No transaction product filtering errors (transaction exists and is accessible)${NC}"
    echo -e "${YELLOW}  - Both referenced entities should be present in the response${NC}"
}

# Run main test
main "$@"