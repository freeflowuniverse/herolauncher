#!/bin/bash
set -e

# Change to the script's directory to ensure relative paths work
cd "$(dirname "$0")"

echo "Building Hetzner Installer for Linux on AMD64..."

# Create build directory if it doesn't exist
mkdir -p build

# Build the Hetzner installer binary
echo "Building Hetzner installer..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -trimpath \
  -o build/hetzner_installer \
  main.go # Reference main.go in the current directory

# Set executable permissions
chmod +x build/hetzner_installer

# Output binary info
echo "Build complete!"
ls -lh build/
