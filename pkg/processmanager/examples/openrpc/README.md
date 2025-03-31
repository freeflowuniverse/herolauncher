# Process Manager OpenRPC Example

This example demonstrates how to use the Process Manager OpenRPC interface to interact with the process manager. It provides a complete working example of both server and client implementations, with a mock process manager for testing purposes.

## Overview

The example includes:

1. A server that exposes the process manager functionality via OpenRPC over a Unix socket
2. A client that connects to the server and performs various operations
3. A mock process manager implementation for testing without requiring actual processes

## Structure

- `main.go`: The main entry point that runs both the server and client in the same process
- `mock_processmanager.go`: A mock implementation of the `ProcessManagerInterface` that simulates process operations
- `example_client.go`: Client implementation demonstrating how to use the OpenRPC client to interact with the process manager

## How to Run

```bash
# From the herolauncher root directory
go run ./pkg/processmanager/examples/openrpc
```

The example will:
1. Start a server using a Unix socket at `/tmp/process-manager.sock`
2. Run a series of client operations against the server
3. Display the results of each operation
4. Clean up and shut down when complete

## What It Demonstrates

This example shows:

1. How to initialize and start an OpenRPC server for the process manager
2. How to create and use an OpenRPC client to interact with the process manager
3. How to perform common operations like:
   - Starting a process (`StartProcess`)
   - Getting process status (`GetProcessStatus`)
   - Listing all processes (`ListProcesses`)
   - Getting process logs (`GetProcessLogs`)
   - Restarting a process (`RestartProcess`)
   - Stopping a process (`StopProcess`)
   - Deleting a process (`DeleteProcess`)
4. How to handle the response types returned by each operation
5. Proper error handling for RPC operations

## Notes

- This example uses a mock process manager implementation for demonstration purposes
- In a real application, you would use the actual process manager implementation
- The server and client run in the same process for simplicity, but they could be in separate processes
- The Unix socket communication can be replaced with other transport mechanisms if needed

## Implementation Details

### Server

The server is implemented using the `openrpc.Server` type, which wraps the OpenRPC manager and Unix socket server. It:

1. Creates a Unix socket at the specified path
2. Registers handlers for each RPC method
3. Authenticates requests using a secret
4. Marshals/unmarshals JSON-RPC requests and responses

### Client

The client is implemented using the `openrpc.Client` type, which provides methods for each operation. It:

1. Connects to the Unix socket
2. Sends JSON-RPC requests with the appropriate parameters
3. Handles authentication using the secret
4. Parses the responses into appropriate result types

### Mock Process Manager

The mock process manager implements the `interfaces.ProcessManagerInterface` and simulates:

1. Process creation and management
2. Status tracking
3. Log collection
4. Error handling

This allows testing the OpenRPC interface without requiring actual processes to be run.
