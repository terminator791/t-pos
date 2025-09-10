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
- [x] Create sync DTOs and request/response structures (`internal/domain/dto/sync.go`)
- [x] Implement basic sync service structure with core business logic (`internal/application/services/sync_service.go`)
- [x] Add conflict resolution algorithms (Last Write Wins)
- [x] Create missing repository interfaces (StockHistoryRepository exists and working)
- [x] Implement license-based data filtering framework
- [x] Add basic unit tests for sync service

**Phase 3: API Layer Implementation ✅**
- [x] Create sync handler for HTTP endpoints (`internal/interfaces/http/handlers/sync_handler.go`)
- [x] Add proper error handling and validation
- [x] Create sync routes and middleware integration (integrated in `routes.go`)
- [x] Implement JWT authentication and license-based authorization
- [x] Add comprehensive input validation (entity limits, UUID validation)
- [x] Integrate sync handler into main application (`cmd/main.go`)

## Undone Progress

**Phase 4: Complete Entity Implementation**
- [x] Cart entity sync implementation (partial - push/pull logic implemented)
- [ ] Categories entity sync implementation (skeleton exists, needs completion)
- [ ] Products entity sync implementation (skeleton exists, needs completion) 
- [ ] Transactions entity sync implementation (skeleton exists, needs completion)
- [ ] Payments entity sync implementation (not started)
- [ ] Receipts entity sync implementation (not started)
- [ ] Histories entity sync implementation (not started)
- [ ] Expenses entity sync implementation (not started)
- [ ] Shops entity sync implementation (not started)
- [ ] StockHistories entity sync implementation (not started)
- [ ] TransactionProducts entity sync implementation (not started)
- [ ] Users entity sync implementation (not started)
- [ ] Add comprehensive conflict resolution for all entity relationships
- [ ] Optimize database queries for large datasets

**Phase 5: Documentation and API Enhancement**
- [x] Basic API documentation exists (`backend/docs/SYNC_API.md`)
- [ ] Enhance SYNC_API.md with comprehensive request/response examples
- [ ] Add detailed conflict resolution examples
- [ ] Document error handling scenarios
- [ ] Add usage examples for different client scenarios

**Phase 6: Advanced Features (Optional)**
- [ ] Implement async worker for sync processing (currently synchronous)
- [ ] Add sync job status tracking
- [ ] Implement batch sync optimization
- [ ] Add sync monitoring and metrics

**Phase 7: Integration and Testing**
- [x] Basic unit tests for sync service exist
- [x] Basic unit tests for sync handler exist  
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

## Future Task Recommendations & Next Steps

### **Immediate Priority (Next Session)**

#### 1. Complete Core Entity Implementations ⚡
- **Categories sync**: Complete the `pushCategories` and `pullCategories` methods
- **Products sync**: Implement full product synchronization with stock updates
- **Transactions sync**: Implement complex transaction sync with related entities
- **Payment & Receipt sync**: Link payment processing with transaction completion

#### 2. Entity Relationship Management 🔗
- **Foreign key validation**: Ensure related entities exist before sync
- **Cascade sync logic**: When syncing transactions, automatically sync related transaction_products
- **Dependency ordering**: Sync categories before products, products before transactions

#### 3. Enhanced Conflict Resolution 💥
```go
// Implement field-level conflict resolution
type FieldConflict struct {
    Field       string `json:"field"`
    ServerValue any    `json:"server_value"`
    ClientValue any    `json:"client_value"`
    Resolution  string `json:"resolution"`
}
```

### **Medium Priority (Future Development)**

#### 4. Performance Optimizations 🚀
- **Batch processing**: Process entities in configurable batches (e.g., 100 entities per batch)
- **Parallel processing**: Use goroutines for concurrent entity processing
- **Database indexing**: Optimize queries with proper indexes on `updated_at`, `license_id`
- **Connection pooling**: Implement database connection pooling for high concurrency

#### 5. Advanced Sync Features 🛠️
```go
// Selective sync - allow clients to sync only specific entity types
type SelectiveSyncRequest struct {
    LastSyncTimestamp *time.Time `json:"last_sync_timestamp"`
    EntityTypes      []string    `json:"entity_types"` // ["products", "transactions"]
    MaxEntities      *int        `json:"max_entities"` // Limit response size
}
```

