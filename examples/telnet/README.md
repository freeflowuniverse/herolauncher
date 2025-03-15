# Telnet Server Example

This example demonstrates how to use the telnet server package with the process manager to create a telnet interface for managing processes.

## Features

- Telnet server that listens on port 8023
- Authentication using a secret
- Command handling for process management
- Interactive and non-interactive modes

## Running the Example

To run this example, use the following command:

```bash
go run main.go
```

Then connect to the telnet server using:

```bash
telnet localhost 8023
```

When prompted, enter the secret: `mysecret`

## Available Commands

Once connected, you can use the following commands:

- `!!help` - Show help information
- `!!process.list` - List all processes
- `!!process.start` - Start a new process
- `!!process.stop` - Stop a running process
- `!!process.restart` - Restart a process
- `!!process.delete` - Delete a process
- `!!process.status` - Get status of a process
- `!!process.log` - View process logs

## Implementation Details

The example creates a process manager and a telnet adapter, then starts the telnet server on port 8023. The telnet adapter handles commands from clients and executes them using the process manager.

For a more detailed implementation, see the code in `main.go` and the telnet package documentation.
