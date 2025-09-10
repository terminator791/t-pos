# Comprehensive ACL and Auth Flow Analysis & Refactor

## Overview

This document provides a complete analysis of the ACL (Access Control List) and authentication/authorization implementation in the T-POS backend system using JWT authentication and Casbin RBAC (Role-Based Access Control) with domain support.

## Current System Architecture

### 1. Authentication Flow

- **JWT Token-based authentication**
- **Single role per user** (simplified from multi-role)
- **Domain-based tenancy** using license serial numbers
- **PIN-based credentials** for POS operations

### 2. Authorization Flow (Casbin RBAC)

- **Subject**: User ID (UUID)
- **Domain**: License serial number or shop ID (e.g., `LIC-001-DEMO`, `shop:shop_uuid`)
- **Object**: API endpoint pattern (e.g., `/api/v1/products`)
- **Action**: HTTP method (GET, POST, PUT, DELETE)

### 3. Role Hierarchy & Permissions

#### **Super Admin** (`*` domain = all domains)

- Full system access across all shops and licenses
- Can manage licenses, users, and system-wide resources
- Domain: `*` (wildcard for all)

#### **Admin** (`*` domain = multi-shop admin)

- Administrative access across multiple shops
- Cannot manage licenses or create super admins
- Can manage users, products, and operations across shops
- Domain: `*` but limited scope

#### **Owner Business** (license-based domain)

- Business owner with full shop management within their license
- Can create/manage cashiers within their shops
- Domain: License serial number (e.g., `LIC-001-DEMO`)

#### **Cashier** (`shop:*` domain = shop-specific)

- Point of sale operations within assigned shop
- Limited to operational functions, cannot manage users
- Domain: `shop:shop_uuid`

## Current ACL Implementation Status

### ✅ **IMPLEMENTED Endpoints with ACL:**

#### **Products** - ALL ROLES

```
POST   /api/v1/products                    // All roles (shop-scoped for cashier)
GET    /api/v1/products                    // All roles (shop-scoped for cashier)
GET    /api/v1/products/:id                // All roles (shop-scoped for cashier)
PUT    /api/v1/products/:id                // All roles (shop-scoped for cashier)
DELETE /api/v1/products/:id                // All roles (shop-scoped for cashier)
GET    /api/v1/products/search             // All roles (shop-scoped for cashier)
GET    /api/v1/products/barcode/:barcode   // All roles (shop-scoped for cashier)
GET    /api/v1/products/low-stock          // All roles (shop-scoped for cashier)
POST   /api/v1/products/upload             // All roles (shop-scoped for cashier)
```

#### **Transactions** - ALL ROLES

```
POST   /api/v1/transactions                // All roles
GET    /api/v1/transactions/:id            // All roles
POST   /api/v1/transactions/:id/pay        // All roles
POST   /api/v1/transactions/:id/cancel     // All roles
GET    /api/v1/transactions                // Super admin, admin only
GET    /api/v1/transactions/shop/:shopId   // All roles (shop-scoped)
```

#### **Auth** - ALL AUTHENTICATED USERS

```
POST   /api/v1/auth/login                  // Public
POST   /api/v1/auth/logout                 // Authenticated
GET    /api/v1/auth/profile                // Authenticated
GET    /api/v1/auth/permissions            // Authenticated
POST   /api/v1/auth/pin                    // Authenticated
PUT    /api/v1/auth/pin                    // Authenticated
DELETE /api/v1/auth/pin                    // Authenticated
```

#### **Customers** - ROLE-BASED

```
GET    /api/v1/customers                   // All roles (shop-scoped for cashier)
POST   /api/v1/customers                   // Owner business, admin, super admin only
DELETE /api/v1/customers/:id               // Owner business, admin, super admin only
```

#### **Users** - ADMIN ONLY

```
GET    /api/v1/users                       // Super admin, admin
POST   /api/v1/users                       // Super admin only
PUT    /api/v1/users/:id                   // Super admin, admin
DELETE /api/v1/users/:id                   // Super admin only
```

### ❌ **MISSING ACL Implementation (Fixed in Refactor):**

#### **Categories** - NOW IMPLEMENTED

```
POST   /api/v1/categories                  // All roles (shop-scoped for cashier)
GET    /api/v1/categories                  // All roles (shop-scoped for cashier)
PUT    /api/v1/categories/:id              // All roles (shop-scoped for cashier)
DELETE /api/v1/categories/:id              // All roles (shop-scoped for cashier)
```

#### **Carts** - NOW IMPLEMENTED

```
POST   /api/v1/carts                       // All roles (shop-scoped for cashier)
GET    /api/v1/carts                       // All roles (shop-scoped for cashier)
PUT    /api/v1/carts/:id                   // All roles (shop-scoped for cashier)
DELETE /api/v1/carts/:id                   // All roles (shop-scoped for cashier)
```

#### **Expenses** - NOW IMPLEMENTED

```
GET    /api/v1/expenses                    // All roles (shop-scoped for owner/cashier)
GET    /api/v1/expenses/shop/:shopId       // All roles (shop-scoped for owner/cashier)
```

#### **Payments** - NOW IMPLEMENTED

```
GET    /api/v1/payments                    // All roles (shop-scoped for owner/cashier)
GET    /api/v1/payments/shop/:shopId       // All roles (shop-scoped for owner/cashier)
```

#### **Histories** - NOW IMPLEMENTED

```
GET    /api/v1/histories                   // All roles (shop-scoped for owner/cashier)
GET    /api/v1/histories/shop/:shopId      // All roles (shop-scoped for owner/cashier)
```

