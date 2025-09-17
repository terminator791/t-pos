package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

type ShopHandler struct {
	shopUseCase *usecases.ShopUseCase
	roleRepo    repositories.RoleRepository
	shopRepo    repositories.ShopRepository
}

func NewShopHandler(shopUseCase *usecases.ShopUseCase, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *ShopHandler {
	return &ShopHandler{
		shopUseCase: shopUseCase,
		roleRepo:    roleRepo,
		shopRepo:    shopRepo,
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

	// Validate user has access to the license before creating shop
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user access info", err.Error())
		return
	}

	if !domainAccess.CanAccessLicense(req.LicenseID) {
		response.Error(c, http.StatusForbidden, "Cannot create shop for this license", map[string]interface{}{
			"license_id": req.LicenseID,
			"user_id":    domainAccess.UserID,
			"role":       domainAccess.Role,
		})
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

	// Retrieve the created shop with relationships
	createdShop, err := h.shopUseCase.GetShop(c.Request.Context(), shop.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve created shop", err.Error())
		return
	}

	shopResponse := dto.ToShopResponse(*createdShop)
	c.JSON(http.StatusCreated, shopResponse)
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

	shopResponse := dto.ToShopResponse(*shop)
	c.JSON(http.StatusOK, shopResponse)
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

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user access info", err.Error())
		return
	}

	var shops []*entities.Shop

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all shops
		shops, err = h.shopUseCase.ListShops(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop/license IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		licenseFilter := domainAccess.GetLicenseFilter()

		if len(shopFilter) == 0 && len(licenseFilter) == 0 {
			// User has no accessible shops or licenses
			shops = []*entities.Shop{}
			err = nil
		} else {
			shops, err = h.shopUseCase.ListShopsFiltered(c.Request.Context(), shopFilter, licenseFilter, limit, offset)
		}
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve shops", err.Error())
		return
	}

	shopsResponse := dto.ToShopsListResponse(shops)
	c.JSON(http.StatusOK, shopsResponse)
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

	// Retrieve the updated shop with relationships
	updatedShop, err := h.shopUseCase.GetShop(c.Request.Context(), existingShop.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve updated shop", err.Error())
		return
	}

	shopResponse := dto.ToShopResponse(*updatedShop)
	c.JSON(http.StatusOK, shopResponse)
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

	shopsResponse := dto.ToShopsListResponse(shops)
	c.JSON(http.StatusOK, shopsResponse)
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

	shopsResponse := dto.ToShopsListResponse(shops)
	c.JSON(http.StatusOK, shopsResponse)
}
