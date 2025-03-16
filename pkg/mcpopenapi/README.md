# MCP OpenAPI Server

This package implements an MCP (Model Context Protocol) server for OpenAPI specification validation using the libopenapi library.

## Features

- Validates OpenAPI specifications
- Returns detailed validation errors
- Lists schema information for valid specifications

## Usage

### As a Command Line Tool

```bash
# Build and run the MCP server
cd pkg/mcpopenapi/cmd/mcpopenapi
go build
./mcpopenapi
```

### As a Library

```go
package main

import (
	"fmt"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/mcpopenapi"
)

func main() {
	// Read an OpenAPI spec file
	specContent, err := os.ReadFile("path/to/openapi.json")
	if err != nil {
		panic(err)
	}

	// Validate the OpenAPI spec
	result, err := mcpopenapi.ValidateOpenAPISpec(specContent)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
```

## Example

See the `examples` directory for a complete example of how to use this package.
