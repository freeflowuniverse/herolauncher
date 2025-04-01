package webui

import (
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
	// Get data from the store
	companies := h.store.CompanyHandler.GetAll()
	activeCompanies := models.GetActiveCompanies() // Fetch active companies directly
	allShareholders := h.store.ShareholderHandler.GetAll() // Fetch all shareholders directly
	
	return c.Render("index", fiber.Map{
		"title": "Dashboard",
		"companiesCount": len(companies), // Pass count explicitly
		"activeCompaniesCount": len(activeCompanies), // Pass count explicitly
		"shareholdersCount": len(allShareholders), // Pass count explicitly
	})
}

// GetCompanies renders the companies list page
func (h *ModelHandlers) GetCompanies(c *fiber.Ctx) error {
	// Get companies from the store
	companies := h.store.CompanyHandler.GetAll()
	
	return c.Render("companies/index", fiber.Map{
		"title": "Companies",
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
		"title": company.Name,
		"company": company,
	})
}

// GetCompaniesAPI returns companies data as JSON for API consumption
func (h *ModelHandlers) GetCompaniesAPI(c *fiber.Ctx) error {
	companies := h.store.CompanyHandler.GetAll()
	return c.JSON(companies)
}
