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
      "id": "cart-uuid-1",
      "shop_id": "shop-uuid-1",
      "product_id": "product-uuid-1",
      "user_id": "user-uuid-1",
      "quantity": 2,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "categories": [
    {
      "id": "category-uuid-1",
      "shop_id": "shop-uuid-1",
      "name": "Electronics",
      "created_at": "2024-01-15T09:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "products": [],
  "transactions": [],
  "payments": [],
  "receipts": [],
  "histories": [],
  "expenses": [],
  "shops": [],
  "stock_histories": [],
  "transaction_products": [],
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
        "id": "cart-uuid-2",
        "shop_id": "shop-uuid-1",
        "product_id": "product-uuid-2",
        "user_id": "user-uuid-2",
        "quantity": 1,
        "created_at": "2024-01-15T11:00:00Z",
        "updated_at": "2024-01-15T11:30:00Z"
      }
    ],
    "categories": [],
    "products": [],
    "transactions": [],
    "payments": [],
    "receipts": [],
    "histories": [],
    "expenses": [],
    "shops": [],
    "stock_histories": [],
    "transaction_products": [],
    "users": [],
    "conflicts": [
      {
        "entity_type": "cart",
        "entity_id": "cart-uuid-1",
        "conflict_type": "timestamp_mismatch",
        "resolution": "server_wins",
        "details": "Server version is newer",
        "server_data": { /* server version of the entity */ },
        "client_data": { /* client version of the entity */ }
      }
    ],
    "errors": [
      {
        "entity_type": "product",
        "entity_id": "product-uuid-1",
        "error_code": "validation_failed",
        "message": "Product name is required",
        "details": "Field validation error"
      }
    ],
    "stats": {
      "processed_entities": {
        "carts": 1,
        "categories": 1
      },
      "created_entities": {
        "carts": 1
      },
      "updated_entities": {
        "categories": 1
      },
      "conflict_count": 1,
      "error_count": 1,
      "processing_time_ms": 150
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
  "message": "User is not associated with a license",
  "errors": null
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
// Example sync request
const syncData = {
  last_sync_timestamp: localStorage.getItem('lastSyncTime'),
  carts: getLocalCarts(),
  categories: getLocalCategories(),
  products: getLocalProducts(),
  // ... other entities
};

fetch('/api/v1/sync', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${getJWTToken()}`
  },
  body: JSON.stringify(syncData)
})
.then(response => response.json())
.then(data => {
  if (data.status === 'success') {
    // Apply server changes to local database
    applyServerChanges(data.data);
    
    // Store new sync timestamp
    localStorage.setItem('lastSyncTime', data.data.sync_timestamp);
    
    // Handle conflicts if any
    if (data.data.conflicts.length > 0) {
      handleConflicts(data.data.conflicts);
    }
    
    // Handle errors if any
    if (data.data.errors.length > 0) {
      handleSyncErrors(data.data.errors);
    }
  }
})
.catch(error => {
  console.error('Sync failed:', error);
});
```