package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"gorm.io/gorm"
)

// EntityProcessor defines the interface for generic entity processing
type EntityProcessor[T any] interface {
	// Validate validates the entity for the given operation
	Validate(ctx context.Context, entity T, operation string) error
	
	// Create creates a new entity in the database
	Create(ctx context.Context, tx *gorm.DB, entity T) error
	
	// Update updates an existing entity in the database
	Update(ctx context.Context, tx *gorm.DB, entity T) error
	
	// FindExisting finds an existing entity by ID
	FindExisting(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*T, error)
	
	// ResolveConflict handles conflicts between existing and incoming entities
	ResolveConflict(existing, incoming T) *dto.ConflictInfo
	
	// GetEntityType returns the entity type name for logging and error reporting
	GetEntityType() string
	
	// GetEntityID extracts the ID from the entity
	GetEntityID(entity T) uuid.UUID
}

// EntityProcessingConfig contains configuration for entity processing
type EntityProcessingConfig struct {
	BatchSize        int
	EnableValidation bool
	EnableSavepoints bool
	ErrorPolicy      dto.SyncErrorPolicy
	MaxRetries       int
	RetryDelay       time.Duration
}

// DefaultEntityProcessingConfig returns default configuration
func DefaultEntityProcessingConfig() EntityProcessingConfig {
	return EntityProcessingConfig{
		BatchSize:        100,
		EnableValidation: true,
		EnableSavepoints: true,
		ErrorPolicy:      dto.ContinueOnError,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}
}

// EntityProcessingContext contains context information for entity processing
type EntityProcessingContext struct {
	UserID        uuid.UUID
	LicenseID     uuid.UUID
	LastSyncTime  time.Time
	SyncTimestamp time.Time
	UserRole      string
}

// GenericEntityProcessor provides common functionality for entity processing
type GenericEntityProcessor[T any] struct {
	config    EntityProcessingConfig
	validator func(ctx context.Context, entity T, operation string) error
}

// NewGenericEntityProcessor creates a new generic entity processor
func NewGenericEntityProcessor[T any](
	config EntityProcessingConfig,
	validator func(ctx context.Context, entity T, operation string) error,
) *GenericEntityProcessor[T] {
	return &GenericEntityProcessor[T]{
		config:    config,
		validator: validator,
	}
}

// ProcessEntities processes a batch of entities using the generic framework
func ProcessEntities[T any](
	ctx context.Context,
	tx *gorm.DB,
	processor EntityProcessor[T],
	entities []T,
	processingContext EntityProcessingContext,
	response *dto.SyncResponse,
	config EntityProcessingConfig,
) error {
	entityType := processor.GetEntityType()
	startTime := time.Now()
	
	log.Printf("Processing %d %s entities with generic framework", len(entities), entityType)
	
	// Process entities in batches
	for i := 0; i < len(entities); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(entities) {
			end = len(entities)
		}
		
		batch := entities[i:end]
		if err := processBatch(ctx, tx, processor, batch, processingContext, response, config); err != nil {
			return fmt.Errorf("failed to process %s batch %d-%d: %w", entityType, i, end-1, err)
		}
	}
	
	processingTime := time.Since(startTime)
	log.Printf("Completed processing %d %s entities in %v", len(entities), entityType, processingTime)
	
	// Update response statistics
	updateResponseStats(response, entityType, len(entities), processingTime)
	
	return nil
}

// processBatch processes a single batch of entities
func processBatch[T any](
	ctx context.Context,
	tx *gorm.DB,
	processor EntityProcessor[T],
	batch []T,
	processingContext EntityProcessingContext,
	response *dto.SyncResponse,
	config EntityProcessingConfig,
) error {
	for _, entity := range batch {
		if err := processEntityWithErrorHandling(ctx, tx, processor, entity, processingContext, response, config); err != nil {
			// Apply error policy
			if config.ErrorPolicy == dto.AbortOnError {
				return fmt.Errorf("entity processing failed (abort policy): %w", err)
			}
			// Continue on error - error already added to response
		}
	}
	return nil
}

// processEntityWithErrorHandling processes a single entity with comprehensive error handling
func processEntityWithErrorHandling[T any](
	ctx context.Context,
	tx *gorm.DB,
	processor EntityProcessor[T],
	entity T,
	processingContext EntityProcessingContext,
	response *dto.SyncResponse,
	config EntityProcessingConfig,
) error {
	entityID := processor.GetEntityID(entity)
	
	// Use savepoint for error isolation if enabled
	if config.EnableSavepoints {
		return processEntityWithSavepoint(ctx, tx, entityID, func() error {
			return processEntity(ctx, tx, processor, entity, processingContext, response, config)
		})
	}
	
	return processEntity(ctx, tx, processor, entity, processingContext, response, config)
}

