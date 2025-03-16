# OpenAPI Package

This package provides tools for working with OpenAPI specifications, including parsing OpenAPI documents and generating Fiber server code with mock implementations using a template-based approach.

## Features

- Parse OpenAPI 3.1 specifications from files or byte slices
- Extract paths, operations, and examples from OpenAPI specifications
- Generate Fiber server code based on OpenAPI specifications using Go templates
- Create mock implementations using examples from the OpenAPI spec
- Properly handle complex example types from OpenAPI specifications
- Command-line tool for testing and demonstration

## Usage

### Parsing an OpenAPI Specification

```go
// Parse from file
spec, err := openapi.ParseFromFile("path/to/openapi.json")
if err != nil {
    log.Fatalf("Failed to parse OpenAPI spec: %v", err)
}

// Parse from bytes
data := []byte(`{"openapi": "3.1.0", ...}`)
spec, err := openapi.ParseFromBytes(data)
if err != nil {
    log.Fatalf("Failed to parse OpenAPI spec: %v", err)
}
```

### Generating a Fiber Server

```go
// Parse the OpenAPI spec
spec, err := openapi.ParseFromFile("path/to/openapi.json")
if err != nil {
    log.Fatalf("Failed to parse OpenAPI spec: %v", err)
}

// Create a server generator
generator := openapi.NewServerGenerator(spec)

// Generate and run the server
app := generator.GenerateServer()
app.Listen(":8080")
```

### Generating Server Code as String

You can also generate the server code as a string, which can be useful for saving to a file or further processing:

```go
// Parse the OpenAPI spec
spec, err := openapi.ParseFromFile("path/to/openapi.json")
if err != nil {
    log.Fatalf("Failed to parse OpenAPI spec: %v", err)
}

// Create a server generator
generator := openapi.NewServerGenerator(spec)

// Generate server code as string
serverCode := generator.GenerateServerCode()

// Write to file if needed
os.WriteFile("server.go", []byte(serverCode), 0644)
```

### Customizing Templates

The package uses Go templates for generating server code. The templates are located in the `templates` directory and can be customized to fit your needs. The following templates are available:

- `server.tmpl` - Main server template
- `route.tmpl` - Template for route handlers
- `response.tmpl` - Template for response handling
- `app.tmpl` - Template for Fiber app setup
- `handler.tmpl` - Template for handler functions
- `types.tmpl` - Template for type definitions
- `middleware.tmpl` - Template for middleware functions

To customize the templates, you can modify the existing templates or create your own and place them in a directory that will be searched by the `loadTemplate` function.
```

### Generating Server Code

```go
// Parse the OpenAPI spec
spec, err := openapi.ParseFromFile("path/to/openapi.json")
if err != nil {
    log.Fatalf("Failed to parse OpenAPI spec: %v", err)
}

// Create a server generator
generator := openapi.NewServerGenerator(spec)

// Generate server code as a string
serverCode := generator.GenerateServerCode()

// Write to file
os.WriteFile("server.go", []byte(serverCode), 0644)
```

## Command-Line Tool

The package includes a command-line tool for testing and demonstration purposes.

### Building the Command-Line Tool

```bash
cd pkg/openapi/cmd
go build -o openapi-server
```

### Running the Command-Line Tool

```bash
# Run a server based on an OpenAPI spec
./openapi-server -spec path/to/openapi.json -port 8080

# Generate server code without running the server
./openapi-server -spec path/to/openapi.json -generate-only -output server.go
```

## Example

See the `example.go` file for a complete example of using the OpenAPI package.
