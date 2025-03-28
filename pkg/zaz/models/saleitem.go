package models

import (
	"errors"
)

// GetSaleItemsBySaleID returns all items for a sale
func GetSaleItemsBySaleID(saleID int64) []SaleItem {
	var items []SaleItem
	GetDB().Where("sale_id = ?", saleID).Find(&items)
	return items
}

// GetSaleItemByID returns a sale item by ID
func GetSaleItemByID(id int64) (SaleItem, error) {
	var item SaleItem
	result := GetDB().First(&item, id)
	if result.Error != nil {
		return SaleItem{}, errors.New("sale item not found")
	}
	return item, nil
}

// AddSaleItem adds a new sale item
func AddSaleItem(item SaleItem) int64 {
	result := GetDB().Create(&item)
	if result.Error != nil {
		return 0
	}
	return item.ID
}

// UpdateSaleItem updates an existing sale item
func UpdateSaleItem(item SaleItem) error {
	// Check if item exists
	var existingItem SaleItem
	if err := GetDB().First(&existingItem, item.ID).Error; err != nil {
		return errors.New("sale item not found")
	}
	
	// Update item
	result := GetDB().Model(&item).Updates(map[string]interface{}{
		"sale_id":    item.SaleID,
		"product_id": item.ProductID,
		"name":       item.Name,
		"quantity":   item.Quantity,
		"unit_price": item.UnitPrice,
		"currency":   item.Currency,
		"subtotal":   item.Subtotal,
	})
	
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteSaleItem deletes a sale item
func DeleteSaleItem(id int64) error {
	// Check if item exists
	var item SaleItem
	if err := GetDB().First(&item, id).Error; err != nil {
		return errors.New("sale item not found")
	}
	
	// Delete item
	result := GetDB().Delete(&SaleItem{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}
