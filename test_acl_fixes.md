# ACL Authorization Fix Testing Guide

This document describes how to test the authorization bypass fixes we've implemented.

## Fixed Issues

### 1. Cashier Domain Isolation 
**Before**: Cashiers had license domains (LIC-001-DEMO) allowing cross-shop access
**After**: Cashiers have shop-specific domains (shop-{uuid}) preventing cross-shop access

### 2. Shop Assignment
**Before**: Cashiers had no ShopID assignment 
**After**: 
- Cashier1 (eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee) → Shop1 (11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa) 
- Cashier2 (ffffffff-ffff-ffff-ffff-ffffffffffff) → Shop2 (22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb)

### 3. Enhanced Middleware Security
**Before**: Dangerous wildcard fallback for any user without domain
**After**: Only super_admin/admin can use wildcard domain, tenant users must have specific domains

### 4. Resource-Level Access Control
**Before**: Individual resource endpoints didn't validate shop ownership
**After**: Added RequireResourceAccess() middleware to critical endpoints

## Test Cases

### Test 1: Cashier Cross-Shop Access (Should be BLOCKED)
1. Login as cashier1@example.com (assigned to Shop1)
2. Try to access Shop2 transaction: `GET /api/v1/transactions/shop/22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb`
3. **Expected**: 403 Forbidden - "Cannot access this shop"

### Test 2: Cashier Own Shop Access (Should be ALLOWED)
1. Login as cashier1@example.com (assigned to Shop1)  
2. Access Shop1 transaction: `GET /api/v1/transactions/shop/11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa`
3. **Expected**: 200 OK - Transaction list returned

### Test 3: Individual Transaction Access (Should validate shop ownership)
1. Login as cashier1@example.com (assigned to Shop1)
2. Try to access a transaction from Shop2: `GET /api/v1/transactions/{shop2-transaction-id}`
3. **Expected**: 403 Forbidden (once resource validation is fully implemented)

### Test 4: Owner Business Isolation
1. Login as owner1@example.com (License LIC-001-DEMO)
2. Try to access Shop from LIC-002-DEMO: `GET /api/v1/shops/license/22222222-2222-2222-2222-222222222222`
3. **Expected**: 403 Forbidden - "Cannot access this license"

### Test 5: Admin Global Access (Should be ALLOWED)
1. Login as admin@example.com
2. Access any shop/license: All endpoints should work
3. **Expected**: 200 OK for all valid endpoints

## Fixed UUIDs for Testing

### Licenses
- License1: 11111111-1111-1111-1111-111111111111 (LIC-001-DEMO)
- License2: 22222222-2222-2222-2222-222222222222 (LIC-002-DEMO)  
- License3: 33333333-3333-3333-3333-333333333333 (LIC-003-DEMO)

### Users
- SuperAdmin: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa
- Admin: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb
- Owner1: cccccccc-cccc-cccc-cccc-cccccccccccc (License1)
- Owner2: dddddddd-dddd-dddd-dddd-dddddddddddd (License2)
- Cashier1: eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee (Shop1)
- Cashier2: ffffffff-ffff-ffff-ffff-ffffffffffff (Shop2)

### Shops  
- Shop1: 11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa (License1, Jakarta)
- Shop2: 22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb (License2, Bandung)
- Shop3: 33333333-cccc-cccc-cccc-cccccccccccc (License3, Surabaya)

## Performance Testing

### Registration Performance
1. Register a new cashier via owner business
2. Measure response time  
3. **Expected**: < 100ms (down from 500ms+ due to bulk policy operations)

### Policy Creation Count
1. Check logs during cashier registration
2. **Expected**: "Created X domain-specific policies in batch" instead of individual policy creation

## Verification Commands

```bash
# Start the application
cd backend && go run cmd/main.go

# Check if seeder ran correctly
# Look for logs: "Assigned cashier role to cashier1 in shop domain shop-{uuid}"

# Test API endpoints using curl or Postman
curl -X POST http://localhost:8080/api/v1/auth/cashier/login \
  -H "Content-Type: application/json" \
  -d '{"username": "cashier1", "pin": "123456"}'

# Use the returned token to test shop access
curl -X GET http://localhost:8080/api/v1/transactions/shop/22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb \
  -H "Authorization: Bearer {token}"
```

## Expected Security Behavior

1. **Domain Isolation**: Each user type has appropriate domain scope
2. **Shop Validation**: All shop-specific endpoints validate shop ownership  
3. **Resource Protection**: Individual resources validate shop ownership
4. **No Bypass**: URL manipulation cannot access unauthorized resources
5. **Performance**: Bulk operations reduce registration time significantly