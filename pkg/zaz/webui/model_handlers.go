package webui

import (
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// ModelHandlers provides HTTP handlers that use the models package
type ModelHandlers struct {
	store *models.Store
}

// NewModelHandlers creates a new ModelHandlers instance
func NewModelHandlers(store *models.Store) *ModelHandlers {
	return &ModelHandlers{
		store: store,
	}
}

// GetDashboard renders the dashboard page
func (h *ModelHandlers) GetDashboard(c *fiber.Ctx) error {
	fmt.Println("ModelHandlers.GetDashboard method called")
	
	// Get data from the store
	companies := h.store.CompanyHandler.GetAll()
	shareholders := h.store.ShareholderHandler.GetAll()
	
	// Count active companies
	activeCompanies := 0
	for _, company := range companies {
		if company.Status == "Active" {
			activeCompanies++
		}
	}

	// Create a simple array for upcoming meetings with properly formatted dates
	upcomingMeetings := []fiber.Map{}
	
	// Add a dummy meeting for testing
	upcomingMeetings = append(upcomingMeetings, fiber.Map{
		"ID":    int64(1),
		"Date":  time.Now().Format("2006-01-02 15:04"),
		"Title": "Test Meeting",
		"Company": fiber.Map{
			"Name": "Test Company",
		},
	})
	
	fmt.Printf("Debug - upcomingMeetings structure: %#v\n", upcomingMeetings)
	
	// Create the data map
	data := fiber.Map{
		"title":                "Dashboard",
		"companiesCount":       len(companies),
		"activeCompaniesCount": activeCompanies,
		"shareholdersCount":    len(shareholders),
		"upcomingMeetings":     upcomingMeetings,
		"recentActivities":     nil,
	}
	
	fmt.Printf("Debug - Full data map being passed to template: %#v\n", data)
	
	// Use direct rendering for now to debug
	return c.Render("index", data)
}

// GetCompanies renders the companies list page
func (h *ModelHandlers) GetCompanies(c *fiber.Ctx) error {
	// Get companies from the store
	companies := h.store.CompanyHandler.GetAll()

	return c.Render("companies/index", fiber.Map{
		"title":     "Companies",
		"companies": companies,
	})
}

// GetCompanyDetails renders the company details page
func (h *ModelHandlers) GetCompanyDetails(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}

	company, err := h.store.CompanyHandler.GetByID(int64(id))
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}

	return c.Render("companies/details", fiber.Map{
		"title":   company.Name,
		"company": company,
	})
}

// GetCompaniesAPI returns companies data as JSON for API consumption
func (h *ModelHandlers) GetCompaniesAPI(c *fiber.Ctx) error {
	companies := h.store.CompanyHandler.GetAll()
	return c.JSON(companies)
}
