#!/bin/bash

# Test script for WebDAV server with all features
# This script demonstrates how to use the WebDAV server with debug mode, authentication, and HTTPS

# Create test directory for WebDAV files
TEST_DIR="/tmp/webdav_test"
mkdir -p "$TEST_DIR"
echo "Test content" > "$TEST_DIR/test.txt"

# Generate self-signed certificate for HTTPS testing if it doesn't exist
CERT_DIR="./certs"
if [ ! -f "$CERT_DIR/webdav.crt" ] || [ ! -f "$CERT_DIR/webdav.key" ]; then
    echo "Generating self-signed certificate for HTTPS testing..."
    ./scripts/generate_cert.sh
fi

# Function to display section headers
section() {
    echo ""
    echo "====================================="
    echo "  $1"
    echo "====================================="
}

# Function to run a test case
run_test() {
    local name="$1"
    local cmd="$2"
    local connect_cmd="$3"
    
    section "TEST: $name"
    echo "Running command: $cmd"
    
    # Run the WebDAV server in the background
    eval "$cmd" &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 2
    
    # Display connection command
    if [ -n "$connect_cmd" ]; then
        echo ""
        echo "To connect to this server, run:"
        echo "$connect_cmd"
        echo ""
        echo "Press Enter to continue to the next test..."
        read
    else
        echo "Server is running. Press Enter to continue to the next test..."
        read
    fi
    
    # Kill the server
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null
    echo "Server stopped."
}

# Test 1: Basic WebDAV server
run_test "Basic WebDAV Server" \
    "./bin/webdavserver -fs $TEST_DIR" \
    "./scripts/open_webdav_osx.sh"

# Test 2: WebDAV with debug mode
run_test "WebDAV Server with Debug Mode" \
    "./bin/webdavserver -fs $TEST_DIR -debug" \
    "./scripts/open_webdav_osx.sh"

# Test 3: WebDAV with authentication
run_test "WebDAV Server with Authentication" \
    "./bin/webdavserver -fs $TEST_DIR -auth -username testuser -password testpass" \
    "./scripts/open_webdav_osx.sh -u testuser -pw testpass"

# Test 4: WebDAV with HTTPS
run_test "WebDAV Server with HTTPS" \
    "./bin/webdavserver -fs $TEST_DIR -https -cert $CERT_DIR/webdav.crt -key $CERT_DIR/webdav.key" \
    "./scripts/open_webdav_osx.sh -s"

# Test 5: WebDAV with all features
run_test "WebDAV Server with All Features" \
    "./bin/webdavserver -fs $TEST_DIR -debug -auth -username testuser -password testpass -https -cert $CERT_DIR/webdav.crt -key $CERT_DIR/webdav.key" \
    "./scripts/open_webdav_osx.sh -s -u testuser -pw testpass"

section "All tests completed"
echo "The WebDAV server has been tested with all features:"
echo "- Basic functionality"
echo "- Debug mode"
echo "- Authentication"
echo "- HTTPS support"
echo ""
echo "To run the WebDAV server in production mode, use:"
echo "./bin/webdavserver -fs /path/to/your/files -auth -username your_username -password your_password -https -cert /path/to/cert.pem -key /path/to/key.pem"
