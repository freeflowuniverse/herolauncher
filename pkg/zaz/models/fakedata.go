package models

import (
	"fmt"
	"math/rand"
	"time"
)

// FakeDataGenerator generates fake data for all models
type FakeDataGenerator struct {
	rnd *rand.Rand
}

// NewFakeDataGenerator creates a new fake data generator
func NewFakeDataGenerator() *FakeDataGenerator {
	source := rand.NewSource(time.Now().UnixNano())
	return &FakeDataGenerator{
		rnd: rand.New(source),
	}
}

// Industries returns a list of sample industries
func (g *FakeDataGenerator) Industries() []string {
	return []string{
		"Technology",
		"Finance",
		"Healthcare",
		"Education",
		"Entertainment",
		"Manufacturing",
		"Retail",
		"Energy",
		"Transportation",
		"Agriculture",
	}
}

// BusinessTypes returns a list of sample business types
func (g *FakeDataGenerator) BusinessTypes() []string {
	return []string{
		"Limited Liability Company",
		"Corporation",
		"Partnership",
		"Sole Proprietorship",
		"Non-Profit Organization",
	}
}

// RandomElement returns a random element from a slice
func (g *FakeDataGenerator) RandomElement(slice []string) string {
	return slice[g.rnd.Intn(len(slice))]
}

// RandomCompanyName generates a random company name
func (g *FakeDataGenerator) RandomCompanyName() string {
	prefixes := []string{
		"Tech", "Global", "Eco", "Smart", "Bright", "Future", "Next", "Green",
		"Blue", "Red", "Alpha", "Beta", "Nova", "Quantum", "Spark", "Horizon",
	}
	suffixes := []string{
		"Systems", "Solutions", "Technologies", "Group", "Innovations", "Labs",
		"Networks", "Enterprises", "Industries", "Corp", "Inc", "Ltd",
	}
	
	return fmt.Sprintf("%s %s", 
		g.RandomElement(prefixes), 
		g.RandomElement(suffixes),
	)
}

// RandomPersonName generates a random person name
func (g *FakeDataGenerator) RandomPersonName() string {
	firstNames := []string{
		"John", "Jane", "Michael", "Emily", "David", "Sarah", "James", "Emma",
		"Robert", "Anna", "William", "Sophia", "Thomas", "Olivia", "Daniel", "Lisa",
	}
	lastNames := []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Garcia",
		"Wilson", "Anderson", "Taylor", "Thomas", "Moore", "Martin", "Lee", "Perez",
	}
	
	return fmt.Sprintf("%s %s", 
		g.RandomElement(firstNames), 
		g.RandomElement(lastNames),
	)
}

// RandomEmail generates a random email based on name
func (g *FakeDataGenerator) RandomEmail(name string) string {
	domains := []string{
		"example.com", "mail.com", "gmail.com", "outlook.com", "yahoo.com",
		"company.co", "business.org", "tech.io",
	}
	
	// Simplify name for email
	simpleNameRunes := []rune{}
	for _, r := range name {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == ' ' {
			simpleNameRunes = append(simpleNameRunes, r)
		}
	}
	
	simpleName := string(simpleNameRunes)
	simpleName = fmt.Sprintf("%s%d", simpleName, g.rnd.Intn(100))
	simpleName = fmt.Sprintf("%s@%s", simpleName, g.RandomElement(domains))
	
	return simpleName
}

// RandomURL generates a random URL for a company
func (g *FakeDataGenerator) RandomURL(companyName string) string {
	tlds := []string{".com", ".co", ".io", ".org", ".net"}
	
	// Simplify company name for URL
	simpleNameRunes := []rune{}
	for _, r := range companyName {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			simpleNameRunes = append(simpleNameRunes, r)
		}
	}
	
	simpleName := string(simpleNameRunes)
	simpleName = fmt.Sprintf("https://%s%s", simpleName, g.RandomElement(tlds))
	
	return simpleName
}

// RandomPhone generates a random phone number
func (g *FakeDataGenerator) RandomPhone() string {
	return fmt.Sprintf("+%d %d %d %d", 
		g.rnd.Intn(90)+10, 
		g.rnd.Intn(900)+100, 
		g.rnd.Intn(900)+100, 
		g.rnd.Intn(9000)+1000,
	)
}

