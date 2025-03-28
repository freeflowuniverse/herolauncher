#!/bin/bash
set -e

# Change to the script's directory to ensure relative paths work
cd "$(dirname "$0")"

echo "Building PostgreSQL Builder for Linux on AMD64..."

# Create build directory if it doesn't exist
mkdir -p build

# Build the PostgreSQL builder
echo "Building PostgreSQL builder..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -trimpath \
  -o build/postgresql_builder \
  ../cmd/main.go

# Set executable permissions
chmod +x build/postgresql_builder

# Output binary info
echo "Build complete!"
ls -lh build/
