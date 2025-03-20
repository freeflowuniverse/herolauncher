#!/bin/bash

# Set architecture to 64-bit to avoid 32-bit compilation issues
export GOARCH=amd64

# Initialize debug flag to false
DEBUG=false

# Check for --debug flag
for arg in "$@"; do
    if [ "$arg" = "--debug" ]; then
        DEBUG=true
        break
    fi
done

# Function to print debug messages
debug_echo() {
    if [ "$DEBUG" = true ]; then
        echo "$@"
    fi
}

# Array of module paths to exclude from running tests
# Paths are relative to the root of the repo
EXCLUDED_MODULES=(
    # "pkg/doctree"
    "pkg/executor"
    "pkg/handlerfactory"
    "pkg/herolauncher"
    "pkg/heroscript"
    "pkg/imapserver"
    "pkg/mcp"
    "pkg/openapi"
    "pkg/ourdb"
    "pkg/packagemanager"
    "pkg/processmanager"
    "pkg/radixtree"
    "pkg/redisserver"
    "pkg/smtpserver"
    "pkg/system"
    "pkg/telnetserver"
    "pkg/tools"
    "pkg/ui"
    "pkg/vfs"
    "pkg/vlang"
    "pkg/vm"
    "pkg/webdavserver"
    "pkg/wire"
)

# Function to join array elements with a delimiter
join_array() {
    local IFS="$1"
    shift
    echo "$*"
}

# Debug: Print the excluded modules
debug_echo "Excluded modules:"
if [ "$DEBUG" = true ]; then
    for module in "${EXCLUDED_MODULES[@]}"; do
        echo "  $module"
    done
fi

# Convert excluded modules into a grep -v pattern (e.g., "pkg/doctree\|pkg/executor")
EXCLUDE_PATTERN=$(join_array "|" "${EXCLUDED_MODULES[@]}")
debug_echo "Exclude pattern: $EXCLUDE_PATTERN"

# Find all Go test files (*.go files containing "Test" in their name or content)
ALL_TEST_FILES=$(find . -type f -name "*_test.go")
debug_echo "All test files found:"
if [ "$DEBUG" = true ]; then
    echo "$ALL_TEST_FILES"
fi

# Exclude the modules specified in EXCLUDED_MODULES
# Add word boundaries to ensure exact directory matching
TEST_FILES=$(echo "$ALL_TEST_FILES" | grep -v -E "($(join_array "|" "${EXCLUDED_MODULES[@]}"))(/|$)")

# Check if there are any test files to run
if [ -z "$TEST_FILES" ]; then
    echo "No test files found after excluding specified modules."
    exit 0
fi

# Run go test on the remaining modules
debug_echo "Running tests for the following files:"
if [ "$DEBUG" = true ]; then
    echo "$TEST_FILES"
fi
echo "---------------------------------------"

# Convert test files to package paths and run go test
for file in $TEST_FILES; do
    # Get the directory of the test file (package path)
    pkg=$(dirname "$file")
    if [ "$DEBUG" = true ]; then
        echo "Testing package: $pkg"
    fi
    go test -v "$pkg"
done

echo "---------------------------------------"
echo "All applicable tests completed."
