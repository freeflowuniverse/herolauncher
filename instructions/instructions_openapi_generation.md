# OpenAPI Generation Instructions

## Overview

The OpenAPI package in `pkg/openapi` provides functionality to generate server code from OpenAPI specifications. This document explains how to use this package to generate and host multiple APIs under a single server with Swagger UI integration.

## Implementation Status

We have successfully implemented:

1. A proper test in `pkg/openapi/examples` that generates code from OpenAPI specifications
2. Code generation for two example APIs:
   - `petstoreapi` (from `petstore.yaml`)
   - `actionsapi` (from `actions.yaml`)
3. A webserver that hosts multiple generated APIs
4. Swagger UI integration for API documentation
5. A home page with links to the APIs and their documentation

All APIs are hosted under `$serverurl:$port/api` with a clean navigation structure.

## Directory Structure

```
pkg/openapi/
├── examples/
│   ├── actions.yaml        # OpenAPI spec for Actions API
│   ├── actionsapi/         # Generated code for Actions API
│   ├── main.go             # Main server implementation
│   ├── petstore.yaml       # OpenAPI spec for Petstore API
│   ├── petstoreapi/        # Generated code for Petstore API
│   ├── README.md           # Documentation for examples
│   ├── run_test.sh         # Script to run tests and server
│   └── test/               # Tests for OpenAPI generation
├── generator.go            # Server code generator
├── parser.go               # OpenAPI spec parser
├── example.go              # Example usage
└── templates/              # Code generation templates
    └── server.tmpl         # Server template
```

## How to Use

### Running the Example

To run the example implementation:

1. Navigate to the examples directory:
   ```bash
   cd pkg/openapi/examples
   ```

2. Run the test script:
   ```bash
   ./run_test.sh
   ```

3. Access the APIs:
   - API Home: http://localhost:9091/api
   - Petstore API: http://localhost:9091/api/petstore
   - Petstore API Documentation: http://localhost:9091/api/swagger/petstore
   - Actions API: http://localhost:9091/api/actions
   - Actions API Documentation: http://localhost:9091/api/swagger/actions

### Generating Code from Your Own OpenAPI Spec

To generate code from your own OpenAPI specification:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
)

func main() {
	// Parse the OpenAPI specification
	spec, err := openapi.ParseFromFile("your-api.yaml")
	if err != nil {
		fmt.Printf("Failed to parse OpenAPI specification: %v\n", err)
		os.Exit(1)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputPath := "generated-server.go"
	err = os.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		fmt.Printf("Failed to write server code: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated server code in %s\n", outputPath)
}
```

### Hosting Multiple APIs

To host multiple APIs under a single server:

```go
package main

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create the main server
	app := fiber.New()

	// Setup API routes
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.SendString("API Home Page")
	})

	// Mount the first API
	spec1, _ := openapi.ParseFromFile("api1.yaml")
	generator1 := openapi.NewServerGenerator(spec1)
	apiServer1 := generator1.GenerateServer()
	app.Mount("/api/api1", apiServer1)

	// Mount the second API
	spec2, _ := openapi.ParseFromFile("api2.yaml")
	generator2 := openapi.NewServerGenerator(spec2)
	apiServer2 := generator2.GenerateServer()
	app.Mount("/api/api2", apiServer2)

	// Start the server
	app.Listen(":8080")
}
```

### Adding Swagger UI

To add Swagger UI for API documentation:

```go
// Serve OpenAPI specs
app.Static("/api/api1/openapi.yaml", "api1.yaml")
app.Static("/api/api2/openapi.yaml", "api2.yaml")

// API1 Swagger UI
app.Get("/api/swagger/api1", func(c *fiber.Ctx) error {
	return c.SendString(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>API1 - Swagger UI</title>
			<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
			<style>
				html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
				*, *:before, *:after { box-sizing: inherit; }
				body { margin: 0; background: #fafafa; }
				.swagger-ui .topbar { display: none; }
			</style>
		</head>
		<body>
			<div id="swagger-ui"></div>
			<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
			<script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
			<script>
				window.onload = function() {
					const ui = SwaggerUIBundle({
						url: "/api/api1/openapi.yaml",
						dom_id: '#swagger-ui',
						deepLinking: true,
						presets: [
							SwaggerUIBundle.presets.apis,
							SwaggerUIStandalonePreset
						],
						layout: "StandaloneLayout"
					});
					window.ui = ui;
				};
			</script>
		</body>
		</html>
	`)
})
```

## Features

### OpenAPI Parsing

The package can parse OpenAPI 3.0 and 3.1 specifications from files or byte slices.

### Code Generation

The package generates Fiber server code with mock implementations based on examples in the OpenAPI spec.

### Mock Implementations

Mock implementations are created using examples from the OpenAPI spec, making it easy to test APIs without writing any code.

### Multiple API Hosting

The package supports hosting multiple APIs under a single server, with each API mounted at a different path.

### Swagger UI Integration

The package includes Swagger UI integration for API documentation, making it easy to explore and test APIs.

## Best Practices

1. **Organize Your Code**: Keep your OpenAPI specs, generated code, and server implementation in separate directories.

2. **Use Examples**: Include examples in your OpenAPI spec to generate better mock implementations.

3. **Test Your APIs**: Write tests to verify that your APIs work as expected.

4. **Document Your APIs**: Use Swagger UI to document your APIs and make them easier to use.

5. **Use Version Control**: Keep your OpenAPI specs and generated code in version control to track changes.

## Troubleshooting

- **Parse Error**: If you get a parse error, check that your OpenAPI spec is valid. You can use tools like [Swagger Editor](https://editor.swagger.io/) to validate your spec.

- **Generation Error**: If code generation fails, check that your OpenAPI spec includes all required fields and that examples are properly formatted.

- **Server Error**: If the server fails to start, check that the port is not already in use and that all required dependencies are installed.

