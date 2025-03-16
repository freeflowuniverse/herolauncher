package openapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"gopkg.in/yaml.v3"
)

// ServerGenerator generates a Fiber server from an OpenAPI specification
type ServerGenerator struct {
	Spec *OpenAPISpec
}

// NewServerGenerator creates a new ServerGenerator
func NewServerGenerator(spec *OpenAPISpec) *ServerGenerator {
	return &ServerGenerator{
		Spec: spec,
	}
}

// GenerateServer creates a Fiber server with routes based on the OpenAPI spec
func (g *ServerGenerator) GenerateServer() *fiber.App {
	app := fiber.New()

	// Add middleware for logging
	app.Use(func(c *fiber.Ctx) error {
		fmt.Printf("[%s] %s\n", c.Method(), c.Path())
		return c.Next()
	})

	// Register all paths and operations
	for path, pathItem := range g.Spec.GetPaths() {
		fiberPath := convertPathParams(path)

		if pathItem.Get != nil {
			g.registerOperation(app, http.MethodGet, fiberPath, pathItem.Get)
		}
		if pathItem.Post != nil {
			g.registerOperation(app, http.MethodPost, fiberPath, pathItem.Post)
		}
		if pathItem.Put != nil {
			g.registerOperation(app, http.MethodPut, fiberPath, pathItem.Put)
		}
		if pathItem.Delete != nil {
			g.registerOperation(app, http.MethodDelete, fiberPath, pathItem.Delete)
		}
		if pathItem.Options != nil {
			g.registerOperation(app, http.MethodOptions, fiberPath, pathItem.Options)
		}
		if pathItem.Head != nil {
			g.registerOperation(app, http.MethodHead, fiberPath, pathItem.Head)
		}
		if pathItem.Patch != nil {
			g.registerOperation(app, http.MethodPatch, fiberPath, pathItem.Patch)
		}
	}

	return app
}

// registerOperation registers a single operation with the Fiber app
func (g *ServerGenerator) registerOperation(app *fiber.App, method, path string, operation *v3.Operation) {
	handler := g.createMockHandler(operation)

	switch method {
	case http.MethodGet:
		app.Get(path, handler)
	case http.MethodPost:
		app.Post(path, handler)
	case http.MethodPut:
		app.Put(path, handler)
	case http.MethodDelete:
		app.Delete(path, handler)
	case http.MethodOptions:
		app.Options(path, handler)
	case http.MethodHead:
		app.Head(path, handler)
	case http.MethodPatch:
		app.Patch(path, handler)
	}

	operationID := "unknown"
	if operation.OperationId != "" {
		operationID = operation.OperationId
	}

	fmt.Printf("Registered %s %s (OperationID: %s)\n", method, path, operationID)
}