// RandomDate generates a random date within a range from now
func (g *FakeDataGenerator) RandomDate(minYearsAgo, maxYearsAgo int) time.Time {
	now := time.Now()
	minDaysAgo := minYearsAgo * 365
	maxDaysAgo := maxYearsAgo * 365
	
	daysToSubtract := g.rnd.Intn(maxDaysAgo-minDaysAgo) + minDaysAgo
	return now.AddDate(0, 0, -daysToSubtract)
}

// RandomFutureDate generates a random date in the future
func (g *FakeDataGenerator) RandomFutureDate(maxDaysInFuture int) time.Time {
	now := time.Now()
	daysToAdd := g.rnd.Intn(maxDaysInFuture) + 1
	return now.AddDate(0, 0, daysToAdd)
}

// GenerateUser creates a fake user
func (g *FakeDataGenerator) GenerateUser() User {
	name := g.RandomPersonName()
	email := g.RandomEmail(name)
	now := time.Now()
	
	return User{
		ID:        int64(g.rnd.Intn(10000) + 1),
		Name:      name,
		Email:     email,
		Password:  "password", // Not a real password, just for demo
		Company:   g.RandomCompanyName(),
		Role:      g.RandomElement([]string{"Admin", "User", "Shareholder", "Director"}),
		CreatedAt: g.RandomDate(0, 2),
		UpdatedAt: now,
	}
}

// GenerateCompany creates a fake company
func (g *FakeDataGenerator) GenerateCompany() Company {
	companyName := g.RandomCompanyName()
	incorporationDate := g.RandomDate(1, 10)
	
	// Generate a valid registration number
	regNumber := fmt.Sprintf("BRN%d", g.rnd.Intn(90000000)+10000000)
	
	company := Company{
		ID:                int64(g.rnd.Intn(10000) + 1),
		Name:              companyName,
		RegistrationNumber: regNumber,
		IncorporationDate: incorporationDate,
		FiscalYearEnd:     "December 31",
		Email:             g.RandomEmail(companyName),
		Phone:             g.RandomPhone(),
		Website:           g.RandomURL(companyName),
		Address:           fmt.Sprintf("%d Main Street, City", g.rnd.Intn(1000)+1),
		BusinessType:      g.RandomElement(g.BusinessTypes()),
		Industry:          g.RandomElement(g.Industries()),
		Description:       fmt.Sprintf("%s is a leading company specializing in innovative solutions.", companyName),
		Status:            g.RandomElement([]string{"Active", "Inactive", "Suspended"}),
		CreatedAt:         g.RandomDate(0, 2),
		UpdatedAt:         time.Now(),
	}
	
	// Add shareholders
	shareholderCount := g.rnd.Intn(5) + 1
	company.Shareholders = make([]Shareholder, shareholderCount)
	
	totalShares := 1000 // Total shares for the company
	remainingShares := totalShares
	
	for i := 0; i < shareholderCount; i++ {
		// Last shareholder gets all remaining shares
		shares := 0
		if i == shareholderCount-1 {
			shares = remainingShares
		} else {
			// Distribute shares randomly but ensure we don't give away too many
			maxShare := remainingShares / 2
			if maxShare == 0 {
				maxShare = remainingShares
			}
			shares = g.rnd.Intn(maxShare) + 1
		}
		
		remainingShares -= shares
		percentage := float64(shares) / float64(totalShares) * 100.0
		
		company.Shareholders[i] = g.GenerateShareholderForCompany(company.ID, shares, percentage)
	}
	
	// Add board meetings
	meetingCount := g.rnd.Intn(3) + 1
	company.BoardMeetings = make([]BoardMeeting, meetingCount)
	
	for i := 0; i < meetingCount; i++ {
		company.BoardMeetings[i] = g.GenerateBoardMeetingForCompany(company.ID)
	}
	
	return company
}

// GenerateShareholderForCompany creates a fake shareholder for a specific company
func (g *FakeDataGenerator) GenerateShareholderForCompany(companyID int64, shares int, percentage float64) Shareholder {
	name := g.RandomPersonName()
	since := g.RandomDate(0, 5)
	
	return Shareholder{
		ID:        int64(g.rnd.Intn(10000) + 1),
		CompanyID: companyID,
		UserID:    int64(g.rnd.Intn(10000) + 1),
		Name:      name,
		Shares:    shares,
		Percentage: percentage,
		Type:      g.RandomElement([]string{"Individual", "Corporate"}),
		Since:     since,
		CreatedAt: since,
		UpdatedAt: time.Now(),
	}
}

