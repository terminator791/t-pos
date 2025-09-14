package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"gorm.io/gorm"
)

// CartUseCase handles cart-related business logic
type CartUseCase struct {
	db          *gorm.DB
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
	userRepo    repositories.UserRepository
	shopRepo    repositories.ShopRepository
}

// NewCartUseCase creates a new CartUseCase
func NewCartUseCase(db *gorm.DB, cartRepo repositories.CartRepository, productRepo repositories.ProductRepository, userRepo repositories.UserRepository, shopRepo repositories.ShopRepository) *CartUseCase {
	return &CartUseCase{
		db:          db,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
		shopRepo:    shopRepo,
	}
}

// AddToCart adds a product to user's cart or updates quantity if already exists
func (uc *CartUseCase) AddToCart(ctx context.Context, userID, productID, shopID uuid.UUID, quantity int) (*entities.Cart, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate product exists
	_, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("product not found")
	}

	// Validate user exists
	_, err = uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("user not found")
	}

	// Validate shop exists
	_, err = uc.shopRepo.GetByID(ctx, shopID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("shop not found")
	}

	// Check if product is already in cart with row lock
	var existingCart entities.Cart
	err = tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("user_id = ? AND product_id = ?", userID, productID).First(&existingCart).Error
	if err == nil {
		// Update quantity if item already exists
		existingCart.Quantity += quantity
		err = tx.WithContext(ctx).Save(&existingCart).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return uc.cartRepo.GetByID(ctx, existingCart.ID)
	}

	// Create new cart item
	cart := &entities.Cart{
		UserID:    userID,
		ProductID: productID,
		ShopID:    shopID,
		Quantity:  quantity,
	}

	err = tx.WithContext(ctx).Create(cart).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Return cart with loaded relationships
	return uc.cartRepo.GetByID(ctx, cart.ID)
}

// GetUserCart retrieves all cart items for a user
func (uc *CartUseCase) GetUserCart(ctx context.Context, userID uuid.UUID) ([]*entities.Cart, error) {
	return uc.cartRepo.GetByUserID(ctx, userID)
}

// RemoveFromCart removes a product from user's cart
func (uc *CartUseCase) RemoveFromCart(ctx context.Context, cartID uuid.UUID, userID uuid.UUID) error {
	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get cart item with row lock
	var cart entities.Cart
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", cartID).First(&cart).Error
	if err != nil {
		tx.Rollback()
		return errors.New("cart item not found")
	}

	// Verify the cart belongs to the user
	if cart.UserID != userID {
		tx.Rollback()
		return errors.New("unauthorized: cart item doesn't belong to user")
	}

	// Delete cart item
	err = tx.WithContext(ctx).Delete(&entities.Cart{}, "id = ?", cartID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// ClearUserCart removes all items from user's cart
func (uc *CartUseCase) ClearUserCart(ctx context.Context, userID uuid.UUID) error {
	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all cart items for user
	err := tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.Cart{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetCartItem retrieves a specific cart item
func (uc *CartUseCase) GetCartItem(ctx context.Context, cartID uuid.UUID) (*entities.Cart, error) {
	return uc.cartRepo.GetByID(ctx, cartID)
}

// UpdateCartQuantity updates the quantity of a cart item
func (uc *CartUseCase) UpdateCartQuantity(ctx context.Context, cartID uuid.UUID, userID uuid.UUID, quantity int) (*entities.Cart, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get cart item with row lock
	var cart entities.Cart
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", cartID).First(&cart).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("cart item not found")
	}

	// Verify the cart belongs to the user
	if cart.UserID != userID {
		tx.Rollback()
		return nil, errors.New("unauthorized: cart item doesn't belong to user")
	}

	cart.Quantity = quantity
	err = tx.WithContext(ctx).Save(&cart).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return uc.cartRepo.GetByID(ctx, cartID)
}

// ListCarts retrieves a list of all cart items (admin function)
func (uc *CartUseCase) ListCarts(ctx context.Context, limit, offset int) ([]*entities.Cart, error) {
	return uc.cartRepo.List(ctx, limit, offset)
}

// ListCartsByShopIDs retrieves cart items filtered by accessible shop IDs for multi-tenant access
func (uc *CartUseCase) ListCartsByShopIDs(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Cart, error) {
	return uc.cartRepo.ListByShopIDs(ctx, shopIDs, limit, offset)
}
