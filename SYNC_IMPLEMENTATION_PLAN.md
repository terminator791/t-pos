# Data Synchronization Implementation Plan

This document outlines the comprehensive implementation plan for robust, scalable, and conflict-resilient data synchronization between the mobile client (SQLite) and the backend server (PostgreSQL) for the t-pos application.

## Project Overview

The t-pos application requires a two-way data synchronization mechanism that allows mobile clients to:
1. **Push**: Send all local changes (new, updated records) to the backend
2. **Pull**: Receive all server-side changes since last sync
3. **Resolve conflicts** using Last Write Wins (LWW) strategy
4. **Handle offline data** created with UUIDs to prevent key collisions
5. **Maintain multi-tenancy** through license_id filtering

## Synchronizable Entities

The following tables require synchronization between mobile and backend:
- `carts` - Shopping cart items
- `categories` - Product categories per shop
- `expenses` - Shop expense records
- `histories` - Transaction history links
- `payments` - Payment records
- `products` - Product catalog
- `receipts` - Receipt records
- `shops` - Shop information
- `stock_histories` - Product stock change history
- `transaction_products` - Transaction line items
- `transactions` - Sales transactions
- `users` - User accounts (cashiers, owner_business)

## Technical Strategy

### 1. Authentication & Authorization
- Use existing JWT token with `license_id` claim for multi-tenancy
- Sync only data belonging to the same license (can contain multiple shops)
- Leverage existing Casbin ACL system for permission control

### 2. Conflict Resolution
- **Last Write Wins (LWW)**: Use `updated_at` timestamps to resolve conflicts
- Server always wins in case of timestamp conflicts
- Track sync metadata for delta synchronization

### 3. Data Change Tracking
- Utilize existing `created_at` and `updated_at` timestamps
- Support UUID primary keys for offline record creation
- Implement soft deletes awareness with `deleted_at`

### 4. Performance & Scalability
- Async worker pattern for sync processing
- Batch operations for efficient data transfer
- Delta sync based on last sync timestamp
- Connection pooling and transaction management

## Implementation Phases

## Current Progress

**Phase 1: Planning and Infrastructure Setup ✅**
- [x] Analyze existing backend architecture and database schema
- [x] Understand authentication and authorization flow
- [x] Identify synchronizable entities and their relationships
- [x] Design sync strategy and conflict resolution approach
- [x] Create comprehensive implementation plan

**Phase 2: Core Sync Infrastructure ✅**
- [x] Create sync DTOs and request/response structures
- [x] Implement basic sync service structure with core business logic
- [x] Add conflict resolution algorithms (Last Write Wins)
- [x] Create missing repository interfaces (StockHistoryRepository)
- [x] Implement license-based data filtering framework
- [x] Add basic unit tests for sync service

## Undone Progress

**Phase 3: API Layer Implementation**
- [x] Create sync handler for HTTP endpoints
- [x] Add proper error handling and validation
- [x] Create sync routes and middleware integration
- [ ] Implement async worker for sync processing
- [ ] Enhance license ID resolution from user context
- [ ] Add comprehensive input validation

**Phase 4: Complete Entity Implementation**
- [ ] Complete sync implementation for all entities (Categories, Products, Transactions, etc.)
- [ ] Implement proper license-based filtering for all entity types
- [ ] Add comprehensive conflict resolution for all entity relationships
- [ ] Optimize database queries for large datasets

**Phase 5: Integration and Testing**
- [ ] Write comprehensive integration tests for sync API
- [ ] Test conflict resolution scenarios with real data
- [ ] Performance testing and optimization
- [ ] Add monitoring and logging

## Detailed Implementation Tasks

### Phase 2: Core Sync Infrastructure

#### 2.1 Sync DTOs (Data Transfer Objects)
```go
// Location: internal/domain/dto/sync.go
type SyncRequest struct {
    LastSyncTimestamp *time.Time         `json:"last_sync_timestamp"`
    Carts             []entities.Cart    `json:"carts"`
    Categories        []entities.Category `json:"categories"`
    Expenses          []entities.Expense  `json:"expenses"`
    // ... other entities
}

type SyncResponse struct {
    SyncTimestamp     time.Time           `json:"sync_timestamp"`
    Carts             []entities.Cart     `json:"carts"`
    Categories        []entities.Category `json:"categories"`
    Conflicts         []ConflictInfo      `json:"conflicts"`
    Errors            []SyncError         `json:"errors"`
}
```

