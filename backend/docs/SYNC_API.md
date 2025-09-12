# Sync API Documentation

The T-POS application provides a robust data synchronization API that enables mobile clients to synchronize data with the backend server using a two-way (push-then-pull) strategy.

## Authentication

All sync endpoints require JWT authentication. Include the Authorization header with your requests:

```
Authorization: Bearer <your-jwt-token>
```

The JWT token must contain valid user information with an associated license ID.

## Endpoints

### 1. Health Check

Check if the sync service is healthy and operational.

**Endpoint:** `GET /api/v1/sync/health`

**Response:**
```json
{
  "status": "healthy",
  "service": "sync",
  "version": "1.0.0"
}
```

### 2. Sync Information

Get information about sync capabilities and current user context.

**Endpoint:** `GET /api/v1/sync/info`

**Response:**
```json
{
  "status": "success",
  "message": "Sync info retrieved successfully",
  "data": {
    "sync_version": "1.0.0",
    "supported_entities": [
      "carts", "categories", "expenses", "histories",
      "payments", "products", "receipts", "shops",
      "stock_histories", "transaction_products", 
      "transactions", "users"
    ],
    "conflict_resolution": "last_write_wins",
    "max_entities_per_request": 1000,
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "license_id": "550e8400-e29b-41d4-a716-446655440000",
    "domain": "LIC-001-DEMO"
  }
}
```

### 3. Process Sync

Main synchronization endpoint for two-way data sync.

**Endpoint:** `POST /api/v1/sync`

