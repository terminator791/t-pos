package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// ShopUseCase handles shop-related business logic
type ShopUseCase struct {
	shopRepo    repositories.ShopRepository
	licenseRepo repositories.LicenseRepository
	userRepo    repositories.UserRepository
}

// NewShopUseCase creates a new ShopUseCase
func NewShopUseCase(shopRepo repositories.ShopRepository, licenseRepo repositories.LicenseRepository, userRepo repositories.UserRepository) *ShopUseCase {
	return &ShopUseCase{
		shopRepo:    shopRepo,
		licenseRepo: licenseRepo,
		userRepo:    userRepo,
	}
}

// CreateShop creates a new shop
func (uc *ShopUseCase) CreateShop(ctx context.Context, shop *entities.Shop) error {
	if shop.Name == "" {
		return errors.New("shop name is required")
	}

	// Check if license exists
	_, err := uc.licenseRepo.GetByID(ctx, shop.LicenseID)
	if err != nil {
		return errors.New("invalid license ID")
	}

	// Check if owner exists
	_, err = uc.userRepo.GetByID(ctx, shop.UserID)
	if err != nil {
		return errors.New("invalid owner ID")
	}

	return uc.shopRepo.Create(ctx, shop)
}

// GetShop retrieves a shop by ID
func (uc *ShopUseCase) GetShop(ctx context.Context, id uuid.UUID) (*entities.Shop, error) {
	return uc.shopRepo.GetByID(ctx, id)
}

// GetShopsByLicense retrieves shops by license ID
func (uc *ShopUseCase) GetShopsByLicense(ctx context.Context, licenseID uuid.UUID) ([]*entities.Shop, error) {
	return uc.shopRepo.GetByLicenseID(ctx, licenseID)
}

// GetShopsByOwner retrieves shops by owner ID
func (uc *ShopUseCase) GetShopsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*entities.Shop, error) {
	return uc.shopRepo.GetByOwnerID(ctx, ownerID)
}

// UpdateShop updates an existing shop
func (uc *ShopUseCase) UpdateShop(ctx context.Context, shop *entities.Shop) error {
	if shop.ID == uuid.Nil {
		return errors.New("shop ID is required")
	}

	existing, err := uc.shopRepo.GetByID(ctx, shop.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("shop not found")
	}

	return uc.shopRepo.Update(ctx, shop)
}

// DeleteShop deletes a shop
func (uc *ShopUseCase) DeleteShop(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.shopRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("shop not found")
	}

	return uc.shopRepo.Delete(ctx, id)
}

// ListShops retrieves a list of shops
func (uc *ShopUseCase) ListShops(ctx context.Context, limit, offset int) ([]*entities.Shop, error) {
	return uc.shopRepo.List(ctx, limit, offset)
}
