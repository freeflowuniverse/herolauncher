package models

// InitializeHandlers initializes all handlers for a store
func InitializeHandlers(store *Store) {
	store.UserHandler = NewUserHandler(store.CircleID, store.DB)
	store.CompanyHandler = NewCompanyHandler(store.CircleID, store.DB)
	store.VoteHandler = NewVoteHandler(store.CircleID, store.DB)
	store.ProductHandler = NewProductHandler(store.CircleID, store.DB)
	store.SaleHandler = NewSaleHandler(store.CircleID, store.DB)
	store.ShareholderHandler = NewShareholderHandler(store.CircleID, store.DB)
	store.BoardMeetingHandler = NewBoardMeetingHandler(store.CircleID, store.DB)
}
