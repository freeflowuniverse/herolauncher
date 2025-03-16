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

Validates an OpenAPI specification file and returns information about its schemas or any validation errors.

#### Parameters

```json
{
  "filePath": "The path to the OpenAPI specification file to validate"
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
      "filePath": "/path/to/openapi.json"
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
