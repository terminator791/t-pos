package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/interfaces/http/handlers"
)

// SetupRoutes sets up all the API routes
func SetupRoutes(
	router *gin.Engine,
	productHandler *handlers.ProductHandler,
	checkoutHandler *handlers.CheckoutHandler,
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
			products.GET("/barcode/:barcode", productHandler.GetProductByBarcode)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Checkout and transaction routes
		checkout := v1.Group("/checkout")
		{
			checkout.POST("", checkoutHandler.ProcessCheckout)
			checkout.POST("/:transactionId/complete", checkoutHandler.CompletePayment)
			checkout.POST("/:transactionId/cancel", checkoutHandler.CancelTransaction)
		}

		// Transaction routes
		transactions := v1.Group("/transactions")
		{
			transactions.GET("/:transactionId", checkoutHandler.GetTransactionDetails)
			transactions.GET("/shop/:shopId/today", checkoutHandler.GetTodaysTransactions)
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