**Request Body:**
```json
{
  "last_sync_timestamp": "2024-01-15T10:30:00Z",
  "carts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "product_id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "550e8400-e29b-41d4-a716-446655440003",
      "quantity": 2,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "categories": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Electronics",
      "description": "Electronic devices and accessories",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "category_id": "550e8400-e29b-41d4-a716-446655440004",
      "name": "Wireless Headphones",
      "description": "Bluetooth wireless headphones with noise cancellation",
      "barcode": "1234567890123",
      "price": 299.99,
      "cost": 150.00,
      "stock": 50,
      "min_stock": 10,
      "image_url": "https://example.com/images/headphones.jpg",
      "is_active": true,
      "created_at": "2024-01-15T08:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "transactions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440005",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "cashier_id": "550e8400-e29b-41d4-a716-446655440003",
      "customer_name": "John Doe",
      "discount": 10.00,
      "discount_percentage": 5.0,
      "additional_cost": 0.00,
      "status": "completed",
      "total_price": 589.98,
      "profit_transaction": 289.98,
      "cashier_name": "Jane Smith",
      "change": 10.02,
      "amount": 600,
      "initial_payment_status": "paid",
      "created_at": "2024-01-15T12:00:00Z",
      "updated_at": "2024-01-15T12:05:00Z"
    }
  ],
  "transaction_products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440006",
      "transaction_id": "550e8400-e29b-41d4-a716-446655440005",
      "product_id": "550e8400-e29b-41d4-a716-446655440002",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2,
      "price": 299.99,
      "cost": 150.00,
      "subtotal": 599.98,
      "created_at": "2024-01-15T12:00:00Z",
      "updated_at": "2024-01-15T12:00:00Z"
    }
  ],
  "payments": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440007",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440003",
      "transaction_id": "550e8400-e29b-41d4-a716-446655440005",
      "status": "completed",
      "total": 589.98,
      "created_at": "2024-01-15T12:05:00Z",
      "updated_at": "2024-01-15T12:05:00Z"
    }
  ],
  "receipts": [],
  "histories": [],
  "expenses": [],
  "shops": [],
  "stock_histories": [],
  "users": []
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Sync completed successfully",
  "data": {
    "sync_timestamp": "2024-01-15T12:00:00Z",
    "carts": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440008",
        "shop_id": "550e8400-e29b-41d4-a716-446655440000",
        "product_id": "550e8400-e29b-41d4-a716-446655440009",
        "user_id": "550e8400-e29b-41d4-a716-446655440010",
        "quantity": 1,
        "created_at": "2024-01-15T11:00:00Z",
        "updated_at": "2024-01-15T11:30:00Z"
      }
    ],
    "categories": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440011",
        "shop_id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Home & Garden",
        "description": "Home improvement and garden supplies",
        "created_at": "2024-01-15T11:00:00Z",
        "updated_at": "2024-01-15T11:30:00Z"
      }
    ],
    "products": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440009",
        "shop_id": "550e8400-e29b-41d4-a716-446655440000",
        "category_id": "550e8400-e29b-41d4-a716-446655440011",
        "name": "Garden Hose",
        "description": "50ft expandable garden hose",
        "barcode": "9876543210987",
        "price": 49.99,
        "cost": 25.00,
        "stock": 25,
        "min_stock": 5,
        "image_url": "https://example.com/images/hose.jpg",
        "is_active": true,
        "created_at": "2024-01-15T11:15:00Z",
        "updated_at": "2024-01-15T11:30:00Z"
      }
    ],
    "transactions": [],
    "transaction_products": [],
    "payments": [],
    "receipts": [],
    "histories": [],
    "expenses": [],
    "shops": [],
    "stock_histories": [],
    "users": [],
    "conflicts": [
      {
        "entity_type": "cart",
        "entity_id": "550e8400-e29b-41d4-a716-446655440001",
        "conflict_type": "timestamp_mismatch",
        "resolution": "server_wins",
        "details": "Server version is newer (updated_at: 2024-01-15T11:00:00Z vs 2024-01-15T10:30:00Z)",
        "server_data": {
          "id": "550e8400-e29b-41d4-a716-446655440001",
          "shop_id": "550e8400-e29b-41d4-a716-446655440000",
          "product_id": "550e8400-e29b-41d4-a716-446655440002",
          "user_id": "550e8400-e29b-41d4-a716-446655440003",
          "quantity": 3,
          "created_at": "2024-01-15T10:00:00Z",
          "updated_at": "2024-01-15T11:00:00Z"
        },
        "client_data": {
          "id": "550e8400-e29b-41d4-a716-446655440001",
          "shop_id": "550e8400-e29b-41d4-a716-446655440000",
          "product_id": "550e8400-e29b-41d4-a716-446655440002",
          "user_id": "550e8400-e29b-41d4-a716-446655440003",
          "quantity": 2,
          "created_at": "2024-01-15T10:00:00Z",
          "updated_at": "2024-01-15T10:30:00Z"
        }
      }
    ],
    "errors": [
      {
        "entity_type": "product",
        "entity_id": "550e8400-e29b-41d4-a716-446655440002",
        "error_code": "validation_failed",
        "message": "Product price cannot be negative",
        "details": "Field validation error: price=-10.00"
      },
      {
        "entity_type": "transaction",
        "entity_id": "550e8400-e29b-41d4-a716-446655440012",
        "error_code": "unauthorized",
        "message": "Transaction does not belong to license",
        "details": "License validation failed"
      }
    ],
    "stats": {
      "processed_entities": {
        "carts": 1,
        "categories": 1,
        "products": 1,
        "transactions": 1,
        "transaction_products": 1,
        "payments": 1
      },
      "created_entities": {
        "carts": 0,
        "categories": 1,
        "products": 1,
        "transactions": 1
      },
      "updated_entities": {
        "carts": 1,
        "products": 0,
        "categories": 0
      },
      "conflict_count": 1,
      "error_count": 2,
      "processing_time_ms": 245
    }
  }
}
```

## Sync Strategy

### 1. Push Phase
- Client sends all local changes (created, updated entities) to server
- Server validates and processes each entity
- Conflicts are resolved using Last Write Wins (LWW) strategy
- Server updates database with valid changes

