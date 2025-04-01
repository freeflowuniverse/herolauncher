#!/bin/bash

# Generate Swagger documentation for HeroLauncher API
# This script ensures all API endpoints are properly documented

# Ensure swag is installed
if ! command -v /root/go/bin/swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Navigate to project root from the script location
cd "$(dirname "$0")/../../.."
PROJECT_ROOT=$(pwd)

# Set the correct paths for docs generation
API_DIR="$PROJECT_ROOT/pkg/herolauncher/api"
DOCS_DIR="$PROJECT_ROOT/pkg/herolauncher/docs"

# Remove existing docs
rm -rf "$DOCS_DIR"
mkdir -p "$DOCS_DIR"

echo "Project root: $PROJECT_ROOT"
echo "API directory: $API_DIR"
echo "Docs directory: $DOCS_DIR"

# Generate Swagger documentation with all API endpoints
echo "Generating Swagger documentation..."

# First, find all API files with @Router annotations to ensure they're included
echo "Scanning API files for Swagger annotations..."
API_FILES_WITH_ROUTES=$(grep -l "@Router" --include="*.go" -r "$API_DIR")
echo "Found $(echo "$API_FILES_WITH_ROUTES" | wc -l) files with API routes"

# Use the main.go file in the api directory as the general info source
# The --output flag specifies where to output the generated files
/root/go/bin/swag init \
  -g main.go \
  -d "$API_DIR" \
  -o "$DOCS_DIR" \
  --parseDependency \
  --parseInternal \
  --parseDepth 3 \
  --generatedTime \
  --requiredByDefault \
  --parseGoList

# Check if Swagger documentation was generated successfully
if [ -f "$DOCS_DIR/swagger.json" ]; then
    echo "Swagger documentation generated successfully!"
    echo "You can access the Swagger UI at http://localhost:9001/swagger/ when the server is running."
    
    # Count the number of endpoints in the generated swagger.json
    ENDPOINT_COUNT=$(grep -c '"get"\|"post"\|"put"\|"delete"' "$DOCS_DIR/swagger.json")
    echo "Generated documentation for $ENDPOINT_COUNT API endpoints."
    
    # List the API endpoints that were documented
    echo "\nAPI Endpoints documented:"
    grep -o '"\(/[^"]*\)"' "$DOCS_DIR/swagger.json" | sort | uniq | sed 's/"//g'
    
    # Check if we're missing any endpoints from the API files
    echo "\nChecking for potentially missing endpoints..."
    MISSING_ROUTES=$(grep "@Router" --include="*.go" -r "$API_DIR" | grep -v "$(grep -o '"\(/[^"]*\)"' "$DOCS_DIR/swagger.json" | tr '\n' '|' | sed 's/|$//')")
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

# Ensure the package name in docs.go is correct
if [ -f "$DOCS_DIR/docs.go" ]; then
    echo "Fixing package name in docs.go..."
    sed -i 's/package docs/package docs/' "$DOCS_DIR/docs.go"
    echo "Package name fixed."
fi
