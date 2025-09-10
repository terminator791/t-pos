# New API Endpoints Documentation

This document lists all the new API endpoints created for listing and retrieving details of various entities in the T-POS system.

## Authentication & Authorization

All endpoints require:

- Authentication via JWT token in Authorization header: `Bearer <token>`
- Proper RBAC permissions based on user role and domain

## Transaction Endpoints

### List All Transactions (Super Admin & Admin Only)

```
GET /api/v1/transactions
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Transactions by Shop

```
GET /api/v1/transactions/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Transactions by Shop and Status

```
GET /api/v1/transactions/shop/{shopId}/status/{status}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)

Status values: pending, completed, cancelled, failed
```

### Get Transaction Details

```
GET /api/v1/transactions/{id}
```

## Expense Endpoints

### List All Expenses (Super Admin & Admin Only)

```
GET /api/v1/expenses
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Expenses by Shop

```
GET /api/v1/expenses/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Expenses by Shop and Status

```
GET /api/v1/expenses/shop/{shopId}/status/{status}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)

Status values: pending, completed, failed, cancelled
```

### Get Expense Details

```
GET /api/v1/expenses/{id}
```

## Payment Endpoints

### List All Payments (Super Admin & Admin Only)

```
GET /api/v1/payments
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Payments by Shop

```
GET /api/v1/payments/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Payments by Shop and Status

```
GET /api/v1/payments/shop/{shopId}/status/{status}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)

Status values: pending, completed, failed, cancelled
```

### Get Payment Details

```
GET /api/v1/payments/{id}
```

## History Endpoints

### List All Histories (Super Admin & Admin Only)

```
GET /api/v1/histories
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Histories by Shop

```
GET /api/v1/histories/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### Get History Details

```
GET /api/v1/histories/{id}
```

## Receipt Endpoints

### List All Receipts (Super Admin & Admin Only)

```
GET /api/v1/receipts
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Receipts by Shop

```
GET /api/v1/receipts/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### Get Receipt Details

```
GET /api/v1/receipts/{id}
```

## Transaction Product Endpoints

### List All Transaction Products (Super Admin & Admin Only)

```
GET /api/v1/transaction-products
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Transaction Products by Transaction

```
GET /api/v1/transaction-products/transaction/{transactionId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### List Transaction Products by Shop

```
GET /api/v1/transaction-products/shop/{shopId}
Query Parameters:
- limit (optional): Number of items per page (default: 20)
- offset (optional): Number of items to skip (default: 0)
```

### Get Transaction Product Details

```
GET /api/v1/transaction-products/{id}
```

## Response Format

All endpoints return a standardized JSON response:

### Success Response

```json
{
  "status": "success",
  "message": "Data retrieved successfully",
  "data": {
    "items": [...],
    "limit": 20,
    "offset": 0,
    "shop_id": "uuid" // if filtered by shop
  }
}
```

### Error Response

```json
{
  "status": "error",
  "message": "Error description",
  "error": "Detailed error message"
}
```

## Authorization Rules

- **Super Admin & Admin**: Can access all endpoints (list all data across all shops)
- **Owner Business**: Can access shop-specific endpoints for shops under their license
- **Cashier**: Can access shop-specific endpoints for their assigned shop only

## Data Relationship Optimizations

The nested data loading has been optimized to prevent circular references and redundant data:

- Transaction details include shop, license, owner, cashier, transaction products (with product & category), and payments
- Payment details include shop, license, owner, transaction, and cashier (without duplicate shop data)
- Product details in transaction products only include necessary category information
- All responses avoid circular references between related entities

## Usage Examples

### Get today's transactions for a specific shop

```bash
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/v1/transactions/shop/12345-67890?limit=50"
```

### Get completed payments for a shop

```bash
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/v1/payments/shop/12345-67890/status/completed"
```

### Get transaction details with all relationships

```bash
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/v1/transactions/12345-67890"
```