#### **Receipts** - NOW IMPLEMENTED

```
GET    /api/v1/receipts                    // All roles (shop-scoped for owner/cashier)
GET    /api/v1/receipts/shop/:shopId       // All roles (shop-scoped for owner/cashier)
```

#### **Transaction Products** - NOW IMPLEMENTED

```
GET    /api/v1/transaction-products        // All roles (shop-scoped for owner/cashier)
```

#### **Roles** - NOW IMPLEMENTED

```
GET    /api/v1/roles                       // Super admin only
GET    /api/v1/roles/:id                   // Super admin only
```

### 🆕 **NEW ACL Management Endpoints:**

#### **ACL Management** - SUPER ADMIN ONLY

```
GET    /api/v1/acl/policies                // Get all Casbin policies
POST   /api/v1/acl/policies                // Add new policy
DELETE /api/v1/acl/policies                // Remove policy
GET    /api/v1/acl/policies/system         // Get system policies from DB

GET    /api/v1/acl/roles                   // Get all Casbin role assignments
GET    /api/v1/acl/roles/system            // Get system roles from DB
GET    /api/v1/acl/users/:userId/roles     // Get user roles
GET    /api/v1/acl/roles/:role/users       // Get users with specific role
POST   /api/v1/acl/users/roles             // Add role to user
DELETE /api/v1/acl/users/roles             // Remove role from user

POST   /api/v1/acl/check                   // Check permission
POST   /api/v1/acl/reload                  // Reload policies from DB
```

## Domain System Explanation

### Domain Types:

1. **`*`** - Global domain (Super admin access)
2. **`LIC-XXX-XXXX`** - License serial number (Owner business scope)
3. **`shop:uuid`** - Specific shop ID (Cashier scope)

### Example Domain Mappings:

```
Super Admin: domain = "*" (access to everything)
Admin: domain = "*" (limited to non-license operations)
Owner Business: domain = "LIC-001-DEMO" (all shops under this license)
Cashier: domain = "shop:550e8400-e29b-41d4-a716-446655440001" (specific shop)
```

## ACL Policy Examples

### Adding/Modifying Role Permissions:

**Example 1: Restrict Cashier Product Management**

```go
// Current: Cashier can CRUD products
{"cashier", "shop:*", "/api/v1/products", "GET|POST|PUT|DELETE"}

// Modified: Cashier can only read products
{"cashier", "shop:*", "/api/v1/products", "GET"}
```

**Impact:**

- Cashiers lose ability to create/update/delete products
- Must use admin/owner account for product management
- Runtime enforcement via Casbin middleware

**Example 2: Add new shop domain**

```go
// Add cashier to new shop
enforcerService.AddRoleForUser("user_123", "cashier", "shop:new-shop-uuid")
```

**Impact:**

- User gains access to new shop
- Inherits all cashier permissions for that shop
- Cannot access other shops

## Key Database Tables

### Casbin Tables:

1. **`casbin_rule`** - Core Casbin policies and role assignments
2. **`roles`** - System role definitions
3. **`policies`** - Policy definitions with metadata
4. **`user_domains`** - User domain assignments
5. **`users`** - User accounts with single role

## Testing ACL Changes

### 1. Testing Permission Changes:

```bash
# Test with different role tokens
curl -H "Authorization: Bearer <cashier_token>" \
     -X POST http://localhost:8080/api/v1/products
# Expected: 403 Forbidden (if product creation restricted)

curl -H "Authorization: Bearer <owner_token>" \
     -X POST http://localhost:8080/api/v1/products
# Expected: 200 OK
```

### 2. Testing Domain Isolation:

```bash
# Cashier trying to access different shop
curl -H "Authorization: Bearer <shop1_cashier_token>" \
     -X GET http://localhost:8080/api/v1/transactions/shop/shop2_uuid
# Expected: 403 Forbidden
```

### 3. ACL Management Testing:

```bash
# Check user permissions
curl -H "Authorization: Bearer <admin_token>" \
     -X POST http://localhost:8080/api/v1/acl/check \
     -d '{"user_id":"user_123","domain":"shop:abc","object":"/api/v1/products","action":"POST"}'
```

## Refactor Changes Summary

### 1. **Fixed Missing ACL Policies:**

- Added comprehensive policies for all endpoints
- Implemented proper domain-based restrictions
- Added role-specific permission matrices

### 2. **Added ACL Management System:**

- New ACL handler for policy management
- Runtime permission checking endpoints
- Policy reload capabilities

### 3. **Enhanced Repository Methods:**

- Added `GetAll()` methods to repositories
- Extended policy and role management

### 4. **Updated Routes Configuration:**

- Added ACL management routes
- Proper middleware application
- Organized endpoint grouping

### 5. **Complete Permission Matrix:**

- Super Admin: Full system access (`*` domain)
- Admin: Multi-shop management (`*` domain, limited scope)
- Owner Business: License-based shop management (license domain)
- Cashier: Shop-specific operations (`shop:*` domain)

## Best Practices Implemented

1. **Principle of Least Privilege**: Each role has minimum required permissions
2. **Domain Isolation**: Shop-based data segregation
3. **Hierarchical Access**: Clear role hierarchy with appropriate escalation
4. **Audit Trail**: All policy changes tracked in database
5. **Runtime Management**: Dynamic policy updates without restart
6. **Comprehensive Testing**: ACL validation endpoints for debugging

This refactor provides a robust, scalable ACL system that properly secures all endpoints while maintaining flexibility for future role and permission modifications.
