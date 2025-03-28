package models

import (
	"errors"
	"sync"
)

// Store represents an in-memory database for all models
type Store struct {
	Users        []User
	Companies    []Company
	Votes        []Vote
	Products     []Product
	Sales        []Sale
	BoardMeetings []BoardMeeting
	mu           sync.RWMutex
}

// NewStore creates a new in-memory database with initial data
func NewStore() *Store {
	users, companies, votes, products, sales := GenerateFakeData()
	
	// Extract board meetings from companies
	var boardMeetings []BoardMeeting
	for _, company := range companies {
		boardMeetings = append(boardMeetings, company.BoardMeetings...)
	}
	
	return &Store{
		Users:        users,
		Companies:    companies,
		Votes:        votes,
		Products:     products,
		Sales:        sales,
		BoardMeetings: boardMeetings,
	}
}

// GetAllUsers returns all users
func (s *Store) GetAllUsers() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Users
}

// GetAllShareholders returns all shareholders across all companies
func (s *Store) GetAllShareholders() []Shareholder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var allShareholders []Shareholder
	for _, company := range s.Companies {
		allShareholders = append(allShareholders, company.Shareholders...)
	}
	
	return allShareholders
}

// GetUserByID returns a user by ID
func (s *Store) GetUserByID(id int64) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, user := range s.Users {
		if user.ID == id {
			return user, nil
		}
	}
	
	return User{}, errors.New("user not found")
}

// AddUser adds a new user
func (s *Store) AddUser(user User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Users = append(s.Users, user)
}

// GetAllCompanies returns all companies
func (s *Store) GetAllCompanies() []Company {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Companies
}

// GetActiveCompanies returns all active companies
func (s *Store) GetActiveCompanies() []Company {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var activeCompanies []Company
	for _, company := range s.Companies {
		if company.Status == "Active" {
			activeCompanies = append(activeCompanies, company)
		}
	}
	
	return activeCompanies
}

// GetCompanyByID returns a company by ID
func (s *Store) GetCompanyByID(id int64) (Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, company := range s.Companies {
		if company.ID == id {
			return company, nil
		}
	}
	
	return Company{}, errors.New("company not found")
}

// AddCompany adds a new company
func (s *Store) AddCompany(company Company) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Companies = append(s.Companies, company)
}

// GetAllBoardMeetings returns all board meetings
func (s *Store) GetAllBoardMeetings() []BoardMeeting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BoardMeetings
}

// GetBoardMeetingByID returns a board meeting by ID
func (s *Store) GetBoardMeetingByID(id int64) (BoardMeeting, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, meeting := range s.BoardMeetings {
		if meeting.ID == id {
			return meeting, nil
		}
	}
	
	return BoardMeeting{}, errors.New("board meeting not found")
}

// GetBoardMeetingsByCompanyID returns all board meetings for a company
func (s *Store) GetBoardMeetingsByCompanyID(companyID int64) []BoardMeeting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var meetings []BoardMeeting
	for _, meeting := range s.BoardMeetings {
		if meeting.CompanyID == companyID {
			meetings = append(meetings, meeting)
		}
	}
	
	return meetings
}

// AddBoardMeeting adds a new board meeting
func (s *Store) AddBoardMeeting(meeting BoardMeeting) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BoardMeetings = append(s.BoardMeetings, meeting)
}

// UpdateBoardMeeting updates an existing board meeting
func (s *Store) UpdateBoardMeeting(meeting BoardMeeting) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for i, m := range s.BoardMeetings {
		if m.ID == meeting.ID {
			s.BoardMeetings[i] = meeting
			return nil
		}
	}
	
	return errors.New("board meeting not found")
}

// GetAllVotes returns all votes
func (s *Store) GetAllVotes() []Vote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Votes
}

// GetVoteByID returns a vote by ID
func (s *Store) GetVoteByID(id int64) (Vote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, vote := range s.Votes {
		if vote.ID == id {
			return vote, nil
		}
	}
	
	return Vote{}, errors.New("vote not found")
}

// GetVotesByCompanyID returns all votes for a company
func (s *Store) GetVotesByCompanyID(companyID int64) []Vote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var votes []Vote
	for _, vote := range s.Votes {
		if vote.CompanyID == companyID {
			votes = append(votes, vote)
		}
	}
	
	return votes
}

// AddVote adds a new vote
func (s *Store) AddVote(vote Vote) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Votes = append(s.Votes, vote)
}

// GetAllProducts returns all products
func (s *Store) GetAllProducts() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Products
}

// GetProductByID returns a product by ID
func (s *Store) GetProductByID(id int64) (Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, product := range s.Products {
		if product.ID == id {
			return product, nil
		}
	}
	
	return Product{}, errors.New("product not found")
}

// GetProductsByType returns all products of a specific type
func (s *Store) GetProductsByType(productType string) []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var products []Product
	for _, product := range s.Products {
		if product.Type == productType {
			products = append(products, product)
		}
	}
	
	return products
}

// AddProduct adds a new product
func (s *Store) AddProduct(product Product) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Products = append(s.Products, product)
}

// GetAllSales returns all sales
func (s *Store) GetAllSales() []Sale {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Sales
}

// GetSaleByID returns a sale by ID
func (s *Store) GetSaleByID(id int64) (Sale, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, sale := range s.Sales {
		if sale.ID == id {
			return sale, nil
		}
	}
	
	return Sale{}, errors.New("sale not found")
}

// GetSalesByCompanyID returns all sales for a company
func (s *Store) GetSalesByCompanyID(companyID int64) []Sale {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var sales []Sale
	for _, sale := range s.Sales {
		if sale.CompanyID == companyID {
			sales = append(sales, sale)
		}
	}
	
	return sales
}

// AddSale adds a new sale
func (s *Store) AddSale(sale Sale) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sales = append(s.Sales, sale)
}
