# Telnet Server Package

This package provides a clean abstraction for creating telnet servers that can operate in both interactive and non-interactive modes. It handles all the complexities of telnet protocol, terminal control, and ANSI color codes.

## Features

- **Dual Mode Support**: Works in both interactive (with colors) and non-interactive (plain text) modes
- **Terminal Handling**: Manages terminal control sequences, cursor movement, and line editing
- **Command History**: Provides command history navigation with up/down arrows
- **Color Support**: Includes primitives for colored text output that automatically adapts to the current mode
- **Authentication**: Built-in support for secret-based authentication
- **Clean API**: Simple, intuitive API for sending formatted text to clients

## Usage

### Creating a Server

```go
// Create a new telnet server with authentication and command handlers
server := telnet.NewServer(
    // Authentication handler
    func(secret string) bool {
        return secret == "mysecret" // Replace with actual secret
    },
    // Command handler
    handleCommand,
)

// Start the server
err := server.Start(":8023")
if err != nil {
    log.Fatalf("Failed to start telnet server: %v", err)
}
defer server.Stop()
```

### Handling Commands

```go
// handleCommand processes commands from clients
func handleCommand(client *telnet.Client, command string) error {
    // Process the command
    if command == "hello" {
        client.PrintlnGreen("Hello there!")
        return nil
    }
    
    // Unknown command
    client.PrintlnYellow(fmt.Sprintf("Unknown command: %s", command))
    return nil
}
```

### Client Output Methods

The `Client` struct provides various methods for sending text to the client:

```go
// Basic output
client.Write("Some text")
client.Println("A line of text")

// Colored output (automatically adapts to interactive/non-interactive mode)
client.PrintlnRed("Error message")
client.PrintlnGreen("Success message")
client.PrintlnYellow("Warning message")
client.PrintlnBlue("Information message")
client.PrintlnCyan("Highlighted message")
client.PrintlnPurple("Special message")
client.PrintlnBold("Important message")
```

### Formatting Utilities

The package includes utilities for formatting common output types:

```go
// Format a result with RESULT/ENDRESULT markers
result := telnet.FormatResult("Operation completed", "job-123", client.IsInteractive())

// Format an error message
errorMsg := telnet.FormatError(err, client.IsInteractive())

// Format a success message
successMsg := telnet.FormatSuccess("Operation succeeded", client.IsInteractive())

// Format a table
headers := []string{"Name", "Status", "Value"}
rows := [][]string{
    {"item1", "active", "100"},
    {"item2", "inactive", "50"},
}
table := telnet.FormatTable(headers, rows, client.IsInteractive())
```

## Interactive vs Non-Interactive Mode

- **Interactive Mode**: Includes ANSI color codes and formatting for better readability in terminal clients
- **Non-Interactive Mode**: Strips all ANSI codes for clean output when used programmatically or with non-terminal clients

Clients can toggle between modes using the built-in `!!interactive` or `!!i` command.

## Built-in Commands

The server automatically handles these commands:

- `!!help`, `?`, `h` - Show help text
- `!!interactive`, `!!i` - Toggle interactive mode
- `!!exit`, `!!quit`, `q`, `Ctrl+C` - Close the connection
- Up/Down arrows - Navigate command history
