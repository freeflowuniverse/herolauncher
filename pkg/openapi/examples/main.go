package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"gopkg.in/yaml.v3"

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
	generateApiFromSpec("Petstore", petstoreSpecPath, petstoreApiDir)
	generateApiFromSpec("Actions", actionsSpecPath, actionsApiDir)

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

// generateApiFromSpec generates API code from an OpenAPI specification file
func generateApiFromSpec(apiName string, specPath string, outputDir string) {
	fmt.Printf("Generating %s API code...\n", apiName)

	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(specPath)
	if err != nil {
		log.Fatalf("Failed to parse %s OpenAPI specification: %v", apiName, err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputPath := filepath.Join(outputDir, "server.go")
	err = os.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write %s server code: %v", apiName, err)
	}

	// Copy the OpenAPI spec to the output directory
	specOutputPath := filepath.Join(outputDir, "openapi.yaml")
	copyFile(specPath, specOutputPath)

	fmt.Printf("Generated %s API code in %s\n", apiName, outputPath)
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

	// Setup Swagger UI for each API
	setupSwaggerUI(app, "Petstore", petstoreApiDir, petstoreSpecPath)
	setupSwaggerUI(app, "Actions", actionsApiDir, actionsSpecPath)

	// OpenAPI specs are now served dynamically by setupSwaggerUI

	// Mount the APIs
	setupApi(app, "Petstore", petstoreSpecPath, "/api/petstore")
	setupApi(app, "Actions", actionsSpecPath, "/api/actions")
}

// setupSwaggerUI sets up the Swagger UI for an API
func setupSwaggerUI(app *fiber.App, apiName string, apiDir string, specPath string) {
	// API Swagger UI
	app.Get(fmt.Sprintf("/api/swagger/%s", strings.ToLower(apiName)), func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		// Set the baseUrl to include the API prefix
		baseUrl := fmt.Sprintf("http://localhost:%d/api/%s", port, strings.ToLower(apiName))
		specUrl := fmt.Sprintf("/api/%s/openapi.yaml", strings.ToLower(apiName))
		return c.SendString(getSwaggerUIHtml(fmt.Sprintf("%s API", apiName), specUrl, baseUrl))
	})

	// Create a modified version of the OpenAPI spec with server information
	app.Get(fmt.Sprintf("/api/%s/openapi.yaml", strings.ToLower(apiName)), func(c *fiber.Ctx) error {
		// Read the original spec file
		specData, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("failed to read OpenAPI spec file: %w", err)
		}

		// Parse the YAML into a map
		var specMap map[string]interface{}
		if err := yaml.Unmarshal(specData, &specMap); err != nil {
			return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}

		// Add or update the servers section
		specMap["servers"] = []map[string]interface{}{
			{
				"url": fmt.Sprintf("/api/%s", strings.ToLower(apiName)),
				"description": fmt.Sprintf("%s API Server", apiName),
			},
		}

		// Convert back to YAML
		modifiedSpec, err := yaml.Marshal(specMap)
		if err != nil {
			return fmt.Errorf("failed to marshal modified OpenAPI spec: %w", err)
		}

		// Send the modified spec
		c.Set("Content-Type", "application/yaml")
		return c.Send(modifiedSpec)
	})
}

// setupApi sets up an API based on an OpenAPI specification
func setupApi(app *fiber.App, apiName string, specPath string, basePath string) {
	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile(specPath)
	if err != nil {
		log.Fatalf("Failed to parse %s OpenAPI specification: %v", apiName, err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate the server
	apiServer := generator.GenerateServer()

	// Create a handler for the root path
	app.Get(basePath, func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": fmt.Sprintf("Welcome to the %s API", apiName)})
	})

	// Register all routes from the generated server directly on the main app with the basePath prefix
	for _, route := range apiServer.GetRoutes() {
		// Skip the root path as we've already handled it
		if route.Path == "/" {
			continue
		}

		// Add the route to the main app with the basePath prefix
		app.Add(route.Method, basePath+route.Path, route.Handlers...)
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
