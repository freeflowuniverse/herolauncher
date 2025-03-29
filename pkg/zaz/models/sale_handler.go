package models

import (
	"time"
)

// SaleHandler handles all operations related to Sale model
type SaleHandler struct {
	circleID string
	db       *Db
}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler(circleID string, db *Db) *SaleHandler {
	return &SaleHandler{
		circleID: circleID,
		db:       db,
	}
}

// Sale represents a sale of products or services
type Sale struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	CompanyID   int64      `gorm:"index;not null" json:"company_id"`
	BuyerName   string     `gorm:"not null" json:"buyer_name"`
	BuyerEmail  string     `json:"buyer_email"`
	TotalAmount float64    `gorm:"not null" json:"total_amount"`
	Currency    string     `gorm:"not null" json:"currency"`
	Status      string     `gorm:"not null" json:"status"` // Pending, Completed, Cancelled
	SaleDate    time.Time  `gorm:"not null" json:"sale_date"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Items       []SaleItem `gorm:"foreignKey:SaleID" json:"items,omitempty"`
	Company     Company    `gorm:"foreignKey:CompanyID" json:"-"`
}

// SaleItem represents an item in a sale
type SaleItem struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	SaleID    int64   `gorm:"index;not null" json:"sale_id"`
	ProductID int64   `gorm:"index;not null" json:"product_id"`
	Name      string  `gorm:"not null" json:"name"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	UnitPrice float64 `gorm:"not null" json:"unit_price"`
	Currency  string  `gorm:"not null" json:"currency"`
	Subtotal  float64 `gorm:"not null" json:"subtotal"`
	Sale      Sale    `gorm:"foreignKey:SaleID" json:"-"`
	Product   Product `gorm:"foreignKey:ProductID" json:"-"`
}

// GetAll returns all sales
func (h *SaleHandler) GetAll() []Sale {
	var sales []Sale
	GetDB().Preload("Items").Find(&sales)
	return sales
}

// GetByID returns a sale by ID
func (h *SaleHandler) GetByID(id int64) (Sale, error) {
	var sale Sale
	result := GetDB().Preload("Items").First(&sale, id)
	if result.Error != nil {
		return Sale{}, ErrRecordNotFound
	}
	return sale, nil
}

// GetByCompanyID returns all sales for a company
func (h *SaleHandler) GetByCompanyID(companyID int64) []Sale {
	var sales []Sale
	GetDB().Preload("Items").Where("company_id = ?", companyID).Find(&sales)
	return sales
}

// Create adds a new sale
func (h *SaleHandler) Create(sale Sale) (int64, error) {
	if sale.CreatedAt.IsZero() {
		sale.CreatedAt = time.Now()
	}
	if sale.UpdatedAt.IsZero() {
		sale.UpdatedAt = time.Now()
	}

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// Create sale
	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Create sale items
	for i := range sale.Items {
		sale.Items[i].SaleID = sale.ID
		if err := tx.Create(&sale.Items[i]).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return sale.ID, nil
}

// Update updates an existing sale
func (h *SaleHandler) Update(sale Sale) error {
	sale.UpdatedAt = time.Now()

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if sale exists
	var existingSale Sale
	if err := tx.First(&existingSale, sale.ID).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
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

	// Update sale items (delete and re-insert)
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

// Delete deletes a sale and all associated items
func (h *SaleHandler) Delete(id int64) error {
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if sale exists
	var sale Sale
	if err := tx.First(&sale, id).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
	}

	// Delete sale items
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
