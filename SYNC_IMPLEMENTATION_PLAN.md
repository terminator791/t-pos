# Data Synchronization Implementation Plan

## Project Overview

This document outlines the comprehensive implementation plan for the data synchronization system between the t-pos mobile application (SQLite) and the backend server (PostgreSQL). The synchronization system will enable seamless two-way data sync for offline-capable mobile POS operations.

## Current Backend Analysis

### Database Entities Requiring Synchronization
The following entities are identified for synchronization with identical structures between mobile and server:
- **Carts** - Shopping cart items before transaction
- **Categories** - Product categorization per shop
- **Expenses** - Shop expense records
- **Histories** - Transaction history links
- **Payments** - Payment records linked to transactions
- **Products** - Sellable items with stock management
- **Receipts** - Receipt records pointing to payments
- **Shops** - Merchant shops under license
- **Stock_Histories** - Append-only stock change tracking
- **Transaction_Products** - Line items per transaction
- **Transactions** - Sales transaction headers
- **Users** - Application users (filtered by license/shop access)

### Current Architecture Strengths
- ✅ Clean architecture with domain entities, repositories, services, handlers
- ✅ GORM ORM with PostgreSQL database
- ✅ UUID primary keys across all entities (perfect for offline creation)
- ✅ Consistent `created_at` and `updated_at` timestamps
- ✅ JWT authentication with license_id support for multi-tenancy
- ✅ Casbin-based authorization with role-based access control
- ✅ Gin HTTP framework with middleware support

---

## Implementation Plan

### Phase 1: Core Sync Infrastructure 🔧

#### 1.1 Sync Data Models and DTOs
**File:** `backend/internal/domain/sync/models.go`
- [ ] Create sync request/response DTOs for each entity
- [ ] Define SyncMetadata structure for tracking sync state
- [ ] Implement ConflictResolution models
- [ ] Create batch operation structures

#### 1.2 Conflict Resolution Framework
**File:** `backend/internal/domain/sync/conflicts.go`
- [ ] Implement Last-Write-Wins (LWW) conflict resolution
- [ ] Create timestamp comparison utilities
- [ ] Define conflict detection logic
- [ ] Implement data validation rules

### Phase 2: Service Layer Implementation 🏗️

#### 2.1 Core Sync Service
**File:** `backend/internal/application/services/sync_service.go`
- [ ] Create main SyncService with license-based filtering
- [ ] Implement push operation processing
- [ ] Implement pull operation with delta sync
- [ ] Add comprehensive error handling and logging
- [ ] Create transaction management for atomic operations

#### 2.2 Repository Layer for Batch Operations
**File:** `backend/internal/infrastructure/repositories/sync_repository.go`
- [ ] Implement efficient batch insert/update operations
- [ ] Create delta queries using timestamp filters
- [ ] Add license-based data filtering
- [ ] Implement shop-level access control for cashiers
- [ ] Create optimized queries for large datasets

### Phase 3: API Layer Implementation 🌐

#### 3.1 Sync HTTP Handlers
**File:** `backend/internal/interfaces/http/handlers/sync_handler.go`
- [ ] Implement POST /api/v1/sync/push endpoint
- [ ] Implement GET /api/v1/sync/pull endpoint
- [ ] Implement POST /api/v1/sync/full endpoint (push + pull)
- [ ] Add comprehensive request validation
- [ ] Implement proper HTTP status codes and error responses

#### 3.2 Route Configuration
**File:** `backend/internal/interfaces/http/routes/routes.go`
- [ ] Add sync routes to protected API group
- [ ] Configure authentication and authorization middleware
- [ ] Set up proper permission checks for sync operations

### Phase 4: Business Logic and Security 🔐

#### 4.1 Multi-Tenancy and Access Control
- [ ] Implement license-based data isolation
- [ ] Add shop-level restrictions for cashier roles
- [ ] Create owner_business access to all shops under license
- [ ] Validate data ownership before sync operations

#### 4.2 Data Integrity and Validation
- [ ] Implement comprehensive data validation
- [ ] Add referential integrity checks
- [ ] Create business rule validations
- [ ] Handle foreign key relationships properly

### Phase 5: Performance and Scalability 🚀

#### 5.1 Async Processing Foundation
- [ ] Design worker pattern for background processing
- [ ] Create job queue structure (foundation for future Redis/RabbitMQ)
- [ ] Implement timeout handling and retry logic
- [ ] Add progress tracking for long-running sync operations

#### 5.2 Optimization and Monitoring
- [ ] Implement query optimization for large datasets
- [ ] Add sync operation metrics and logging
- [ ] Create performance monitoring endpoints
- [ ] Implement rate limiting for sync operations

