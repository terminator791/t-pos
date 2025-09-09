package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/internal/interfaces/http/handlers"
)

// SetupRoutes sets up all the API routes
func SetupRoutes(
	router *gin.Engine,
	productHandler *handlers.ProductHandler,
	checkoutHandler *handlers.CheckoutHandler,
	authHandler *handlers.AuthHandler,
	authMiddleware *auth.AuthMiddleware,
	authzMiddleware *casbin.AuthzMiddleware,
) {
	// API version 1
	v1 := router.Group("/api/v1")
	{
		// Public authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Protected authentication routes
		authProtected := v1.Group("/auth")
		authProtected.Use(authMiddleware.RequireAuth())
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.POST("/refresh", authHandler.RefreshToken)
			authProtected.GET("/profile", authHandler.Profile)
			authProtected.GET("/permissions", authHandler.GetPermissions)
			authProtected.POST("/pin", authHandler.CreatePin)
			authProtected.PUT("/pin", authHandler.UpdatePin)
			authProtected.DELETE("/pin", authHandler.DeletePin)
		}

		// Protected API routes
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		protected.Use(authzMiddleware.RequirePermission())
		{
			// Product routes
			products := protected.Group("/products")
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
			checkout := protected.Group("/checkout")
			{
				checkout.POST("", checkoutHandler.ProcessCheckout)
				checkout.POST("/:transactionId/complete", checkoutHandler.CompletePayment)
				checkout.POST("/:transactionId/cancel", checkoutHandler.CancelTransaction)
			}

			// Transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.GET("/:transactionId", checkoutHandler.GetTransactionDetails)
				transactions.GET("/shop/:shopId/today", checkoutHandler.GetTodaysTransactions)
			}
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