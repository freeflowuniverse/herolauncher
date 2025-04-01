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

# First, find all API files with @Router annotations to ensure they're included
echo "Scanning API files for Swagger annotations..."
API_FILES_WITH_ROUTES=$(grep -l "@Router" --include="*.go" -r api/)
echo "Found $(echo "$API_FILES_WITH_ROUTES" | wc -l) files with API routes"

# Use the docs.go file in the api directory as the main source for Swagger documentation
/root/go/bin/swag init \
  -g docs.go \
  -d api \
  -o docs \
  --parseDependency \
  --parseInternal \
  --parseDepth 3 \
  --generatedTime \
  --requiredByDefault \
  --parseGoList \
  --outputTypes json,yaml

# Check if Swagger documentation was generated successfully
if [ -f "docs/swagger.json" ]; then
    echo "Swagger documentation generated successfully!"
    echo "You can access the Swagger UI at http://localhost:9001/swagger/ when the server is running."
    
    # Count the number of endpoints in the generated swagger.json
    ENDPOINT_COUNT=$(grep -c '"get"\|"post"\|"put"\|"delete"' docs/swagger.json)
    echo "Generated documentation for $ENDPOINT_COUNT API endpoints."
    
    # List the API endpoints that were documented
    echo "\nAPI Endpoints documented:"
    grep -o '"\(/[^"]*\)"' docs/swagger.json | sort | uniq | sed 's/"//g'
    
    # Check if we're missing any endpoints from the API files
    echo "\nChecking for potentially missing endpoints..."
    MISSING_ROUTES=$(grep "@Router" --include="*.go" -r api/ | grep -v "$(grep -o '"\(/[^"]*\)"' docs/swagger.json | tr '\n' '|' | sed 's/|$//')")
    if [ -n "$MISSING_ROUTES" ]; then
        echo "Warning: Some routes might be missing from the Swagger documentation:"
        echo "$MISSING_ROUTES"
    else
        echo "All routes with @Router annotations appear to be documented."
    fi
else
    echo "Error: Swagger documentation generation failed."
    exit 1
fi
