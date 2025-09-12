# Cashier Permission Fix

## Issue Description

Cashiers were receiving "Insufficient permissions" errors when trying to access shop resources because of a domain mismatch in the ACL policies.

## Root Cause

The original cashier policies used `shop:*` as the domain:

```
{"cashier", "shop:*", "/api/v1/categories/*", "GET"}
```

But when cashiers logged in, their JWT token contained a specific shop domain:

```
"domain": "shop-71eda649-0f59-4de9-a7dc-e4c4a8bcfbe4"
```

Casbin was doing an exact match between `shop:*` and `shop-71eda649-0f59-4de9-a7dc-e4c4a8bcfbe4`, which failed.

## Solution

Changed all cashier policies to use `*` domain, which matches any domain:

```
{"cashier", "*", "/api/v1/categories/*", "GET"}
```

## Files Modified

1. **auth_seeder.go**: Updated all cashier policies from `shop:*` to `*` domain
2. **auth_handler.go**: Removed unnecessary shop permission assignment
3. **reset_policies.sql**: Script to clean up old policies

## Domain Isolation

Even though cashier policies use `*` domain for Casbin matching, shop-level data isolation is still enforced at the:

1. **Application Layer**: Repository methods filter by `shop_id` from JWT claims
2. **Middleware Layer**: `GetUserShopIDFromContext()` provides shop context
3. **Business Logic**: Each cashier can only access their assigned shop's data

## Verification

After applying the fix, cashiers should be able to:

- Access categories: `GET /api/v1/categories/:id`
- Manage products: `GET/POST/PUT/DELETE /api/v1/products/*`
- Handle transactions: `POST /api/v1/transactions`
- View shop data: `GET /api/v1/shops/:id`

The domain in error responses should match their shop domain:

```json
{
  "user": "0a5ba04f-ebdd-450a-aec0-d5247bbd63cf",
  "domain": "shop-71eda649-0f59-4de9-a7dc-e4c4a8bcfbe4",
  "object": "/api/v1/categories/:id",
  "action": "GET"
}
```
