package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/pkg/response"
)

type ShopHandler struct {
	shopUseCase *usecases.ShopUseCase
}

func NewShopHandler(shopUseCase *usecases.ShopUseCase) *ShopHandler {
	return &ShopHandler{
		shopUseCase: shopUseCase,
	}
}

// CreateShop creates a new shop
func (h *ShopHandler) CreateShop(c *gin.Context) {
	var req struct {
		Name            string    `json:"name" binding:"required"`
		Photo           *string   `json:"photo"`
		Address         *string   `json:"address"`
		Slogan          *string   `json:"slogan"`
		LicenseID       uuid.UUID `json:"license_id" binding:"required"`
		UserID          uuid.UUID `json:"user_id" binding:"required"`
		Domain          string    `json:"domain"`
		ProfitCalculate int64     `json:"profit_calculate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	shop := &entities.Shop{
		ID:              uuid.New(),
		Name:            req.Name,
		Photo:           req.Photo,
		Address:         req.Address,
		Slogan:          req.Slogan,
		LicenseID:       req.LicenseID,
		UserID:          req.UserID,
		Domain:          req.Domain,
		ProfitCalculate: req.ProfitCalculate,
	}

	if err := h.shopUseCase.CreateShop(c.Request.Context(), shop); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create shop", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Shop created successfully", shop)
}

// GetShop retrieves a shop by ID
func (h *ShopHandler) GetShop(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid shop ID", err.Error())
		return
	}

	shop, err := h.shopUseCase.GetShop(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Shop not found", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shop retrieved successfully", shop)
}

// ListShops retrieves a list of shops
func (h *ShopHandler) ListShops(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid limit parameter", err.Error())
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid offset parameter", err.Error())
		return
	}

	shops, err := h.shopUseCase.ListShops(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shops", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shops retrieved successfully", gin.H{
		"shops": shops,
		"count": len(shops),
	})
}

// UpdateShop updates an existing shop
func (h *ShopHandler) UpdateShop(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid shop ID", err.Error())
		return
	}

	var req struct {
		Name            string  `json:"name"`
		Photo           *string `json:"photo"`
		Address         *string `json:"address"`
		Slogan          *string `json:"slogan"`
		Domain          string  `json:"domain"`
		ProfitCalculate int64   `json:"profit_calculate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Get existing shop
	existingShop, err := h.shopUseCase.GetShop(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Shop not found", err.Error())
		return
	}

	// Update fields
	if req.Name != "" {
		existingShop.Name = req.Name
	}
	if req.Photo != nil {
		existingShop.Photo = req.Photo
	}
	if req.Address != nil {
		existingShop.Address = req.Address
	}
	if req.Slogan != nil {
		existingShop.Slogan = req.Slogan
	}
	if req.Domain != "" {
		existingShop.Domain = req.Domain
	}
	if req.ProfitCalculate != 0 {
		existingShop.ProfitCalculate = req.ProfitCalculate
	}

	if err := h.shopUseCase.UpdateShop(c.Request.Context(), existingShop); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to update shop", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shop updated successfully", existingShop)
}

// DeleteShop deletes a shop
func (h *ShopHandler) DeleteShop(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid shop ID", err.Error())
		return
	}

	if err := h.shopUseCase.DeleteShop(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete shop", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shop deleted successfully", nil)
}

// GetShopsByLicense retrieves shops by license ID
func (h *ShopHandler) GetShopsByLicense(c *gin.Context) {
	licenseIDStr := c.Param("licenseId")
	licenseID, err := uuid.Parse(licenseIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid license ID", err.Error())
		return
	}

	shops, err := h.shopUseCase.GetShopsByLicense(c.Request.Context(), licenseID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shops", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shops retrieved successfully", gin.H{
		"shops": shops,
		"count": len(shops),
	})
}

// GetShopsByOwner retrieves shops by owner ID
func (h *ShopHandler) GetShopsByOwner(c *gin.Context) {
	ownerIDStr := c.Param("ownerId")
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid owner ID", err.Error())
		return
	}

	shops, err := h.shopUseCase.GetShopsByOwner(c.Request.Context(), ownerID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shops", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Shops retrieved successfully", gin.H{
		"shops": shops,
		"count": len(shops),
	})
}
