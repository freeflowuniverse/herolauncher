# Instruction Guide: Building MCP Servers with REST Interface

## Overview
This guide provides a step-by-step approach to building MCP servers using the REST interface. It consolidates key concepts, architecture, and examples from the MCP library documentation.

---

## What is MCP?
MCP (Model Context Protocol) is a standardized protocol for building servers that interact with AI models. The `mcp-golang` library provides a robust framework for implementing MCP servers in Go.

### Key Features
- **Batteries Included**: Quickly set up servers with tools, resources, and prompts.
- **Type Safety**: Full Go type safety with JSON schema generation.
- **Composable**: Use only the components you need (transport, protocol, or server).
- **Customizable Transport**: Supports default transports (stdio, HTTP) and custom implementations.

---

## Architecture
The MCP library is divided into three main layers:
1. **Transport Layer**: Handles communication (e.g., HTTP, WebSocket) and converts messages to/from JSON-RPC.
2. **Protocol Layer**: Manages JSON-RPC messages, routing, and error handling.
3. **Server Layer**: Combines transport and protocol layers to build a high-level server API.

User-defined handlers implement the business logic for tools, resources, and prompts.

---

## Setting Up an MCP Server

### Installation
Add the MCP library to your project:
```bash
go get github.com/metoro-io/mcp-golang
```

### Basic Server Example
Here’s how to create a simple MCP server:

```go
package main

import (
    "fmt"
    "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/stdio"
)

type MyToolArgs struct {
    Name string `json:"name"`
}

func main() {
    server := mcp.NewServer(stdio.NewStdioServerTransport())

    server.RegisterTool("greet", "Greets a user", func(args MyToolArgs) (*mcp.ToolResponse, error) {
        message := fmt.Sprintf("Hello, %s!", args.Name)
        return mcp.NewToolResponse(mcp.NewTextContent(message)), nil
    })

    if err := server.Serve(); err != nil {
        panic(err)
    }
}
```

### HTTP Server Example
For an HTTP-based server:

```go
package main

import (
    "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/http"
)

func main() {
    transport := http.NewHTTPTransport("/mcp")
    transport.WithAddr(":8080")

    server := mcp.NewServer(transport)
    server.RegisterTool("hello", &HelloTool{})

    if err := server.Serve(); err != nil {
        panic(err)
    }
}
```

---

## Prompt Example

Prompts in MCP allow for structured interactions with AI models. Here’s how to implement a prompt:

```go
package main

import (
	"context"
	"fmt"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

type MyPromptArgs struct {
	Input string `json:"input"`
}

func main() {
	transport := http.NewHTTPTransport("/mcp")
	transport.WithAddr(":8080")

	server := mcp.NewServer(transport)

	server.RegisterPrompt("my-prompt", "My example prompt", func(args MyPromptArgs) (*mcp.PromptResponse, error) {
		// Process the prompt input
		response := fmt.Sprintf("Received input: %s", args.Input)
		return mcp.NewPromptResponse("success", mcp.NewPromptMessage(mcp.NewTextContent(response), mcp.RoleAssistant)), nil
	})

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
```

This example demonstrates how to register a prompt handler that processes input and returns a structured response.

---

## Sampling Example

Sampling in MCP allows servers to request LLM completions through the client, enabling advanced agentic behaviors. Here’s how to implement sampling:

### How Sampling Works
1. The server sends a `sampling/createMessage` request to the client.
2. The client reviews the request and may modify it.
3. The client samples from an LLM.
4. The client reviews the completion.
5. The client returns the result to the server.

This process ensures human oversight and control over the LLM interactions.

### Example Implementation

```go
package main

import (
	"context"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

func main() {
	transport := http.NewHTTPTransport("/mcp")
	transport.WithAddr(":8080")

	server := mcp.NewServer(transport)

	server.RegisterTool("sampling-example", "Example of sampling", func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResponse, error) {
		request := map[string]interface{}{
			"method": "sampling/createMessage",
			"params": map[string]interface{}{
				"messages": []map[string]interface{}{
					{
						"role": "user",
						"content": map[string]interface{}{
							"type": "text",
							"text": "What is the weather today?",
						},
					},
				},
				"systemPrompt": "You are a helpful assistant providing weather updates.",
				"includeContext": "thisServer",
				"maxTokens": 50,
			},
		}

		response, err := server.CallClient(ctx, request)
		if err != nil {
			return nil, err
		}

		return mcp.NewToolResponse(response), nil
	})

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
```

