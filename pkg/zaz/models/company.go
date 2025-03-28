package models

import (
	"errors"
	"time"
)

// GetAllCompanies returns all companies
func GetAllCompanies() []Company {
	var companies []Company
	GetDB().Preload("Shareholders").Preload("BoardMeetings").Find(&companies)
	return companies
}

// GetActiveCompanies returns all active companies
func GetActiveCompanies() []Company {
	var companies []Company
	GetDB().Preload("Shareholders").Preload("BoardMeetings").Where("status = ?", "Active").Find(&companies)
	return companies
}

// GetCompanyByID returns a company by ID
func GetCompanyByID(id int64) (Company, error) {
	var company Company
	result := GetDB().Preload("Shareholders").Preload("BoardMeetings").First(&company, id)
	if result.Error != nil {
		return Company{}, errors.New("company not found")
	}
	return company, nil
}

// AddCompany adds a new company
func AddCompany(company Company) int64 {
	// Set timestamps if not already set
	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now()
	}
	if company.UpdatedAt.IsZero() {
		company.UpdatedAt = time.Now()
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0
	}
	
	// Create company
	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		return 0
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0
	}
	
	return company.ID
}

// UpdateCompany updates an existing company
func UpdateCompany(company Company) error {
	// Set update timestamp
	company.UpdatedAt = time.Now()
	
	// Check if company exists
	var existingCompany Company
	if err := GetDB().First(&existingCompany, company.ID).Error; err != nil {
		return errors.New("company not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Update company
	if err := tx.Model(&company).Updates(map[string]interface{}{
		"name":                company.Name,
		"registration_number": company.RegistrationNumber,
		"incorporation_date":  company.IncorporationDate,
		"fiscal_year_end":     company.FiscalYearEnd,
		"email":               company.Email,
		"phone":               company.Phone,
		"website":             company.Website,
		"address":             company.Address,
		"business_type":       company.BusinessType,
		"industry":            company.Industry,
		"description":         company.Description,
		"status":              company.Status,
		"updated_at":          company.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}

// DeleteCompany deletes a company
func DeleteCompany(id int64) error {
	// Check if company exists
	var company Company
	if err := GetDB().First(&company, id).Error; err != nil {
		return errors.New("company not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Delete shareholders
	if err := tx.Where("company_id = ?", id).Delete(&Shareholder{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete board meetings
	if err := tx.Where("company_id = ?", id).Delete(&BoardMeeting{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete company
	if err := tx.Delete(&Company{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}