### 2. Pull Phase
- Server returns all changes that occurred after `last_sync_timestamp`
- Client receives updated entities from server
- Client applies server changes to local database

### 3. Conflict Resolution
- **Last Write Wins (LWW)**: Entity with most recent `updated_at` timestamp wins
- Server always wins in case of timestamp ties
- Conflicts are reported in the response for client awareness

## Error Handling

### Authentication Errors
```json
{
  "status": "failed",
  "message": "User not authenticated",
  "errors": null
}
```

### Authorization Errors
```json
{
  "status": "failed",
  "message": "User is not associated with a license",
  "errors": null
}
```

### Validation Errors
```json
{
  "status": "failed",
  "message": "Invalid sync request data",
  "errors": "too many products in sync request (max 1000)"
}
```

### License Errors
```json
{
  "status": "failed",
  "message": "Entity does not belong to user's license",
  "errors": "Product shop_id does not match user's license"
}
```

### Internal Server Errors
```json
{
  "status": "failed",
  "message": "Sync processing failed",
  "errors": "database connection timeout"
}
```

## Conflict Resolution Examples

### Timestamp Conflict
When the same entity is modified on both client and server:

**Scenario**: Cart item quantity changed on both sides
- **Client version**: quantity=2, updated_at="2024-01-15T10:30:00Z"
- **Server version**: quantity=3, updated_at="2024-01-15T11:00:00Z"

**Resolution**: Server wins (Last Write Wins strategy)
```json
{
  "entity_type": "cart",
  "entity_id": "550e8400-e29b-41d4-a716-446655440001",
  "conflict_type": "timestamp_mismatch",
  "resolution": "server_wins",
  "details": "Server version is newer (11:00 vs 10:30)"
}
```

### Entity Creation Conflict
When the same UUID is used for different entities:

**Resolution**: Server entity is preserved, client gets error
```json
{
  "entity_type": "product",
  "entity_id": "550e8400-e29b-41d4-a716-446655440002",
  "conflict_type": "duplicate_entity",
  "resolution": "server_wins",
  "details": "Entity already exists on server with different data"
}
```

## Entity-Specific Examples

### Complete Transaction Sync
Example showing a complete sales transaction with all related entities:

```json
{
  "last_sync_timestamp": "2024-01-15T10:00:00Z",
  "transactions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440020",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "cashier_id": "550e8400-e29b-41d4-a716-446655440003",
      "customer_name": "Sarah Johnson",
      "discount": 25.00,
      "discount_percentage": 0.0,
      "additional_cost": 5.99,
      "status": "completed",
      "total_price": 154.99,
      "profit_transaction": 75.00,
      "cashier_name": "Mike Wilson",
      "change": 45.01,
      "amount": 200,
      "initial_payment_status": "paid",
      "created_at": "2024-01-15T14:30:00Z",
      "updated_at": "2024-01-15T14:35:00Z"
    }
  ],
  "transaction_products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440021",
      "transaction_id": "550e8400-e29b-41d4-a716-446655440020",
      "product_id": "550e8400-e29b-41d4-a716-446655440002",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 1,
      "price": 149.99,
      "cost": 75.00,
      "subtotal": 149.99,
      "created_at": "2024-01-15T14:30:00Z",
      "updated_at": "2024-01-15T14:30:00Z"
    }
  ],
  "payments": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440022",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440003",
      "transaction_id": "550e8400-e29b-41d4-a716-446655440020",
      "status": "completed",
      "total": 154.99,
      "created_at": "2024-01-15T14:35:00Z",
      "updated_at": "2024-01-15T14:35:00Z"
    }
  ]
}
```

### Stock Management Sync
Example showing stock history and inventory updates:

