#!/bin/bash

# Test script for WebDAV server with auto-certificate generation
# This script demonstrates how the WebDAV server automatically generates certificates when needed

# Create test directory for WebDAV files
TEST_DIR="/tmp/herolauncher_test"
CERT_DIR="/tmp/certificates"

# Clean up any existing test directories
if [ -d "$TEST_DIR" ]; then
    echo "Removing existing test directory to start fresh..."
    rm -rf "$TEST_DIR"
fi

# Clean up any existing certificates
if [ -d "$CERT_DIR" ]; then
    echo "Removing existing certificates to test auto-generation..."
    rm -rf "$CERT_DIR"
fi

# Create test directory and add a test file
mkdir -p "$TEST_DIR"
echo "Test content" > "$TEST_DIR/test.txt"

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

# Test 1: WebDAV with HTTPS and auto-generated certificates
run_test "WebDAV Server with Auto-Generated Certificates" \
    "./bin/webdavserver -fs $TEST_DIR -debug -https" \
    "./scripts/open_webdav_osx.sh -s"

# Test 2: Check if certificates were generated
section "Checking Generated Certificates"
if [ -f "$CERT_DIR/webdav.crt" ] && [ -f "$CERT_DIR/webdav.key" ]; then
    echo "✅ Certificates were successfully auto-generated:"
    echo "   - $CERT_DIR/webdav.crt"
    echo "   - $CERT_DIR/webdav.key"
    
    # Display certificate information
    echo ""
    echo "Certificate information:"
    openssl x509 -in "$CERT_DIR/webdav.crt" -text -noout | grep -E "Subject:|Issuer:|Not Before:|Not After :|DNS:"
else
    echo "❌ Certificates were not generated properly"
fi

# Test 3: WebDAV with HTTPS using the generated certificates
run_test "WebDAV Server with Previously Generated Certificates" \
    "./bin/webdavserver -fs $TEST_DIR -debug -https -cert $CERT_DIR/webdav.crt -key $CERT_DIR/webdav.key" \
    "./scripts/open_webdav_osx.sh -s"

# Test 4: WebDAV with HTTPS, authentication and custom certificate settings
run_test "WebDAV Server with Custom Certificate Settings" \
    "./bin/webdavserver -fs $TEST_DIR -debug -https -auth -username testuser -password testpass -cert-validity 30 -cert-org \"Test Organization\"" \
    "./scripts/open_webdav_osx.sh -s -u testuser -pw testpass"

section "All tests completed"
echo "The WebDAV server has been tested with auto-certificate generation:"
echo "- Basic auto-generation"
echo "- Reusing generated certificates"
echo "- Custom certificate settings"
echo ""
echo "For production use, consider:"
echo "1. Using properly signed certificates instead of self-signed ones"
echo "2. Setting a longer validity period (default is 365 days)"
echo "3. Always enabling authentication with strong credentials"