#### 6. Async Sync Processing (Optional) ⚙️
```go
// Implement async worker pattern
type SyncWorker struct {
    jobQueue    chan SyncJob
    workerPool  chan chan SyncJob
    quit        chan bool
    maxWorkers  int
}

// For large datasets, return job ID and process asynchronously
func (h *SyncHandler) ProcessSyncAsync(c *gin.Context) {
    jobID := uuid.New()
    // Enqueue sync job
    // Return job ID immediately
    response.SuccessOK(c, "Sync job queued", gin.H{"job_id": jobID})
}
```

### **Long-term Enhancements**

#### 7. Monitoring & Observability 📊
- **Sync metrics**: Track sync frequency, duration, success rates per license
- **Conflict analytics**: Analyze conflict patterns to improve UX
- **Performance monitoring**: Database query performance, memory usage
- **Alerting**: Email/SMS alerts for sync failures, high conflict rates

#### 8. Data Validation & Quality 🛡️
```go
// Enhanced validation rules
type EntityValidator interface {
    ValidateCreate(entity interface{}) []ValidationError
    ValidateUpdate(existing, incoming interface{}) []ValidationError
    ValidateBusinessRules(entity interface{}) []ValidationError
}
```

#### 9. Sync Strategies & Policies 📋
- **Sync policies**: Configurable sync strategies per license (aggressive, conservative, manual)
- **Conflict resolution strategies**: Allow different strategies per entity type
- **Data retention**: Configurable data retention policies for sync history

#### 10. Advanced Security Features 🔒
- **Field-level encryption**: Encrypt sensitive data in sync payloads
- **Audit logging**: Comprehensive audit logs for all sync operations
- **Rate limiting**: Prevent sync API abuse with configurable rate limits
- **Data anonymization**: Option to anonymize customer data in sync

### **Technical Debt & Code Quality**

#### 11. Testing & Documentation 🧪
```go
// Comprehensive integration tests
func TestFullSyncFlow(t *testing.T) {
    // Setup test database with sample data
    // Perform full sync cycle
    // Verify data consistency
    // Test conflict resolution
    // Validate performance metrics
}
```

#### 12. Error Handling & Resilience 🔄
- **Retry mechanism**: Automatic retry with exponential backoff
- **Partial sync recovery**: Resume interrupted sync operations
- **Data consistency checks**: Verify data integrity after sync
- **Rollback capability**: Ability to rollback failed sync operations

#### 13. API Versioning & Backward Compatibility 🔄
```go
// Support multiple API versions
type SyncRequestV2 struct {
    Version           string     `json:"version"` // "2.0"
    LastSyncTimestamp *time.Time `json:"last_sync_timestamp"`
    // Enhanced fields...
}
```

### **Performance Benchmarks & Goals**

#### Target Performance Metrics:
- **Sync processing**: < 500ms for 100 entities
- **Large sync**: < 5 seconds for 1000 entities
- **Conflict resolution**: < 50ms per conflict
- **Database queries**: < 100ms per entity type pull
- **Memory usage**: < 100MB per sync operation
- **Concurrent syncs**: Support 50+ simultaneous sync operations

#### Scalability Targets:
- **Users per license**: 100+ concurrent users
- **Entities per sync**: 1000+ entities per request
- **Daily sync volume**: 10,000+ sync operations per day
- **Data size**: Support shops with 100,000+ products

### **Development Workflow Recommendations**

1. **Start with Categories**: Simplest entity, good for testing the sync pattern
2. **Move to Products**: More complex with stock management
3. **Implement Transactions**: Most complex with multiple relationships
4. **Add comprehensive tests**: Integration tests for each entity type
5. **Performance testing**: Load test with realistic data volumes
6. **Documentation updates**: Keep API docs current with implementation

### **Code Organization Suggestions**

```
internal/application/services/sync/
├── sync_service.go           # Main service orchestrator
├── entity_processors/        # Individual entity processors
│   ├── cart_processor.go
│   ├── category_processor.go
│   ├── product_processor.go
│   └── transaction_processor.go
├── conflict_resolution/      # Conflict resolution strategies
│   ├── resolver.go
│   └── strategies.go
├── validators/              # Entity-specific validators
│   ├── cart_validator.go
│   └── product_validator.go
└── workers/                # Async processing (optional)
    ├── sync_worker.go
    └── job_queue.go
```

This organization promotes:
- **Single Responsibility**: Each processor handles one entity type
- **Testability**: Easy to unit test individual components
- **Maintainability**: Changes to one entity don't affect others
- **Scalability**: Easy to add new entity types or conflict strategies

---

*This document will continue to be updated as implementation progresses and requirements evolve.*