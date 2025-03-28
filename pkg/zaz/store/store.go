package store

import (
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
)

// Store is an in-memory data store for the Freezone Manager
type Store struct {
	mu        sync.RWMutex
	Users     map[int64]models.User
	Companies map[int64]models.Company
	Votes     map[int64]models.Vote
	Products  map[int64]models.Product
	Sales     map[int64]models.Sale
	
	// Simple indexes for quicker lookups
	CompanyByName      map[string]int64
	ShareholdersByCompany map[int64][]models.Shareholder
	BoardMeetingsByCompany map[int64][]models.BoardMeeting
	VotesByCompany    map[int64][]models.Vote
	SalesByCompany    map[int64][]models.Sale
}

// NewStore creates a new in-memory data store
func NewStore() *Store {
	return &Store{
		Users:     make(map[int64]models.User),
		Companies: make(map[int64]models.Company),
		Votes:     make(map[int64]models.Vote),
		Products:  make(map[int64]models.Product),
		Sales:     make(map[int64]models.Sale),
		
		CompanyByName:      make(map[string]int64),
		ShareholdersByCompany: make(map[int64][]models.Shareholder),
		BoardMeetingsByCompany: make(map[int64][]models.BoardMeeting),
		VotesByCompany:    make(map[int64][]models.Vote),
		SalesByCompany:    make(map[int64][]models.Sale),
	}
}

// LoadFakeData populates the store with fake data
func (s *Store) LoadFakeData() {
	users, companies, votes, products, sales := models.GenerateFakeData()
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Load users
	for _, user := range users {
		s.Users[user.ID] = user
	}
	
	// Load companies and build indexes
	for _, company := range companies {
		s.Companies[company.ID] = company
		s.CompanyByName[company.Name] = company.ID
		
		// Index shareholders
		s.ShareholdersByCompany[company.ID] = company.Shareholders
		
		// Index board meetings
		s.BoardMeetingsByCompany[company.ID] = company.BoardMeetings
	}
	
	// Load votes and index by company
	for _, vote := range votes {
		s.Votes[vote.ID] = vote
		s.VotesByCompany[vote.CompanyID] = append(s.VotesByCompany[vote.CompanyID], vote)
	}
	
	// Load products
	for _, product := range products {
		s.Products[product.ID] = product
	}
	
	// Load sales and index by company
	for _, sale := range sales {
		s.Sales[sale.ID] = sale
		s.SalesByCompany[sale.CompanyID] = append(s.SalesByCompany[sale.CompanyID], sale)
	}
}

// GetCompany returns a company by ID
func (s *Store) GetCompany(id int64) (models.Company, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	company, exists := s.Companies[id]
	return company, exists
}

// GetCompanyByName returns a company by name
func (s *Store) GetCompanyByName(name string) (models.Company, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	id, exists := s.CompanyByName[name]
	if !exists {
		return models.Company{}, false
	}
	
	company, exists := s.Companies[id]
	return company, exists
}

// GetAllCompanies returns all companies
func (s *Store) GetAllCompanies() []models.Company {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	companies := make([]models.Company, 0, len(s.Companies))
	for _, company := range s.Companies {
		companies = append(companies, company)
	}
	
	return companies
}

// GetActiveCompanies returns all active companies
func (s *Store) GetActiveCompanies() []models.Company {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	companies := make([]models.Company, 0)
	for _, company := range s.Companies {
		if company.Status == "Active" {
			companies = append(companies, company)
		}
	}
	
	return companies
}

// GetAllShareholders returns all shareholders
func (s *Store) GetAllShareholders() []models.Shareholder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	shareholders := make([]models.Shareholder, 0)
	for _, shareholderList := range s.ShareholdersByCompany {
		shareholders = append(shareholders, shareholderList...)
	}
	
	return shareholders
}

// GetAllVotes returns all votes
func (s *Store) GetAllVotes() []models.Vote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	votes := make([]models.Vote, 0, len(s.Votes))
	for _, vote := range s.Votes {
		votes = append(votes, vote)
	}
	
	return votes
}

// GetAllProducts returns all products
func (s *Store) GetAllProducts() []models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	products := make([]models.Product, 0, len(s.Products))
	for _, product := range s.Products {
		products = append(products, product)
	}
	
	return products
}

// GetAllSales returns all sales
func (s *Store) GetAllSales() []models.Sale {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	sales := make([]models.Sale, 0, len(s.Sales))
	for _, sale := range s.Sales {
		sales = append(sales, sale)
	}
	
	return sales
}

// GetStats returns basic statistics about the data
func (s *Store) GetStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := make(map[string]int)
	stats["users"] = len(s.Users)
	stats["companies"] = len(s.Companies)
	
	// Count active companies
	activeCompanies := 0
	for _, company := range s.Companies {
		if company.Status == "Active" {
			activeCompanies++
		}
	}
	stats["active_companies"] = activeCompanies
	
	// Count shareholders
	shareholderCount := 0
	for _, shareholderList := range s.ShareholdersByCompany {
		shareholderCount += len(shareholderList)
	}
	stats["shareholders"] = shareholderCount
	
	stats["votes"] = len(s.Votes)
	stats["products"] = len(s.Products)
	stats["sales"] = len(s.Sales)
	
	return stats
}
