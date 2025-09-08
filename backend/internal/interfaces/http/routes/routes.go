package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/interfaces/http/handlers"
)

// SetupRoutes sets up all the API routes
func SetupRoutes(
	router *gin.Engine,
	productHandler *handlers.ProductHandler,
	orderHandler *handlers.OrderHandler,
) {
	// API version 1
	v1 := router.Group("/api/v1")
	{
		// Product routes
		products := v1.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("", productHandler.ListProducts)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/low-stock", productHandler.GetLowStockProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/sku/:sku", productHandler.GetProductBySKU)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/today", orderHandler.GetTodaysOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("/number/:orderNumber", orderHandler.GetOrderByNumber)
		}
	}

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "T-POS API is running",
		})
	})
}