// processEntity contains the core entity processing logic
func processEntity[T any](
	ctx context.Context,
	tx *gorm.DB,
	processor EntityProcessor[T],
	entity T,
	processingContext EntityProcessingContext,
	response *dto.SyncResponse,
	config EntityProcessingConfig,
) error {
	entityID := processor.GetEntityID(entity)
	entityType := processor.GetEntityType()
	
	// 1. Validation (if enabled)
	if config.EnableValidation {
		if err := processor.Validate(ctx, entity, "create"); err != nil {
			addEntityError(response, entityType, entityID, "validation_failed", err, map[string]interface{}{
				"operation": "create",
				"stage":     "validation",
			})
			return err
		}
	}
	
	// 2. Check if entity exists
	existing, err := processor.FindExisting(ctx, tx, entityID)
	if err != nil && err != gorm.ErrRecordNotFound {
		addEntityError(response, entityType, entityID, "lookup_failed", err, map[string]interface{}{
			"operation": "lookup",
			"stage":     "existence_check",
		})
		return err
	}
	
	// 3. Process based on existence
	if existing == nil {
		// Create new entity
		if err := processor.Create(ctx, tx, entity); err != nil {
			addEntityError(response, entityType, entityID, "create_failed", err, map[string]interface{}{
				"operation": "create",
				"stage":     "database_insert",
			})
			return err
		}
		updateEntityStats(response, entityType, "created")
	} else {
		// Check for conflicts and update
		if conflict := processor.ResolveConflict(*existing, entity); conflict != nil {
			response.Conflicts = append(response.Conflicts, *conflict)
			updateEntityStats(response, entityType, "conflicted")
		}
		
		// Validate for update if enabled
		if config.EnableValidation {
			if err := processor.Validate(ctx, entity, "update"); err != nil {
				addEntityError(response, entityType, entityID, "update_validation_failed", err, map[string]interface{}{
					"operation": "update",
					"stage":     "validation",
				})
				return err
			}
		}
		
		if err := processor.Update(ctx, tx, entity); err != nil {
			addEntityError(response, entityType, entityID, "update_failed", err, map[string]interface{}{
				"operation": "update",
				"stage":     "database_update",
			})
			return err
		}
		updateEntityStats(response, entityType, "updated")
	}
	
	return nil
}

// processEntityWithSavepoint executes entity processing within a savepoint for error isolation
func processEntityWithSavepoint(ctx context.Context, tx *gorm.DB, entityID uuid.UUID, processFunc func() error) error {
	savepointName := fmt.Sprintf("sp_%s", entityID.String()[:8])
	
	// Create savepoint
	if err := tx.SavePoint(savepointName).Error; err != nil {
		log.Printf("Failed to create savepoint %s, proceeding without: %v", savepointName, err)
		return processFunc()
	}
	
	// Execute processing
	if err := processFunc(); err != nil {
		// Rollback to savepoint on error
		if rollbackErr := tx.RollbackTo(savepointName).Error; rollbackErr != nil {
			log.Printf("Failed to rollback to savepoint %s: %v", savepointName, rollbackErr)
		}
		return err
	}
	
	return nil
}

// addEntityError adds an error to the sync response
func addEntityError(response *dto.SyncResponse, entityType string, entityID uuid.UUID, errorCode string, err error, details map[string]interface{}) {
	syncError := dto.SyncError{
		EntityType: entityType,
		EntityID:   entityID,
		ErrorCode:  errorCode,
		Message:    err.Error(),
		Details:    fmt.Sprintf("Details: %+v", details),
	}
	response.Errors = append(response.Errors, syncError)
	response.Stats.ErrorCount++
}

// updateEntityStats updates the processing statistics in the response
func updateEntityStats(response *dto.SyncResponse, entityType string, operation string) {
	if response.Stats.ProcessedEntities == nil {
		response.Stats.ProcessedEntities = make(map[string]int)
	}
	if response.Stats.CreatedEntities == nil {
		response.Stats.CreatedEntities = make(map[string]int)
	}
	if response.Stats.UpdatedEntities == nil {
		response.Stats.UpdatedEntities = make(map[string]int)
	}
	
	response.Stats.ProcessedEntities[entityType]++
	
	switch operation {
	case "created":
		response.Stats.CreatedEntities[entityType]++
	case "updated":
		response.Stats.UpdatedEntities[entityType]++
	case "conflicted":
		response.Stats.ConflictCount++
	}
}

// updateResponseStats updates overall processing statistics
func updateResponseStats(response *dto.SyncResponse, entityType string, count int, processingTime time.Duration) {
	if response.Stats.ProcessedEntities == nil {
		response.Stats.ProcessedEntities = make(map[string]int)
	}
	
	response.Stats.ProcessedEntities[entityType] += count
	response.Stats.ProcessingTimeMs += processingTime.Milliseconds()
}