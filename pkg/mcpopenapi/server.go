package mcpopenapi

import (
	"fmt"
	"io"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

// HelloArguments represents the arguments for the hello tool
type HelloArguments struct {
	Submitter string `json:"submitter" jsonschema:"required,description=The name of the person to greet"`
}

// OpenAPIValidationArgs represents the arguments for the OpenAPI validation tool
type OpenAPIValidationArgs struct {
	Spec string `json:"spec" jsonschema:"required,description=The OpenAPI specification content to validate"`
}

// NewMCPServer creates a new MCP server for OpenAPI validation
func NewMCPServer(stdin io.Reader, stdout io.Writer) (*mcp.Server, error) {
	// Create a transport with specified IO
	transport := stdio.NewStdioServerTransportWithIO(stdin, stdout)
	
	// Create the server with the transport
	server := mcp.NewServer(transport)

	// Register the hello tool as an example
	err := server.RegisterTool("hello", "Say hello to a person", 
		func(arguments HelloArguments) (*mcp.ToolResponse, error) {
			return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Hello, %s!", arguments.Submitter))), nil
		})
	
	if err != nil {
		return nil, fmt.Errorf("failed to register hello tool: %w", err)
	}

	// Register the OpenAPI validation tool
	err = server.RegisterTool("validate_openapi", "Validate an OpenAPI specification", 
		func(args OpenAPIValidationArgs) (*mcp.ToolResponse, error) {
			result, err := ValidateOpenAPISpec([]byte(args.Spec))
			if err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Error validating OpenAPI spec: %v", err))), nil
			}
			
			if result == "" {
				return mcp.NewToolResponse(mcp.NewTextContent("OpenAPI specification is valid. No schemas found.")), nil
			}
			
			return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
		})
	
	if err != nil {
		return nil, fmt.Errorf("failed to register validate_openapi tool: %w", err)
	}

	return server, nil
}
