# MCP V Language Specs Server

This is a Model Context Protocol (MCP) server for extracting specifications from V language files. It follows the [Model Context Protocol specification](https://spec.modelcontextprotocol.io/) to enable AI models to interact with V language specification tools.

## Features

- Implements the Model Context Protocol for AI model interaction
- Provides tools for extracting public structs, enums, and methods from V language files
- Recursively scans directories to find all V language files
- Skips generated files (files with names ending in `_.v` or starting with `_`)
- Preserves documentation comments
- Uses standard I/O for communication following the MCP transport specification

## Installation

```bash
go build
```

## Usage

Simply run the executable to start the MCP server:

```bash
./mcpv
```

The server will start and listen for MCP protocol requests on standard input/output. You can interact with it using any MCP client.

## MCP Tools

The server exposes the following MCP tools:

### V Language Specs Tool

Extracts public structs, enums, and methods from V language files.

#### Parameters

```json
{
  "path": "The path to the V language files to process"
}
```

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tool.call",
  "params": {
    "name": "get_specs",
    "arguments": {
      "path": "/path/to/vlang/files"
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
      "text": "// From file: /path/to/vlang/files/user.v\npub struct User {\n  id string\n  name string\n  email string\n}\n\npub fn (u &User) CreateUser() {}\n"
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
- [vlangspecs](https://github.com/freeflowuniverse/herolauncher/pkg/vlang/vlangspecs) - V language specification extraction library

## MCP Specification Compliance

This server implements the following parts of the MCP specification:

- **Transport**: Uses stdio transport as defined in the specification
- **Tools**: Implements tool discovery and execution
- **Error Handling**: Provides proper error responses for invalid requests

For more information on the MCP specification, visit [spec.modelcontextprotocol.io](https://spec.modelcontextprotocol.io/).
