package models

import (
	"time"
)

// Shareholder represents a shareholder of a company
type Shareholder struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	CompanyID  int64     `gorm:"index;not null" json:"company_id"`
	UserID     int64     `json:"user_id,omitempty"`
	Name       string    `gorm:"not null" json:"name"`
	Shares     int       `gorm:"not null" json:"shares"`
	Percentage float64   `gorm:"not null" json:"percentage"`
	Type       string    `gorm:"not null" json:"type"` // Individual, Corporate
	Since      time.Time `json:"since"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Company    Company   `gorm:"foreignKey:CompanyID" json:"-"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
}

// ShareholderHandler handles all operations related to Shareholder model
type ShareholderHandler struct {
	circleID string
	db       *Db
}

// NewShareholderHandler creates a new ShareholderHandler
func NewShareholderHandler(circleID string, db *Db) *ShareholderHandler {
	return &ShareholderHandler{
		circleID: circleID,
		db:       db,
	}
}

// GetAll returns all shareholders
func (h *ShareholderHandler) GetAll() []Shareholder {
	var shareholders []Shareholder
	GetDB().Find(&shareholders)
	return shareholders
}

// GetByID returns a shareholder by ID
func (h *ShareholderHandler) GetByID(id int64) (Shareholder, error) {
	var shareholder Shareholder
	result := GetDB().First(&shareholder, id)
	if result.Error != nil {
		return Shareholder{}, ErrRecordNotFound
	}
	return shareholder, nil
}

// GetByCompanyID returns all shareholders for a company
func (h *ShareholderHandler) GetByCompanyID(companyID int64) []Shareholder {
	var shareholders []Shareholder
	GetDB().Where("company_id = ?", companyID).Find(&shareholders)
	return shareholders
}

// Create adds a new shareholder
func (h *ShareholderHandler) Create(shareholder Shareholder) (int64, error) {
	if shareholder.CreatedAt.IsZero() {
		shareholder.CreatedAt = time.Now()
	}
	if shareholder.UpdatedAt.IsZero() {
		shareholder.UpdatedAt = time.Now()
	}

	result := GetDB().Create(&shareholder)
	if result.Error != nil {
		return 0, result.Error
	}
	return shareholder.ID, nil
}

// Update updates an existing shareholder
func (h *ShareholderHandler) Update(shareholder Shareholder) error {
	shareholder.UpdatedAt = time.Now()

	var existingShareholder Shareholder
	if err := GetDB().First(&existingShareholder, shareholder.ID).Error; err != nil {
		return ErrRecordNotFound
	}

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

// Delete deletes a shareholder
func (h *ShareholderHandler) Delete(id int64) error {
	var shareholder Shareholder
	if err := GetDB().First(&shareholder, id).Error; err != nil {
		return ErrRecordNotFound
	}

	result := GetDB().Delete(&Shareholder{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
