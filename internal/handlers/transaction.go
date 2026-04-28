package handlers

import (
	"fmt"
	"net/http"

	"github.com/example/pos/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type TransactionHandler struct {
	db *sqlx.DB
}

func NewTransactionHandler(db *sqlx.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

// POST /checkout - คิดเงิน
func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate stock and calculate total
	type checkoutItem struct {
		ProductID   int
		ProductName string
		UnitPrice   float64
		Quantity    int
		Subtotal    float64
	}

	items := make([]checkoutItem, 0, len(req.Items))
	var totalAmount float64

	for _, item := range req.Items {
		var product models.Product
		err := h.db.Get(&product, "SELECT id, name, price, stock FROM products WHERE id = $1", item.ProductID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Product ID %d not found", item.ProductID)})
			return
		}

		if product.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Insufficient stock for product: %s (available: %d)", product.Name, product.Stock)})
			return
		}

		subtotal := product.Price * float64(item.Quantity)
		items = append(items, checkoutItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			UnitPrice:   product.Price,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
		totalAmount += subtotal
	}

	// Begin transaction
	tx, err := h.db.Beginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Create transaction record
	var transactionID int
	err = tx.QueryRow(
		"INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id",
		totalAmount,
	).Scan(&transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Insert transaction items and update stock
	for _, item := range items {
		_, err = tx.Exec(
			"INSERT INTO transaction_items (transaction_id, product_id, product_name, unit_price, quantity, subtotal) VALUES ($1, $2, $3, $4, $5, $6)",
			transactionID, item.ProductID, item.ProductName, item.UnitPrice, item.Quantity, item.Subtotal,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add transaction item"})
			return
		}

		_, err = tx.Exec(
			"UPDATE products SET stock = stock - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2",
			item.Quantity, item.ProductID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Fetch the complete transaction with items
	var transaction models.Transaction
	h.db.Get(&transaction, "SELECT * FROM transactions WHERE id = $1", transactionID)

	var transactionItems []models.TransactionItem
	h.db.Select(&transactionItems, "SELECT * FROM transaction_items WHERE transaction_id = $1", transactionID)
	transaction.Items = transactionItems

	c.JSON(http.StatusCreated, transaction)
}

// GET /transactions - ดูประวัติรายการซื้อขายทั้งหมด
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	var transactions []models.Transaction
	err := h.db.Select(&transactions, "SELECT * FROM transactions ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	// Fetch items for each transaction
	for i := range transactions {
		var items []models.TransactionItem
		h.db.Select(&items, "SELECT * FROM transaction_items WHERE transaction_id = $1", transactions[i].ID)
		transactions[i].Items = items
	}

	c.JSON(http.StatusOK, transactions)
}

// GET /transactions/:id - ดูรายการซื้อขายราย transaction
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")

	var transaction models.Transaction
	err := h.db.Get(&transaction, "SELECT * FROM transactions WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	var items []models.TransactionItem
	h.db.Select(&items, "SELECT * FROM transaction_items WHERE transaction_id = $1", id)
	transaction.Items = items

	c.JSON(http.StatusOK, transaction)
}
