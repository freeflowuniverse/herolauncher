# HeroHandler Example

This package demonstrates how to implement and use a handler for HeroScript in the HeroLauncher project.

## Overview

The HeroHandler example provides a simple key-value store implementation that showcases how to:

1. Create a custom handler that extends the base handler functionality
2. Implement action methods that can be called via HeroScript
3. Parse parameters from HeroScript actions
4. Process and execute HeroScript commands

## Project Structure

```
./herohandler/
├── README.md           # This documentation file
├── main.go             # Main executable that uses the example handler
└── internal/           # Internal package for the example handler implementation
    └── example_handler.go  # Example handler implementation
```

## Handler Actions

The example handler supports the following actions:

- `example.set`: Store a key-value pair
  - Parameters: `key`, `value`
- `example.get`: Retrieve a value by key
  - Parameters: `key`
- `example.list`: List all stored key-value pairs
  - No parameters
- `example.delete`: Remove a key-value pair
  - Parameters: `key`

## Usage

You can run the example handler using the provided `main.go`:

```bash
# Build the binary
cd pkg/heroscript/cmd/herohandler
go build -o herohandler

# Set a key-value pair
./herohandler "example.set key:mykey value:myvalue"

# Get a value by key
./herohandler "example.get key:mykey"

# List all stored key-value pairs
./herohandler "example.list"

# Delete a key-value pair
./herohandler "example.delete key:mykey"
```

### Important Note on State Persistence

The example handler maintains its key-value store in memory only for the duration of a single command execution. Each time you run the `herohandler` command, a new instance of the handler is created with an empty data store. This is by design to keep the example simple.

In a real-world application, you would typically implement persistence using a database, file storage, or other mechanisms to maintain state between command executions.

### Multi-Command Example

To execute multiple commands in a single script, you can create a HeroScript file and pass it to the handler. For example:

```bash
# Create a script file
cat > example.hero << EOF
!!example.set key:user value:john
!!example.set key:role value:admin
!!example.list
EOF

# Run the script
cat example.hero | ./herohandler
```

This would process all commands in a single execution, allowing the in-memory state to be shared between commands.

## Implementation Details

The example handler demonstrates several important concepts:

1. **Handler Structure**: The `ExampleHandler` extends the `handlers.BaseHandler` which provides common functionality for all handlers.

2. **Action Methods**: Each action is implemented as a method on the handler struct (e.g., `Set`, `Get`, `List`, `Delete`).

3. **Parameter Parsing**: The `ParseParams` method from `BaseHandler` is used to extract parameters from HeroScript.

4. **Action Execution**: The `Play` method from `BaseHandler` uses reflection to find and call the appropriate method based on the action name.

5. **In-Memory Storage**: The example handler maintains a simple in-memory key-value store using a map.

## Extending the Example

To create your own handler:

1. Create a new struct that embeds the `handlers.BaseHandler`
2. Implement methods for each action your handler will support
3. Create a constructor function that initializes your handler with the appropriate actor name
4. Use the `Play` method to process HeroScript commands

For more complex handlers, you might need to add additional fields to store state or configuration.
