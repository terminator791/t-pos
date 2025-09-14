# Cashier Authentication Flow

This document describes the implementation of the cashier authentication system in T-POS.

## Overview

The cashier authentication system allows `owner_business` users to create cashier accounts that are bound to specific shops. Cashiers can only access data and perform operations within their assigned shop.

## Authentication Flows

### 1. Owner Business Flow (Existing)

- **Register**: `POST /api/v1/auth/owner/register` with `{username, serial_number}`
- **Create PIN**: `POST /api/v1/auth/pin` with `{pin}`
- **Login**: `POST /api/v1/auth/owner/login` with `{username, pin}`

### 2. Cashier Flow (New)

- **Register** (by owner_business): `POST /api/v1/auth/cashier/register` with `{username, shop_id, name}`
- **Create PIN**: `POST /api/v1/auth/pin` with `{pin}` (same as owner)
- **Login**: `POST /api/v1/auth/cashier/login` with `{username, pin}`

## API Endpoints

### Cashier Registration

```http
POST /api/v1/auth/cashier/register
Authorization: Bearer <owner_business_token>
Content-Type: application/json

{
    "username": "cashier1",
    "shop_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe"
}
```

**Response:**

```json
{
  "status": "success",
  "message": "Cashier registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "cashier1",
      "name": "John Doe",
      "license_id": "550e8400-e29b-41d4-a716-446655440002",
      "role_id": "550e8400-e29b-41d4-a716-446655440003",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000"
    },
    "roles": ["cashier"],
    "domain": "shop-550e8400-e29b-41d4-a716-446655440000",
    "expires_at": 1726142400
  }
}
```

### Cashier Login

```http
POST /api/v1/auth/cashier/login
Content-Type: application/json

{
    "username": "cashier1",
    "pin": "123456"
}
```

**Response:**

```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "cashier1",
      "name": "John Doe",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000"
    },
    "roles": ["cashier"],
    "domain": "shop-550e8400-e29b-41d4-a716-446655440000",
    "expires_at": 1726142400
  }
}
```

## Authorization and Permissions

### Cashier Permissions

Cashiers have shop-specific permissions with domain `shop:<shop_id>`:

- **Products**: Full CRUD within assigned shop
- **Categories**: Full CRUD within assigned shop
- **Carts**: Full CRUD within assigned shop
- **Transactions**: Full CRUD within assigned shop
- **Expenses**: Read-only within assigned shop
- **Payments**: Read-only within assigned shop
- **Histories**: Read-only within assigned shop
- **Receipts**: Read-only within assigned shop
- **Customers**: Read-only
- **Sync**: Within assigned shop only

### Domain Binding

- Owner business gets domain from license serial number (e.g., `LIC-001-DEMO`)
- Cashiers get domain from shop (e.g., `shop-550e8400-e29b-41d4-a716-446655440000`)
- ACL policies use these domains to enforce access control

## Security Features

1. **License Validation**: Only owner_business with valid license can create cashiers
2. **Shop Binding**: Cashiers are bound to specific shops and cannot access other shops
3. **Role Verification**: Registration and login endpoints verify user roles
4. **JWT Tokens**: Include shop_id for cashiers to enable shop-specific access
5. **Domain Isolation**: Each shop operates in its own domain for multi-tenancy

## Database Schema

### User Entity Updates

```go
type User struct {
    // ... existing fields
    ShopID *uuid.UUID `gorm:"type:uuid" json:"shop_id"` // Shop binding for cashiers
}
```

### Domain Structure

- `LicenseID` for owner_business users
- `ShopID` for cashier users
- `UserDomain` table for domain access mapping

## Implementation Details

### RegisterCashier Handler

1. Validates owner_business authentication and permissions
2. Verifies shop belongs to same license as creating user
3. Creates cashier user with shop binding
4. Sets up domain permissions in Casbin
5. Returns JWT token with shop context

### LoginCashier Handler

