#!/bin/bash

# Navigate to the vfs9p directory
cd /Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/vfs/vfs9p

# Build and run the vfs9p server
echo "Building and running the vfs9p server..."
go run cmd/main.go -addr=0.0.0.0:5640
