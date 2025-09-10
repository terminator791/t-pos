# Data Synchronization API Usage Examples

This document provides practical examples of how to use the data synchronization API endpoints.

## Authentication

All sync endpoints require JWT authentication. Include the bearer token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

The system automatically extracts `license_id` from the authenticated user's profile.

## API Endpoints

### 1. Push Sync (Mobile → Server)

Upload offline data from mobile client to server.

**Endpoint:** `POST /api/v1/sync/push`

**Request Body:**
```json
{
  "metadata": {
    "last_sync_timestamp": "2025-01-10T10:00:00Z",
    "client_id": "mobile_app_v1.0"
  },
  "data": {
    "products": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "shop_id": "987fcdeb-51a2-43d7-8329-123456789abc",
        "name": "Coffee Beans Premium",
        "sale": 25.50,
        "buy": 15.00,
        "stock": 100,
        "created_at": "2025-01-10T09:30:00Z",
        "updated_at": "2025-01-10T10:15:00Z"
      }
    ],
    "transactions": [
      {
        "id": "456e7890-e89b-12d3-a456-426614174001",
        "shop_id": "987fcdeb-51a2-43d7-8329-123456789abc",
        "cashier_id": "user_id_here",
        "total_price": 51.00,
        "status": "completed",
        "created_at": "2025-01-10T10:00:00Z",
        "updated_at": "2025-01-10T10:05:00Z"
      }
    ]
  }
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Push sync completed successfully",
  "data": {
    "success": true,
    "sync_timestamp": "2025-01-10T10:30:00Z",
    "records_processed": 2,
    "conflicts_resolved": 0,
    "errors": []
  }
}
```

### 2. Pull Sync (Server → Mobile)

Retrieve server changes since last sync.

**Endpoint:** `GET /api/v1/sync/pull?since=2025-01-10T10:00:00Z&limit=1000`

**Query Parameters:**
- `since`: ISO 8601 timestamp (optional) - get changes since this time
- `limit`: Maximum records to return (optional, default: 1000, max: 1000)
- `offset`: Pagination offset (optional, default: 0)

**Response:**
```json
{
  "status": "success",
  "message": "Pull sync completed successfully",
  "data": {
    "success": true,
    "sync_timestamp": "2025-01-10T10:30:00Z",
    "records_processed": 15,
    "data": {
      "products": [
        {
          "id": "789e1234-e89b-12d3-a456-426614174002",
          "shop_id": "987fcdeb-51a2-43d7-8329-123456789abc",
          "name": "Updated Product Name",
          "sale": 30.00,
          "updated_at": "2025-01-10T10:25:00Z"
        }
      ],
      "categories": [
        {
          "id": "321e9876-e89b-12d3-a456-426614174003",
          "shop_id": "987fcdeb-51a2-43d7-8329-123456789abc",
          "name": "New Category",
          "created_at": "2025-01-10T10:20:00Z",
          "updated_at": "2025-01-10T10:20:00Z"
        }
      ]
    }
  }
}
```

### 3. Full Sync (Two-way)

Perform both push and pull in a single operation.

**Endpoint:** `POST /api/v1/sync/full`

**Request Body:**
```json
{
  "push_data": {
    "metadata": {
      "last_sync_timestamp": "2025-01-10T09:00:00Z"
    },
    "data": {
      "products": [
        {
          "id": "new_product_id",
          "name": "New Offline Product",
          "created_at": "2025-01-10T09:30:00Z"
        }
      ]
    }
  },
  "pull_since": "2025-01-10T09:00:00Z"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Full sync completed successfully",
  "data": {
    "push_result": {
      "success": true,
      "records_processed": 1,
      "conflicts_resolved": 0
    },
    "pull_result": {
      "success": true,
      "records_processed": 8,
      "data": {
        "products": [...],
        "categories": [...]
      }
    }
  }
}
```

### 4. Sync Status

Get current sync status and server timestamp.

**Endpoint:** `GET /api/v1/sync/status`

**Response:**
```json
{
  "status": "success",
  "message": "Sync status retrieved",
  "data": {
    "user_id": "user_uuid_here",
    "shop_id": "shop_uuid_here",
    "server_timestamp": "2025-01-10T10:30:00Z",
    "sync_available": true,
    "last_sync": null,
    "pending_changes": 0
  }
}
```

### 5. Validate Sync Data

Validate sync data without actually processing it.

**Endpoint:** `POST /api/v1/sync/validate`

**Request Body:** Same as push sync request

**Response:**
```json
{
  "status": "success",
  "message": "Sync data is valid",
  "data": []
}
```

Or with validation errors:
```json
{
  "status": "success",
  "message": "Sync data has validation errors",
  "data": [
    {
      "entity_type": "product",
      "entity_id": "123e4567-e89b-12d3-a456-426614174000",
      "field": "sale",
      "message": "sale price must be greater than 0",
      "error_type": "validation"
    }
  ]
}
```

## Error Handling

All endpoints return consistent error responses:

```json
{
  "status": "failed",
  "message": "Error description",
  "errors": "Detailed error information"
}
```

Common HTTP status codes:
- `200`: Success (even with sync conflicts resolved)
- `400`: Bad Request (invalid data, validation errors)
- `401`: Unauthorized (missing or invalid JWT token)
- `403`: Forbidden (insufficient permissions)
- `500`: Internal Server Error

## Conflict Resolution

The system uses **Last-Write-Wins (LWW)** conflict resolution:

1. Compare `updated_at` timestamps
2. More recent timestamp wins
3. Conflicts are automatically resolved and logged
4. Special handling for certain entity types (e.g., transactions)

## Multi-tenancy

- All data is automatically filtered by user's `license_id`
- Cashier users only see data from their assigned shop
- Owner/business users see all shops under their license

## Performance Considerations

- Use delta sync with `since` parameter for efficiency
- Batch size is limited to 1000 records per request
- Use pagination for large datasets
- Server timestamps are in UTC

## Mobile Client Integration Tips

1. **Store last sync timestamp** locally for delta sync
2. **Handle offline UUIDs** - generate UUIDs offline for new records
3. **Implement retry logic** for failed sync operations
4. **Batch operations** to minimize API calls
5. **Validate data** before syncing to catch errors early
6. **Handle conflicts gracefully** - inform users when server data overwrites local changes