// GenerateBoardMeetingForCompany creates a fake board meeting for a specific company
func (g *FakeDataGenerator) GenerateBoardMeetingForCompany(companyID int64) BoardMeeting {
	// 50% chance of past meeting, 50% chance of future meeting
	var date time.Time
	var status string
	
	if g.rnd.Intn(2) == 0 {
		// Past meeting
		date = g.RandomDate(0, 1)
		status = "Completed"
	} else {
		// Future meeting
		date = g.RandomFutureDate(90)
		status = "Scheduled"
	}
	
	meeting := BoardMeeting{
		ID:          int64(g.rnd.Intn(10000) + 1),
		CompanyID:   companyID,
		Title:       g.RandomElement([]string{
			"Quarterly Board Meeting", 
			"Annual General Meeting",
			"Special Board Meeting",
			"Strategic Planning Session",
			"Budget Review Meeting",
		}),
		Date:        date,
		Location:    g.RandomElement([]string{
			"Company Headquarters",
			"Conference Room A",
			"Virtual Meeting",
			"Downtown Office",
			"Hotel Conference Center",
		}),
		Description: "Meeting to discuss company strategy and performance.",
		Status:      status,
		CreatedAt:   g.RandomDate(0, 1),
		UpdatedAt:   time.Now(),
	}
	
	// Add attendees
	attendeeCount := g.rnd.Intn(5) + 3
	meeting.Attendees = make([]Attendee, attendeeCount)
	
	for i := 0; i < attendeeCount; i++ {
		meeting.Attendees[i] = g.GenerateAttendeeForMeeting(meeting.ID)
	}
	
	// Add minutes if completed
	if status == "Completed" {
		meeting.Minutes = "The board discussed Q1 results and approved the strategic plan for the next year."
	}
	
	return meeting
}

// GenerateAttendeeForMeeting creates a fake attendee for a specific meeting
func (g *FakeDataGenerator) GenerateAttendeeForMeeting(meetingID int64) Attendee {
	return Attendee{
		ID:            int64(g.rnd.Intn(10000) + 1),
		BoardMeetingID: meetingID,
		UserID:        int64(g.rnd.Intn(10000) + 1),
		Name:          g.RandomPersonName(),
		Role:          g.RandomElement([]string{
			"Board Member", 
			"CEO", 
			"CFO", 
			"Secretary",
			"Legal Counsel",
			"Shareholder",
		}),
		Status:        g.RandomElement([]string{"Confirmed", "Pending", "Declined"}),
		CreatedAt:     g.RandomDate(0, 1),
	}
}

// GenerateVote creates a fake vote
func (g *FakeDataGenerator) GenerateVote(companyID int64) Vote {
	startDate := g.RandomDate(0, 1)
	endDate := startDate.AddDate(0, 0, g.rnd.Intn(14)+1)
	now := time.Now()
	
	var status string
	if endDate.Before(now) {
		status = "Closed"
	} else if startDate.After(now) {
		status = "Scheduled"
	} else {
		status = "Open"
	}
	
	vote := Vote{
		ID:          int64(g.rnd.Intn(10000) + 1),
		CompanyID:   companyID,
		Title:       g.RandomElement([]string{
			"Approval of Annual Financial Statements",
			"Election of Board Members",
			"Dividend Distribution Proposal",
			"Merger Approval",
			"Amendment to Company Bylaws",
		}),
		Description: "Please cast your vote on this important matter.",
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      status,
		CreatedAt:   startDate.AddDate(0, 0, -g.rnd.Intn(7)),
		UpdatedAt:   time.Now(),
	}
	
	// Add options
	optionCount := g.rnd.Intn(3) + 2
	vote.Options = make([]VoteOption, optionCount)
	
	for i := 0; i < optionCount; i++ {
		vote.Options[i] = VoteOption{
			ID:     int64(g.rnd.Intn(10000) + 1),
			VoteID: vote.ID,
			Text:   fmt.Sprintf("Option %d", i+1),
			Count:  g.rnd.Intn(100),
		}
	}
	
	// Add ballots
	ballotCount := g.rnd.Intn(10) + 5
	vote.Ballots = make([]Ballot, ballotCount)
	
	for i := 0; i < ballotCount; i++ {
		optionIndex := g.rnd.Intn(len(vote.Options))
		vote.Ballots[i] = Ballot{
			ID:           int64(g.rnd.Intn(10000) + 1),
			VoteID:       vote.ID,
			UserID:       int64(g.rnd.Intn(10000) + 1),
			VoteOptionID: vote.Options[optionIndex].ID,
			SharesCount:  g.rnd.Intn(100) + 1,
			CreatedAt:    g.RandomDate(0, 1),
		}
	}
	
	return vote
}

