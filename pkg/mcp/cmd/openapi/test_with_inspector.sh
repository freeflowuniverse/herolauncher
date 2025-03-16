#!/bin/bash -e

NAME=mcpopenapi

cd "$(dirname "${BASH_SOURCE[0]}")"

# Run the build script
./build.sh

# Check if the build was successful
if [ $? -ne 0 ]; then
    echo "Failed to build $NAME server"
    exit 1
fi

echo "Starting $NAME server with MCP Inspector..."
echo "This will open a browser window with the MCP Inspector interface"
echo "Press Ctrl+C to stop the server when done"

# Use the full path to the binary
npx -y @modelcontextprotocol/inspector ${HOME}/hero/bin/${NAME}
