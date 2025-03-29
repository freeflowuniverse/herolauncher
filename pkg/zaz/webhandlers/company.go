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
	
	return RenderWithDefaults(c, "index", fiber.Map{
		"title": "Dashboard",
		"companiesCount": len(companies),
		"activeCompaniesCount": activeCompanies,
		"shareholdersCount": len(shareholders),
	})
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
		"searchQuery": searchQuery,
	})
}

// GetCreateCompany renders the company creation page
func (h *CompanyHandler) GetCreateCompany(c *fiber.Ctx) error {
	return RenderWithDefaults(c, "companies_create", fiber.Map{
		"title": "Create Company",
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
