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
	"github.com/gofiber/template/jet/v2"
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

	// Initialize Jet template engine
	engine := jet.New(config.TemplatesPath, ".jet")
	engine.Reload(true) // Optional: enable reload for development
	engine.Debug(true)   // Optional: enable debug mode for development

	// Initialize handlers
	authHandler := webhandlers.NewAuthHandler(store)
	companyHandler := webhandlers.NewCompanyHandler(store)
	shareholderHandler := webhandlers.NewShareholderHandler(store)
	boardMeetingHandler := webhandlers.NewBoardMeetingHandler(store)
	voteHandler := webhandlers.NewVoteHandler(store)

	// Create Fiber app with Jet engine
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Log the error
			log.Printf("Error: %v", err)

			// Send custom error page
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			// Attempt to render an error template
			// Ensure you have an error.jet template
			renderErr := c.Status(code).Render("error", fiber.Map{
				"ErrorCode": code,
				"ErrorMsg":  err.Error(),
			})
			if renderErr != nil {
				// Fallback if error template fails
				return c.Status(code).SendString(fmt.Sprintf("%d: %s", code, err.Error()))
			}
			return nil
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

// RunServer starts the Freezone Manager UI server
func (s *Server) RunServer() {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("Starting Freezone Manager UI server on %s", addr)
	
	// Record start time
	s.startTime = time.Now()

	// Start the server
	log.Fatal(s.app.Listen(addr))
}
