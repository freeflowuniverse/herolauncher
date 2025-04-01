package webhandlers

import (
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// ShareholderHandler handles shareholder-related routes
type ShareholderHandler struct {
	store *models.Store
}

// NewShareholderHandler creates a new ShareholderHandler
func NewShareholderHandler(store *models.Store) *ShareholderHandler {
	return &ShareholderHandler{
		store: store,
	}
}

// GetShareholders renders the shareholders list page
func (h *ShareholderHandler) GetShareholders(c *fiber.Ctx) error {
	// Get shareholders directly from the store
	shareholders := h.store.ShareholderHandler.GetAll()

	return RenderWithDefaults(c, "shareholders", fiber.Map{
		"title": "Shareholders",
		"shareholders": shareholders,
		"search": c.Query("search"),
		"currentYear": time.Now().Year(),
	})
}

// GetCreateShareholder renders the shareholder creation page
func (h *ShareholderHandler) GetCreateShareholder(c *fiber.Ctx) error {
	// Get list of companies directly from the store
	companies := h.store.CompanyHandler.GetAll()

	return RenderWithDefaults(c, "shareholders_create", fiber.Map{
		"title": "Create Shareholder",
		"companies": companies,
		"formErrors": []string{}, // Empty errors for initial load
		"csrfToken": "sample-token", // This would normally be generated
		"form": fiber.Map{}, // Empty form for initial load
		"currentYear": time.Now().Year(),
	})
}

// PostCreateShareholder handles shareholder creation form submission
func (h *ShareholderHandler) PostCreateShareholder(c *fiber.Ctx) error {
	// Parse form data
	name := c.FormValue("name")
	companyIDStr := c.FormValue("company_id")
	sharesStr := c.FormValue("shares")
	percentageStr := c.FormValue("percentage")
	
	// Simple validation
	if name == "" || companyIDStr == "" {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_create", fiber.Map{
			"title": "Create Shareholder",
			"companies": companies,
			"error": "Name and company are required",
		})
	}
	
	// Parse numeric values
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_create", fiber.Map{
			"title": "Create Shareholder",
			"companies": companies,
			"error": "Invalid company ID",
		})
	}
	
	shares := 0
	if sharesStr != "" {
		shares, _ = strconv.Atoi(sharesStr)
	}
	
	percentage := 0.0
	if percentageStr != "" {
		percentage, _ = strconv.ParseFloat(percentageStr, 64)
	}
	
	// Create shareholder
	shareholder := models.Shareholder{
		CompanyID:  companyID,
		Name:       name,
		Shares:     shares,
		Percentage: percentage,
		Type:       c.FormValue("type"),
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Save to database using the store
	_, err = h.store.ShareholderHandler.Create(shareholder)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_create", fiber.Map{
			"title": "Create Shareholder",
			"companies": companies,
			"error": "Failed to create shareholder: " + err.Error(),
		})
	}
	
	// Redirect to shareholders list
	return c.Redirect("/shareholders")
}

