package models

import (
	"time"
)

// User represents a user in the Freezone Manager system
type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Don't serialize password
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

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

// BoardMeeting represents a board meeting of a company
type BoardMeeting struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	CompanyID   int64      `gorm:"index;not null" json:"company_id"`
	Title       string     `gorm:"not null" json:"title"`
	Date        time.Time  `gorm:"not null" json:"date"`
	Location    string     `json:"location"`
	Description string     `json:"description"`
	Status      string     `gorm:"not null" json:"status"` // Scheduled, Completed, Cancelled
	Minutes     string     `json:"minutes,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Attendees   []Attendee `gorm:"foreignKey:BoardMeetingID" json:"attendees,omitempty"`
	Company     Company    `gorm:"foreignKey:CompanyID" json:"-"`
}

// Attendee represents an attendee of a board meeting
type Attendee struct {
	ID             int64        `gorm:"primaryKey" json:"id"`
	BoardMeetingID int64        `gorm:"index;not null" json:"board_meeting_id"`
	UserID         int64        `gorm:"index;not null" json:"user_id"`
	Name           string       `gorm:"not null" json:"name"`
	Role           string       `json:"role"`
	Status         string       `gorm:"not null" json:"status"` // Confirmed, Pending, Declined
	CreatedAt      time.Time    `gorm:"autoCreateTime" json:"created_at"`
	BoardMeeting   BoardMeeting `gorm:"foreignKey:BoardMeetingID" json:"-"`
	User           User         `gorm:"foreignKey:UserID" json:"-"`
}

// Vote represents a voting item in the Freezone
type Vote struct {
	ID          int64        `gorm:"primaryKey" json:"id"`
	CompanyID   int64        `gorm:"index;not null" json:"company_id"`
	Title       string       `gorm:"not null" json:"title"`
	Description string       `json:"description"`
	StartDate   time.Time    `gorm:"not null" json:"start_date"`
	EndDate     time.Time    `gorm:"not null" json:"end_date"`
	Status      string       `gorm:"not null" json:"status"` // Open, Closed, Cancelled
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	Options     []VoteOption `gorm:"foreignKey:VoteID" json:"options,omitempty"`
	Ballots     []Ballot     `gorm:"foreignKey:VoteID" json:"ballots,omitempty"`
	Company     Company      `gorm:"foreignKey:CompanyID" json:"-"`
}

// VoteOption represents an option in a vote
type VoteOption struct {
	ID     int64  `gorm:"primaryKey" json:"id"`
	VoteID int64  `gorm:"index;not null" json:"vote_id"`
	Text   string `gorm:"not null" json:"text"`
	Count  int    `gorm:"default:0" json:"count"`
	Vote   Vote   `gorm:"foreignKey:VoteID" json:"-"`
}

// Ballot represents a cast ballot in a vote
type Ballot struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	VoteID       int64      `gorm:"index;not null" json:"vote_id"`
	UserID       int64      `gorm:"index;not null" json:"user_id"`
	VoteOptionID int64      `gorm:"index;not null" json:"vote_option_id"`
	SharesCount  int        `gorm:"not null" json:"shares_count"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	Vote         Vote       `gorm:"foreignKey:VoteID" json:"-"`
	User         User       `gorm:"foreignKey:UserID" json:"-"`
	VoteOption   VoteOption `gorm:"foreignKey:VoteOptionID" json:"-"`
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