// createMockHandler creates a mock handler for an operation
func (g *ServerGenerator) createMockHandler(operation *v3.Operation) fiber.Handler {
	return func(c *fiber.Ctx) error {
		operationID := operation.OperationId
		if operationID == "" {
			operationID = "unknown"
		}

		// Find the first success response (2xx)
		var successResponse *v3.Response
		var statusCode string

		for pair := operation.Responses.Codes.First(); pair != nil; pair = pair.Next() {
			code := pair.Key()
			response := pair.Value()
			if strings.HasPrefix(code, "2") {
				successResponse = response
				statusCode = code
				break
			}
		}

		if successResponse == nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "No success response defined in OpenAPI spec",
			})
		}

		// Look for examples in the response
		if successResponse.Content != nil {
			for contentPair := successResponse.Content.First(); contentPair != nil; contentPair = contentPair.Next() {
				mediaType := contentPair.Key()
				mediaTypeObj := contentPair.Value()

				// Check if we have an example directly in the media type
				if mediaTypeObj.Example != nil {
					c.Set("Content-Type", mediaType)
					
					// Extract the example value from the YAML node
					var exampleData interface{}
					
					// Check if it's a scalar value (string, number, etc.)
					if mediaTypeObj.Example.Kind == yaml.ScalarNode {
						// Parse the scalar value as JSON if it looks like JSON
						value := mediaTypeObj.Example.Value
						if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
							// Try to parse as JSON
							err := json.Unmarshal([]byte(value), &exampleData)
							if err != nil {
								// If parsing fails, use as raw string
								exampleData = value
							}
						} else {
							// Use as raw string or number
							exampleData = value
						}
					} else {
						// For complex types (maps, arrays), decode directly
						err := mediaTypeObj.Example.Decode(&exampleData)
						if err != nil {
							// If decoding fails, use a default empty object
							exampleData = map[string]interface{}{}
						}
					}
					
					return c.Status(getStatusCode(statusCode)).JSON(exampleData)
				}

				// Check if we have examples in the media type
				if mediaTypeObj.Examples != nil {
					for examplePair := mediaTypeObj.Examples.First(); examplePair != nil; examplePair = examplePair.Next() {
						example := examplePair.Value()
						if example.Value != nil {
							c.Set("Content-Type", mediaType)
							
							// Extract the example value
							var exampleData interface{}
							
							// The Value field is already a *yaml.Node
							yamlNode := example.Value
							
							// Check if it's a scalar value (string, number, etc.)
							if yamlNode.Kind == yaml.ScalarNode {
								// Parse the scalar value as JSON if it looks like JSON
								value := yamlNode.Value
								if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
									// Try to parse as JSON
									err := json.Unmarshal([]byte(value), &exampleData)
									if err != nil {
										// If parsing fails, use as raw string
										exampleData = value
									}
								} else {
									// Use as raw string or number
									exampleData = value
								}
							} else {
								// For complex types (maps, arrays), decode directly
								err := yamlNode.Decode(&exampleData)
								if err != nil {
									// If decoding fails, use a default empty object
									exampleData = map[string]interface{}{}
								}
							}
							
							return c.Status(getStatusCode(statusCode)).JSON(exampleData)
						}
					}
				}

				// If we have a schema but no examples, generate a mock response based on the schema
				if mediaTypeObj.Schema != nil {
					// Convert SchemaProxy to a map for mock generation
					schemaMap := orderedmap.New[string, interface{}]()
					schemaMap.Set("type", "object")
					mockResponse := generateMockFromSchema(schemaMap)
					c.Set("Content-Type", mediaType)
					return c.Status(getStatusCode(statusCode)).JSON(mockResponse)
				}
			}
		}

		// If no examples or schemas found, return a generic response
		return c.Status(getStatusCode(statusCode)).JSON(fiber.Map{
			"message": fmt.Sprintf("Mock response for operation: %s", operationID),
		})
	}
}

// convertPathParams converts OpenAPI path parameters to Fiber path parameters
// OpenAPI: /users/{userId}/posts/{postId}
// Fiber:   /users/:userId/posts/:postId
func convertPathParams(path string) string {
	return strings.ReplaceAll(strings.ReplaceAll(path, "{", ":"), "}", "")
}

// getStatusCode converts a string status code to an integer
func getStatusCode(code string) int {
	switch code {
	case "200":
		return http.StatusOK
	case "201":
		return http.StatusCreated
	case "202":
		return http.StatusAccepted
	case "204":
		return http.StatusNoContent
	default:
		// Default to 200 OK for unknown success codes
		return http.StatusOK
	}
}

// generateMockFromSchema generates a mock response based on a schema
func generateMockFromSchema(schema *orderedmap.Map[string, interface{}]) interface{} {
	if schema == nil {
		return nil
	}

	// Handle different types
	typeValue, exists := schema.Get("type")
	if exists {
		switch typeValue.(string) {
		case "object":
			return generateMockObject(schema)
		case "array":
			return generateMockArray(schema)
		case "string":
			return "string"
		case "integer", "number":
			return 0
		case "boolean":
			return false
		}
	}

	// Default mock
	return map[string]string{"mock": "response"}
}

// generateMockObject generates a mock object based on a schema
func generateMockObject(schema *orderedmap.Map[string, interface{}]) map[string]interface{} {
	result := make(map[string]interface{})

	properties, exists := schema.Get("properties")
	if exists && properties != nil {
		if propsMap, ok := properties.(*orderedmap.Map[string, interface{}]); ok {
			for pair := propsMap.First(); pair != nil; pair = pair.Next() {
				name := pair.Key()
				propSchema, _ := pair.Value().(*orderedmap.Map[string, interface{}])
				result[name] = generateMockFromSchema(propSchema)
			}
		}
	}

	return result
}

// generateMockArray generates a mock array based on a schema
func generateMockArray(schema *orderedmap.Map[string, interface{}]) []interface{} {
	result := make([]interface{}, 0)

	// Generate one item for the array
	items, exists := schema.Get("items")
	if exists && items != nil {
		if itemsMap, ok := items.(*orderedmap.Map[string, interface{}]); ok {
			item := generateMockFromSchema(itemsMap)
			result = append(result, item)
		}
	}

	return result
}

// TemplateData holds the data for the server template
type TemplateData struct {
	Routes []RouteData
}

// RouteData holds the data for a route template
type RouteData struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Responses   []ResponseData
}

// ResponseData holds the data for a response template
type ResponseData struct {
	StatusCode string
	MediaType  string
	Example    string
}

