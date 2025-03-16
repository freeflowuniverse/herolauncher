package mcpv

import (
	"fmt"
	"io"
	"os"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	vlangspecs "github.com/freeflowuniverse/herolauncher/pkg/vlang/vlangspecs"
)

// GetSpecsArgs represents the arguments for the get_specs tool
type GetSpecsArgs struct {
	Path string `json:"path" jsonschema:"required,description=The path to the V language files to process"`
}

// NewMCPServer creates a new MCP server for V language specs processing
func NewMCPServer(stdin io.Reader, stdout io.Writer) (*mcp.Server, error) {
	// Create a transport with specified IO
	transport := stdio.NewStdioServerTransportWithIO(stdin, stdout)
	
	// Create the server with the transport
	server := mcp.NewServer(transport)

	// Register the get_specs tool
	err := server.RegisterTool("get_specs", "Extract public structs, enums, and methods from V language files", 
		func(args GetSpecsArgs) (*mcp.ToolResponse, error) {
			// Check if the path exists
			if _, err := os.Stat(args.Path); os.IsNotExist(err) {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Error: Path does not exist: %s", args.Path))), nil
			}
			
			// Create a new VlangProcessor
			processor := vlangspecs.NewVlangProcessor()
			
			// Get the specification
			spec, err := processor.GetSpec(args.Path)
			if err != nil {
				return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Error processing V files: %v", err))), nil
			}
			
			if spec == "" {
				return mcp.NewToolResponse(mcp.NewTextContent("No public structs, enums, or methods found in the specified path.")), nil
			}
			
			return mcp.NewToolResponse(mcp.NewTextContent(spec)), nil
		})
	
	if err != nil {
		return nil, fmt.Errorf("failed to register get_specs tool: %w", err)
	}

	return server, nil
}
