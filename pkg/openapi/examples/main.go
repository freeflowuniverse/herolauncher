package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const (
	petstoreSpecPath = "petstore.yaml"
	actionsSpecPath  = "actions.yaml"
	petstoreApiDir   = "petstoreapi"
	actionsApiDir    = "actionsapi"
	port             = 9092
)

func main() {
	// Generate code for both APIs
	generatePetstoreApi()
	generateActionsApi()

	// Create the main server
	app := fiber.New(fiber.Config{
		AppName: "OpenAPI Test Server",
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Setup API routes
	setupApiRoutes(app)

	// Start the server
	fmt.Printf("\nStarting server on port %d...\n", port)
	fmt.Printf("Visit http://localhost:%d/api to see the API home page\n", port)
	fmt.Printf("Press Ctrl+C to stop the server\n")

	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// generatePetstoreApi generates the Petstore API code
func generatePetstoreApi() {
	fmt.Println("Generating Petstore API code...")

	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(petstoreSpecPath)
	if err != nil {
		log.Fatalf("Failed to parse Petstore OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputPath := filepath.Join(petstoreApiDir, "server.go")
	err = os.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write Petstore server code: %v", err)
	}

	// Copy the OpenAPI spec to the output directory
	specOutputPath := filepath.Join(petstoreApiDir, "openapi.yaml")
	copyFile(petstoreSpecPath, specOutputPath)

	fmt.Printf("Generated Petstore API code in %s\n", outputPath)
}

// generateActionsApi generates the Actions API code
func generateActionsApi() {
	fmt.Println("Generating Actions API code...")

	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(actionsSpecPath)
	if err != nil {
		log.Fatalf("Failed to parse Actions OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputPath := filepath.Join(actionsApiDir, "server.go")
	err = ioutil.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write Actions server code: %v", err)
	}

	// Copy the OpenAPI spec to the output directory
	specOutputPath := filepath.Join(actionsApiDir, "openapi.yaml")
	copyFile(actionsSpecPath, specOutputPath)

	fmt.Printf("Generated Actions API code in %s\n", outputPath)
}

// APIHomeData holds the data for the API home template
type APIHomeData struct {
	Title string
}

// getAPIHomeHtml returns the HTML for the API home page using a template
func getAPIHomeHtml() string {
	// Define the template path
	tmplPath := filepath.Join("templates", "api-home.html")
	
	// Parse the template
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing API home template: %v", err)
		// Fallback to a simple HTML if template parsing fails
		return `
			<!DOCTYPE html>
			<html>
			<head><title>API Home</title></head>
			<body>
				<h1>API Home</h1>
				<p>Error loading template. Please check the server logs.</p>
				<div>
					<h2>Petstore API</h2>
					<ul>
						<li><a href="/api/petstore">Petstore API Server</a></li>
						<li><a href="/api/swagger/petstore">Petstore API Documentation</a></li>
					</ul>
				</div>
				<div>
					<h2>Actions API</h2>
					<ul>
						<li><a href="/api/actions">Actions API Server</a></li>
						<li><a href="/api/swagger/actions">Actions API Documentation</a></li>
					</ul>
				</div>
			</body>
			</html>
		`
	}
	
	// Prepare the template data
	data := APIHomeData{
		Title: "API Home",
	}
	
	// Execute the template
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		log.Printf("Error executing API home template: %v", err)
		// Fallback to a simple HTML if template execution fails
		return `
			<!DOCTYPE html>
			<html>
			<head><title>API Home</title></head>
			<body>
				<h1>API Home</h1>
				<p>Error executing template. Please check the server logs.</p>
			</body>
			</html>
		`
	}
	
	return buf.String()
}

// setupApiRoutes sets up the API routes for the main server
func setupApiRoutes(app *fiber.App) {
	// API home page
	app.Get("/api", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(getAPIHomeHtml())
	})

	// Serve Swagger UI
	app.Static("/api/swagger-ui", "./swagger-ui")

	// Petstore API Swagger UI
	app.Get("/api/swagger/petstore", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(getSwaggerUIHtml("Petstore API", "/api/petstore/openapi.yaml", fmt.Sprintf("http://localhost:%d/api/petstore", port)))
	})

	// Actions API Swagger UI
	app.Get("/api/swagger/actions", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(getSwaggerUIHtml("Actions API", "/api/actions/openapi.yaml", fmt.Sprintf("http://localhost:%d/api/actions", port)))
	})

	// Serve OpenAPI specs
	app.Static("/api/petstore/openapi.yaml", filepath.Join(petstoreApiDir, "openapi.yaml"))
	app.Static("/api/actions/openapi.yaml", filepath.Join(actionsApiDir, "openapi.yaml"))

	// Mount the Petstore API
	setupPetstoreApi(app)

	// Mount the Actions API
	setupActionsApi(app)
}

// setupPetstoreApi sets up the Petstore API
func setupPetstoreApi(app *fiber.App) {
	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(petstoreSpecPath)
	if err != nil {
		log.Fatalf("Failed to parse Petstore OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate the server
	apiServer := generator.GenerateServer()

	// Create a handler for the root path
	app.Get("/api/petstore", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "Welcome to the Petstore API"})
	})

	// Register all routes from the generated server directly on the main app with the /api/petstore prefix
	for _, route := range apiServer.GetRoutes() {
		// Skip the root path as we've already handled it
		if route.Path == "/" {
			continue
		}

		// Add the route to the main app with the /api/petstore prefix
		app.Add(route.Method, "/api/petstore"+route.Path, route.Handlers...)
	}
}

// setupActionsApi sets up the Actions API
func setupActionsApi(app *fiber.App) {
	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(actionsSpecPath)
	if err != nil {
		log.Fatalf("Failed to parse Actions OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate the server
	apiServer := generator.GenerateServer()

	// Create a handler for the root path
	app.Get("/api/actions", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "Welcome to the Actions API"})
	})

	// Register all routes from the generated server directly on the main app with the /api/actions prefix
	for _, route := range apiServer.GetRoutes() {
		// Skip the root path as we've already handled it
		if route.Path == "/" {
			continue
		}

		// Add the route to the main app with the /api/actions prefix
		app.Add(route.Method, "/api/actions"+route.Path, route.Handlers...)
	}
}

// SwaggerUIData holds the data for the Swagger UI template
type SwaggerUIData struct {
	Title   string
	SpecUrl string
	BaseUrl string
}

// getSwaggerUIHtml returns the HTML for the Swagger UI using a template
func getSwaggerUIHtml(title string, specUrl string, baseUrl string) string {
	// Define the template path
	tmplPath := filepath.Join("templates", "swagger-ui.html")

	// Parse the template
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing Swagger UI template: %v", err)
		// Fallback to a simple HTML if template parsing fails
		return fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head><title>%s</title></head>
			<body>
				<h1>Error loading Swagger UI</h1>
				<p>Could not load template: %v</p>
			</body>
			</html>
		`, title, err)
	}

	// Prepare the template data
	data := SwaggerUIData{
		Title:   title,
		SpecUrl: specUrl,
		BaseUrl: baseUrl,
	}

	// Execute the template
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		log.Printf("Error executing Swagger UI template: %v", err)
		// Fallback to a simple HTML if template execution fails
		return fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head><title>%s</title></head>
			<body>
				<h1>Error loading Swagger UI</h1>
				<p>Could not execute template: %v</p>
			</body>
			</html>
		`, title, err)
	}

	return buf.String()
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Ensure the destination directory exists
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	// Read the source file
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to the destination file
	return os.WriteFile(dst, data, 0644)
}
