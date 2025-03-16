# Testing the MCP OpenAPI Server

This guide explains how to test the MCP OpenAPI server using the MCP Inspector tool and manual requests.

## Using the MCP Inspector

The MCP Inspector provides a graphical interface for testing MCP servers. To use it:

1. Make sure you have Node.js installed
2. Run the provided test script:
   ```bash
   chmod +x test_with_inspector.sh
   ./test_with_inspector.sh
   ```

This will build the server and launch it with the MCP Inspector, which opens a browser interface.

## Testing the "hello" Tool

In the MCP Inspector interface:

1. Navigate to the "Tools" tab
2. Find the "hello" tool in the list
3. Click on it to view its details
4. Enter a value for the "submitter" parameter (e.g., "World")
5. Click "Execute" to run the tool
6. You should see a response: "Hello, World!"

## Testing the OpenAPI Validation Tool

In the MCP Inspector interface:

1. Navigate to the "Tools" tab
2. Find the "validate_openapi" tool in the list
3. Click on it to view its details
4. For the "spec" parameter, paste the contents of the `sample_openapi.json` file
5. Click "Execute" to run the tool
6. You should see validation results showing schema information

## Manual Testing with curl

You can also test the server manually using curl commands:

### Testing the "hello" Tool

```bash
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tool.call",
  "params": {
    "name": "hello",
    "arguments": {
      "submitter": "World"
    }
  }
}' | curl -s -X POST -H "Content-Type: application/json" --data @- http://localhost:8080/
```

### Testing the OpenAPI Validation Tool

```bash
echo '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tool.call",
  "params": {
    "name": "validate_openapi",
    "arguments": {
      "spec": "$(cat sample_openapi.json)"
    }
  }
}' | curl -s -X POST -H "Content-Type: application/json" --data @- http://localhost:8080/
```

Note: When using curl, you'll need to start the server separately without the inspector:

```bash
./mcpopenapi
```

## Expected Results

### Hello Tool Response

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

### OpenAPI Validation Response

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

## Troubleshooting

If you encounter issues:

1. Make sure the server is running
2. Check that you're using the correct format for tool arguments
3. Verify that the OpenAPI specification is valid JSON
4. Check the server logs for any error messages
