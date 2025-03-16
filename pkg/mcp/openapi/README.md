# MCP OpenAPI Server

This package implements an MCP (Model Context Protocol) server for OpenAPI specification validation using the libopenapi library.

## Features

- Validates OpenAPI specifications
- Returns detailed validation errors
- Lists schema information for valid specifications

## Usage

### As a Command Line Tool

```bash
# Build
cd ${HOME}/code/github/freeflowuniverse/herolauncher/pkg/mcp/openapi/cmd/openapi
go build -o mcpopenapi
cp mcpopenapi ${HOME}/hero/bin/
```

### As a Library

```go
package main

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/mcp/openapi"
)

func main() {
	// Specify the path to an OpenAPI spec file
	filePath := "path/to/openapi.json"

	// Validate the OpenAPI spec
	result, err := openapi.ValidateOpenAPISpec(filePath)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
```

## Example

See the `examples` directory for a complete example of how to use this package.
