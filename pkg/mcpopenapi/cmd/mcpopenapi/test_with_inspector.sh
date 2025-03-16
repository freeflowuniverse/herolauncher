#!/bin/bash

# Test script for mcpopenapi MCP server using the MCP Inspector
# This script will build the server and run it with the MCP Inspector

echo "Building mcpopenapi server..."
go build

# Check if the build was successful
if [ $? -ne 0 ]; then
    echo "Failed to build mcpopenapi server"
    exit 1
fi

echo "Starting mcpopenapi server with MCP Inspector..."
echo "This will open a browser window with the MCP Inspector interface"
echo "Press Ctrl+C to stop the server when done"

# Run the server with the MCP Inspector
npx -y @modelcontextprotocol/inspector ./mcpopenapi

# Note: The MCP Inspector will automatically open a browser window
# You can interact with the server through the Inspector interface
