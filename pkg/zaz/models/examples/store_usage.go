package examples

import (
	"fmt"
	"log"
	"time"
)

// This example demonstrates how to use the Store and its model handlers

func UseStore() {
	// Initialize the database
	_, err := InitDB("./data/test.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a new store with a circle ID
	store := NewStore("circle123")

	// Using the VoteHandler
	useVoteHandler(store)

	// Using the CompanyHandler
	useCompanyHandler(store)

	// Using the ProductHandler
	useProductHandler(store)
}

func useVoteHandler(store *Store) {
	// Create a new vote
	vote := Vote{
		CompanyID:   1,
		Title:       "Board Election 2025",
		Description: "Election for new board members",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(7 * 24 * time.Hour), // 1 week from now
		Status:      "Open",
		Options: []VoteOption{
			{Text: "Candidate A", Count: 0},
			{Text: "Candidate B", Count: 0},
			{Text: "Candidate C", Count: 0},
		},
	}

	// Create the vote using the VoteHandler
	voteID, err := store.VoteHandler.Create(vote)
	if err != nil {
		log.Printf("Failed to create vote: %v", err)
		return
	}
	fmt.Printf("Created vote with ID: %d\n", voteID)

	// Get the vote by ID
	createdVote, err := store.VoteHandler.GetByID(voteID)
	if err != nil {
		log.Printf("Failed to get vote: %v", err)
		return
	}
	fmt.Printf("Retrieved vote: %s\n", createdVote.Title)

	// Cast a ballot
	ballot := Ballot{
		VoteID:       voteID,
		UserID:       1,
		VoteOptionID: createdVote.Options[0].ID,
		SharesCount:  100,
	}
	err = store.VoteHandler.CastBallot(ballot)
	if err != nil {
		log.Printf("Failed to cast ballot: %v", err)
		return
	}
	fmt.Println("Ballot cast successfully")

	// Update the vote
	createdVote.Status = "Closed"
	err = store.VoteHandler.Update(createdVote)
	if err != nil {
		log.Printf("Failed to update vote: %v", err)
		return
	}
	fmt.Println("Vote updated successfully")

	// Get all votes for a company
	companyVotes := store.VoteHandler.GetByCompanyID(1)
	fmt.Printf("Found %d votes for company ID 1\n", len(companyVotes))
}

func useCompanyHandler(store *Store) {
	// Create a new company
	company := Company{
		Name:               "Acme Corp",
		RegistrationNumber: "REG12345",
		IncorporationDate:  time.Now(),
		FiscalYearEnd:      "December",
		Email:              "info@acmecorp.com",
		Status:             "Active",
	}

	// Create the company using the CompanyHandler
	companyID, err := store.CompanyHandler.Create(company)
	if err != nil {
		log.Printf("Failed to create company: %v", err)
		return
	}
	fmt.Printf("Created company with ID: %d\n", companyID)

	// Add a shareholder to the company
	shareholder := Shareholder{
		CompanyID:  companyID,
		Name:       "John Doe",
		Shares:     1000,
		Percentage: 10.0,
		Type:       "Individual",
		Since:      time.Now(),
	}

	shareholderID, err := store.ShareholderHandler.Create(shareholder)
	if err != nil {
		log.Printf("Failed to create shareholder: %v", err)
		return
	}
	fmt.Printf("Created shareholder with ID: %d\n", shareholderID)

	// Get the company with its shareholders
	createdCompany, err := store.CompanyHandler.GetByID(companyID)
	if err != nil {
		log.Printf("Failed to get company: %v", err)
		return
	}
	fmt.Printf("Retrieved company: %s with %d shareholders\n", createdCompany.Name, len(createdCompany.Shareholders))
}

func useProductHandler(store *Store) {
	// Create a new product
	product := Product{
		Name:        "Widget Pro",
		Description: "Professional grade widget",
		Price:       99.99,
		Currency:    "USD",
		Type:        "Product",
		Category:    "Electronics",
		Status:      "Available",
	}

	// Create the product using the ProductHandler
	productID, err := store.ProductHandler.Create(product)
	if err != nil {
		log.Printf("Failed to create product: %v", err)
		return
	}
	fmt.Printf("Created product with ID: %d\n", productID)

	// Create a sale with the product
	sale := Sale{
		CompanyID:   1,
		BuyerName:   "Jane Smith",
		BuyerEmail:  "jane@example.com",
		TotalAmount: 99.99,
		Currency:    "USD",
		Status:      "Completed",
		SaleDate:    time.Now(),
		Items: []SaleItem{
			{
				ProductID: productID,
				Name:      "Widget Pro",
				Quantity:  1,
				UnitPrice: 99.99,
				Currency:  "USD",
				Subtotal:  99.99,
			},
		},
	}

	// Create the sale using the SaleHandler
	saleID, err := store.SaleHandler.Create(sale)
	if err != nil {
		log.Printf("Failed to create sale: %v", err)
		return
	}
	fmt.Printf("Created sale with ID: %d\n", saleID)

	// Get the sale with its items
	createdSale, err := store.SaleHandler.GetByID(saleID)
	if err != nil {
		log.Printf("Failed to get sale: %v", err)
		return
	}
	fmt.Printf("Retrieved sale: Buyer %s, Total: %.2f %s\n",
		createdSale.BuyerName, createdSale.TotalAmount, createdSale.Currency)
}
