package models

import ()

// Store represents a database interface for all models
type Store struct {
	CircleID            string
	UserHandler         *UserHandler
	CompanyHandler      *CompanyHandler
	VoteHandler         *VoteHandler
	ProductHandler      *ProductHandler
	SaleHandler         *SaleHandler
	ShareholderHandler  *ShareholderHandler
	BoardMeetingHandler *BoardMeetingHandler
	DB                  *Db
}

// NewStore creates a new store with connection to the database
func NewStore(circleID string) *Store {
	db := GetDbObject()

	store := &Store{
		CircleID: circleID,
		DB:       db,
	}

	// Initialize all handlers
	store.UserHandler = NewUserHandler(circleID, db)
	store.CompanyHandler = NewCompanyHandler(circleID, db)
	store.VoteHandler = NewVoteHandler(circleID, db)
	store.ProductHandler = NewProductHandler(circleID, db)
	store.SaleHandler = NewSaleHandler(circleID, db)
	store.ShareholderHandler = NewShareholderHandler(circleID, db)
	store.BoardMeetingHandler = NewBoardMeetingHandler(circleID, db)

	return store
}