#### 2.2 Sync Service Implementation
```go
// Location: internal/application/services/sync_service.go
type SyncService struct {
    // Repository dependencies
    cartRepo           repositories.CartRepository
    categoryRepo       repositories.CategoryRepository
    // ... other repos
    
    // Core methods
    ProcessSync(ctx context.Context, req SyncRequest, licenseID uuid.UUID) (*SyncResponse, error)
    PushChanges(ctx context.Context, data interface{}, licenseID uuid.UUID) error
    PullChanges(ctx context.Context, lastSync *time.Time, licenseID uuid.UUID) (*SyncResponse, error)
    ResolveConflicts(existing, incoming interface{}) (interface{}, *ConflictInfo)
}
```

#### 2.3 Conflict Resolution Engine
```go
// Location: internal/domain/sync/conflict_resolver.go
type ConflictResolver struct {
    strategy ConflictResolutionStrategy
}

func (cr *ConflictResolver) Resolve(existing, incoming Syncable) (Syncable, *ConflictInfo)
```

### Phase 3: API Layer Implementation

#### 3.1 Sync Handler
```go
// Location: internal/interfaces/http/handlers/sync_handler.go
type SyncHandler struct {
    syncService *services.SyncService
    worker      *SyncWorker
}

func (h *SyncHandler) ProcessSync(c *gin.Context)
func (h *SyncHandler) GetSyncStatus(c *gin.Context)
```

#### 3.2 Async Worker
```go
// Location: internal/infrastructure/workers/sync_worker.go
type SyncWorker struct {
    jobQueue    chan SyncJob
    workerPool  chan chan SyncJob
    quit        chan bool
}
```

#### 3.3 API Routes
```go
// Add to routes.go
sync := protected.Group("/sync")
{
    sync.POST("", syncHandler.ProcessSync)
    sync.GET("/status/:jobId", syncHandler.GetSyncStatus)
}
```

## API Design

### POST /api/v1/sync
**Description**: Main synchronization endpoint supporting push-then-pull pattern

**Request Format**:
```json
{
  "last_sync_timestamp": "2024-01-15T10:30:00Z",
  "carts": [...],
  "categories": [...],
  "products": [...],
  "transactions": [...],
  // ... other entities
}
```

**Response Format**:
```json
{
  "status": "success",
  "message": "Sync completed successfully",
  "data": {
    "sync_timestamp": "2024-01-15T12:00:00Z",
    "carts": [...],
    "categories": [...],
    "conflicts": [
      {
        "entity_type": "product",
        "entity_id": "uuid",
        "conflict_type": "timestamp",
        "resolution": "server_wins",
        "details": "..."
      }
    ],
    "errors": [
      {
        "entity_type": "transaction",
        "entity_id": "uuid", 
        "error_code": "validation_failed",
        "message": "Invalid transaction data"
      }
    ]
  }
}
```

## Security Considerations

1. **Multi-tenancy**: Strict license_id based data isolation
2. **Authentication**: Validate JWT tokens on all sync requests
3. **Authorization**: Use Casbin policies for operation permissions
4. **Data validation**: Validate all incoming sync data
5. **Rate limiting**: Prevent sync API abuse
6. **Audit logging**: Track all sync operations

## Performance Optimizations

1. **Batch processing**: Handle multiple records efficiently
2. **Connection pooling**: Optimize database connections
3. **Indexing**: Ensure proper indexes on timestamp columns
4. **Caching**: Cache frequently accessed sync metadata
5. **Compression**: Compress large sync payloads
6. **Pagination**: Support large dataset synchronization

## Error Handling Strategy

1. **Partial sync support**: Continue processing even if some records fail
2. **Detailed error reporting**: Provide specific error information per record
3. **Retry mechanism**: Support retry for transient failures
4. **Rollback support**: Ability to rollback partial sync on critical errors
5. **Monitoring**: Track sync success/failure rates

## Testing Strategy

1. **Unit tests**: Test sync service logic and conflict resolution
2. **Integration tests**: Test complete sync flow with database
3. **Performance tests**: Test sync with large datasets
4. **Conflict resolution tests**: Test various conflict scenarios
5. **Multi-tenancy tests**: Ensure proper data isolation

## Monitoring and Observability

1. **Sync metrics**: Track sync frequency, duration, and success rates
2. **Conflict tracking**: Monitor conflict resolution patterns
3. **Performance monitoring**: Track sync performance and bottlenecks
4. **Error tracking**: Monitor and alert on sync failures
5. **Audit logs**: Comprehensive logging of sync operations

## Next Steps

1. Begin implementation with Phase 2 (Core Sync Infrastructure)
2. Create sync DTOs and basic service structure
3. Implement conflict resolution algorithms
4. Add comprehensive unit tests
5. Move to API layer implementation

---

*This document will be updated as implementation progresses and requirements evolve.*