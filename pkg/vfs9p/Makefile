# Makefile for VFS9P package

.PHONY: all build test clean

# Default target
all: build

# Build the command-line tool
build:
	go build -o bin/vfs9p-server ./cmd

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run the server with default settings
run: build
	./bin/vfs9p-server

# Install the command-line tool
install:
	go install ./cmd

# Generate documentation
doc:
	godoc -http=:6060

# Help target
help:
	@echo "Available targets:"
	@echo "  all      - Build the command-line tool (default)"
	@echo "  build    - Build the command-line tool"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run the server with default settings"
	@echo "  install  - Install the command-line tool"
	@echo "  doc      - Generate documentation"
	@echo "  help     - Show this help message"
