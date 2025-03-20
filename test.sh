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
    # Uncomment modules you want to exclude
    # "pkg/doctree"
    # "pkg/executor"
    # "pkg/handlerfactory"
    # "pkg/herolauncher"
    # "pkg/heroscript"
    # "pkg/imapserver"
    # "pkg/mcp"
    # "pkg/openapi"
    # "pkg/ourdb"
    # "pkg/packagemanager"
    # "pkg/processmanager"
    # "pkg/radixtree"
    # "pkg/redisserver"
    # "pkg/smtpserver"
    # "pkg/system"
    # "pkg/telnetserver"
    # "pkg/tools"
    # "pkg/ui"
    # "pkg/vfs"
    # "pkg/vlang"
    # "pkg/vm"
    # "pkg/webdavserver"
    # "pkg/wire"
)

# Function to join array elements with a delimiter
join_array() {
    local IFS="$1"
    shift
    echo "$*"
}

# Debug: Print the excluded modules
debug_echo "Excluded modules:"
debug_echo "Array size: ${#EXCLUDED_MODULES[@]}"
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
if [ ${#EXCLUDED_MODULES[@]} -eq 0 ]; then
    # If no modules are excluded, use all test files
    TEST_FILES="$ALL_TEST_FILES"
    debug_echo "No modules excluded, using all test files"
else
    # Otherwise, exclude the specified modules
    TEST_FILES=$(echo "$ALL_TEST_FILES" | grep -v -E "($(join_array "|" "${EXCLUDED_MODULES[@]}"))(/|$)")
    debug_echo "Excluding specified modules"
fi

# Check if there are any test files to run
if [ -z "$TEST_FILES" ]; then
    echo "No test files found after excluding specified modules."
    exit 0
fi

# Initialize counters for test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

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
    
    # Run the test and capture the output
    TEST_OUTPUT=$(go test -v "$pkg")
    TEST_RESULT=$?
    
    # Display the test output
    echo "$TEST_OUTPUT"
    
    # Count the number of tests run, passed, and failed
    TESTS_RUN=$(echo "$TEST_OUTPUT" | grep -c "=== RUN")
    TOTAL_TESTS=$((TOTAL_TESTS + TESTS_RUN))
    
    if [ $TEST_RESULT -eq 0 ]; then
        # All tests passed
        PASSED_TESTS=$((PASSED_TESTS + TESTS_RUN))
    else
        # Count failed tests - use grep with a pattern that escapes the hyphens
        TESTS_FAILED=$(echo "$TEST_OUTPUT" | grep -c "\-\-\- FAIL")
        FAILED_TESTS=$((FAILED_TESTS + TESTS_FAILED))
        PASSED_TESTS=$((PASSED_TESTS + TESTS_RUN - TESTS_FAILED))
    fi
done

echo "---------------------------------------"
echo "TEST SUMMARY"
echo "---------------------------------------"
echo "Total Tests: $TOTAL_TESTS"
echo "Passed Tests: $PASSED_TESTS"
echo "Failed Tests: $FAILED_TESTS"
echo "---------------------------------------"
echo "All applicable tests completed."
