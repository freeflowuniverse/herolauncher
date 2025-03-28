package book

import (
	"html/template"
	"log"
	"path/filepath"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
)

// WikiServer represents the book server with its configuration and VFS backend
type WikiServer struct {
	App            *fiber.App
	Config         Configuration
	VFS            vfs.VFSImplementation
	ContentPath    string
	AbsContentPath string
}

// NewWikiServer creates a new book server with the given configuration and VFS backend
func NewWikiServer(contentPath string, configPath string, vfsBackend vfs.VFSImplementation) (*WikiServer, error) {
	// Initialize template engine
	templatesPath, _ := filepath.Abs("./web/templates")
	engine := pug.New(templatesPath, ".pug")

	// Add template functions
	engine.AddFunc("title", func(s string) string {
		return strings.Title(s)
	})

	// Add function to render unescaped HTML
	engine.AddFunc("unescaped", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Static files
	app.Static("/", "./web/static")
	app.Static("/css", "./web/static/css")
	app.Static("/js", "./web/static/js")
	app.Static("/content", contentPath)

	// Load configuration from file if provided, otherwise use empty configuration
	var config Configuration
	var err error

	if configPath != "" {
		config, err = loadConfiguration(configPath)
		if err != nil {
			log.Printf("Error loading configuration: %v", err)
			// If config file loading fails, use empty sidebar
			config = Configuration{
				Sidebar: []SidebarSection{},
				Title:   "Book",
			}
			log.Printf("Warning: Using empty sidebar due to configuration loading error")
		}
	} else {
		// No configuration file provided, use empty sidebar
		config = Configuration{
			Sidebar: []SidebarSection{},
			Title:   "Book",
		}
		log.Printf("Warning: No configuration file provided. Please generate a configuration file with sidebar data.")
	}

	// Create the book server
	book := &WikiServer{
		App:            app,
		Config:         config,
		VFS:            vfsBackend,
		ContentPath:    contentPath,
		AbsContentPath: contentPath, // This will be the absolute path
	}

	return book, nil
}

// Start starts the book server on the given port
func (w *WikiServer) Start(port string) error {
	log.Printf("Starting book server on port %s", port)
	return w.App.Listen(":" + port)
}

// Note: loadConfiguration function is defined in utils.go
