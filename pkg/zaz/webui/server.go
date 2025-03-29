package webui

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/freeflowuniverse/herolauncher/pkg/zaz/webhandlers"
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
	
	// Store for database operations
	store     *models.Store
	
	// Handlers
	authHandler        *webhandlers.AuthHandler
	companyHandler     *webhandlers.CompanyHandler
	shareholderHandler *webhandlers.ShareholderHandler
	boardMeetingHandler *webhandlers.BoardMeetingHandler
	voteHandler        *webhandlers.VoteHandler
}

// NewServer creates a new instance of the Freezone Manager UI server
func NewServer(config Config) *Server {
	// Initialize the database and generate fake data
	dbPath := config.DatabasePath
	_, err := models.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Create store
	store := models.NewStore("default")
	
	// Generate fake data
	models.GenerateFakeData()
	log.Printf("Database initialized with fake data")

	// Initialize template engine
	engine := pug.New(config.TemplatesPath, ".pug")
	engine.Debug(true)  // Enable debug mode
	engine.Reload(true) // Reload templates on each render
	
	// Initialize handlers
	authHandler := webhandlers.NewAuthHandler(store)
	companyHandler := webhandlers.NewCompanyHandler(store)
	shareholderHandler := webhandlers.NewShareholderHandler(store)
	boardMeetingHandler := webhandlers.NewBoardMeetingHandler(store)
	voteHandler := webhandlers.NewVoteHandler(store)
	
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
		// Get companies using model function directly
		companies := models.GetAllCompanies()
		result := make([]map[string]interface{}, len(companies))
		
		// Log the number of companies being passed to the template
		log.Printf("Passing %d companies to the template", len(companies))
		
		for i, company := range companies {
			result[i] = map[string]interface{}{
				"ID": company.ID,
				"Name": company.Name,
				"IncorporationDate": company.IncorporationDate,
				"Status": company.Status,
				"Industry": company.Industry,
				"Shareholders": company.Shareholders,
			}
		}
		
		return result
	})
	
	// Search function for the template
	engine.AddFunc("search", func() string {
		return ""
	})
	
	// Sum function to sum values in a map
	engine.AddFunc("sum", func(values map[string]float64) float64 {
		var total float64
		for _, v := range values {
			total += v
		}
		return total
	})
	
	// Div function to divide two numbers
	engine.AddFunc("div", func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	})
	
	// TopCompany function to get the company with highest sales
	engine.AddFunc("topCompany", func(companySales map[string]float64) string {
		var topCompany string
		var maxSales float64
		for company, sales := range companySales {
			if sales > maxSales {
				maxSales = sales
				topCompany = company
			}
		}
		return topCompany
	})
	
	// TopCompanyAmount function to get the amount of the top company
	engine.AddFunc("topCompanyAmount", func(companySales map[string]float64) float64 {
		var maxSales float64
		for _, sales := range companySales {
			if sales > maxSales {
				maxSales = sales
			}
		}
		return maxSales
	})
	
	// CompaniesCount function for the template
	engine.AddFunc("companiesCount", func() int {
		return len(models.GetAllCompanies())
	})
	
	// ActiveCompaniesCount function for the template
	engine.AddFunc("activeCompaniesCount", func() int {
		return len(models.GetActiveCompanies())
	})
	
	// ShareholdersCount function for the template
	engine.AddFunc("shareholdersCount", func() int {
		return len(models.GetAllShareholders())
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
		companies := models.GetAllCompanies()
		if len(companies) == 0 {
			return map[string]interface{}{}
		}
		
		// Use the first company in the store
		company := companies[0]
		
		// Format shareholders for template
		shareholders := make([]map[string]interface{}, len(company.Shareholders))
		for i, sh := range company.Shareholders {
			shareholders[i] = map[string]interface{}{
				"id": sh.ID,
				"name": sh.Name,
				"shares": sh.Shares,
				"percentage": sh.Percentage,
				"since": sh.Since.Format("2006-01-02"),
			}
		}
		
		// Format board meetings for template
		meetings := make([]map[string]interface{}, len(company.BoardMeetings))
		for i, bm := range company.BoardMeetings {
			meetings[i] = map[string]interface{}{
				"id": bm.ID,
				"date": bm.Date.Format("2006-01-02"),
				"title": bm.Title,
				"attendees_count": len(bm.Attendees),
				"status": bm.Status,
			}
		}
		
		// Return formatted company data
		return map[string]interface{}{
			"id": company.ID,
			"name": company.Name,
			"registration_number": company.RegistrationNumber,
			"incorporation_date": company.IncorporationDate.Format("2006-01-02"),
			"status": company.Status,
			"business_type": company.BusinessType,
			"industry": company.Industry,
			"email": company.Email,
			"phone": company.Phone,
			"website": company.Website,
			"description": company.Description,
			"shareholders": shareholders,
			"boardmeetings": meetings,
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

	// Note: we already initialized the store earlier

	// Create server instance
	srv := &Server{
		app:                app,
		config:             config,
		startTime:          time.Now(),
		store:              store,
		authHandler:        authHandler,
		companyHandler:     companyHandler,
		shareholderHandler: shareholderHandler,
		boardMeetingHandler: boardMeetingHandler,
		voteHandler:        voteHandler,
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
	s.app.Get("/shareholders/:id", s.shareholderHandler.GetShareholderDetails)
	s.app.Get("/shareholders/:id/edit", s.shareholderHandler.GetEditShareholder)
	s.app.Post("/shareholders/:id/edit", s.shareholderHandler.PostEditShareholder)
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



	// API routes
	api := s.app.Group("/api")
	api.Get("/companies", s.companyHandler.GetCompaniesAPI)
	api.Get("/companies/:id", s.companyHandler.GetCompanyDetailsAPI)
	api.Get("/shareholders", s.shareholderHandler.GetShareholdersAPI)
	api.Get("/boardmeetings", s.boardMeetingHandler.GetBoardMeetingsAPI)
	api.Get("/votes", s.voteHandler.GetVotesAPI)
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
