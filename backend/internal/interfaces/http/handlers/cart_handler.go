package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	cartUseCase *usecases.CartUseCase
	roleRepo    repositories.RoleRepository
	shopRepo    repositories.ShopRepository
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartUseCase *usecases.CartUseCase, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *CartHandler {
	return &CartHandler{
		cartUseCase: cartUseCase,
		roleRepo:    roleRepo,
		shopRepo:    shopRepo,
	}
}

// AddToCartRequest represents the request structure for adding to cart
type AddToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	ShopID    uuid.UUID `json:"shop_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

// UpdateCartQuantityRequest represents the request structure for updating cart quantity
type UpdateCartQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// AddToCart handles POST /carts
func (h *CartHandler) AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	cart, err := h.cartUseCase.AddToCart(c.Request.Context(), userID, req.ProductID, req.ShopID, req.Quantity)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to add to cart", err.Error())
		return
	}

	response.SuccessCreated(c, "Product added to cart successfully", cart)
}

// GetUserCart handles GET /carts
func (h *CartHandler) GetUserCart(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	carts, err := h.cartUseCase.GetUserCart(c.Request.Context(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve cart", err.Error())
		return
	}

	response.SuccessOK(c, "Cart retrieved successfully", gin.H{
		"cart_items": carts,
		"total":      len(carts),
	})
}

// GetCartItem handles GET /carts/:id
func (h *CartHandler) GetCartItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid cart ID", err.Error())
		return
	}

	cart, err := h.cartUseCase.GetCartItem(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Cart item not found", err.Error())
		return
	}

	response.SuccessOK(c, "Cart item retrieved successfully", cart)
}

// UpdateCartQuantity handles PUT /carts/:id
func (h *CartHandler) UpdateCartQuantity(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid cart ID", err.Error())
		return
	}

	var req UpdateCartQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	cart, err := h.cartUseCase.UpdateCartQuantity(c.Request.Context(), id, userID, req.Quantity)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to update cart quantity", err.Error())
		return
	}

	response.SuccessOK(c, "Cart quantity updated successfully", cart)
}

// RemoveFromCart handles DELETE /carts/:id
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid cart ID", err.Error())
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	err = h.cartUseCase.RemoveFromCart(c.Request.Context(), id, userID)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to remove from cart", err.Error())
		return
	}

	response.SuccessOK(c, "Product removed from cart successfully", nil)
}

// ClearCart handles DELETE /carts
func (h *CartHandler) ClearCart(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	err := h.cartUseCase.ClearUserCart(c.Request.Context(), userID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to clear cart", err.Error())
		return
	}

	response.SuccessOK(c, "Cart cleared successfully", nil)
}

// ListAllCarts handles GET /carts/all - with domain-specific filtering
func (h *CartHandler) ListAllCarts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
		return
	}

	var carts []*entities.Cart

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all carts
		carts, err = h.cartUseCase.ListCarts(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		if len(shopFilter) == 0 {
			// User has no accessible shops
			carts = []*entities.Cart{}
			err = nil
		} else {
			carts, err = h.cartUseCase.ListCartsByShopIDs(c.Request.Context(), shopFilter, limit, offset)
		}
	}

	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve carts", err.Error())
		return
	}

	response.SuccessOK(c, "Carts retrieved successfully", gin.H{
		"carts":  carts,
		"limit":  limit,
		"offset": offset,
	})
}
