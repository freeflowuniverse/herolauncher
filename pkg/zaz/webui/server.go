package webui

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	handlerpkg "github.com/freeflowuniverse/herolauncher/pkg/zaz/webui/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
)

// Server represents the Freezone Manager web UI server
type Server struct {
	app       *fiber.App
	config    Config
	startTime time.Time
	
	// Handlers
	authHandler       AuthHandler
	companyHandler    CompanyHandler
	shareholderHandler ShareholderHandler
	boardMeetingHandler BoardMeetingHandler
	voteHandler      VoteHandler
	saleHandler      SaleHandler
}

// NewServer creates a new instance of the Freezone Manager UI server
func NewServer(config Config) *Server {
	// Initialize template engine
	engine := pug.New(config.TemplatesPath, ".pug")
	engine.Debug(true)  // Enable debug mode
	engine.Reload(true) // Reload templates on each render
	
	// Add template functions
	// User function returns a mock user for demonstration
	engine.AddFunc("user", func() map[string]interface{} {
		return map[string]interface{}{
			"id": 1,
			"name": "Demo User",
			"email": "demo@example.com",
			"role": "Admin",
		}
	})
	
	// Companies function returns a list of sample companies
	engine.AddFunc("companies", func() []map[string]interface{} {
		return []map[string]interface{}{
			{
				"id": 1,
				"name": "TechCorp Inc.",
				"registration_date": "2024-09-28",
				"status": "Active",
				"shareholders_count": 5,
			},
			{
				"id": 2,
				"name": "GreenEnergy Ltd.",
				"registration_date": "2023-03-15",
				"status": "Active",
				"shareholders_count": 3,
			},
		}
	})
	
	// Search function for the template
	engine.AddFunc("search", func() string {
		return ""
	})
	
	// CompaniesCount function for the template
	engine.AddFunc("companiesCount", func() int {
		return 2
	})
	
	// ActiveCompaniesCount function for the template
	engine.AddFunc("activeCompaniesCount", func() int {
		return 2
	})
	
	// ShareholdersCount function for the template
	engine.AddFunc("shareholdersCount", func() int {
		return 7
	})
	
	// Error function for the login template
	engine.AddFunc("error", func() string {
		return ""
	})
	
	// CurrentPath function to help with highlighting active navigation links
	engine.AddFunc("currentPath", func() string {
		return "/" // Default to home path
	})
	
	// Company function returns a sample company for the details page
	engine.AddFunc("company", func() map[string]interface{} {
		return map[string]interface{}{
			"id": 1,
			"name": "TechCorp Inc.",
			"registration_number": "BRN12345678",
			"incorporation_date": "2024-09-28",
			"status": "Active",
			"business_type": "Limited Liability Company",
			"industry": "Technology",
			"email": "info@techcorp.com",
			"phone": "+1 234 567 8900",
			"website": "https://techcorp.com",
			"description": "TechCorp is a leading technology company specializing in AI solutions and cloud services.",
			"shareholders": []map[string]interface{}{
				{
					"id": 1,
					"name": "John Smith",
					"shares": 1000,
					"percentage": 50,
					"since": "2024-09-28",
				},
				{
					"id": 2,
					"name": "Jane Doe",
					"shares": 500,
					"percentage": 25,
					"since": "2024-09-28",
				},
			},
			"boardmeetings": []map[string]interface{}{
				{
					"id": 1,
					"date": "2025-04-15",
					"title": "Quarterly Board Meeting",
					"attendees_count": 5,
					"status": "Scheduled",
				},
				{
					"id": 2,
					"date": "2025-02-28",
					"title": "Annual General Meeting",
					"attendees_count": 8,
					"status": "Completed",
				},
			},
		}
	})
	



	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Handle API errors
			if c.Path() != "/" && c.Path() != "/login" && c.Path() != "/register" {
				if c.Path() == "/api/*" {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": err.Error(),
					})
				}
			}
			
			// Handle view errors
			return c.Status(fiber.StatusInternalServerError).Render("error", fiber.Map{
				"title": "Error",
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Static files
	app.Static("/", config.StaticFilesPath)
	app.Static("/css", filepath.Join(config.StaticFilesPath, "css"))
	app.Static("/js", filepath.Join(config.StaticFilesPath, "js"))
	app.Static("/img", filepath.Join(config.StaticFilesPath, "img"))
	app.Static("/favicon.ico", filepath.Join(config.StaticFilesPath, "favicon.ico"))

	// Initialize handlers
	authHandler := handlerpkg.NewAuthHandler()
	companyHandler := handlerpkg.NewCompanyHandler()
	shareholderHandler := handlerpkg.NewShareholderHandler()
	boardMeetingHandler := handlerpkg.NewBoardMeetingHandler()
	voteHandler := handlerpkg.NewVoteHandler()
	saleHandler := handlerpkg.NewSaleHandler()

	// Create server instance
	srv := &Server{
		app:       app,
		config:    config,
		startTime: time.Now(),
		authHandler: authHandler,
		companyHandler: companyHandler,
		shareholderHandler: shareholderHandler,
		boardMeetingHandler: boardMeetingHandler,
		voteHandler: voteHandler,
		saleHandler: saleHandler,
	}

	// Setup routes
	srv.setupRoutes()

	return srv
}

// setupRoutes initializes and registers all route handlers
func (s *Server) setupRoutes() {
	// Use the handlers initialized in NewServer

	// Register routes
	
	// Auth routes
	s.app.Get("/login", s.authHandler.GetLogin)
	s.app.Post("/login", s.authHandler.PostLogin)
	s.app.Get("/register", s.authHandler.GetRegister)
	s.app.Post("/register", s.authHandler.PostRegister)
	s.app.Get("/logout", s.authHandler.Logout)
	s.app.Get("/forgot-password", s.authHandler.GetForgotPassword)
	s.app.Post("/forgot-password", s.authHandler.PostForgotPassword)

	// Company routes
	s.app.Get("/", s.companyHandler.GetDashboard)
	s.app.Get("/companies", s.companyHandler.GetCompanies)
	s.app.Get("/companies/create", s.companyHandler.GetCreateCompany)
	s.app.Post("/companies/create", s.companyHandler.PostCreateCompany)
	s.app.Get("/companies/:id", s.companyHandler.GetCompanyDetails)
	s.app.Get("/companies/:id/edit", s.companyHandler.GetEditCompany)
	s.app.Post("/companies/:id/edit", s.companyHandler.PostEditCompany)

	// Shareholder routes
	s.app.Get("/shareholders", s.shareholderHandler.GetShareholders)
	s.app.Get("/shareholders/create", s.shareholderHandler.GetCreateShareholder)
	s.app.Post("/shareholders/create", s.shareholderHandler.PostCreateShareholder)
	s.app.Get("/companies/:companyId/shareholders/add", s.shareholderHandler.GetAddShareholder)
	s.app.Post("/companies/:companyId/shareholders/add", s.shareholderHandler.PostAddShareholder)

	// Board Meeting routes
	s.app.Get("/boardmeetings", s.boardMeetingHandler.GetBoardMeetings)
	s.app.Get("/boardmeetings/create", s.boardMeetingHandler.GetCreateBoardMeeting)
	s.app.Post("/boardmeetings/create", s.boardMeetingHandler.PostCreateBoardMeeting)
	s.app.Get("/boardmeetings/:id", s.boardMeetingHandler.GetBoardMeetingDetails)
	s.app.Get("/boardmeetings/:id/minutes", s.boardMeetingHandler.GetBoardMeetingMinutes)
	s.app.Post("/boardmeetings/:id/minutes", s.boardMeetingHandler.PostBoardMeetingMinutes)

	// Vote routes
	s.app.Get("/votes", s.voteHandler.GetVotes)
	s.app.Get("/votes/create", s.voteHandler.GetCreateVote)
	s.app.Post("/votes/create", s.voteHandler.PostCreateVote)
	s.app.Get("/votes/:id", s.voteHandler.GetVoteDetails)
	s.app.Get("/votes/results", s.voteHandler.GetVoteResults)

	// Sales routes
	s.app.Get("/sales", s.saleHandler.GetSales)
	s.app.Get("/sales/products", s.saleHandler.GetProducts)
	s.app.Get("/sales/services", s.saleHandler.GetServices)
	s.app.Get("/sales/create", s.saleHandler.GetCreateSale)
	s.app.Post("/sales/create", s.saleHandler.PostCreateSale)
	s.app.Get("/sales/:id", s.saleHandler.GetSaleDetails)
	s.app.Get("/sales/reports", s.saleHandler.GetSalesReports)

	// API routes
	api := s.app.Group("/api")
	api.Get("/companies", s.companyHandler.GetCompaniesAPI)
	api.Get("/companies/:id", s.companyHandler.GetCompanyDetailsAPI)
	api.Get("/shareholders", s.shareholderHandler.GetShareholdersAPI)
	api.Get("/boardmeetings", s.boardMeetingHandler.GetBoardMeetingsAPI)
	api.Get("/votes", s.voteHandler.GetVotesAPI)
	api.Get("/sales", s.saleHandler.GetSalesAPI)
}

// GetUptime returns the uptime of the Freezone Manager UI server
func (s *Server) GetUptime() string {
	uptimeDuration := time.Since(s.startTime)
	
	totalSeconds := int(uptimeDuration.Seconds())
	days := totalSeconds / (24 * 3600)
	hours := (totalSeconds % (24 * 3600)) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}

// Start starts the Freezone Manager UI server
func (s *Server) Start() error {
	log.Printf("Starting Freezone Manager UI server on port %s", s.config.Port)
	return s.app.Listen(":" + s.config.Port)
}
