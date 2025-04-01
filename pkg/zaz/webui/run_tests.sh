#!/bin/bash

# Check if the server is running
if ! curl -s http://localhost:9030/ > /dev/null; then
  echo "❌ Server is not running on port 9030. Please start the server before running tests."
  echo "Run: cd /Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/zaz/cmd/webui && go run main.go"
  exit 1
fi

# Run the tests
cd /Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/zaz/webui
go test -v -run TestGetEndpoints
go test -v -run TestPostEndpoints

# Check if tests passed
if [ $? -eq 0 ]; then
  echo "✅ All tests passed!"
else
  echo "❌ Some tests failed."
fi