// GetAddShareholder renders the add shareholder to company page
func (h *ShareholderHandler) GetAddShareholder(c *fiber.Ctx) error {
	// Parse company ID from URL
	companyID, err := strconv.ParseInt(c.Params("companyId"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(companyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	return RenderWithDefaults(c, "shareholders_add", fiber.Map{
		"title": "Add Shareholder to " + company.Name,
		"company": company,
	})
}

// PostAddShareholder handles adding shareholder to company form submission
func (h *ShareholderHandler) PostAddShareholder(c *fiber.Ctx) error {
	// Parse company ID from URL
	companyID, err := strconv.ParseInt(c.Params("companyId"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid company ID")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(companyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Parse form data
	name := c.FormValue("name")
	sharesStr := c.FormValue("shares")
	percentageStr := c.FormValue("percentage")
	
	// Simple validation
	if name == "" {
		return RenderWithDefaults(c, "shareholders_add", fiber.Map{
			"title": "Add Shareholder to " + company.Name,
			"company": company,
			"error": "Name is required",
		})
	}
	
	// Parse numeric values
	shares := 0
	if sharesStr != "" {
		shares, _ = strconv.Atoi(sharesStr)
	}
	
	percentage := 0.0
	if percentageStr != "" {
		percentage, _ = strconv.ParseFloat(percentageStr, 64)
	}
	
	// Create shareholder
	shareholder := models.Shareholder{
		CompanyID:  companyID,
		Name:       name,
		Shares:     shares,
		Percentage: percentage,
		Type:       c.FormValue("type"),
		Since:      time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Save to database using the store
	_, err = h.store.ShareholderHandler.Create(shareholder)
	if err != nil {
		return RenderWithDefaults(c, "shareholders_add", fiber.Map{
			"title": "Add Shareholder to " + company.Name,
			"company": company,
			"error": "Failed to add shareholder: " + err.Error(),
		})
	}
	
	// Redirect to company details
	return c.Redirect("/companies/" + c.Params("companyId"))
}

// GetShareholderDetails renders the shareholder details page
func (h *ShareholderHandler) GetShareholderDetails(c *fiber.Ctx) error {
	// Parse shareholder ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid shareholder ID")
	}
	
	// Get shareholder from store
	shareholder, err := h.store.ShareholderHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Shareholder not found")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(shareholder.CompanyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// For now, we'll create sample activities and documents
	// In a real application, these would come from the database
	activities := []map[string]interface{}{
		{
			"Date":    time.Now().AddDate(0, -1, 0),
			"Action":  "Share Purchase",
			"Details": "Purchased 100 additional shares",
		},
		{
			"Date":    time.Now().AddDate(0, -3, 0),
			"Action":  "Dividend Payment",
			"Details": "Received quarterly dividend",
		},
	}
	
	documents := []map[string]interface{}{
		{
			"Name": "Share Certificate",
			"URL":  "#",
			"Date": time.Now().AddDate(0, -6, 0),
		},
		{
			"Name": "Shareholder Agreement",
			"URL":  "#",
			"Date": time.Now().AddDate(-1, 0, 0),
		},
	}
	
	return RenderWithDefaults(c, "shareholders_details", fiber.Map{
		"title":       shareholder.Name,
		"shareholder":  shareholder,
		"company":     company,
		"activities":  activities,
		"documents":   documents,
	})
}

// GetEditShareholder renders the shareholder edit page
func (h *ShareholderHandler) GetEditShareholder(c *fiber.Ctx) error {
	// Parse shareholder ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid shareholder ID")
	}
	
	// Get shareholder from store
	shareholder, err := h.store.ShareholderHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Shareholder not found")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(shareholder.CompanyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Get all companies for the dropdown
	companies := h.store.CompanyHandler.GetAll()
	
	return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
		"title": "Edit " + shareholder.Name,
		"shareholder": shareholder,
		"company": company,
		"companies": companies,
	})
}

// PostEditShareholder handles shareholder edit form submission
func (h *ShareholderHandler) PostEditShareholder(c *fiber.Ctx) error {
	// Parse shareholder ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid shareholder ID")
	}
	
	// Get existing shareholder from store
	shareholder, err := h.store.ShareholderHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Shareholder not found")
	}
	
	// Get company for error handling
	company, err := h.store.CompanyHandler.GetByID(shareholder.CompanyID)
	if err != nil {
		company = models.Company{} // Empty company if not found
	}
	
	// Parse form data
	name := c.FormValue("name")
	companyIDStr := c.FormValue("company_id")
	sharesStr := c.FormValue("shares")
	percentageStr := c.FormValue("percentage")
	shareholderType := c.FormValue("type")
	sinceStr := c.FormValue("since")
	
	// Simple validation
	if name == "" {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
			"title": "Edit " + shareholder.Name,
			"shareholder": shareholder,
			"company": company,
			"companies": companies,
			"error": "Name is required",
		})
	}
	
	// Parse company ID
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
			"title": "Edit " + shareholder.Name,
			"shareholder": shareholder,
			"company": company,
			"companies": companies,
			"error": "Invalid company ID",
		})
	}
	
	// Parse shares
	shares := 0
	if sharesStr != "" {
		shares, err = strconv.Atoi(sharesStr)
		if err != nil {
			companies := h.store.CompanyHandler.GetAll()
			return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
				"title": "Edit " + shareholder.Name,
				"shareholder": shareholder,
				"company": company,
				"companies": companies,
				"error": "Invalid shares value",
			})
		}
	}
	
	// Parse percentage
	percentage := 0.0
	if percentageStr != "" {
		percentage, err = strconv.ParseFloat(percentageStr, 64)
		if err != nil {
			companies := h.store.CompanyHandler.GetAll()
			return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
				"title": "Edit " + shareholder.Name,
				"shareholder": shareholder,
				"company": company,
				"companies": companies,
				"error": "Invalid percentage value",
			})
		}
	}
	
	// Parse since date
	var since time.Time
	if sinceStr != "" {
		since, err = time.Parse("2006-01-02", sinceStr)
		if err != nil {
			companies := h.store.CompanyHandler.GetAll()
			return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
				"title": "Edit " + shareholder.Name,
				"shareholder": shareholder,
				"company": company,
				"companies": companies,
				"error": "Invalid date format. Please use YYYY-MM-DD",
			})
		}
	} else {
		// Keep the existing date if not provided
		since = shareholder.Since
	}
	
	// Update shareholder fields
	shareholder.Name = name
	shareholder.CompanyID = companyID
	shareholder.Shares = shares
	shareholder.Percentage = percentage
	shareholder.Type = shareholderType
	shareholder.Since = since
	shareholder.UpdatedAt = time.Now()
	
	// Save to database using the store
	err = h.store.ShareholderHandler.Update(shareholder)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "shareholders_edit", fiber.Map{
			"title": "Edit " + shareholder.Name,
			"shareholder": shareholder,
			"company": company,
			"companies": companies,
			"error": "Failed to update shareholder: " + err.Error(),
		})
	}
	
	// Redirect to shareholder details
	return c.Redirect("/shareholders/" + c.Params("id"))
}

// GetShareholdersAPI returns shareholders data as JSON for API consumption
func (h *ShareholderHandler) GetShareholdersAPI(c *fiber.Ctx) error {
	// Check if company ID is provided as query parameter
	companyIDStr := c.Query("company_id")
	if companyIDStr != "" {
		companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid company ID",
			})
		}
		
		// Get shareholders for specific company
		shareholders := h.store.ShareholderHandler.GetByCompanyID(companyID)
		return c.JSON(shareholders)
	}
	
	// Get all shareholders
	shareholders := h.store.ShareholderHandler.GetAll()
	return c.JSON(shareholders)
}
