# Handler Factory Module

The Handler Factory module provides a framework for creating and managing handlers that process HeroScript commands through a telnet interface. It allows for both Unix socket and TCP connections, making it flexible for different use cases.

## Overview

The Handler Factory module consists of several components:

1. **HandlerFactory**: Core component that manages a collection of handlers, each responsible for processing specific actor commands.
2. **TelnetServer**: Provides a telnet interface for interacting with the handlers, supporting both Unix socket and TCP connections.
3. **HeroHandler**: Main handler that initializes and manages the HandlerFactory and TelnetServer.
4. **ProcessManagerHandler**: Example handler implementation that manages processes through HeroScript commands.

## Architecture

The module follows a plugin-based architecture where:

- The `HandlerFactory` maintains a registry of handlers
- Each handler implements the `Handler` interface and is responsible for a specific actor
- The `TelnetServer` provides a communication interface to send HeroScript commands
- HeroScript commands are parsed and routed to the appropriate handler based on the actor name

## Connecting to the Handler Factory

The Handler Factory exposes two interfaces for communication:

1. Unix Socket (default: `/tmp/hero.sock`)
2. TCP Port (default: `localhost:8023`)

to get started

```bash
cd /root/code/github/freeflowuniverse/herolauncher/pkg/handlerfactory/herohandler/cmd
go run .
```

### Using Telnet to Connect

You can use the standard telnet client to connect to the TCP port:

```bash
# Connect to the default TCP port
telnet localhost 8023

# Once connected, you can send HeroScript commands
# For example:
!!process.list
```

### Using Netcat to Connect

Netcat (nc) can be used to connect to both the Unix socket and TCP port:

#### Connecting to TCP Port

```bash
# Connect to the TCP port
nc localhost 8023

# Send HeroScript commands
!!process.list
```

#### Connecting to Unix Socket

```bash
# Connect to the Unix socket
nc -U /tmp/hero.sock

# Send HeroScript commands
!!process.list
```

## HeroScript Command Format

Commands follow the HeroScript format:

```
!!actor.action param1:"value1" param2:"value2"
```

For example:

```
!!process.start name:"web_server" command:"python -m http.server 8080" log:true
!!process.status name:"web_server"
!!process.stop name:"web_server"
```

## Available Commands

The Handler Factory comes with a built-in ProcessManagerHandler that supports the following commands:

- `!!process.start` - Start a new process
- `!!process.stop` - Stop a running process
- `!!process.restart` - Restart a process
- `!!process.delete` - Delete a process
- `!!process.list` - List all processes
- `!!process.status` - Get status of a specific process
- `!!process.logs` - Get logs of a specific process
- `!!process.help` - Show help information

You can get help on available commands by typing `!!help` in the telnet/netcat session.

## Authentication

The telnet server supports optional authentication with secrets. If secrets are provided when starting the server, clients will need to authenticate using one of these secrets before they can execute commands.

## Extending with Custom Handlers

You can extend the functionality by implementing your own handlers:

1. Create a new handler that implements the `Handler` interface
2. Register the handler with the HandlerFactory
3. Access the handler's functionality through the telnet interface

## Example Usage

Here's a complete example of connecting and using the telnet interface:

```bash
# Connect to the telnet server
nc localhost 8023

# List all processes
!!process.list

# Start a new process
!!process.start name:"web_server" command:"python -m http.server 8080"

# Check the status of the process
!!process.status name:"web_server"

# View the logs
!!process.logs name:"web_server" lines:50

# Stop the process
!!process.stop name:"web_server"
```

## Implementation Details

The Handler Factory module is implemented in pure Go and follows the Go project structure conventions. It uses standard Go libraries for networking and does not have external dependencies for its core functionality.
