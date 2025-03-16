# MCP OpenAPI Server

This is a Model Context Protocol (MCP) server for validating OpenAPI specifications. It follows the [Model Context Protocol specification](https://spec.modelcontextprotocol.io/) to enable AI models to interact with OpenAPI validation tools.

## Features

- Implements the Model Context Protocol for AI model interaction
- Provides tools for basic interaction and OpenAPI validation
- Returns detailed information about schemas in valid OpenAPI specifications
- Reports validation errors for invalid specifications
- Supports MCP discovery for tools
- Uses standard I/O for communication following the MCP transport specification

## Installation

```bash
go build
```

## Usage

Simply run the executable to start the MCP server:

```bash
./mcpopenapi
```

The server will start and listen for MCP protocol requests on standard input/output. You can interact with it using any MCP client.

## MCP Tools

The server exposes the following MCP tools:

### 1. Hello Tool

A simple greeting tool that demonstrates basic MCP tool functionality.

#### Parameters

```json
{
  "submitter": "The name of the person to greet"
}
```

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tool.call",
  "params": {
    "name": "hello",
    "arguments": {
      "submitter": "World"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": {
      "type": "text",
      "text": "Hello, World!"
    }
  }
}
```

### 2. OpenAPI Validation Tool

Validates an OpenAPI specification and returns information about its schemas or any validation errors.

#### Parameters

```json
{
  "spec": "The OpenAPI specification content to validate (as a string)"
}
```

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tool.call",
  "params": {
    "name": "validate_openapi",
    "arguments": {
      "spec": "{ ... OpenAPI specification JSON ... }"
    }
  }
}
```

#### Example Response

For a valid OpenAPI specification:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": {
      "type": "text",
      "text": "Schema 'User' has 3 properties"
    }
  }
}
```

## Testing

We provide several methods to test the MCP server:

1. **MCP Inspector**: Use the provided `test_with_inspector.sh` script to test the server with a graphical interface.

2. **JavaScript Test Script**: Run `node test_mcp_server.js` to test the server programmatically.

3. **Manual Testing**: See the `TESTING.md` file for detailed instructions on manual testing.

## Development

This server is built using:

- [mcp-golang](https://github.com/metoro-io/mcp-golang) - Go implementation of the Model Context Protocol
- [libopenapi](https://github.com/pb33f/libopenapi) - OpenAPI specification validation library

## MCP Specification Compliance

This server implements the following parts of the MCP specification:

- **Transport**: Uses stdio transport as defined in the specification
- **Tools**: Implements tool discovery and execution
- **Error Handling**: Provides proper error responses for invalid requests

For more information on the MCP specification, visit [spec.modelcontextprotocol.io](https://spec.modelcontextprotocol.io/).
