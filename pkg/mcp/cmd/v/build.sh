#!/bin/bash -e

NAME=mcpv

cd "$(dirname "${BASH_SOURCE[0]}")"

# Build script for $NAME MCP server
# This script builds the server and installs it to the hero/bin directory

# Create hero/bin directory if it doesn't exist
mkdir -p ${HOME}/hero/bin

echo "Building $NAME server..."

if [ -f ${HOME}/hero/bin/$NAME ]; then
  echo "A previous version of $NAME has been deleted at ${HOME}/hero/bin/$NAME"
  rm ${HOME}/hero/bin/$NAME
fi

# Build the server
go build -o $NAME

# Copy to hero/bin directory
mv $NAME ${HOME}/hero/bin/


echo "$NAME server built and installed to ${HOME}/hero/bin/"
