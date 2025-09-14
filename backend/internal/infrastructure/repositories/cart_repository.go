package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// CartRepositoryImpl implements CartRepository interface
type CartRepositoryImpl struct {
	db *gorm.DB
}

// NewCartRepository creates a new cart repository
func NewCartRepository(db *gorm.DB) *CartRepositoryImpl {
	return &CartRepositoryImpl{db: db}
}

// Create creates a new cart item
func (r *CartRepositoryImpl) Create(ctx context.Context, cart *entities.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

// GetByID retrieves a cart item by ID
func (r *CartRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart
	err := r.db.WithContext(ctx).Preload("Product").Preload("Shop").Preload("User").First(&cart, id).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// GetByUserID retrieves cart items by user ID
func (r *CartRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Cart, error) {
	var carts []*entities.Cart
	err := r.db.WithContext(ctx).Preload("Product").Preload("Shop").Where("user_id = ?", userID).Find(&carts).Error
	return carts, err
}

// GetByUserAndProduct retrieves a cart item by user and product
func (r *CartRepositoryImpl) GetByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entities.Cart, error) {
	var cart entities.Cart
	err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// Update updates an existing cart item
func (r *CartRepositoryImpl) Update(ctx context.Context, cart *entities.Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

// Delete deletes a cart item (soft delete)
func (r *CartRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Cart{}, id).Error
}

// DeleteByUserID deletes all cart items for a user (soft delete)
func (r *CartRepositoryImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.Cart{}).Error
}

// List retrieves a list of cart items with pagination
func (r *CartRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Cart, error) {
	var carts []*entities.Cart
	err := r.db.WithContext(ctx).Preload("Product").Preload("Shop").Preload("User").Limit(limit).Offset(offset).Find(&carts).Error
	return carts, err
}

// ListByShopIDs retrieves cart items filtered by accessible shop IDs for multi-tenant access
func (r *CartRepositoryImpl) ListByShopIDs(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Cart, error) {
	var carts []*entities.Cart
	if len(shopIDs) == 0 {
		return carts, nil // Return empty slice if no accessible shops
	}
	
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Shop").
		Preload("User").
		Joins("JOIN products ON carts.product_id = products.id").
		Where("products.shop_id IN (?)", shopIDs).
		Limit(limit).
		Offset(offset).
		Find(&carts).Error
		
	return carts, err
}

// GetByShopIDs retrieves all cart items for specified shop IDs (no pagination)
func (r *CartRepositoryImpl) GetByShopIDs(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Cart, error) {
	var carts []*entities.Cart
	if len(shopIDs) == 0 {
		return carts, nil // Return empty slice if no accessible shops
	}
	
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Shop").
		Preload("User").
		Joins("JOIN products ON carts.product_id = products.id").
		Where("products.shop_id IN (?)", shopIDs).
		Find(&carts).Error
		
	return carts, err
}
