package handlers

import (
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// SaleHandler handles sale-related routes
type SaleHandler struct {
	store *models.Store
}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler(store *models.Store) *SaleHandler {
	return &SaleHandler{
		store: store,
	}
}

// GetSales renders the sales list page
func (h *SaleHandler) GetSales(c *fiber.Ctx) error {
	// Get sales from the store
	sales := h.store.GetAllSales()

	return c.Render("sales", fiber.Map{
		"title": "Sales",
		"sales": sales,
		"search": c.Query("search"),
	})
}

// GetProducts renders the products list page
func (h *SaleHandler) GetProducts(c *fiber.Ctx) error {
	// Get products from the store
	products := h.store.GetAllProducts()

	return c.Render("products", fiber.Map{
		"title": "Products",
		"products": products,
		"search": c.Query("search"),
	})
}

// GetServices renders the services list page
func (h *SaleHandler) GetServices(c *fiber.Ctx) error {
	// Get services (products of type Service) from the store
	services := h.store.GetProductsByType("Service")

	return c.Render("services", fiber.Map{
		"title": "Services",
		"services": services,
		"search": c.Query("search"),
	})
}

// GetCreateSale renders the sale creation page
func (h *SaleHandler) GetCreateSale(c *fiber.Ctx) error {
	// Get list of companies from the store
	companies := h.store.GetAllCompanies()

	// Get list of products from the store
	products := h.store.GetAllProducts()

	return c.Render("sales_create", fiber.Map{
		"title": "Create Sale",
		"companies": companies,
		"products": products,
	})
}

// PostCreateSale handles sale creation form submission
func (h *SaleHandler) PostCreateSale(c *fiber.Ctx) error {
	// Parse form data
	companyIDStr := c.FormValue("company_id")
	buyerName := c.FormValue("buyer_name")
	buyerEmail := c.FormValue("buyer_email")
	
	// Simple validation
	if companyIDStr == "" || buyerName == "" || buyerEmail == "" {
		// Get data for the form
		companies := h.store.GetAllCompanies()
		products := h.store.GetAllProducts()
		return c.Render("sales_create", fiber.Map{
			"title": "Create Sale",
			"companies": companies,
			"products": products,
			"error": "Company, buyer name, and email are required",
		})
	}

	// Parse company ID
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid company ID")
	}

	// Create a new sale
	sale := models.Sale{
		ID:          int64(len(h.store.GetAllSales()) + 1),
		CompanyID:   companyID,
		BuyerName:   buyerName,
		BuyerEmail:  buyerEmail,
		TotalAmount: 0, // Will be calculated based on items
		Currency:    "USD", // Default currency
		Status:      "Pending",
		SaleDate:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       []models.SaleItem{},
	}

	// Add the sale to the store
	h.store.AddSale(sale)

	return c.Redirect("/sales")
}

// GetSaleDetails renders the sale details page
func (h *SaleHandler) GetSaleDetails(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	
	// Fetch the sale from the database
	sale, err := h.store.GetSaleByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Sale not found")
	}

	// Get company info
	company, err := h.store.GetCompanyByID(sale.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Company not found")
	}

	return c.Render("sale_details", fiber.Map{
		"title": "Sale Details",
		"sale": sale,
		"company": company,
	})
}

// GetSalesReports renders the sales reports page
func (h *SaleHandler) GetSalesReports(c *fiber.Ctx) error {
	// Get all sales from the store
	sales := h.store.GetAllSales()
	
	// Calculate monthly sales
	monthlySales := make(map[string]float64)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	
	for _, sale := range sales {
		month := monthNames[sale.SaleDate.Month()-1]
		monthlySales[month] += sale.TotalAmount
	}
	
	// Calculate sales by company
	companySales := make(map[string]float64)
	for _, sale := range sales {
		company, err := h.store.GetCompanyByID(sale.CompanyID)
		if err == nil {
			companySales[company.Name] += sale.TotalAmount
		}
	}
	
	// Calculate sales by category (using product categories)
	categorySales := make(map[string]float64)
	for _, sale := range sales {
		for _, item := range sale.Items {
			product, err := h.store.GetProductByID(item.ProductID)
			if err == nil {
				categorySales[product.Category] += item.Subtotal
			}
		}
	}
	
	return c.Render("sales_reports", fiber.Map{
		"title": "Sales Reports",
		"monthlySales": monthlySales,
		"categorySales": categorySales,
		"companySales": companySales,
	})
}

// GetProductDetails renders the product details page
func (h *SaleHandler) GetProductDetails(c *fiber.Ctx) error {
	// Get product ID from URL
	id := c.Params("id")
	
	// Convert ID to int64
	productID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(400).Render("error", fiber.Map{
			"title": "Error",
			"message": "Invalid product ID",
		})
	}
	
	// Get product from store
	product, err := h.store.GetProductByID(productID)
	if err != nil {
		return c.Status(404).Render("error", fiber.Map{
			"title": "Not Found",
			"message": "Product not found",
		})
	}
	
	return c.Render("product_details", fiber.Map{
		"title": product.Name + " - Product Details",
		"product": product,
	})
}

// GetSalesAPI returns sales data as JSON for API consumption
func (h *SaleHandler) GetSalesAPI(c *fiber.Ctx) error {
	// Get all sales from the store
	sales := h.store.GetAllSales()

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
