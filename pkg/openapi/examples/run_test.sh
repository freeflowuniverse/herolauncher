#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

# Check if anything is running on port 9092
if command -v lsof >/dev/null 2>&1; then
    # OSX
    if [ "$(lsof -i :9092 | wc -l)" -gt 0 ]; then
        echo "Killing process running on port 9092"
        lsof -i :9092 | awk 'NR!=1 {print $2}' | xargs kill
    fi
elif command -v fuser >/dev/null 2>&1; then
    # Linux
    if [ "$(fuser 9092/tcp | wc -l)" -gt 0 ]; then
        echo "Killing process running on port 9092"
        fuser -k 9092/tcp
    fi
fi


# Run the OpenAPI test
echo "Running OpenAPI test..."
cd "$(dirname "$0")"

# Ensure directories exist
mkdir -p petstoreapi actionsapi

# Run the tests
echo "Running tests..."
cd test
go test -v
cd ..

# Run the server
echo -e "\nStarting the multi-API server..."
go run main.go