1. Validates cashier credentials (username + PIN)
2. Verifies user has cashier role
3. Confirms shop assignment
4. Returns JWT token with shop domain

### ACL Integration

- Casbin policies use shop domains: `{"cashier", "shop:<shop_id>", "/api/v1/products", "GET"}`
- Authorization middleware checks domain context from JWT
- Shop-specific API endpoints filter data by shop_id

## Usage Examples

### 1. Owner Creates Cashier

```bash
# Owner logs in
curl -X POST http://localhost:8080/api/v1/auth/owner/login \
  -H "Content-Type: application/json" \
  -d '{"username": "owner1", "pin": "123456"}'

# Owner creates cashier
curl -X POST http://localhost:8080/api/v1/auth/cashier/register \
  -H "Authorization: Bearer <owner_token>" \
  -H "Content-Type: application/json" \
  -d '{"username": "cashier1", "shop_id": "550e8400-e29b-41d4-a716-446655440000", "name": "John Doe"}'
```

### 2. Cashier Creates PIN and Logs In

```bash
# Cashier creates PIN (using token from registration)
curl -X POST http://localhost:8080/api/v1/auth/pin \
  -H "Authorization: Bearer <cashier_token>" \
  -H "Content-Type: application/json" \
  -d '{"pin": "654321"}'

# Cashier logs in
curl -X POST http://localhost:8080/api/v1/auth/cashier/login \
  -H "Content-Type: application/json" \
  -d '{"username": "cashier1", "pin": "654321"}'
```

### 3. Cashier Accesses Shop Data

```bash
# Get products in assigned shop
curl -X GET http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <cashier_token>"

# Create transaction in assigned shop
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer <cashier_token>" \
  -H "Content-Type: application/json" \
  -d '{"items": [{"product_id": "...", "quantity": 1}]}'
```

## Error Handling

- `400 Bad Request`: Invalid request data, username exists, shop not found
- `401 Unauthorized`: Invalid credentials, PIN not set, wrong role
- `403 Forbidden`: Only owner business can create cashiers, shop doesn't belong to license
- `500 Internal Server Error`: Database errors, token generation failures

## Troubleshooting

### Issue: Cashier gets "Insufficient permissions" error

**Problem**: Cashier receives 403 Forbidden with "Insufficient permissions" when trying to access shop resources.

**Cause**: This typically happens when there's a mismatch between the ACL policies and the user's domain assignment.

**Solution**:

1. **Check Policy Domain**: Ensure cashier policies use `*` domain (not `shop:*`):

   ```sql
   SELECT * FROM casbin_rule WHERE ptype = 'p' AND v0 = 'cashier';
   ```

2. **Reset Policies** (if needed):

   ```sql
   -- Run the reset script
   source backend/scripts/reset_policies.sql

   -- Or delete manually
   DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = 'cashier' AND v1 = 'shop:*';
   ```

3. **Reload Policies**:

   ```bash
   curl -X POST http://localhost:8080/api/v1/acl/reload \
     -H "Authorization: Bearer <admin_token>"
   ```

4. **Re-seed Database** (if necessary):
   ```bash
   cd backend
   go run cmd/main.go # Will auto-seed on startup
   ```

### Issue: Domain Mismatch

**Problem**: User domain doesn't match policy domain.

**Check User Domain**:

```bash
curl -X GET http://localhost:8080/api/v1/auth/permissions \
  -H "Authorization: Bearer <cashier_token>"
```

**Expected Response**:

```json
{
  "role": "cashier",
  "current_domain": "shop-550e8400-e29b-41d4-a716-446655440000",
  "accessible_domains": ["shop-550e8400-e29b-41d4-a716-446655440000"]
}
```

### Issue: ACL Policies Not Loading

**Problem**: Changes to auth seeder don't take effect.

**Solution**:

1. Delete existing policies from database
2. Restart application to trigger auto-seeding
3. Or manually call seeder endpoints if available
