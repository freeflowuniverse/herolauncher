package models

import (
	"errors"
	"time"
)

// GetAllShareholders returns all shareholders across all companies
func GetAllShareholders() []Shareholder {
	var shareholders []Shareholder
	GetDB().Preload("Company").Find(&shareholders)
	return shareholders
}

// GetShareholderByID returns a shareholder by ID
func GetShareholderByID(id int64) (Shareholder, error) {
	var shareholder Shareholder
	result := GetDB().Preload("Company").First(&shareholder, id)
	if result.Error != nil {
		return Shareholder{}, errors.New("shareholder not found")
	}
	return shareholder, nil
}

// GetShareholdersByCompanyID returns all shareholders for a company
func GetShareholdersByCompanyID(companyID int64) []Shareholder {
	var shareholders []Shareholder
	GetDB().Where("company_id = ?", companyID).Find(&shareholders)
	return shareholders
}

// AddShareholder adds a new shareholder
func AddShareholder(shareholder Shareholder) int64 {
	// Set timestamps if not already set
	if shareholder.CreatedAt.IsZero() {
		shareholder.CreatedAt = time.Now()
	}
	if shareholder.UpdatedAt.IsZero() {
		shareholder.UpdatedAt = time.Now()
	}
	
	result := GetDB().Create(&shareholder)
	if result.Error != nil {
		return 0
	}
	return shareholder.ID
}

// UpdateShareholder updates an existing shareholder
func UpdateShareholder(shareholder Shareholder) error {
	// Set update timestamp
	shareholder.UpdatedAt = time.Now()
	
	// Check if shareholder exists
	var existingShareholder Shareholder
	if err := GetDB().First(&existingShareholder, shareholder.ID).Error; err != nil {
		return errors.New("shareholder not found")
	}
	
	// Update shareholder
	result := GetDB().Model(&shareholder).Updates(map[string]interface{}{
		"company_id": shareholder.CompanyID,
		"user_id":    shareholder.UserID,
		"name":       shareholder.Name,
		"shares":     shareholder.Shares,
		"percentage": shareholder.Percentage,
		"type":       shareholder.Type,
		"since":      shareholder.Since,
		"updated_at": shareholder.UpdatedAt,
	})
	
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteShareholder deletes a shareholder
func DeleteShareholder(id int64) error {
	// Check if shareholder exists
	var shareholder Shareholder
	if err := GetDB().First(&shareholder, id).Error; err != nil {
		return errors.New("shareholder not found")
	}
	
	// Delete shareholder
	result := GetDB().Delete(&Shareholder{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}
