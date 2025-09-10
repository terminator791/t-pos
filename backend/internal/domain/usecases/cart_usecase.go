package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// CartUseCase handles cart-related business logic
type CartUseCase struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
	userRepo    repositories.UserRepository
	shopRepo    repositories.ShopRepository
}

// NewCartUseCase creates a new CartUseCase
func NewCartUseCase(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository, userRepo repositories.UserRepository, shopRepo repositories.ShopRepository) *CartUseCase {
	return &CartUseCase{
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

	// Validate product exists
	_, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// Validate user exists
	_, err = uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Validate shop exists
	_, err = uc.shopRepo.GetByID(ctx, shopID)
	if err != nil {
		return nil, errors.New("shop not found")
	}

	// Check if product is already in cart
	existingCart, err := uc.cartRepo.GetByUserAndProduct(ctx, userID, productID)
	if err == nil && existingCart != nil {
		// Update quantity if item already exists
		existingCart.Quantity += quantity
		err = uc.cartRepo.Update(ctx, existingCart)
		if err != nil {
			return nil, err
		}
		return uc.cartRepo.GetByID(ctx, existingCart.ID)
	}

	cart := &entities.Cart{
		UserID:    userID,
		ProductID: productID,
		ShopID:    shopID,
		Quantity:  quantity,
	}

	err = uc.cartRepo.Create(ctx, cart)
	if err != nil {
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
	cart, err := uc.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return errors.New("cart item not found")
	}

	// Verify the cart belongs to the user
	if cart.UserID != userID {
		return errors.New("unauthorized: cart item doesn't belong to user")
	}

	return uc.cartRepo.Delete(ctx, cartID)
}

// ClearUserCart removes all items from user's cart
func (uc *CartUseCase) ClearUserCart(ctx context.Context, userID uuid.UUID) error {
	return uc.cartRepo.DeleteByUserID(ctx, userID)
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

	cart, err := uc.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	// Verify the cart belongs to the user
	if cart.UserID != userID {
		return nil, errors.New("unauthorized: cart item doesn't belong to user")
	}

	cart.Quantity = quantity
	err = uc.cartRepo.Update(ctx, cart)
	if err != nil {
		return nil, err
	}

	return uc.cartRepo.GetByID(ctx, cartID)
}

// ListCarts retrieves a list of all cart items (admin function)
func (uc *CartUseCase) ListCarts(ctx context.Context, limit, offset int) ([]*entities.Cart, error) {
	return uc.cartRepo.List(ctx, limit, offset)
}