// createRouteData creates a RouteData object from an Operation
func createRouteData(method, path string, operation *v3.Operation) RouteData {
	operationID := operation.OperationId
	if operationID == "" {
		operationID = "unknown"
	}

	route := RouteData{
		Method:      method,
		Path:        path,
		OperationID: operationID,
		Summary:     operation.Summary,
		Description: operation.Description,
		Responses:   []ResponseData{},
	}

	// Add example responses
	for pair := operation.Responses.Codes.First(); pair != nil; pair = pair.Next() {
		statusCode := pair.Key()
		response := pair.Value()

		if strings.HasPrefix(statusCode, "2") {
			if response.Content != nil {
				for contentPair := response.Content.First(); contentPair != nil; contentPair = contentPair.Next() {
					mediaType := contentPair.Key()
					mediaTypeObj := contentPair.Value()

					if strings.Contains(mediaType, "json") && mediaTypeObj.Example != nil {
						// Extract the example value from the YAML node
						var exampleStr string
						
						// Check if it's a scalar value (string, number, etc.)
						if mediaTypeObj.Example.Kind == yaml.ScalarNode {
							// Use the scalar value directly if it's already a JSON string
							exampleStr = mediaTypeObj.Example.Value
							// Trim any trailing newlines that might be in the YAML
							exampleStr = strings.TrimSpace(exampleStr)
						} else {
							// For complex types (maps, arrays), decode to interface{} and then marshal to JSON
							var data interface{}
							err := mediaTypeObj.Example.Decode(&data)
							if err == nil {
								// Marshal to pretty JSON for better readability
								exampleJSON, err := json.MarshalIndent(data, "", "  ")
								if err == nil {
									exampleStr = string(exampleJSON)
								}
							}
						}

						// Only add if we have a valid example string
						if exampleStr != "" {
							route.Responses = append(route.Responses, ResponseData{
								StatusCode: statusCode,
								MediaType:  mediaType,
								Example:    exampleStr,
							})
							// Only use the first valid example
							break
						}
					}
				}
			}
		}
	}

	return route
}

// loadTemplate loads a template from a file with fallback paths
func loadTemplate(name string) (string, error) {
	// Get the executable directory
	execDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}
	
	// Try different paths for the template
	paths := []string{
		fmt.Sprintf("%s/templates/%s.tmpl", execDir, name),
		fmt.Sprintf("%s/pkg/openapi/templates/%s.tmpl", execDir, name),
		fmt.Sprintf("%s/../templates/%s.tmpl", execDir, name),
		fmt.Sprintf("%s/../../templates/%s.tmpl", execDir, name),
		// Absolute paths for the repo structure
		fmt.Sprintf("%s/pkg/openapi/templates/%s.tmpl", strings.TrimSuffix(execDir, "/pkg/openapi/examples"), name),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
	}

	return "", fmt.Errorf("template %s not found (searched in %v)", name, paths)
}

// GenerateServerCode generates Fiber server code as a string
// This can be used to write the server code to a file
func (g *ServerGenerator) GenerateServerCode() string {
	// Load all templates
	templateFiles := []string{"server", "route", "response", "app", "handler", "types", "middleware"}
	templates := make(map[string]string)

	for _, name := range templateFiles {
		content, err := loadTemplate(name)
		if err != nil {
			return fmt.Sprintf("Error loading template %s: %v", name, err)
		}
		templates[name] = content
	}

	// Create template with all files
	tmpl := template.New("server")
	
	// Add each template to the template set
	for name, content := range templates {
		_, err := tmpl.New(name).Parse(content)
		if err != nil {
			return fmt.Sprintf("Error parsing template %s: %v", name, err)
		}
	}

	// Prepare template data
	templateData := TemplateData{
		Routes: []RouteData{},
	}

	// Add routes to template data
	for pathPair := g.Spec.Document.Paths.PathItems.First(); pathPair != nil; pathPair = pathPair.Next() {
		path := pathPair.Key()
		pathItem := pathPair.Value()
		fiberPath := convertPathParams(path)

		if pathItem.Get != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Get", fiberPath, pathItem.Get))
		}
		if pathItem.Post != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Post", fiberPath, pathItem.Post))
		}
		if pathItem.Put != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Put", fiberPath, pathItem.Put))
		}
		if pathItem.Delete != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Delete", fiberPath, pathItem.Delete))
		}
		if pathItem.Options != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Options", fiberPath, pathItem.Options))
		}
		if pathItem.Head != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Head", fiberPath, pathItem.Head))
		}
		if pathItem.Patch != nil {
			templateData.Routes = append(templateData.Routes, createRouteData("Patch", fiberPath, pathItem.Patch))
		}
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "server", templateData); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return buf.String()
}
