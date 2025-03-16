#!/bin/bash 

pkill -f "modelcontextprotocol/inspector"

set -ex

NAME=mcpv

# Move to the directory of this script
cd "$(dirname "${BASH_SOURCE[0]}")"

# Test script for mcpv MCP server using the MCP Inspector
# This script will build the server and run it with the MCP Inspector

echo "Building mcpv server..."
./build.sh
# Check if the build was successful
if [ $? -ne 0 ]; then
    echo "Failed to build mcpv server"
    exit 1
fi

echo "Starting mcpv server with MCP Inspector..."
echo "This will open a browser window with the MCP Inspector interface"
echo "Press Ctrl+C to stop the server when done"

CLIENT_PORT=6000 SERVER_PORT=6001

# Run the server with the MCP Inspector 
npx -y @modelcontextprotocol/inspector ${HOME}/hero/bin/${NAME} "${HOME}/hero/bin/${NAME}"

