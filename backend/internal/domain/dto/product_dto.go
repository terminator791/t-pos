package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ProductListDTO represents a product with simplified relations for listing
type ProductListDTO struct {
	ID          uuid.UUID  `json:"id"`
	ShopID      uuid.UUID  `json:"shop_id"`
	CatID       *uuid.UUID `json:"cat_id"`
	Photo       *string    `json:"photo"`
	Name        string     `json:"name"`
	Barcode     *string    `json:"barcode"`
	Unit        *string    `json:"unit"`
	PPN         *float64   `json:"ppn"`
	Sale        float64    `json:"sale"`
	Buy         float64    `json:"buy"`
	Profit      *float64   `json:"profit"`
	Stock       int        `json:"stock"`
	IsSchedule  bool       `json:"is_schedule"`
	Schedule    *string    `json:"schedule"`
	Qty         *int       `json:"qty"`
	IsHaveStock bool       `json:"is_have_stock"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Simplified relationships - only direct hierarchy
	Shop     *ShopListDTO     `json:"shop,omitempty"`
	Category *CategoryListDTO `json:"category,omitempty"`
}

// ShopListDTO represents a simplified shop for product listing
type ShopListDTO struct {
	ID              uuid.UUID `json:"id"`
	LicenseID       uuid.UUID `json:"license_id"`
	UserID          uuid.UUID `json:"user_id"`
	Name            string    `json:"name"`
	Domain          string    `json:"domain"`
	Photo           *string   `json:"photo"`
	Address         *string   `json:"address"`
	Slogan          *string   `json:"slogan"`
	ProfitCalculate int64     `json:"profit_calculate"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CategoryListDTO represents a category with simplified shop relation for listing
type CategoryListDTO struct {
	ID        uuid.UUID    `json:"id"`
	ShopID    uuid.UUID    `json:"shop_id"`
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Shop      *ShopListDTO `json:"shop,omitempty"`
}

// CategoryResponseDTO represents a category for API responses without nested relationships
type CategoryResponseDTO struct {
	ID        uuid.UUID `json:"id"`
	ShopID    uuid.UUID `json:"shop_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductToListDTO converts a Product entity to ProductListDTO
func ProductToListDTO(product *entities.Product) *ProductListDTO {
	dto := &ProductListDTO{
		ID:          product.ID,
		ShopID:      product.ShopID,
		CatID:       product.CatID,
		Photo:       product.Photo,
		Name:        product.Name,
		Barcode:     product.Barcode,
		Unit:        product.Unit,
		PPN:         product.PPN,
		Sale:        product.Sale,
		Buy:         product.Buy,
		Profit:      product.Profit,
		Stock:       product.Stock,
		IsSchedule:  product.IsSchedule,
		Schedule:    product.Schedule,
		Qty:         product.Qty,
		IsHaveStock: product.IsHaveStock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	// Convert shop if loaded
	if product.Shop.ID != uuid.Nil {
		dto.Shop = &ShopListDTO{
			ID:              product.Shop.ID,
			LicenseID:       product.Shop.LicenseID,
			UserID:          product.Shop.UserID,
			Name:            product.Shop.Name,
			Domain:          product.Shop.Domain,
			Photo:           product.Shop.Photo,
			Address:         product.Shop.Address,
			Slogan:          product.Shop.Slogan,
			ProfitCalculate: product.Shop.ProfitCalculate,
			CreatedAt:       product.Shop.CreatedAt,
			UpdatedAt:       product.Shop.UpdatedAt,
		}
	}

	// Convert category if loaded
	if product.Category != nil {
		dto.Category = &CategoryListDTO{
			ID:        product.Category.ID,
			ShopID:    product.Category.ShopID,
			Name:      product.Category.Name,
			CreatedAt: product.Category.CreatedAt,
			UpdatedAt: product.Category.UpdatedAt,
		}
	}

	return dto
}

// ProductsToListDTO converts a slice of Product entities to ProductListDTO
func ProductsToListDTO(products []*entities.Product) []*ProductListDTO {
	dtos := make([]*ProductListDTO, len(products))
	for i, product := range products {
		dtos[i] = ProductToListDTO(product)
	}
	return dtos
}

// CategoryToListDTO converts a Category entity to CategoryListDTO
func CategoryToListDTO(category *entities.Category) *CategoryListDTO {
	dto := &CategoryListDTO{
		ID:        category.ID,
		ShopID:    category.ShopID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}

	// Convert shop if loaded
	if category.Shop.ID != uuid.Nil {
		dto.Shop = &ShopListDTO{
			ID:              category.Shop.ID,
			LicenseID:       category.Shop.LicenseID,
			UserID:          category.Shop.UserID,
			Name:            category.Shop.Name,
			Domain:          category.Shop.Domain,
			Photo:           category.Shop.Photo,
			Address:         category.Shop.Address,
			Slogan:          category.Shop.Slogan,
			ProfitCalculate: category.Shop.ProfitCalculate,
			CreatedAt:       category.Shop.CreatedAt,
			UpdatedAt:       category.Shop.UpdatedAt,
		}
	}

	return dto
}

// CategoriesToListDTO converts a slice of Category entities to CategoryListDTO
func CategoriesToListDTO(categories []*entities.Category) []*CategoryListDTO {
	dtos := make([]*CategoryListDTO, len(categories))
	for i, category := range categories {
		dtos[i] = CategoryToListDTO(category)
	}
	return dtos
}

// CategoryToResponseDTO converts a Category entity to CategoryResponseDTO
func CategoryToResponseDTO(category *entities.Category) *CategoryResponseDTO {
	return &CategoryResponseDTO{
		ID:        category.ID,
		ShopID:    category.ShopID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
