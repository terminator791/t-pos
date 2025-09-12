package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/pkg/response"
)

// CreateProductRequest represents the request payload for creating a product
type CreateProductRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   *string  `json:"description"`
	Sale          float64  `json:"sale" binding:"required,gt=0"`
	Buy           float64  `json:"buy" binding:"required,gt=0"`
	Unit          *string  `json:"unit"`
	PPN           *float64 `json:"ppn"`
	Photo         *string  `json:"photo"`
	CategoryID    *string  `json:"category_id"`
	Barcode       *string  `json:"barcode"`
	StockQuantity int      `json:"stock_quantity"`
	ShopID        string   `json:"shop_id" binding:"required"`
}

// UpdateProductRequest represents the request payload for updating a product
type UpdateProductRequest struct {
	Name          *string  `json:"name"`
	Description   *string  `json:"description"`
	Sale          *float64 `json:"sale"`
	Buy           *float64 `json:"buy"`
	Unit          *string  `json:"unit"`
	PPN           *float64 `json:"ppn"`
	Photo         *string  `json:"photo"`
	CategoryID    *string  `json:"category_id"`
	Barcode       *string  `json:"barcode"`
	StockQuantity *int     `json:"stock_quantity"`
}

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

// saveUploadedFile saves an uploaded file to the products directory
func (h *ProductHandler) saveUploadedFile(file *multipart.FileHeader) (string, error) {
	// Create directory if it doesn't exist
	uploadDir := "public/images/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", timestamp, "product", ext)
	filePath := filepath.Join(uploadDir, filename)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return relative path
	relativePath := fmt.Sprintf("images/products/%s", filename)
	return relativePath, nil
}

// CreateProductWithFile creates a new product with file upload support
func (h *ProductHandler) CreateProductWithFile(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		response.ErrorBadRequest(c, "Failed to parse form data", err.Error())
		return
	}

	// Extract form values
	name := c.PostForm("name")
	if name == "" {
		response.ErrorBadRequest(c, "Product name is required", nil)
		return
	}

	shopIDStr := c.PostForm("shop_id")
	if shopIDStr == "" {
		response.ErrorBadRequest(c, "Shop ID is required", nil)
		return
	}

	shopID, err := uuid.Parse(shopIDStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	saleStr := c.PostForm("sale")
	if saleStr == "" {
		response.ErrorBadRequest(c, "Sale price is required", nil)
		return
	}
	sale, err := strconv.ParseFloat(saleStr, 64)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid sale price", err.Error())
		return
	}

	buyStr := c.PostForm("buy")
	if buyStr == "" {
		response.ErrorBadRequest(c, "Buy price is required", nil)
		return
	}
	buy, err := strconv.ParseFloat(buyStr, 64)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid buy price", err.Error())
		return
	}

	// Optional fields
	var categoryID *uuid.UUID
	if catIDStr := c.PostForm("category_id"); catIDStr != "" {
		catID, err := uuid.Parse(catIDStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid category ID", err.Error())
			return
		}
		categoryID = &catID
	}

	var unit *string
	if u := c.PostForm("unit"); u != "" {
		unit = &u
	}

	var ppn *float64
	if ppnStr := c.PostForm("ppn"); ppnStr != "" {
		p, err := strconv.ParseFloat(ppnStr, 64)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid PPN value", err.Error())
			return
		}
		ppn = &p
	}

	var barcode *string
	if bc := c.PostForm("barcode"); bc != "" {
		barcode = &bc
	}

	stockQuantity := 0
	if stockStr := c.PostForm("stock_quantity"); stockStr != "" {
		stockQuantity, err = strconv.Atoi(stockStr)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid stock quantity", err.Error())
			return
		}
	}

	// Handle file upload
	var photoPath *string
	if file, err := c.FormFile("photo"); err == nil {
		savedPath, err := h.saveUploadedFile(file)
		if err != nil {
			response.ErrorInternalServer(c, "Failed to save uploaded file", err.Error())
			return
		}
		photoPath = &savedPath
	}

	// Create product entity
	product := &entities.Product{
		ShopID:      shopID,
		CatID:       categoryID,
		Photo:       photoPath,
		Name:        name,
		Barcode:     barcode,
		Unit:        unit,
		PPN:         ppn,
		Sale:        sale,
		Buy:         buy,
		Stock:       stockQuantity,
		IsHaveStock: true,
	}

	err = h.productUseCase.CreateProduct(c.Request.Context(), product)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create product", err.Error())
		return
	}

	response.SuccessCreated(c, "Product created successfully", product)
}
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Parse shop ID
	shopID, err := uuid.Parse(req.ShopID)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	// Parse category ID if provided
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		catID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			response.ErrorBadRequest(c, "Invalid category ID", err.Error())
			return
		}
		categoryID = &catID
	}

	// Handle photo upload/save
	var photoPath *string
	if req.Photo != nil && *req.Photo != "" {
		// Create directory if it doesn't exist
		uploadDir := "public/images/products"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			response.ErrorInternalServer(c, "Failed to create upload directory", err.Error())
			return
		}

		// Generate unique filename
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s", timestamp, filepath.Base(*req.Photo))

		// For now, just save the path. In a real scenario, you'd handle the actual file upload
		relativePath := fmt.Sprintf("images/products/%s", filename)
		photoPath = &relativePath
	}

	// Create product entity
	product := &entities.Product{
		ShopID:      shopID,
		CatID:       categoryID,
		Photo:       photoPath,
		Name:        req.Name,
		Barcode:     req.Barcode,
		Unit:        req.Unit,
		PPN:         req.PPN,
		Sale:        req.Sale,
		Buy:         req.Buy,
		Stock:       req.StockQuantity,
		IsHaveStock: true,
	}

	err = h.productUseCase.CreateProduct(c.Request.Context(), product)
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

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Get existing product
	existingProduct, err := h.productUseCase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Product not found", err.Error())
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingProduct.Name = *req.Name
	}
	if req.Sale != nil {
		existingProduct.Sale = *req.Sale
	}
	if req.Buy != nil {
		existingProduct.Buy = *req.Buy
	}
	if req.Unit != nil {
		existingProduct.Unit = req.Unit
	}
	if req.PPN != nil {
		existingProduct.PPN = req.PPN
	}
	if req.Barcode != nil {
		existingProduct.Barcode = req.Barcode
	}
	if req.StockQuantity != nil {
		existingProduct.Stock = *req.StockQuantity
	}

	// Handle category update
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			existingProduct.CatID = nil
		} else {
			catID, err := uuid.Parse(*req.CategoryID)
			if err != nil {
				response.ErrorBadRequest(c, "Invalid category ID", err.Error())
				return
			}
			existingProduct.CatID = &catID
		}
	}

	// Handle photo update
	if req.Photo != nil {
		if *req.Photo == "" {
			existingProduct.Photo = nil
		} else {
			// Create directory if it doesn't exist
			uploadDir := "public/images/products"
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				response.ErrorInternalServer(c, "Failed to create upload directory", err.Error())
				return
			}

			// Generate unique filename
			timestamp := time.Now().Unix()
			filename := fmt.Sprintf("%d_%s", timestamp, filepath.Base(*req.Photo))

			// For now, just save the path. In a real scenario, you'd handle the actual file upload
			relativePath := fmt.Sprintf("images/products/%s", filename)
			existingProduct.Photo = &relativePath
		}
	}

	err = h.productUseCase.UpdateProduct(c.Request.Context(), existingProduct)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to update product", err.Error())
		return
	}

	response.SuccessOK(c, "Product updated successfully", existingProduct)
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
