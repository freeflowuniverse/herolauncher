#!/bin/bash
set -e

# Colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

PROJECT_ROOT=$(pwd)
SERVER_MAIN="$PROJECT_ROOT/cmd/server/main.go"
OUTPUT_BIN="$PROJECT_ROOT/bin/herolauncher"

echo -e "${YELLOW}Starting build process for HeroLauncher...${NC}"

# Create bin directory if it doesn't exist
mkdir -p "$PROJECT_ROOT/bin"

# Make sure dependencies are up to date
echo -e "${YELLOW}Updating dependencies...${NC}"
go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to update dependencies${NC}"
    exit 1
fi

# Run linter
echo -e "${YELLOW}Running linter...${NC}"
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}Linting issues found. Attempting to fix automatically...${NC}"
        golangci-lint run --fix ./...
    fi
else
    echo -e "${YELLOW}golangci-lint not found, skipping linting. Consider installing it with:${NC}"
    echo -e "${YELLOW}go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest${NC}"
fi

# Build the server
echo -e "${YELLOW}Building server...${NC}"
go build -o "$OUTPUT_BIN" "$SERVER_MAIN"
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"
echo -e "${GREEN}Binary located at: $OUTPUT_BIN${NC}"

# Run the server
echo -e "${YELLOW}Starting server...${NC}"
"$OUTPUT_BIN"
