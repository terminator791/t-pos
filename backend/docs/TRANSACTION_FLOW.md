# Transaction Flow Documentation

## Overview

The T-POS system uses a transaction-based checkout flow where users can process sales, payments, and manage transaction states. The system is designed to handle the complete point-of-sale process from cart to completion.

## Database Schema

### Core Tables

1. **transactions** - Main transaction header
2. **transaction_products** - Line items for each transaction
3. **payments** - Payment records linked to transactions
4. **users** - Contains both staff (cashiers) and customers
5. **products** - Items being sold
6. **shops** - Store/merchant information

### Customer Handling

- Customers are stored in the `users` table (not a separate customer table)
- Transactions can have an optional `user_id` field that references the customer
- This allows for both walk-in customers (no customer ID) and registered customers

## Transaction Flow

### 1. Checkout Process

```
POST /api/v1/checkout
```

**Request:**

```json
{
  "shop_id": "uuid",
  "cashier_id": "uuid",
  "customer_id": "uuid", // Optional
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2
    }
  ],
  "payment_method": "cash", // cash, card, digital
  "discount": 5.0,
  "discount_percentage": 10.0,
  "additional_cost": 2.5,
  "amount_paid": 10000 // Amount in cents or smallest currency unit
}
```

**Process:**

1. Validates shop, cashier, customer (if provided)
2. Validates product availability and stock
3. Calculates subtotal, applies discounts, additional costs
4. Creates transaction with status "pending"
5. Creates transaction_products records
6. Creates payment record with status "pending"
7. Updates product stock
8. Calculates change

**Response:**

```json
{
  "transaction": {
    /* transaction object */
  },
  "payment": {
    /* payment object */
  },
  "change": 5.5,
  "message": "Checkout processed successfully"
}
```

### 2. Complete Payment

```
POST /api/v1/checkout/{transactionId}/complete
```

- Updates transaction status to "completed"
- Updates payment status to "completed"
- Finalizes the sale

### 3. Cancel Transaction

```
POST /api/v1/checkout/{transactionId}/cancel
```

- Updates transaction status to "cancelled"
- Updates payment status to "cancelled"
- Restores product stock
- Only available for pending transactions

### 4. Get Transaction Details

```
GET /api/v1/transactions/{transactionId}
```

- Returns complete transaction with products, payments, customer info

### 5. Get Today's Transactions

```
GET /api/v1/transactions/shop/{shopId}/today
```

- Returns all transactions for a shop for the current day

## Entity Relationships

```
Shop (1) -> (many) Transaction
User (1) -> (many) Transaction [as cashier]
User (1) -> (many) Transaction [as customer, optional]
Transaction (1) -> (many) TransactionProduct
Transaction (1) -> (many) Payment
Product (1) -> (many) TransactionProduct
```

## Transaction States

### Transaction Status

- **pending** - Initial state, payment not confirmed
- **completed** - Payment confirmed, sale finalized
- **cancelled** - Transaction cancelled, stock restored
- **failed** - Payment failed

### Payment Status

- **pending** - Awaiting payment confirmation
- **completed** - Payment confirmed
- **failed** - Payment failed
- **cancelled** - Payment cancelled

## Stock Management

### During Checkout

- Stock is immediately reduced when transaction is created (pending state)
- This prevents overselling during the payment process

### On Completion

- No additional stock changes needed

### On Cancellation

- Stock is restored to previous levels

## Use Cases

### Walk-in Customer Sale

```json
{
  "shop_id": "shop-uuid",
  "cashier_id": "cashier-uuid",
  "customer_id": null, // No customer ID
  "items": [...],
  "payment_method": "cash",
  "amount_paid": 10000
}
```

### Registered Customer Sale

```json
{
  "shop_id": "shop-uuid",
  "cashier_id": "cashier-uuid",
  "customer_id": "customer-uuid", // Known customer
  "items": [...],
  "payment_method": "card",
  "amount_paid": 10000
}
```

### Sale with Discount

```json
{
  "shop_id": "shop-uuid",
  "cashier_id": "cashier-uuid",
  "items": [...],
  "discount": 5.00, // Fixed discount amount
  "discount_percentage": 10.0, // Percentage discount
  "payment_method": "cash",
  "amount_paid": 10000
}
```

## Benefits of This Flow

1. **Atomic Operations** - All checkout operations happen in one transaction
2. **Stock Safety** - Stock is reserved immediately to prevent overselling
3. **Flexible Customer Handling** - Supports both walk-in and registered customers
4. **Payment Tracking** - Separate payment records for accounting
5. **Audit Trail** - Complete history of transaction states
6. **Cancellation Support** - Can reverse transactions with stock restoration
7. **Multi-payment Support** - Can handle multiple payments per transaction (future enhancement)

## Error Handling

The system handles various error scenarios:

- Insufficient stock
- Invalid product/shop/user IDs
- Insufficient payment amount
- Attempting to cancel completed transactions
- Stock restoration failures

All errors return appropriate HTTP status codes and descriptive messages.
