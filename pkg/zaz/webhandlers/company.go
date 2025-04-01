package webhandlers

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// CompanyHandler handles company-related routes
type CompanyHandler struct {
	store *models.Store
}

// NewCompanyHandler creates a new CompanyHandler
func NewCompanyHandler(store *models.Store) *CompanyHandler {
	return &CompanyHandler{
		store: store,
	}
}

// GetDashboard renders the dashboard page
func (h *CompanyHandler) GetDashboard(c *fiber.Ctx) error {
	log.Println("CompanyHandler.GetDashboard method called")
	
	// Get data directly from the store
	companies := h.store.CompanyHandler.GetAll()
	shareholders := h.store.ShareholderHandler.GetAll()
	
	// Count active companies
	activeCompanies := 0
	for _, company := range companies {
		if company.Status == "Active" {
			activeCompanies++
		}
	}
	
	// Get actual board meetings
	boardMeetings := h.store.BoardMeetingHandler.GetAll()
	
	// Create a simple array for upcoming meetings
	upcomingMeetings := []fiber.Map{}
	
	// Process actual board meetings
	for _, meeting := range boardMeetings {
		// Get the company for this meeting
		company, err := h.store.CompanyHandler.GetByID(meeting.CompanyID)
		if err != nil {
			// If company not found, still include the meeting but with nil company
			upcomingMeetings = append(upcomingMeetings, fiber.Map{
				"ID":     meeting.ID,
				"Date":   meeting.Date.Format("2006-01-02 15:04"),
				"Title":  meeting.Title,
				"Company": nil,
			})
		} else {
			// Include meeting with company info
			upcomingMeetings = append(upcomingMeetings, fiber.Map{
				"ID":    meeting.ID,
				"Date":  meeting.Date.Format("2006-01-02 15:04"),
				"Title": meeting.Title,
				"Company": fiber.Map{
					"Name": company.Name,
					"ID":   company.ID,
				},
			})
		}
	}
	
	// If no meetings found, add a dummy one for testing
	if len(upcomingMeetings) == 0 {
		upcomingMeetings = append(upcomingMeetings, fiber.Map{
			"ID":    int64(1),
			"Date":  time.Now().Format("2006-01-02 15:04"),
			"Title": "Test Meeting",
			"Company": fiber.Map{
				"Name": "Test Company",
				"ID":   int64(1),
			},
		})
	}
	
	log.Printf("Debug - upcomingMeetings structure: %#v", upcomingMeetings)
	
	// Create the data map
	data := fiber.Map{
		"title":                "Dashboard",
		"companiesCount":       len(companies),
		"activeCompaniesCount": activeCompanies,
		"shareholdersCount":    len(shareholders),
		"upcomingMeetings":     upcomingMeetings,
		"recentActivities":     nil,
		"currentYear":          time.Now().Year(),
	}
	
	log.Printf("Debug - Full data map being passed to template: %#v", data)
	
	return RenderWithDefaults(c, "index", data)
}

// GetCompanies renders the companies list page
func (h *CompanyHandler) GetCompanies(c *fiber.Ctx) error {
	// Get companies directly from the store
	companies := h.store.CompanyHandler.GetAll()
	log.Printf("Initial companies count from store: %d", len(companies))
	
	// Filter by search query if provided
	searchQuery := c.Query("search")
	if searchQuery != "" {
		filteredCompanies := []models.Company{}
		for _, company := range companies {
			if strings.Contains(strings.ToLower(company.Name), strings.ToLower(searchQuery)) {
				filteredCompanies = append(filteredCompanies, company)
			}
		}
		companies = filteredCompanies
		log.Printf("After filtering by search query '%s': %d companies", searchQuery, len(companies))
	}
	
	// Render template
	return RenderWithDefaults(c, "companies", fiber.Map{
		"title": "Companies",
		"companies": companies,
		"count": len(companies),
		"search": searchQuery,
		"currentYear": time.Now().Year(),
	})
}

