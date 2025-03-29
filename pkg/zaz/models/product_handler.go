package models

import (
	"errors"
	"time"
)

// ProductHandler handles all operations related to Product model
type ProductHandler struct {
	circleID string
	db       *Db
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(circleID string, db *Db) *ProductHandler {
	return &ProductHandler{
		circleID: circleID,
		db:       db,
	}
}

// Product represents a product or service offered by the Freezone
type Product struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Currency    string    `gorm:"not null" json:"currency"`
	Type        string    `gorm:"not null" json:"type"` // Product, Service
	Category    string    `json:"category"`
	Status      string    `gorm:"not null" json:"status"` // Available, Unavailable
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// GetAll returns all products
func (h *ProductHandler) GetAll() []Product {
	var products []Product
	GetDB().Find(&products)
	return products
}

// GetByID returns a product by ID
func (h *ProductHandler) GetByID(id int64) (Product, error) {
	var product Product
	result := GetDB().First(&product, id)
	if result.Error != nil {
		return Product{}, ErrRecordNotFound
	}
	return product, nil
}

// Create adds a new product
func (h *ProductHandler) Create(product Product) (int64, error) {
	if product.CreatedAt.IsZero() {
		product.CreatedAt = time.Now()
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}

	result := GetDB().Create(&product)
	if result.Error != nil {
		return 0, result.Error
	}
	return product.ID, nil
}

// Update updates an existing product
func (h *ProductHandler) Update(product Product) error {
	product.UpdatedAt = time.Now()

	var existingProduct Product
	if err := GetDB().First(&existingProduct, product.ID).Error; err != nil {
		return ErrRecordNotFound
	}

	result := GetDB().Model(&product).Updates(map[string]interface{}{
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"currency":    product.Currency,
		"type":        product.Type,
		"category":    product.Category,
		"status":      product.Status,
		"updated_at":  product.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete deletes a product
func (h *ProductHandler) Delete(id int64) error {
	var product Product
	if err := GetDB().First(&product, id).Error; err != nil {
		return ErrRecordNotFound
	}

	// Check if product is used in any sale
	var count int64
	GetDB().Model(&SaleItem{}).Where("product_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete product that is used in sales")
	}

	result := GetDB().Delete(&Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
