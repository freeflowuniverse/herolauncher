package mail

import (
	"html/template"
	"log"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
)

// NewMailServer creates a new mail server with the given configuration
func NewMailServer(configPath string) (*MailServer, error) {
	// Initialize template engine
	templatesPath, _ := filepath.Abs("/Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/ui/mail/web/templates")
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
	staticPath, _ := filepath.Abs("/Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/ui/mail/web/static")
	app.Static("/", staticPath)
	app.Static("/css", filepath.Join(staticPath, "css"))
	app.Static("/js", filepath.Join(staticPath, "js"))
	app.Static("/icons", filepath.Join(staticPath, "icons"))

	// Load configuration from file if provided, otherwise use empty configuration
	var config Configuration
	var err error

	if configPath != "" {
		config, err = loadConfiguration(configPath)
		if err != nil {
			log.Printf("Error loading configuration: %v", err)
			// If config file loading fails, use empty config
			config = Configuration{
				Title: "Mail",
			}
			log.Printf("Warning: Using default configuration due to loading error")
		}
	} else {
		// No configuration file provided, use default config
		config = Configuration{
			Title: "Mail",
		}
		log.Printf("Warning: No configuration file provided. Using default configuration.")
	}

	// Create the mail server
	mail := &MailServer{
		App:    app,
		Config: config,
	}

	return mail, nil
}

// Start starts the mail server on the given port
func (m *MailServer) Start(port string) error {
	log.Printf("Starting mail server on port %s", port)
	return m.App.Listen(":" + port)
}

// loadConfiguration loads the configuration from a JSON file
func loadConfiguration(configPath string) (Configuration, error) {
	config := Configuration{
		Title: "Mail",
	}
	
	// In a real implementation, this would load from a JSON file
	// For now, we'll just return the default config
	return config, nil
}
