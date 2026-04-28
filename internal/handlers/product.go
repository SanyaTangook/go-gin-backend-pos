package handlers

import (
	"fmt"
	"net/http"

	"github.com/example/pos/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ProductHandler struct {
	db *sqlx.DB
}

func NewProductHandler(db *sqlx.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// GET /products - ดูรายการสินค้าทั้งหมด
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var products []models.Product
	err := h.db.Select(&products, "SELECT * FROM products ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	c.JSON(http.StatusOK, products)
}

// GET /products/:id - ดูสินค้ารายตัว
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	err := h.db.Get(&product, "SELECT * FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// POST /products - เพิ่มสินค้าใหม่
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := h.db.QueryRow(
		"INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id",
		req.Name, req.Price, req.Stock,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	var product models.Product
	h.db.Get(&product, "SELECT * FROM products WHERE id = $1", id)
	c.JSON(http.StatusCreated, product)
}

// PUT /products/:id - แก้ไขสินค้า
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product exists
	var exists bool
	err := h.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Build dynamic update query
	query := "UPDATE products SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCount := 0

	if req.Name != "" {
		argCount++
		query += fmt.Sprintf(", name = $%d", argCount)
		args = append(args, req.Name)
	}
	if req.Price > 0 {
		argCount++
		query += fmt.Sprintf(", price = $%d", argCount)
		args = append(args, req.Price)
	}
	if req.Stock > 0 || req.Stock == 0 {
		argCount++
		query += fmt.Sprintf(", stock = $%d", argCount)
		args = append(args, req.Stock)
	}

	argCount++
	query += fmt.Sprintf(" WHERE id = $%d RETURNING *", argCount)
	args = append(args, id)

	var product models.Product
	err = h.db.Get(&product, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DELETE /products/:id - ลบสินค้า
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	result, err := h.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
