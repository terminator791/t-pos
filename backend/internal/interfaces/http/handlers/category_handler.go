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

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryUseCase *usecases.CategoryUseCase
	roleRepo        repositories.RoleRepository
	shopRepo        repositories.ShopRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryUseCase *usecases.CategoryUseCase, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryUseCase: categoryUseCase,
		roleRepo:        roleRepo,
		shopRepo:        shopRepo,
	}
}

// CreateCategoryRequest represents the request structure for creating a category
type CreateCategoryRequest struct {
	ShopID uuid.UUID `json:"shop_id" binding:"required"`
	Name   string    `json:"name" binding:"required"`
}

// UpdateCategoryRequest represents the request structure for updating a category
type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateCategory handles POST /categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	category := &entities.Category{
		ShopID: req.ShopID,
		Name:   req.Name,
	}

	err := h.categoryUseCase.CreateCategory(c.Request.Context(), category)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create category", err.Error())
		return
	}

	response.SuccessCreated(c, "Category created successfully", category)
}

// GetCategory handles GET /categories/:id
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid category ID", err.Error())
		return
	}

	category, err := h.categoryUseCase.GetCategory(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Category not found", err.Error())
		return
	}

	response.SuccessOK(c, "Category retrieved successfully", category)
}

// ListCategories handles GET /categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	shopIDStr := c.Query("shop_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var categories []*entities.Category

	// If specific shop_id is requested, validate access and filter by that shop
	if shopIDStr != "" {
		shopID, err := uuid.Parse(shopIDStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
			return
		}

		// Get domain access info to validate shop access
		domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
			return
		}

		// Check if user can access the requested shop
		if !domainAccess.CanAccessShop(shopID) {
			response.ErrorForbidden(c, "Cannot access categories for this shop", map[string]interface{}{
				"requested_shop_id": shopID,
				"user_role":        domainAccess.Role,
				"accessible_shops": domainAccess.AccessibleShopIDs,
			})
			return
		}

		categories, err = h.categoryUseCase.GetCategoriesByShop(c.Request.Context(), shopID)
	} else {
		// List all categories with domain-specific filtering
		domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
			return
		}

		// Apply domain-specific filtering
		if domainAccess.HasGlobalAccess {
			// Super admin and admin can see all categories
			categories, err = h.categoryUseCase.ListCategories(c.Request.Context(), limit, offset)
		} else {
			// Filter by accessible shop IDs for tenant users
			shopFilter := domainAccess.GetShopFilter()
			if len(shopFilter) == 0 {
				// User has no accessible shops
				categories = []*entities.Category{}
				err = nil
			} else {
				categories, err = h.categoryUseCase.ListCategoriesFiltered(c.Request.Context(), shopFilter, limit, offset)
			}
		}
	}

	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve categories", err.Error())
		return
	}

	response.SuccessOK(c, "Categories retrieved successfully", gin.H{
		"categories": categories,
		"limit":      limit,
		"offset":     offset,
	})
}

// UpdateCategory handles PUT /categories/:id
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid category ID", err.Error())
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get existing category
	existing, err := h.categoryUseCase.GetCategory(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Category not found", err.Error())
		return
	}

	// Update fields
	existing.Name = req.Name

	err = h.categoryUseCase.UpdateCategory(c.Request.Context(), existing)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to update category", err.Error())
		return
	}

	response.SuccessOK(c, "Category updated successfully", existing)
}

// DeleteCategory handles DELETE /categories/:id
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid category ID", err.Error())
		return
	}

	err = h.categoryUseCase.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to delete category", err.Error())
		return
	}

	response.SuccessOK(c, "Category deleted successfully", nil)
}
