package handlers

import (
	"log"
	"strconv"
	"strings"

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
	// Calculate stats from the store
	companiesCount := len(h.store.GetAllCompanies())
	activeCompaniesCount := len(h.store.GetActiveCompanies())
	shareholderCount := len(h.store.GetAllShareholders())
	
	return RenderWithDefaults(c, "index", fiber.Map{
		"title": "Dashboard",
		"companiesCount": companiesCount,
		"activeCompaniesCount": activeCompaniesCount,
		"shareholdersCount": shareholderCount,
	})
}

// GetCompanies renders the companies list page
func (h *CompanyHandler) GetCompanies(c *fiber.Ctx) error {
	// Get companies from the store
	companies := h.store.GetAllCompanies()
	
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
	}

	// Ensure we have company data
	if len(companies) == 0 {
		// If no companies exist in the store, log a warning
		log.Printf("Warning: No companies found in the store")
	}

	// Debug log to verify companies are being loaded
	log.Printf("Rendering companies page with %d companies", len(companies))
	
	// Log some sample company data for debugging
	if len(companies) > 0 {
		sampleCompany := companies[0]
		log.Printf("Sample company: ID=%d, Name=%s, Status=%s, Industry=%s, Shareholders=%d", 
			sampleCompany.ID, 
			sampleCompany.Name, 
			sampleCompany.Status, 
			sampleCompany.Industry, 
			len(sampleCompany.Shareholders))
	}

	return RenderWithDefaults(c, "companies", fiber.Map{
		"title": "Companies",
		"companies": companies,
		"search": searchQuery,
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
	name := c.FormValue("name")
	_ = c.FormValue("registration_number") // Use the variable to avoid unused variable warning
	_ = c.FormValue("incorporation_date") // Use the variable to avoid unused variable warning
	
	// Simple validation
	if name == "" {
		return RenderWithDefaults(c, "companies_create", fiber.Map{
			"title": "Create Company",
			"error": "Company name is required",
		})
	}

	// TODO: Implement actual company creation
	// For now, just redirect to companies list
	return c.Redirect("/companies")
}

// GetCompanyDetails renders the company details page
func (h *CompanyHandler) GetCompanyDetails(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"title": "Error",
			"error": "Invalid company ID",
		})
	}
	
	// Get company from the store
	company, err := h.store.GetCompanyByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title": "Error",
			"error": "Company not found",
		})
	}
	
	// Get board meetings for this company
	boardMeetings := h.store.GetBoardMeetingsByCompanyID(company.ID)
	company.BoardMeetings = boardMeetings
	
	// Get votes for this company
	votes := h.store.GetVotesByCompanyID(company.ID)

	return RenderWithDefaults(c, "company_details", fiber.Map{
		"title": company.Name,
		"company": company,
		"votes": votes,
	})
}

// GetEditCompany renders the company edit page
func (h *CompanyHandler) GetEditCompany(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).Render("error", fiber.Map{
			"title": "Error",
			"error": "Invalid company ID",
		})
	}
	
	// Get company from the store
	company, err := h.store.GetCompanyByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).Render("error", fiber.Map{
			"title": "Error",
			"error": "Company not found",
		})
	}

	return c.Render("companies_edit", fiber.Map{
		"title": "Edit " + company.Name,
		"company": company,
	})
}

// PostEditCompany handles company edit form submission
func (h *CompanyHandler) PostEditCompany(c *fiber.Ctx) error {
	companyID := c.Params("id")
	
	// Parse form data
	name := c.FormValue("name")
	
	// Simple validation
	if name == "" {
		return c.Render("companies_edit", fiber.Map{
			"title": "Edit Company",
			"error": "Company name is required",
		})
	}

	// TODO: Implement actual company update
	// For now, just redirect to company details
	return c.Redirect("/companies/" + companyID)
}

// GetCompaniesAPI returns companies data as JSON for API consumption
func (h *CompanyHandler) GetCompaniesAPI(c *fiber.Ctx) error {
	// Get companies from the store
	companies := h.store.GetAllCompanies()

	return c.JSON(fiber.Map{
		"success": true,
		"companies": companies,
	})
}

// GetCompanyDetailsAPI returns company details as JSON for API consumption
func (h *CompanyHandler) GetCompanyDetailsAPI(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": "Invalid company ID",
		})
	}
	
	// Get company from the store
	company, err := h.store.GetCompanyByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error": "Company not found",
		})
	}

	// Get board meetings for this company
	boardMeetings := h.store.GetBoardMeetingsByCompanyID(company.ID)
	company.BoardMeetings = boardMeetings

	return c.JSON(fiber.Map{
		"success": true,
		"company": company,
	})
}
