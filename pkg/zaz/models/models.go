package models

import (
	"time"
)

// User represents a user in the Freezone Manager system
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Don't serialize password
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Company represents a company registered in the Freezone
type Company struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	RegistrationNumber string   `json:"registration_number"`
	IncorporationDate time.Time `json:"incorporation_date"`
	FiscalYearEnd     string    `json:"fiscal_year_end"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Website           string    `json:"website"`
	Address           string    `json:"address"`
	BusinessType      string    `json:"business_type"`
	Industry          string    `json:"industry"`
	Description       string    `json:"description"`
	Status            string    `json:"status"` // Active, Inactive, Suspended
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Shareholders      []Shareholder `json:"shareholders,omitempty"`
	BoardMeetings     []BoardMeeting `json:"boardmeetings,omitempty"`
}

// Shareholder represents a shareholder of a company
type Shareholder struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	UserID    int64     `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Shares    int       `json:"shares"`
	Percentage float64   `json:"percentage"`
	Type      string    `json:"type"` // Individual, Corporate
	Since     time.Time `json:"since"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BoardMeeting represents a board meeting of a company
type BoardMeeting struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	Title       string    `json:"title"`
	Date        time.Time `json:"date"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // Scheduled, Completed, Cancelled
	Minutes     string    `json:"minutes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Attendees   []Attendee `json:"attendees,omitempty"`
}

// Attendee represents an attendee of a board meeting
type Attendee struct {
	ID            int64     `json:"id"`
	BoardMeetingID int64     `json:"board_meeting_id"`
	UserID        int64     `json:"user_id"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	Status        string    `json:"status"` // Confirmed, Pending, Declined
	CreatedAt     time.Time `json:"created_at"`
}

// Vote represents a voting item in the Freezone
type Vote struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"` // Open, Closed, Cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Options     []VoteOption `json:"options,omitempty"`
	Ballots     []Ballot `json:"ballots,omitempty"`
}

// VoteOption represents an option in a vote
type VoteOption struct {
	ID      int64  `json:"id"`
	VoteID  int64  `json:"vote_id"`
	Text    string `json:"text"`
	Count   int    `json:"count"`
}

// Ballot represents a cast ballot in a vote
type Ballot struct {
	ID            int64     `json:"id"`
	VoteID        int64     `json:"vote_id"`
	UserID        int64     `json:"user_id"`
	VoteOptionID  int64     `json:"vote_option_id"`
	SharesCount   int       `json:"shares_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// Product represents a product or service offered by the Freezone
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Type        string    `json:"type"` // Product, Service
	Category    string    `json:"category"`
	Status      string    `json:"status"` // Available, Unavailable
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Sale represents a sale of products or services
type Sale struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	BuyerName   string    `json:"buyer_name"`
	BuyerEmail  string    `json:"buyer_email"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"` // Pending, Completed, Cancelled
	SaleDate    time.Time `json:"sale_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Items       []SaleItem `json:"items,omitempty"`
}

// SaleItem represents an item in a sale
type SaleItem struct {
	ID        int64   `json:"id"`
	SaleID    int64   `json:"sale_id"`
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Currency  string  `json:"currency"`
	Subtotal  float64 `json:"subtotal"`
}