```json
{
  "last_sync_timestamp": "2024-01-15T10:00:00Z",
  "stock_histories": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440030",
      "product_id": "550e8400-e29b-41d4-a716-446655440002",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "sale",
      "quantity_change": -2,
      "quantity_before": 50,
      "quantity_after": 48,
      "reference_id": "550e8400-e29b-41d4-a716-446655440020",
      "reference_type": "transaction",
      "notes": "Sale transaction",
      "created_at": "2024-01-15T14:35:00Z",
      "updated_at": "2024-01-15T14:35:00Z"
    }
  ],
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "shop_id": "550e8400-e29b-41d4-a716-446655440000",
      "category_id": "550e8400-e29b-41d4-a716-446655440004",
      "name": "Wireless Headphones",
      "stock": 48,
      "updated_at": "2024-01-15T14:35:00Z"
    }
  ]
}
```

## Best Practices

1. **Incremental Sync**: Always include `last_sync_timestamp` for efficient delta sync
2. **Batch Size**: Keep entity arrays under 1000 items per request
3. **Error Handling**: Check response for conflicts and errors, handle appropriately
4. **Frequency**: Sync when app comes online, after major operations, and periodically
5. **Offline Storage**: Store sync timestamp locally for next sync request

## Usage Example

```javascript
// Example sync request with comprehensive error handling
const syncData = {
  last_sync_timestamp: localStorage.getItem('lastSyncTime'),
  carts: getLocalCarts(),
  categories: getLocalCategories(),
  products: getLocalProducts(),
  transactions: getLocalTransactions(),
  transaction_products: getLocalTransactionProducts(),
  payments: getLocalPayments(),
  // ... other entities
};

// Validate data before sending
if (Object.values(syncData).some(arr => Array.isArray(arr) && arr.length > 1000)) {
  console.error('Sync data exceeds maximum entities per type (1000)');
  return;
}

try {
  const response = await fetch('/api/v1/sync', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getJWTToken()}`
    },
    body: JSON.stringify(syncData)
  });

  const data = await response.json();

  if (data.status === 'success') {
    console.log(`Sync completed in ${data.data.stats.processing_time_ms}ms`);
    
    // Apply server changes to local database
    await applyServerChanges(data.data);
    
    // Store new sync timestamp
    localStorage.setItem('lastSyncTime', data.data.sync_timestamp);
    
    // Handle conflicts if any
    if (data.data.conflicts.length > 0) {
      console.warn(`${data.data.conflicts.length} conflicts resolved:`);
      data.data.conflicts.forEach(conflict => {
        console.warn(`- ${conflict.entity_type}/${conflict.entity_id}: ${conflict.resolution}`);
        handleConflict(conflict);
      });
    }
    
    // Handle errors if any
    if (data.data.errors.length > 0) {
      console.error(`${data.data.errors.length} sync errors:`);
      data.data.errors.forEach(error => {
        console.error(`- ${error.entity_type}/${error.entity_id}: ${error.message}`);
        handleSyncError(error);
      });
    }

    // Log statistics
    console.log('Sync Statistics:', {
      processed: Object.values(data.data.stats.processed_entities).reduce((a, b) => a + b, 0),
      created: Object.values(data.data.stats.created_entities).reduce((a, b) => a + b, 0),
      updated: Object.values(data.data.stats.updated_entities).reduce((a, b) => a + b, 0),
      conflicts: data.data.stats.conflict_count,
      errors: data.data.stats.error_count
    });

  } else {
    // Handle API errors
    console.error('Sync failed:', data.message);
    if (data.errors) {
      console.error('Details:', data.errors);
    }
  }
} catch (error) {
  console.error('Network or parsing error:', error);
}

// Helper functions
async function applyServerChanges(syncData) {
  // Apply changes to local SQLite database
  const db = await openLocalDatabase();
  
  for (const [entityType, entities] of Object.entries(syncData)) {
    if (Array.isArray(entities) && entities.length > 0) {
      for (const entity of entities) {
        await upsertLocalEntity(db, entityType, entity);
      }
    }
  }
}