---

## Technical Specifications

### API Endpoints Design

#### POST /api/v1/sync/push
```json
{
  "last_sync_timestamp": "2025-01-10T10:00:00Z",
  "data": {
    "products": [...],
    "categories": [...],
    "transactions": [...],
    // ... other entities
  }
}
```

#### GET /api/v1/sync/pull?since=2025-01-10T10:00:00Z
```json
{
  "data": {
    "products": [...],
    "categories": [...],
    "transactions": [...]
  },
  "sync_timestamp": "2025-01-10T10:30:00Z"
}
```

### Conflict Resolution Strategy
- **Primary Strategy:** Last-Write-Wins (LWW) using `updated_at` timestamps
- **Validation:** Data integrity checks before applying changes
- **Rollback:** Transaction-based rollback on validation failures
- **Audit:** Log all conflict resolutions for debugging

### Multi-Tenancy Implementation
- **License Isolation:** All queries filtered by user's license_id from JWT
- **Shop Restrictions:** Cashier users restricted to their assigned shop_id
- **Owner Access:** owner_business role can access all shops under their license

### Performance Considerations
- **Batch Size:** Process sync data in configurable batches (default: 1000 records)
- **Pagination:** Support for paginated responses on large datasets
- **Timeout:** Configure reasonable timeouts for mobile clients
- **Async Processing:** Foundation for background processing of large sync operations

---

## Current Progress

### ✅ Completed Tasks
- [x] **Backend Analysis**: Analyzed current database schema and architecture
- [x] **Entity Identification**: Identified 12 entities requiring synchronization
- [x] **Authentication Review**: Confirmed JWT token structure with license_id support
- [x] **Build Verification**: Verified current codebase builds successfully
- [x] **Architecture Design**: Designed clean sync architecture following existing patterns
- [x] **Implementation Planning**: Created comprehensive phased implementation plan

### 🚧 Current Phase: Core Sync Infrastructure
**Next Task**: Create sync data models and DTOs in `backend/internal/domain/sync/models.go`

### 📋 Undone Progress (Next Session Tasks)

#### Immediate Next Steps (Phase 1)
- [ ] Create sync data models and DTOs (`backend/internal/domain/sync/models.go`)
- [ ] Implement conflict resolution framework (`backend/internal/domain/sync/conflicts.go`)
- [ ] Create core sync service (`backend/internal/application/services/sync_service.go`)
- [ ] Implement sync repository layer (`backend/internal/infrastructure/repositories/sync_repository.go`)

#### Subsequent Implementation (Phases 2-5)
- [ ] Create sync HTTP handlers (`backend/internal/interfaces/http/handlers/sync_handler.go`)
- [ ] Add sync routes to router configuration
- [ ] Implement license-based multi-tenancy and access control
- [ ] Add comprehensive data validation and business rules
- [ ] Create async processing foundation for performance optimization
- [ ] Add comprehensive testing and error handling
- [ ] Implement monitoring and metrics collection

---

## Development Guidelines

### Code Quality Standards
- Follow existing project patterns and conventions
- Maintain clean architecture separation
- Use comprehensive error handling with proper logging
- Implement thorough input validation
- Follow Go best practices and idioms

### Testing Strategy
- Unit tests for sync service logic
- Integration tests for repository operations
- API endpoint tests with authentication
- Performance tests for large datasets
- Conflict resolution scenario testing

### Security Considerations
- Validate all incoming sync data
- Ensure proper authorization checks
- Implement rate limiting
- Log all sync operations for audit
- Protect against injection attacks

---

## Future Enhancements

### Phase 6: Advanced Features (Future Scope)
- [ ] Real-time sync notifications using WebSocket/SSE
- [ ] Incremental backup and restore functionality
- [ ] Advanced conflict resolution strategies (operational transforms)
- [ ] Distributed cache integration (Redis)
- [ ] Message queue integration (RabbitMQ/Apache Kafka)
- [ ] Sync analytics and reporting dashboard
- [ ] Mobile sync SDK for easier integration

### Monitoring and Observability
- [ ] Sync operation metrics (Prometheus/Grafana)
- [ ] Distributed tracing for sync operations
- [ ] Health check endpoints for sync services
- [ ] Performance profiling and optimization
- [ ] Alert system for sync failures

---

## Conclusion

This implementation plan provides a comprehensive roadmap for building a robust, scalable, and secure data synchronization system for the t-pos application. The phased approach ensures steady progress while maintaining code quality and system reliability.

The foundation built in the initial phases will support future enhancements and scaling requirements as the application grows.