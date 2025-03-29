package models

import (
	"time"
)

// Company represents a company registered in the Freezone
type Company struct {
	ID                 int64          `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"not null" json:"name"`
	RegistrationNumber string         `gorm:"uniqueIndex;not null" json:"registration_number"`
	IncorporationDate  time.Time      `json:"incorporation_date"`
	FiscalYearEnd      string         `json:"fiscal_year_end"`
	Email              string         `json:"email"`
	Phone              string         `json:"phone"`
	Website            string         `json:"website"`
	Address            string         `json:"address"`
	BusinessType       string         `json:"business_type"`
	Industry           string         `json:"industry"`
	Description        string         `json:"description"`
	Status             string         `json:"status"` // Active, Inactive, Suspended
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	Shareholders       []Shareholder  `gorm:"foreignKey:CompanyID" json:"shareholders,omitempty"`
	BoardMeetings      []BoardMeeting `gorm:"foreignKey:CompanyID" json:"boardmeetings,omitempty"`
}

// CompanyHandler handles all operations related to Company model
type CompanyHandler struct {
	circleID string
	db       *Db
}

// NewCompanyHandler creates a new CompanyHandler
func NewCompanyHandler(circleID string, db *Db) *CompanyHandler {
	return &CompanyHandler{
		circleID: circleID,
		db:       db,
	}
}

// GetAll returns all companies
func (h *CompanyHandler) GetAll() []Company {
	var companies []Company
	GetDB().Find(&companies)
	return companies
}

// GetByID returns a company by ID
func (h *CompanyHandler) GetByID(id int64) (Company, error) {
	var company Company
	result := GetDB().Preload("Shareholders").Preload("BoardMeetings").First(&company, id)
	if result.Error != nil {
		return Company{}, ErrRecordNotFound
	}
	return company, nil
}

// Create adds a new company
func (h *CompanyHandler) Create(company Company) (int64, error) {
	if company.CreatedAt.IsZero() {
		company.CreatedAt = time.Now()
	}
	if company.UpdatedAt.IsZero() {
		company.UpdatedAt = time.Now()
	}

	result := GetDB().Create(&company)
	if result.Error != nil {
		return 0, result.Error
	}
	return company.ID, nil
}

// Update updates an existing company
func (h *CompanyHandler) Update(company Company) error {
	company.UpdatedAt = time.Now()

	var existingCompany Company
	if err := GetDB().First(&existingCompany, company.ID).Error; err != nil {
		return ErrRecordNotFound
	}

	result := GetDB().Model(&company).Updates(map[string]interface{}{
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
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete deletes a company
func (h *CompanyHandler) Delete(id int64) error {
	var company Company
	if err := GetDB().First(&company, id).Error; err != nil {
		return ErrRecordNotFound
	}

	// Begin transaction to delete company and related records
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete related shareholders
	if err := tx.Where("company_id = ?", id).Delete(&Shareholder{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete related board meetings
	if err := tx.Where("company_id = ?", id).Delete(&BoardMeeting{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete company
	if err := tx.Delete(&Company{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
