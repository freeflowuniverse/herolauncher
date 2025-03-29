package models

// GetAllCompanies returns all companies from the database
func GetAllCompanies() []Company {
	store := NewStore("default")
	return store.CompanyHandler.GetAll()
}

// GetActiveCompanies returns all active companies from the database
func GetActiveCompanies() []Company {
	allCompanies := GetAllCompanies()
	var activeCompanies []Company
	
	for _, company := range allCompanies {
		if company.Status == "Active" {
			activeCompanies = append(activeCompanies, company)
		}
	}
	
	return activeCompanies
}

// GetAllShareholders returns all shareholders from the database
func GetAllShareholders() []Shareholder {
	store := NewStore("default")
	return store.ShareholderHandler.GetAll()
}

// GetCompanyByID returns a company by its ID
func GetCompanyByID(id int64) (Company, error) {
	store := NewStore("default")
	return store.CompanyHandler.GetByID(id)
}

// AddUser adds a user to the database
func AddUser(user User) (int64, error) {
	store := NewStore("default")
	return store.UserHandler.Create(user)
}

// AddCompany adds a company to the database
func AddCompany(company Company) int64 {
	store := NewStore("default")
	id, err := store.CompanyHandler.Create(company)
	if err != nil {
		return 0
	}
	return id
}

// AddVote adds a vote to the database
func AddVote(vote Vote) (int64, error) {
	store := NewStore("default")
	return store.VoteHandler.Create(vote)
}

// AddProduct adds a product to the database
func AddProduct(product Product) int64 {
	store := NewStore("default")
	id, err := store.ProductHandler.Create(product)
	if err != nil {
		return 0
	}
	return id
}

// AddSale adds a sale to the database
func AddSale(sale Sale) (int64, error) {
	store := NewStore("default")
	return store.SaleHandler.Create(sale)
}
