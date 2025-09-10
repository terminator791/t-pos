package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/pkg/response"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productUseCase *usecases.ProductUseCase
}

// NewProductHandler creates a new product handler
func NewProductHandler(productUseCase *usecases.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product entities.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	err := h.productUseCase.CreateProduct(c.Request.Context(), &product)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create product", err.Error())
		return
	}

	response.SuccessCreated(c, "Product created successfully", product)
}

// GetProduct retrieves a product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid product ID", err.Error())
		return
	}

	product, err := h.productUseCase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Product not found", err.Error())
		return
	}

	response.SuccessOK(c, "Product retrieved successfully", product)
}

// GetProductByBarcode retrieves a product by barcode
func (h *ProductHandler) GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	if barcode == "" {
		response.ErrorBadRequest(c, "Barcode is required", nil)
		return
	}

	product, err := h.productUseCase.GetProductByBarcode(c.Request.Context(), barcode)
	if err != nil {
		response.ErrorNotFound(c, "Product not found", err.Error())
		return
	}

	response.SuccessOK(c, "Product retrieved successfully", product)
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid product ID", err.Error())
		return
	}

	var product entities.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	product.ID = id
	err = h.productUseCase.UpdateProduct(c.Request.Context(), &product)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to update product", err.Error())
		return
	}

	response.SuccessOK(c, "Product updated successfully", product)
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid product ID", err.Error())
		return
	}

	err = h.productUseCase.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to delete product", err.Error())
		return
	}

	response.SuccessOK(c, "Product deleted successfully", nil)
}

// ListProducts retrieves a list of products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	limitParam := c.DefaultQuery("limit", "20")
	offsetParam := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetParam)
	if err != nil {
		offset = 0
	}

	products, err := h.productUseCase.ListProducts(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve products", err.Error())
		return
	}

	data := gin.H{
		"products": products,
		"count":    len(products),
		"limit":    limit,
		"offset":   offset,
	}
	response.SuccessOK(c, "Products retrieved successfully", data)
}

// SearchProducts searches for products
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	shopIDParam := c.Query("shop_id")
	if query == "" {
		response.ErrorBadRequest(c, "Search query is required", nil)
		return
	}
	if shopIDParam == "" {
		response.ErrorBadRequest(c, "Shop ID is required", nil)
		return
	}

	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	products, err := h.productUseCase.SearchProducts(c.Request.Context(), query, shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to search products", err.Error())
		return
	}

	data := gin.H{
		"products": products,
		"count":    len(products),
		"query":    query,
	}
	response.SuccessOK(c, "Products searched successfully", data)
}

// GetLowStockProducts retrieves products with low stock
func (h *ProductHandler) GetLowStockProducts(c *gin.Context) {
	shopIDParam := c.Query("shop_id")
	if shopIDParam == "" {
		response.ErrorBadRequest(c, "Shop ID is required", nil)
		return
	}

	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	products, err := h.productUseCase.GetLowStockProducts(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get low stock products", err.Error())
		return
	}

	data := gin.H{
		"products": products,
		"count":    len(products),
	}
	response.SuccessOK(c, "Low stock products retrieved successfully", data)
}