// GetCreateCompany renders the company creation page
func (h *CompanyHandler) GetCreateCompany(c *fiber.Ctx) error {
	// Define business types
	businessTypes := []fiber.Map{
		{"Value": "corporation", "Name": "Corporation"},
		{"Value": "llc", "Name": "Limited Liability Company"},
		{"Value": "partnership", "Name": "Partnership"},
		{"Value": "sole_proprietorship", "Name": "Sole Proprietorship"},
	}

	// Define industries
	industries := []fiber.Map{
		{"Value": "technology", "Name": "Technology"},
		{"Value": "finance", "Name": "Finance"},
		{"Value": "healthcare", "Name": "Healthcare"},
		{"Value": "retail", "Name": "Retail"},
		{"Value": "manufacturing", "Name": "Manufacturing"},
		{"Value": "other", "Name": "Other"},
	}

	// Define months for fiscal year end
	months := []fiber.Map{
		{"Value": "1", "Name": "January"},
		{"Value": "2", "Name": "February"},
		{"Value": "3", "Name": "March"},
		{"Value": "4", "Name": "April"},
		{"Value": "5", "Name": "May"},
		{"Value": "6", "Name": "June"},
		{"Value": "7", "Name": "July"},
		{"Value": "8", "Name": "August"},
		{"Value": "9", "Name": "September"},
		{"Value": "10", "Name": "October"},
		{"Value": "11", "Name": "November"},
		{"Value": "12", "Name": "December"},
	}

	return RenderWithDefaults(c, "companies_create", fiber.Map{
		"title": "Create Company",
		"businessTypes": businessTypes,
		"industries": industries,
		"months": months,
		"form": fiber.Map{}, // Empty form for initial load
		"formErrors": []string{}, // Empty errors for initial load
		"csrfToken": "sample-token", // This would normally be generated
		"currentYear": time.Now().Year(),
	})
}

// PostCreateCompany handles company creation form submission
func (h *CompanyHandler) PostCreateCompany(c *fiber.Ctx) error {
	// Parse form data
	company := models.Company{
		Name:               c.FormValue("name"),
		RegistrationNumber: c.FormValue("registration_number"),
		Email:              c.FormValue("email"),
		Phone:              c.FormValue("phone"),
		Website:            c.FormValue("website"),
		Address:            c.FormValue("address"),
		BusinessType:       c.FormValue("business_type"),
		Industry:           c.FormValue("industry"),
		Description:        c.FormValue("description"),
		Status:             "Active",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	
	// Create company using the store
	_, err := h.store.CompanyHandler.Create(company)
	if err != nil {
		return RenderWithDefaults(c, "companies_create", fiber.Map{
			"title": "Create Company",
			"error": "Failed to create company: " + err.Error(),
			"company": company,
		})
	}
	
	// Redirect to companies list
	return c.Redirect("/companies")
}

// GetCompanyDetails renders the company details page
func (h *CompanyHandler) GetCompanyDetails(c *fiber.Ctx) error {
	// Parse company ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Get shareholders for this company
	shareholders := h.store.ShareholderHandler.GetByCompanyID(id)
	
	// Get board meetings for this company
	boardMeetings := h.store.BoardMeetingHandler.GetByCompanyID(id)
	
	// Render template
	return RenderWithDefaults(c, "company_details", fiber.Map{
		"title": company.Name,
		"company": company,
		"shareholders": shareholders,
		"boardMeetings": boardMeetings,
	})
}

// GetEditCompany renders the company edit page
func (h *CompanyHandler) GetEditCompany(c *fiber.Ctx) error {
	// Parse company ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Render template
	return RenderWithDefaults(c, "company_edit", fiber.Map{
		"title": "Edit " + company.Name,
		"company": company,
	})
}

// PostEditCompany handles company edit form submission
func (h *CompanyHandler) PostEditCompany(c *fiber.Ctx) error {
	// Parse company ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}
	
	// Get existing company
	company, err := h.store.CompanyHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Update company fields
	company.Name = c.FormValue("name")
	company.RegistrationNumber = c.FormValue("registration_number")
	company.Email = c.FormValue("email")
	company.Phone = c.FormValue("phone")
	company.Website = c.FormValue("website")
	company.Address = c.FormValue("address")
	company.BusinessType = c.FormValue("business_type")
	company.Industry = c.FormValue("industry")
	company.Description = c.FormValue("description")
	company.Status = c.FormValue("status")
	company.UpdatedAt = time.Now()
	
	// Update company in store
	err = h.store.CompanyHandler.Update(company)
	if err != nil {
		return RenderWithDefaults(c, "company_edit", fiber.Map{
			"title": "Edit " + company.Name,
			"company": company,
			"error": "Failed to update company: " + err.Error(),
		})
	}
	
	// Redirect to company details
	return c.Redirect("/companies/" + c.Params("id"))
}

// GetCompaniesAPI returns companies data as JSON for API consumption
func (h *CompanyHandler) GetCompaniesAPI(c *fiber.Ctx) error {
	companies := h.store.CompanyHandler.GetAll()
	return c.JSON(companies)
}

// GetCompanyDetailsAPI returns company details as JSON for API consumption
func (h *CompanyHandler) GetCompanyDetailsAPI(c *fiber.Ctx) error {
	// Parse company ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Company not found",
		})
	}
	
	return c.JSON(company)
}
