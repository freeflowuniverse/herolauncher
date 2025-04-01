#!/bin/bash

# Generate Swagger documentation for HeroLauncher API
# This script ensures all API endpoints are properly documented

# Ensure swag is installed
if ! command -v /root/go/bin/swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Navigate to project root
cd "$(dirname "$0")/.."

# Remove existing docs
rm -rf docs
mkdir -p docs

# Generate Swagger documentation with all API endpoints
echo "Generating Swagger documentation..."
/root/go/bin/swag init \
  -d api \
  -o docs \
  --parseDependency \
  --parseInternal

echo "Swagger documentation generated successfully!"
echo "You can access the Swagger UI at http://localhost:9001/swagger/ when the server is running."
