package models

import (
	"errors"
	"time"
)

// GetAllSales returns all sales
func GetAllSales() []Sale {
	var sales []Sale
	GetDB().Preload("Items").Preload("Company").Find(&sales)
	return sales
}

// GetSaleByID returns a sale by ID
func GetSaleByID(id int64) (Sale, error) {
	var sale Sale
	result := GetDB().Preload("Items").Preload("Company").First(&sale, id)
	if result.Error != nil {
		return Sale{}, errors.New("sale not found")
	}
	return sale, nil
}

// GetSalesByCompanyID returns all sales for a company
func GetSalesByCompanyID(companyID int64) []Sale {
	var sales []Sale
	GetDB().Preload("Items").Where("company_id = ?", companyID).Find(&sales)
	return sales
}

// AddSale adds a new sale
func AddSale(sale Sale) int64 {
	// Set timestamps if not already set
	if sale.CreatedAt.IsZero() {
		sale.CreatedAt = time.Now()
	}
	if sale.UpdatedAt.IsZero() {
		sale.UpdatedAt = time.Now()
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0
	}
	
	// Create sale
	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		return 0
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0
	}
	
	return sale.ID
}

// UpdateSale updates an existing sale
func UpdateSale(sale Sale) error {
	// Set update timestamp
	sale.UpdatedAt = time.Now()
	
	// Check if sale exists
	var existingSale Sale
	if err := GetDB().First(&existingSale, sale.ID).Error; err != nil {
		return errors.New("sale not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Update sale
	if err := tx.Model(&sale).Updates(map[string]interface{}{
		"company_id":   sale.CompanyID,
		"buyer_name":   sale.BuyerName,
		"buyer_email":  sale.BuyerEmail,
		"total_amount": sale.TotalAmount,
		"currency":     sale.Currency,
		"status":       sale.Status,
		"sale_date":    sale.SaleDate,
		"updated_at":   sale.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Update items (delete and re-insert)
	if err := tx.Where("sale_id = ?", sale.ID).Delete(&SaleItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Insert new items
	for i := range sale.Items {
		sale.Items[i].SaleID = sale.ID
		if err := tx.Create(&sale.Items[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	
	// Commit transaction
	return tx.Commit().Error
}

// DeleteSale deletes a sale
func DeleteSale(id int64) error {
	// Check if sale exists
	var sale Sale
	if err := GetDB().First(&sale, id).Error; err != nil {
		return errors.New("sale not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Delete items
	if err := tx.Where("sale_id = ?", id).Delete(&SaleItem{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete sale
	if err := tx.Delete(&Sale{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}
