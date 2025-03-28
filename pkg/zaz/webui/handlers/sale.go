package handlers

import (
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// SaleHandler handles sale-related routes
type SaleHandler struct {}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler() *SaleHandler {
	return &SaleHandler{}
}

// GetSales renders the sales list page
func (h *SaleHandler) GetSales(c *fiber.Ctx) error {
	// Sample sales for demonstration
	sales := []models.Sale{
		{
			ID: 1,
			CompanyID: 1,
			BuyerName: "Acme Corp",
			BuyerEmail: "purchasing@acmecorp.com",
			TotalAmount: 2500.00,
			Currency: "USD",
			Status: "Completed",
			SaleDate: time.Now().AddDate(0, 0, -5),
		},
		{
			ID: 2,
			CompanyID: 1,
			BuyerName: "Global Enterprises",
			BuyerEmail: "finance@globalent.com",
			TotalAmount: 1800.50,
			Currency: "USD",
			Status: "Pending",
			SaleDate: time.Now().AddDate(0, 0, -2),
		},
		{
			ID: 3,
			CompanyID: 2,
			BuyerName: "Tech Solutions Ltd",
			BuyerEmail: "orders@techsolutions.com",
			TotalAmount: 3200.75,
			Currency: "USD",
			Status: "Completed",
			SaleDate: time.Now().AddDate(0, 0, -10),
		},
	}

	return c.Render("sales", fiber.Map{
		"title": "Sales",
		"sales": sales,
		"search": c.Query("search"),
	})
}

// GetProducts renders the products list page
func (h *SaleHandler) GetProducts(c *fiber.Ctx) error {
	// Sample products for demonstration
	products := []models.Product{
		{
			ID: 1,
			Name: "Business Registration",
			Description: "Complete business registration service in the freezone",
			Price: 1500.00,
			Currency: "USD",
			Type: "Service",
			Category: "Registration",
			Status: "Available",
		},
		{
			ID: 2,
			Name: "Office Space (Small)",
			Description: "Small office space (25 sq.m) in the freezone",
			Price: 500.00,
			Currency: "USD",
			Type: "Service",
			Category: "Real Estate",
			Status: "Available",
		},
		{
			ID: 3,
			Name: "Legal Consultation",
			Description: "One hour of legal consultation with our experts",
			Price: 200.00,
			Currency: "USD",
			Type: "Service",
			Category: "Legal",
			Status: "Available",
		},
		{
			ID: 4,
			Name: "Company Seal",
			Description: "Official company seal for documentation",
			Price: 75.00,
			Currency: "USD",
			Type: "Product",
			Category: "Office Supplies",
			Status: "Available",
		},
	}

	return c.Render("products", fiber.Map{
		"title": "Products",
		"products": products,
		"search": c.Query("search"),
	})
}

// GetServices renders the services list page
func (h *SaleHandler) GetServices(c *fiber.Ctx) error {
	// Filter only service types
	services := []models.Product{
		{
			ID: 1,
			Name: "Business Registration",
			Description: "Complete business registration service in the freezone",
			Price: 1500.00,
			Currency: "USD",
			Type: "Service",
			Category: "Registration",
			Status: "Available",
		},
		{
			ID: 2,
			Name: "Office Space (Small)",
			Description: "Small office space (25 sq.m) in the freezone",
			Price: 500.00,
			Currency: "USD",
			Type: "Service",
			Category: "Real Estate",
			Status: "Available",
		},
		{
			ID: 3,
			Name: "Legal Consultation",
			Description: "One hour of legal consultation with our experts",
			Price: 200.00,
			Currency: "USD",
			Type: "Service",
			Category: "Legal",
			Status: "Available",
		},
		{
			ID: 5,
			Name: "Annual Audit",
			Description: "Mandatory annual audit service",
			Price: 800.00,
			Currency: "USD",
			Type: "Service",
			Category: "Compliance",
			Status: "Available",
		},
	}

	return c.Render("services", fiber.Map{
		"title": "Services",
		"services": services,
		"search": c.Query("search"),
	})
}

// GetCreateSale renders the sale creation page
func (h *SaleHandler) GetCreateSale(c *fiber.Ctx) error {
	// Get list of companies for dropdown
	companies := []models.Company{
		{
			ID: 1,
			Name: "TechCorp Inc.",
		},
		{
			ID: 2,
			Name: "GreenEnergy Ltd.",
		},
		{
			ID: 3,
			Name: "InnoFinance Corp.",
		},
	}

	// Get list of products
	products := []models.Product{
		{
			ID: 1,
			Name: "Business Registration",
			Price: 1500.00,
			Currency: "USD",
		},
		{
			ID: 2,
			Name: "Office Space (Small)",
			Price: 500.00,
			Currency: "USD",
		},
		{
			ID: 3,
			Name: "Legal Consultation",
			Price: 200.00,
			Currency: "USD",
		},
		{
			ID: 4,
			Name: "Company Seal",
			Price: 75.00,
			Currency: "USD",
		},
	}

	return c.Render("sales_create", fiber.Map{
		"title": "Create Sale",
		"companies": companies,
		"products": products,
	})
}

// PostCreateSale handles sale creation form submission
func (h *SaleHandler) PostCreateSale(c *fiber.Ctx) error {
	// Parse form data
	companyID := c.FormValue("company_id")
	buyerName := c.FormValue("buyer_name")
	_ = c.FormValue("buyer_email") // Use the variable to avoid unused variable warning
	
	// Simple validation
	if companyID == "" || buyerName == "" {
		return c.Render("sales_create", fiber.Map{
			"title": "Create Sale",
			"error": "Company and buyer name are required",
		})
	}

	// TODO: Implement actual sale creation
	// For now, just redirect to sales list
	return c.Redirect("/sales")
}

// GetSaleDetails renders the sale details page
func (h *SaleHandler) GetSaleDetails(c *fiber.Ctx) error {
	_ = c.Params("id") // Use the parameter to avoid unused variable warning
	
	// In a real implementation, we would fetch the sale from the database
	sale := models.Sale{
		ID: 1,
		CompanyID: 1,
		BuyerName: "Acme Corp",
		BuyerEmail: "purchasing@acmecorp.com",
		TotalAmount: 2500.00,
		Currency: "USD",
		Status: "Completed",
		SaleDate: time.Now().AddDate(0, 0, -5),
		Items: []models.SaleItem{
			{
				ID: 1,
				SaleID: 1,
				ProductID: 1,
				Name: "Business Registration",
				Quantity: 1,
				UnitPrice: 1500.00,
				Currency: "USD",
				Subtotal: 1500.00,
			},
			{
				ID: 2,
				SaleID: 1,
				ProductID: 2,
				Name: "Office Space (Small)",
				Quantity: 2,
				UnitPrice: 500.00,
				Currency: "USD",
				Subtotal: 1000.00,
			},
		},
	}

	// Get company info
	company := models.Company{
		ID: 1,
		Name: "TechCorp Inc.",
	}

	return c.Render("sale_details", fiber.Map{
		"title": "Sale Details",
		"sale": sale,
		"company": company,
	})
}

// GetSalesReports renders the sales reports page
func (h *SaleHandler) GetSalesReports(c *fiber.Ctx) error {
	// Sample sales data for reports
	monthlySales := map[string]float64{
		"Jan": 5200.00,
		"Feb": 6100.00,
		"Mar": 8500.00,
		"Apr": 9200.00,
		"May": 7800.00,
		"Jun": 10500.00,
	}

	categorySales := map[string]float64{
		"Registration": 15000.00,
		"Real Estate": 25000.00,
		"Legal": 8000.00,
		"Office Supplies": 2000.00,
		"Compliance": 12000.00,
	}

	companySales := map[string]float64{
		"TechCorp Inc.": 22000.00,
		"GreenEnergy Ltd.": 18000.00,
		"InnoFinance Corp.": 12000.00,
	}

	return c.Render("sales_reports", fiber.Map{
		"title": "Sales Reports",
		"monthlySales": monthlySales,
		"categorySales": categorySales,
		"companySales": companySales,
	})
}

// GetSalesAPI returns sales data as JSON for API consumption
func (h *SaleHandler) GetSalesAPI(c *fiber.Ctx) error {
	// Sample sales for demonstration
	sales := []models.Sale{
		{
			ID: 1,
			CompanyID: 1,
			BuyerName: "Acme Corp",
			BuyerEmail: "purchasing@acmecorp.com",
			TotalAmount: 2500.00,
			Currency: "USD",
			Status: "Completed",
			SaleDate: time.Now().AddDate(0, 0, -5),
		},
		{
			ID: 2,
			CompanyID: 1,
			BuyerName: "Global Enterprises",
			BuyerEmail: "finance@globalent.com",
			TotalAmount: 1800.50,
			Currency: "USD",
			Status: "Pending",
			SaleDate: time.Now().AddDate(0, 0, -2),
		},
		{
			ID: 3,
			CompanyID: 2,
			BuyerName: "Tech Solutions Ltd",
			BuyerEmail: "orders@techsolutions.com",
			TotalAmount: 3200.75,
			Currency: "USD",
			Status: "Completed",
			SaleDate: time.Now().AddDate(0, 0, -10),
		},
	}

	// Filter by company if specified
	companyID := c.Query("company_id")
	if companyID != "" {
		var filtered []models.Sale
		for _, s := range sales {
			if s.CompanyID == 1 { // This would actually use the companyID param
				filtered = append(filtered, s)
			}
		}
		sales = filtered
	}

	// Filter by status if specified
	status := c.Query("status")
	if status != "" {
		var filtered []models.Sale
		for _, s := range sales {
			if s.Status == status {
				filtered = append(filtered, s)
			}
		}
		sales = filtered
	}

	return c.JSON(fiber.Map{
		"success": true,
		"sales": sales,
	})
}
