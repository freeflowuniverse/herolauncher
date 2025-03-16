# Testing the MCP V Language Specs Server

This guide explains how to test the MCP V Language Specs server using the MCP Inspector tool and manual requests.

## Using the MCP Inspector

The MCP Inspector provides a graphical interface for testing MCP servers. To use it:

1. Make sure you have Node.js installed
2. Run the provided test script:
   ```bash
   chmod +x test_with_inspector.sh
   ./test_with_inspector.sh
   ```

This will build the server and launch it with the MCP Inspector, which opens a browser interface.

## Testing the "get_specs" Tool

In the MCP Inspector interface:

1. Navigate to the "Tools" tab
2. Find the "get_specs" tool in the list
3. Click on it to view its details
4. Enter a value for the "path" parameter (e.g., "/path/to/vlang/files")
5. Click "Execute" to run the tool
6. You should see a response with the extracted V language specifications

## Manual Testing with curl

You can also test the server manually using curl commands:

### Testing the "get_specs" Tool

```bash
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tool.call",
  "params": {
    "name": "get_specs",
    "arguments": {
      "path": "/path/to/vlang/files"
    }
  }
}' | curl -s -X POST -H "Content-Type: application/json" --data @- http://localhost:8080/
```

Note: When using curl, you'll need to start the server separately without the inspector:

```bash
./mcpv
```

## Expected Results

### Get Specs Tool Response

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

## Troubleshooting

If you encounter issues:

1. Make sure the server is running
2. Check that you're using the correct format for tool arguments
3. Verify that the path to V language files exists
4. Check the server logs for any error messages

## Sample V Language Files

You can use the sample V language files provided in the `sample_vlang` directory for testing purposes. These files contain various V language constructs like structs, enums, and methods that can be processed by the server.
