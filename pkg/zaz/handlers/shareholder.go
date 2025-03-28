package handlers

import (
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// ShareholderHandler handles shareholder-related routes
type ShareholderHandler struct {}

// NewShareholderHandler creates a new ShareholderHandler
func NewShareholderHandler(_ *models.Store) *ShareholderHandler {
	return &ShareholderHandler{}
}

// GetShareholders renders the shareholders list page
func (h *ShareholderHandler) GetShareholders(c *fiber.Ctx) error {
	// Get shareholders using model function directly
	shareholders := models.GetAllShareholders()

	return c.Render("shareholders", fiber.Map{
		"title": "Shareholders",
		"shareholders": shareholders,
		"search": c.Query("search"),
	})
}

// GetCreateShareholder renders the shareholder creation page
func (h *ShareholderHandler) GetCreateShareholder(c *fiber.Ctx) error {
	// Get list of companies using model function directly
	companies := models.GetAllCompanies()

	return c.Render("shareholders_create", fiber.Map{
		"title": "Create Shareholder",
		"companies": companies,
	})
}

// PostCreateShareholder handles shareholder creation form submission
func (h *ShareholderHandler) PostCreateShareholder(c *fiber.Ctx) error {
	// Parse form data
	name := c.FormValue("name")
	companyID := c.FormValue("company_id")
	
	// Simple validation
	if name == "" || companyID == "" {
		return c.Render("shareholders_create", fiber.Map{
			"title": "Create Shareholder",
			"error": "Name and company are required",
		})
	}

	// TODO: Implement actual shareholder creation
	// For now, just redirect to shareholders list
	return c.Redirect("/shareholders")
}

// GetAddShareholder renders the add shareholder to company page
func (h *ShareholderHandler) GetAddShareholder(c *fiber.Ctx) error {
	_ = c.Params("companyId") // Use the parameter to avoid unused variable warning
	
	// In a real implementation, we would fetch the company from the database
	company := models.Company{
		ID: 1,
		Name: "TechCorp Inc.",
	}

	return c.Render("shareholders_add", fiber.Map{
		"title": "Add Shareholder to " + company.Name,
		"company": company,
	})
}

// PostAddShareholder handles adding shareholder to company form submission
func (h *ShareholderHandler) PostAddShareholder(c *fiber.Ctx) error {
	companyID := c.Params("companyId")
	
	// Parse form data
	name := c.FormValue("name")
	shares := c.FormValue("shares")
	
	// Simple validation
	if name == "" || shares == "" {
		return c.Render("shareholders_add", fiber.Map{
			"title": "Add Shareholder",
			"error": "Name and shares are required",
		})
	}

	// TODO: Implement actual shareholder creation
	// For now, just redirect to company details
	return c.Redirect("/companies/" + companyID)
}

// GetShareholderDetails renders the shareholder details page
func (h *ShareholderHandler) GetShareholderDetails(c *fiber.Ctx) error {
	// Get shareholder ID from URL
	id := c.Params("id")
	
	// Convert ID to int64
	shareholderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(400).Render("error", fiber.Map{
			"title": "Error",
			"message": "Invalid shareholder ID",
		})
	}
	
	// Get shareholder using model function directly
	shareholder, err := models.GetShareholderByID(shareholderID)
	if err != nil {
		return c.Status(404).Render("error", fiber.Map{
			"title": "Not Found",
			"message": "Shareholder not found",
		})
	}
	
	// Get company information using model function directly
	company, _ := models.GetCompanyByID(shareholder.CompanyID)
	
	return c.Render("shareholder_details", fiber.Map{
		"title": shareholder.Name + " - Shareholder Details",
		"shareholder": shareholder,
		"company": company,
	})
}

// GetShareholdersAPI returns shareholders data as JSON for API consumption
func (h *ShareholderHandler) GetShareholdersAPI(c *fiber.Ctx) error {
	// Sample shareholders for demonstration
	shareholders := []models.Shareholder{
		{
			ID: 1,
			CompanyID: 1,
			Name: "John Smith",
			Shares: 1000,
			Percentage: 60.0,
			Type: "Individual",
			Since: time.Now().AddDate(0, -6, 0),
		},
		{
			ID: 2,
			CompanyID: 1,
			Name: "Venture Capital LLC",
			Shares: 667,
			Percentage: 40.0,
			Type: "Corporate",
			Since: time.Now().AddDate(0, -3, 0),
		},
		{
			ID: 3,
			CompanyID: 2,
			Name: "Sarah Johnson",
			Shares: 500,
			Percentage: 50.0,
			Type: "Individual",
			Since: time.Now().AddDate(-1, 0, 0),
		},
	}

	// Filter by company if specified
	companyID := c.Query("company_id")
	if companyID != "" {
		var filtered []models.Shareholder
		for _, s := range shareholders {
			if s.CompanyID == 1 { // This would actually use the companyID param
				filtered = append(filtered, s)
			}
		}
		shareholders = filtered
	}

	return c.JSON(fiber.Map{
		"success": true,
		"shareholders": shareholders,
	})
}
