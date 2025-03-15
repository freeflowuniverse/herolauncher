#!/bin/bash

# works.sh - Script to verify that all components of the vfsdav package work correctly
# This script runs all tests, examples, and commands and exits with an appropriate status code

set -e  # Exit immediately if a command exits with a non-zero status
set -o pipefail  # Return the exit status of the last command in the pipe that failed

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        exit $1
    fi
}

# Function to run a command and check its exit status
run_check() {
    echo -e "${YELLOW}Running: $1${NC}"
    eval "$1"
    local status=$?
    print_status $status "$1"
    return $status
}

# Get the root directory of the project
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(cd "$PKG_DIR/../.." && pwd)"

echo "=== Verifying vfsdav package ==="
echo "Package directory: $PKG_DIR"
echo "Root directory: $ROOT_DIR"

# Step 1: Run the unit tests
echo -e "\n${YELLOW}Step 1: Running unit tests${NC}"
cd "$PKG_DIR"
run_check "go test -v ./tests"

# Step 2: Run the basic example
echo -e "\n${YELLOW}Step 2: Running basic example (in background)${NC}"
cd "$PKG_DIR/examples/goclient"
# Modify the code to use a different port
sed -i.bak 's/addr := "localhost:8080"/addr := "localhost:8091"/' main.go
# Run the example in the background and save the PID
go run main.go &
EXAMPLE_PID=$!
# Give it a moment to start
sleep 2

# Check if the server is running by making a request
curl -s -o /dev/null -w "%{http_code}" http://localhost:8091/sample.txt
CURL_STATUS=$?
if [ $CURL_STATUS -eq 0 ]; then
    print_status 0 "Basic example is running"
else
    print_status 1 "Basic example failed to start"
fi

# Kill the example server
kill $EXAMPLE_PID 2>/dev/null || true
sleep 1
# Restore the original file
mv main.go.bak main.go 2>/dev/null || true

# Step 3: Run the client example
echo -e "\n${YELLOW}Step 3: Running client example${NC}"
cd "$PKG_DIR/examples/goclient/client"
# Modify the code to use a different port
sed -i.bak 's/addr := "localhost:8080"/addr := "localhost:8092"/' main.go
run_check "go run main.go"
# Restore the original file
mv main.go.bak main.go 2>/dev/null || true

# Step 4: Run the rclone example if rclone is installed
echo -e "\n${YELLOW}Step 4: Running rclone example${NC}"
if command -v rclone >/dev/null 2>&1; then
    cd "$PKG_DIR/examples/rclone"
    # Modify the code to use a different port
    sed -i.bak 's/addr := "localhost:8080"/addr := "localhost:8093"/' main.go
    run_check "go run main.go"
    # Restore the original file
    mv main.go.bak main.go 2>/dev/null || true
else
    echo -e "${YELLOW}Skipping rclone example - rclone not installed${NC}"
fi

# Step 5: Run the command-line tool
echo -e "\n${YELLOW}Step 5: Running command-line tool (in background)${NC}"
cd "$PKG_DIR/cmd"
# Run the command-line tool in the background
go run main.go --port 8082 --host localhost &
CMD_PID=$!
# Give it a moment to start
sleep 2

# Check if the server is running by making a request
curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/hello.txt
CURL_STATUS=$?
if [ $CURL_STATUS -eq 0 ]; then
    print_status 0 "Command-line tool is running"
else
    print_status 1 "Command-line tool failed to start"
fi

# Kill the command-line tool
kill $CMD_PID 2>/dev/null || true
sleep 1

echo -e "\n${GREEN}All tests passed! The vfsdav package is working correctly.${NC}"
exit 0
