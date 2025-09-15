package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/config"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/validators"
	"gorm.io/gorm"
)

// ProductProcessor implements EntityProcessor for Product entities
type ProductProcessor struct {
	*GenericEntityProcessor[entities.Product]
	validator *validators.SyncEntityValidator
}

// NewProductProcessor creates a new product processor
func NewProductProcessor(validator *validators.SyncEntityValidator, config EntityProcessingConfig) *ProductProcessor {
	genericProcessor := NewGenericEntityProcessor(config, func(ctx context.Context, entity entities.Product, operation string) error {
		return validator.ValidateEntity(ctx, entity, operation, nil)
	})
	
	return &ProductProcessor{
		GenericEntityProcessor: genericProcessor,
		validator:              validator,
	}
}

func (p *ProductProcessor) Validate(ctx context.Context, entity entities.Product, operation string) error {
	return p.validator.ValidateEntity(ctx, entity, operation, nil)
}

func (p *ProductProcessor) Create(ctx context.Context, tx *gorm.DB, entity entities.Product) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Create(&entity).Error
}

func (p *ProductProcessor) Update(ctx context.Context, tx *gorm.DB, entity entities.Product) error {
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Save(&entity).Error
}

func (p *ProductProcessor) FindExisting(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := tx.WithContext(ctx).First(&product, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (p *ProductProcessor) ResolveConflict(existing, incoming entities.Product) *dto.ConflictInfo {
	// Last Write Wins strategy
	if existing.UpdatedAt.After(incoming.UpdatedAt) {
		return &dto.ConflictInfo{
			EntityType:   "products",
			EntityID:     existing.ID,
			ConflictType: "timestamp",
			Resolution:   "keep_local",
			Details:      fmt.Sprintf("Local: %s, Incoming: %s", existing.UpdatedAt.Format(time.RFC3339), incoming.UpdatedAt.Format(time.RFC3339)),
		}
	}
	return nil
}

func (p *ProductProcessor) GetEntityType() string {
	return "products"
}

func (p *ProductProcessor) GetEntityID(entity entities.Product) uuid.UUID {
	return entity.ID
}

// CartProcessor implements EntityProcessor for Cart entities
type CartProcessor struct {
	*GenericEntityProcessor[entities.Cart]
	validator *validators.SyncEntityValidator
}

// NewCartProcessor creates a new cart processor
func NewCartProcessor(validator *validators.SyncEntityValidator, config EntityProcessingConfig) *CartProcessor {
	genericProcessor := NewGenericEntityProcessor(config, func(ctx context.Context, entity entities.Cart, operation string) error {
		return validator.ValidateEntity(ctx, entity, operation, nil)
	})
	
	return &CartProcessor{
		GenericEntityProcessor: genericProcessor,
		validator:              validator,
	}
}

func (c *CartProcessor) Validate(ctx context.Context, entity entities.Cart, operation string) error {
	return c.validator.ValidateEntity(ctx, entity, operation, nil)
}

func (c *CartProcessor) Create(ctx context.Context, tx *gorm.DB, entity entities.Cart) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Create(&entity).Error
}

func (c *CartProcessor) Update(ctx context.Context, tx *gorm.DB, entity entities.Cart) error {
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Save(&entity).Error
}

func (c *CartProcessor) FindExisting(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart
	err := tx.WithContext(ctx).First(&cart, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *CartProcessor) ResolveConflict(existing, incoming entities.Cart) *dto.ConflictInfo {
	// Last Write Wins strategy
	if existing.UpdatedAt.After(incoming.UpdatedAt) {
		return &dto.ConflictInfo{
			EntityType:   "carts",
			EntityID:     existing.ID,
			ConflictType: "timestamp",
			Resolution:   "keep_local",
			Details:      fmt.Sprintf("Local: %s, Incoming: %s", existing.UpdatedAt.Format(time.RFC3339), incoming.UpdatedAt.Format(time.RFC3339)),
		}
	}
	return nil
}

func (c *CartProcessor) GetEntityType() string {
	return "carts"
}

func (c *CartProcessor) GetEntityID(entity entities.Cart) uuid.UUID {
	return entity.ID
}

// CategoryProcessor implements EntityProcessor for Category entities
type CategoryProcessor struct {
	*GenericEntityProcessor[entities.Category]
	validator *validators.SyncEntityValidator
}

// NewCategoryProcessor creates a new category processor
func NewCategoryProcessor(validator *validators.SyncEntityValidator, config EntityProcessingConfig) *CategoryProcessor {
	genericProcessor := NewGenericEntityProcessor(config, func(ctx context.Context, entity entities.Category, operation string) error {
		return validator.ValidateEntity(ctx, entity, operation, nil)
	})
	
	return &CategoryProcessor{
		GenericEntityProcessor: genericProcessor,
		validator:              validator,
	}
}

func (c *CategoryProcessor) Validate(ctx context.Context, entity entities.Category, operation string) error {
	return c.validator.ValidateEntity(ctx, entity, operation, nil)
}

func (c *CategoryProcessor) Create(ctx context.Context, tx *gorm.DB, entity entities.Category) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Create(&entity).Error
}

func (c *CategoryProcessor) Update(ctx context.Context, tx *gorm.DB, entity entities.Category) error {
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Save(&entity).Error
}

func (c *CategoryProcessor) FindExisting(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Category, error) {
	var category entities.Category
	err := tx.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (c *CategoryProcessor) ResolveConflict(existing, incoming entities.Category) *dto.ConflictInfo {
	// Last Write Wins strategy
	if existing.UpdatedAt.After(incoming.UpdatedAt) {
		return &dto.ConflictInfo{
			EntityType:   "categories",
			EntityID:     existing.ID,
			ConflictType: "timestamp",
			Resolution:   "keep_local",
			Details:      fmt.Sprintf("Local: %s, Incoming: %s", existing.UpdatedAt.Format(time.RFC3339), incoming.UpdatedAt.Format(time.RFC3339)),
		}
	}
	return nil
}

func (c *CategoryProcessor) GetEntityType() string {
	return "categories"
}

func (c *CategoryProcessor) GetEntityID(entity entities.Category) uuid.UUID {
	return entity.ID
}

// TransactionProcessor implements EntityProcessor for Transaction entities
type TransactionProcessor struct {
	*GenericEntityProcessor[entities.Transaction]
	validator *validators.SyncEntityValidator
}

// NewTransactionProcessor creates a new transaction processor
func NewTransactionProcessor(validator *validators.SyncEntityValidator, config EntityProcessingConfig) *TransactionProcessor {
	genericProcessor := NewGenericEntityProcessor(config, func(ctx context.Context, entity entities.Transaction, operation string) error {
		return validator.ValidateEntity(ctx, entity, operation, nil)
	})
	
	return &TransactionProcessor{
		GenericEntityProcessor: genericProcessor,
		validator:              validator,
	}
}

func (t *TransactionProcessor) Validate(ctx context.Context, entity entities.Transaction, operation string) error {
	return t.validator.ValidateEntity(ctx, entity, operation, nil)
}

func (t *TransactionProcessor) Create(ctx context.Context, tx *gorm.DB, entity entities.Transaction) error {
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Create(&entity).Error
}

func (t *TransactionProcessor) Update(ctx context.Context, tx *gorm.DB, entity entities.Transaction) error {
	entity.UpdatedAt = time.Now()
	return tx.WithContext(ctx).Save(&entity).Error
}

func (t *TransactionProcessor) FindExisting(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entities.Transaction, error) {
	var transaction entities.Transaction
	err := tx.WithContext(ctx).First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (t *TransactionProcessor) ResolveConflict(existing, incoming entities.Transaction) *dto.ConflictInfo {
	// Last Write Wins strategy
	if existing.UpdatedAt.After(incoming.UpdatedAt) {
		return &dto.ConflictInfo{
			EntityType:   "transactions",
			EntityID:     existing.ID,
			ConflictType: "timestamp",
			Resolution:   "keep_local",
			Details:      fmt.Sprintf("Local: %s, Incoming: %s", existing.UpdatedAt.Format(time.RFC3339), incoming.UpdatedAt.Format(time.RFC3339)),
		}
	}
	return nil
}

func (t *TransactionProcessor) GetEntityType() string {
	return "transactions"
}

func (t *TransactionProcessor) GetEntityID(entity entities.Transaction) uuid.UUID {
	return entity.ID
}

// ProcessorFactory creates and manages entity processors
type ProcessorFactory struct {
	validator *validators.SyncEntityValidator
	config    EntityProcessingConfig
}

// NewProcessorFactory creates a new processor factory
func NewProcessorFactory(validator *validators.SyncEntityValidator, config EntityProcessingConfig) *ProcessorFactory {
	return &ProcessorFactory{
		validator: validator,
		config:    config,
	}
}

// GetProductProcessor returns a configured product processor
func (f *ProcessorFactory) GetProductProcessor() EntityProcessor[entities.Product] {
	return NewProductProcessor(f.validator, f.config)
}

// GetCartProcessor returns a configured cart processor
func (f *ProcessorFactory) GetCartProcessor() EntityProcessor[entities.Cart] {
	return NewCartProcessor(f.validator, f.config)
}

// GetCategoryProcessor returns a configured category processor
func (f *ProcessorFactory) GetCategoryProcessor() EntityProcessor[entities.Category] {
	return NewCategoryProcessor(f.validator, f.config)
}

// GetTransactionProcessor returns a configured transaction processor
func (f *ProcessorFactory) GetTransactionProcessor() EntityProcessor[entities.Transaction] {
	return NewTransactionProcessor(f.validator, f.config)
}

// processEntitiesGeneric is a generic processing wrapper for all entity types
func processEntitiesGeneric[T any](
	ctx context.Context,
	tx *gorm.DB,
	entities []T,
	processor EntityProcessor[T],
	syncContext dto.SyncContext,
	response *dto.SyncResponse,
	syncConfig config.SyncConfig,
) error {
	if len(entities) == 0 {
		return nil
	}
	
	processingContext := EntityProcessingContext{
		UserID:        syncContext.UserID,
		LicenseID:     syncContext.LicenseID,
		LastSyncTime:  syncContext.LastSyncTime,
		SyncTimestamp: time.Now(),
		UserRole:      syncContext.UserRole,
	}
	
	// Use configuration from the sync service
	config := EntityProcessingConfig{
		BatchSize:        syncConfig.BatchSize,
		EnableValidation: syncConfig.EnableComprehensiveValidation,
		EnableSavepoints: true, // Always use savepoints for entity processing
		ErrorPolicy:      dto.ParseSyncErrorPolicy(syncConfig.ErrorPolicy),
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
	}
	
	return ProcessEntities(ctx, tx, processor, entities, processingContext, response, config)
}