// GenerateProduct creates a fake product
func (g *FakeDataGenerator) GenerateProduct() Product {
	productTypes := []string{"Physical", "Service", "Software", "Subscription"}
	productCategories := []string{
		"Software", "Hardware", "Consulting", "Training", "Support", 
		"Cloud Services", "Security", "Analytics", "Infrastructure",
	}
	
	return Product{
		ID:          int64(g.rnd.Intn(10000) + 1),
		Name:        g.RandomElement([]string{
			"Premium License", 
			"Basic Support Package",
			"Enterprise Solution",
			"Developer Toolkit",
			"Annual Subscription",
			"Cloud Storage",
			"Security Audit",
			"Training Workshop",
		}),
		Description: "High-quality product with full support.",
		Price:       float64(g.rnd.Intn(10000)) + 99.99,
		Currency:    "USD",
		Type:        g.RandomElement(productTypes),
		Category:    g.RandomElement(productCategories),
		Status:      g.RandomElement([]string{"Active", "Inactive", "Coming Soon"}),
		CreatedAt:   g.RandomDate(0, 2),
		UpdatedAt:   time.Now(),
	}
}

// GenerateSale creates a fake sale
func (g *FakeDataGenerator) GenerateSale(companyID int64) Sale {
	buyerName := g.RandomPersonName()
	
	sale := Sale{
		ID:          int64(g.rnd.Intn(10000) + 1),
		CompanyID:   companyID,
		BuyerName:   buyerName,
		BuyerEmail:  g.RandomEmail(buyerName),
		Currency:    "USD",
		Status:      g.RandomElement([]string{"Completed", "Pending", "Cancelled"}),
		SaleDate:    g.RandomDate(0, 1),
		CreatedAt:   g.RandomDate(0, 1),
		UpdatedAt:   time.Now(),
	}
	
	// Generate items
	itemCount := g.rnd.Intn(5) + 1
	sale.Items = make([]SaleItem, itemCount)
	
	totalAmount := 0.0
	
	for i := 0; i < itemCount; i++ {
		product := g.GenerateProduct()
		quantity := g.rnd.Intn(5) + 1
		subtotal := product.Price * float64(quantity)
		
		sale.Items[i] = SaleItem{
			ID:        int64(g.rnd.Intn(10000) + 1),
			SaleID:    sale.ID,
			ProductID: product.ID,
			Name:      product.Name,
			Quantity:  quantity,
			UnitPrice: product.Price,
			Currency:  "USD",
			Subtotal:  subtotal,
		}
		
		totalAmount += subtotal
	}
	
	sale.TotalAmount = totalAmount
	
	return sale
}

// GenerateFakeData generates a complete set of fake data and adds it to the database
func GenerateFakeData() {
	generator := NewFakeDataGenerator()
	
	// Generate users
	userCount := 20
	var users []User
	for i := 0; i < userCount; i++ {
		user := generator.GenerateUser()
		AddUser(user)
		users = append(users, user)
	}
	
	// Generate companies
	companyCount := 10
	var companies []Company
	for i := 0; i < companyCount; i++ {
		company := generator.GenerateCompany()
		companyID := AddCompany(company)
		company.ID = companyID
		companies = append(companies, company)
	}
	
	// Generate votes
	voteCount := 15
	for i := 0; i < voteCount; i++ {
		companyIndex := generator.rnd.Intn(len(companies))
		vote := generator.GenerateVote(companies[companyIndex].ID)
		AddVote(vote)
	}
	
	// Generate products
	productCount := 25
	var products []Product
	for i := 0; i < productCount; i++ {
		product := generator.GenerateProduct()
		productID := AddProduct(product)
		product.ID = productID
		products = append(products, product)
	}
	
	// Generate sales
	saleCount := 30
	for i := 0; i < saleCount; i++ {
		companyIndex := generator.rnd.Intn(len(companies))
		sale := generator.GenerateSale(companies[companyIndex].ID)
		AddSale(sale)
	}
}