### Key Parameters
- **messages**: Conversation history sent to the LLM.
- **systemPrompt**: Optional field to guide the LLM’s behavior.
- **includeContext**: Specifies the context to include (`none`, `thisServer`, or `allServers`).
- **maxTokens**: Maximum tokens to generate.

---

## Roots Example

Roots in MCP define the boundaries where servers can operate. They allow clients to inform servers about relevant resources and their locations, such as filesystem paths or HTTP URLs.

### How Roots Work
1. The client declares the `roots` capability during connection.
2. The client provides a list of suggested roots to the server.
3. The server uses these roots to locate and access resources, prioritizing operations within root boundaries.

### Example Implementation

```go
package main

import (
	"context"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

func main() {
	transport := http.NewHTTPTransport("/mcp")
	transport.WithAddr(":8080")

	server := mcp.NewServer(transport)

	server.RegisterTool("list-roots", "List available roots", func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResponse, error) {
		roots := []map[string]interface{}{
			{"uri": "file:///home/user/projects/frontend", "name": "Frontend Repository"},
			{"uri": "https://api.example.com/v1", "name": "API Endpoint"},
		}

		return mcp.NewToolResponse(roots), nil
	})

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
```

### Common Use Cases
- Defining project directories
- Specifying repository locations
- Indicating API endpoints
- Setting configuration locations
- Establishing resource boundaries

---

## Resources Example

Resources in MCP allow servers to expose data and content that can be read by clients and used as context for LLM interactions. Each resource is identified by a unique URI and can contain either text or binary data.

### How Resources Work
1. Resources are identified using URIs, such as `file:///path/to/file` or `https://api.example.com/data`.
2. Clients can discover available resources through the `resources/list` endpoint.
3. Clients can read resource contents using the `resources/read` request.
4. Servers can notify clients about resource updates via `notifications/resources/updated`.

### Example Implementation

```go
package main

import (
	"context"
	"github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
)

func main() {
	transport := http.NewHTTPTransport("/mcp")
	transport.WithAddr(":8080")

	server := mcp.NewServer(transport)

	// List available resources
	server.RegisterTool("list-resources", "List available resources", func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResponse, error) {
		resources := []map[string]interface{}{
			{"uri": "file:///logs/app.log", "name": "Application Logs", "mimeType": "text/plain"},
			{"uri": "https://api.example.com/data", "name": "API Data", "mimeType": "application/json"},
		}
		return mcp.NewToolResponse(resources), nil
	})

	// Read resource contents
	server.RegisterTool("read-resource", "Read resource contents", func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResponse, error) {
		uri := args["uri"].(string)
		if uri == "file:///logs/app.log" {
			content := "Log file contents here..."
			return mcp.NewToolResponse([]map[string]interface{}{
				{"uri": uri, "mimeType": "text/plain", "text": content},
			}), nil
		} else if uri == "https://api.example.com/data" {
			content := "{\"key\": \"value\"}"
			return mcp.NewToolResponse([]map[string]interface{}{
				{"uri": uri, "mimeType": "application/json", "text": content},
			}), nil
		}
		return nil, fmt.Errorf("Resource not found")
	})

	if err := server.Serve(); err != nil {
		panic(err)
	}
}
```

### Common Use Cases
- Exposing file contents (e.g., logs, configuration files)
- Providing API responses
- Sharing live system data
- Delivering images or binary data

### Best Practices
1. Use clear, descriptive resource names and URIs.
2. Include helpful descriptions to guide LLM understanding.
3. Set appropriate MIME types when known.
4. Handle errors gracefully with clear error messages.
5. Consider pagination for large resource lists.
6. Validate URIs before processing.
7. Document your custom URI schemes.

---

## Using the MCP Client
The MCP client allows interaction with MCP servers. Below is an example of initializing a client and calling a tool:

```go
import (
    "context"
    "github.com/metoro-io/mcp-golang"
    "github.com/metoro-io/mcp-golang/transport/http"
)

func main() {
    transport := http.NewHTTPClientTransport("/mcp")
    transport.WithBaseURL("http://localhost:8080")

    client := mcp.NewClient(transport)

    response, err := client.CallTool(context.Background(), "greet", map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```