function handleConflict(conflict) {
  // Log conflict for user awareness
  // In a full implementation, you might want to:
  // 1. Show user notification about the conflict
  // 2. Update local data with server version
  // 3. Log conflict for audit purposes
}

function handleSyncError(error) {
  // Handle specific error types
  switch (error.error_code) {
    case 'validation_failed':
      // Fix validation issues and retry
      console.log('Validation failed, checking data quality...');
      break;
    case 'unauthorized':
      // Handle authorization issues
      console.log('Authorization failed, checking permissions...');
      break;
    default:
      console.log('Unknown error, logging for investigation...');
  }
}
```

## Common Sync Scenarios

### 1. Initial Sync (First Time)
When the mobile app connects for the first time:

```javascript
const initialSync = {
  last_sync_timestamp: null, // No previous sync
  // Send all local data
  carts: getAllLocalCarts(),
  categories: getAllLocalCategories(),
  products: getAllLocalProducts(),
  // ... other entities
};
```

**Server Response**: Returns recent data (last 30 days) and processes all client data.

### 2. Regular Incremental Sync
Normal sync after the app has been offline:

```javascript
const incrementalSync = {
  last_sync_timestamp: "2024-01-15T10:00:00Z", // Last successful sync
  // Only send changed data since last sync
  carts: getChangedCartsAfter("2024-01-15T10:00:00Z"),
  products: getChangedProductsAfter("2024-01-15T10:00:00Z"),
  // ... other entities
};
```

**Server Response**: Returns only server changes after the timestamp.

### 3. Conflict Resolution Sync
When conflicts are detected and need resolution:

```javascript
// After receiving conflicts in previous sync
const conflictResolutionSync = {
  last_sync_timestamp: "2024-01-15T12:00:00Z",
  // Re-send entities that had conflicts with updated data
  carts: getConflictedCarts(),
  products: getConflictedProducts(),
  // ... other entities
};
```

### 4. Error Recovery Sync
When previous sync had errors:

```javascript
const retrySync = {
  last_sync_timestamp: "2024-01-15T12:00:00Z",
  // Re-send only entities that failed in previous sync
  products: getFailedProducts(),
  transactions: getFailedTransactions(),
  // ... other entities
};
```

## Performance Considerations

### 1. Batch Size Optimization
- **Maximum entities per type**: 1000
- **Recommended batch size**: 100-500 entities per type
- **Large datasets**: Split into multiple sync requests

### 2. Network Optimization
- **Compression**: Server supports gzip compression
- **Incremental sync**: Always provide `last_sync_timestamp`
- **Selective sync**: Only send changed entities

### 3. Conflict Minimization
- **Frequent syncing**: Sync every 5-15 minutes when online
- **Immediate sync**: After major operations (transactions, inventory changes)
- **Background sync**: Schedule periodic background sync

## Security Best Practices

### 1. Authentication
- Always include valid JWT token in Authorization header
- Handle token expiration gracefully
- Refresh tokens as needed

### 2. Data Validation
- Validate all data before sending to prevent errors
- Check entity limits to avoid rejection
- Ensure UUIDs are properly formatted

### 3. License Isolation
- Server automatically filters data by license_id
- Client cannot access data from other licenses
- All sync operations are scoped to user's license

## Monitoring and Debugging

### 1. Sync Statistics
Monitor these metrics for sync health:
- **Success rate**: Percentage of successful syncs
- **Processing time**: Average sync processing time
- **Conflict rate**: Frequency of conflicts
- **Error rate**: Frequency of sync errors

### 2. Common Issues
- **High conflict rate**: May indicate clock synchronization issues
- **Frequent validation errors**: Data quality issues on client
- **Authorization errors**: Token or license configuration problems
- **Timeout errors**: Large dataset or network issues

### 3. Debugging Tips
- Check server logs for detailed error information
- Validate JWT token claims and expiration
- Ensure proper license_id association
- Test with smaller datasets to isolate issues