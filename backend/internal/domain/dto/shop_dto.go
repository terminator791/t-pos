package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// LicenseDTO represents the license data in shop responses
type LicenseDTO struct {
	ID           uuid.UUID `json:"id"`
	SerialNumber string    `json:"serial_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserDTO represents the owner/user data in shop responses
type UserDTO struct {
	ID              uuid.UUID  `json:"id"`
	LicenseID       *uuid.UUID `json:"license_id"`
	RoleID          *uuid.UUID `json:"role_id"`
	ShopID          *uuid.UUID `json:"shop_id"`
	Email           *string    `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Username        *string    `json:"username"`
	Name            string     `json:"name"`
	InfoDevice      *string    `json:"info_device"`
	FCMToken        *string    `json:"fcm_token"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ShopDTO represents the shop data in responses
type ShopDTO struct {
	ID              uuid.UUID  `json:"id"`
	LicenseID       uuid.UUID  `json:"license_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Name            string     `json:"name"`
	Domain          string     `json:"domain"`
	Photo           *string    `json:"photo"`
	Address         *string    `json:"address"`
	Slogan          *string    `json:"slogan"`
	ProfitCalculate int64      `json:"profit_calculate"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	License         LicenseDTO `json:"license"`
	Owner           UserDTO    `json:"owner"`
}

// ShopsListResponse represents the response for listing shops
type ShopsListResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Count int       `json:"count"`
		Shops []ShopDTO `json:"shops"`
	} `json:"data"`
}

// ShopResponse represents the response for a single shop
type ShopResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Data    ShopDTO `json:"data"`
}

// ToLicenseDTO converts License entity to LicenseDTO
func ToLicenseDTO(license entities.License) LicenseDTO {
	return LicenseDTO{
		ID:           license.ID,
		SerialNumber: license.SerialNumber,
		CreatedAt:    license.CreatedAt,
		UpdatedAt:    license.UpdatedAt,
	}
}

// ToUserDTO converts User entity to UserDTO
func ToUserDTO(user entities.User) UserDTO {
	return UserDTO{
		ID:              user.ID,
		LicenseID:       user.LicenseID,
		RoleID:          user.RoleID,
		ShopID:          user.ShopID,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		Username:        user.Username,
		Name:            user.Name,
		InfoDevice:      user.InfoDevice,
		FCMToken:        user.FCMToken,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// ToShopDTO converts Shop entity to ShopDTO
func ToShopDTO(shop entities.Shop) ShopDTO {
	return ShopDTO{
		ID:              shop.ID,
		LicenseID:       shop.LicenseID,
		UserID:          shop.UserID,
		Name:            shop.Name,
		Domain:          shop.Domain,
		Photo:           shop.Photo,
		Address:         shop.Address,
		Slogan:          shop.Slogan,
		ProfitCalculate: shop.ProfitCalculate,
		CreatedAt:       shop.CreatedAt,
		UpdatedAt:       shop.UpdatedAt,
		License:         ToLicenseDTO(shop.License),
		Owner:           ToUserDTO(shop.Owner),
	}
}

// ToShopsListResponse creates ShopsListResponse from shops slice
func ToShopsListResponse(shops []*entities.Shop) ShopsListResponse {
	shopDTOs := make([]ShopDTO, len(shops))
	for i, shop := range shops {
		shopDTOs[i] = ToShopDTO(*shop)
	}

	return ShopsListResponse{
		Status:  "success",
		Message: "Shops retrieved successfully",
		Data: struct {
			Count int       `json:"count"`
			Shops []ShopDTO `json:"shops"`
		}{
			Count: len(shopDTOs),
			Shops: shopDTOs,
		},
	}
}

// ToShopResponse creates ShopResponse from shop entity
func ToShopResponse(shop entities.Shop) ShopResponse {
	return ShopResponse{
		Status:  "success",
		Message: "Shop retrieved successfully",
		Data:    ToShopDTO(shop),
	}
}
