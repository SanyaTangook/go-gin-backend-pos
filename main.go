package main

import (
	"fmt"
	"log"

	"github.com/example/pos/internal/config"
	"github.com/example/pos/internal/database"
	"github.com/example/pos/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	productHandler := handlers.NewProductHandler(db)
	transactionHandler := handlers.NewTransactionHandler(db)

	// Setup router
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		// Product routes (คีย์ขาย)
		products := api.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Transaction routes (คิดเงิน และ บันทึกรายการซื้อขาย)
		transactions := api.Group("/transactions")
		{
			transactions.GET("", transactionHandler.ListTransactions)
			transactions.GET("/:id", transactionHandler.GetTransaction)
		}

		// Checkout route
		api.POST("/checkout", transactionHandler.Checkout)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
