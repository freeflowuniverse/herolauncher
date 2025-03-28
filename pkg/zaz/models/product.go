package models

import (
	"errors"
	"time"
)

// GetAllProducts returns all products
func GetAllProducts() []Product {
	var products []Product
	GetDB().Find(&products)
	return products
}

// GetProductByID returns a product by ID
func GetProductByID(id int64) (Product, error) {
	var product Product
	result := GetDB().First(&product, id)
	if result.Error != nil {
		return Product{}, errors.New("product not found")
	}
	return product, nil
}

// GetProductsByType returns all products of a specific type
func GetProductsByType(productType string) []Product {
	var products []Product
	GetDB().Where("type = ?", productType).Find(&products)
	return products
}

// AddProduct adds a new product
func AddProduct(product Product) int64 {
	// Set timestamps if not already set
	if product.CreatedAt.IsZero() {
		product.CreatedAt = time.Now()
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}
	
	result := GetDB().Create(&product)
	if result.Error != nil {
		return 0
	}
	return product.ID
}

// UpdateProduct updates an existing product
func UpdateProduct(product Product) error {
	// Set update timestamp
	product.UpdatedAt = time.Now()
	
	// Check if product exists
	var existingProduct Product
	if err := GetDB().First(&existingProduct, product.ID).Error; err != nil {
		return errors.New("product not found")
	}
	
	// Update product
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

// DeleteProduct deletes a product
func DeleteProduct(id int64) error {
	// Check if product exists
	var product Product
	if err := GetDB().First(&product, id).Error; err != nil {
		return errors.New("product not found")
	}
	
	// Delete product
	result := GetDB().Delete(&Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}
