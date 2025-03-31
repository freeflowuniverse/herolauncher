# Process Manager

The Process Manager is a component that manages and monitors external processes. It provides functionality to start, stop, restart, and monitor processes, as well as retrieve their status information such as CPU and memory usage.

## Features

- Start, stop, restart, and delete processes
- Monitor CPU and memory usage of managed processes
- Set deadlines for process execution
- Support for cron-like scheduling
- Telnet interface for remote management
- Authentication via secret key

## Components

The Process Manager consists of the following components:

1. **Process Manager Core**: Manages processes and their lifecycle
2. **Telnet Server**: Provides a telnet interface for remote management
3. **Client Library**: Provides a Go API for interacting with the Process Manager
4. **Command-line Client**: Provides a command-line interface for managing processes
5. **Interfaces**: Various interface implementations for interacting with the Process Manager

## Interfaces

The Process Manager provides several interfaces for integration and extension:

### ProcessManagerInterface

The core interface that defines all process management operations. This interface allows for dependency injection and easier testing.

```go
type ProcessManagerInterface interface {
	// Authentication
	GetSecret() string

	// Process management
	StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) error
	StopProcess(name string) error
	RestartProcess(name string) error
	DeleteProcess(name string) error
	GetProcessStatus(name string) (*ProcessInfo, error)
	ListProcesses() []*ProcessInfo
	GetProcessLogs(name string, lines int) (string, error)
}
```

### Interface Implementations

The Process Manager includes the following interface implementations:

1. **Telnet Interface**: Provides a telnet-based interface for remote management
   - Located in `pkg/processmanager/interfaces/telnet`
   - Allows command execution via heroscript syntax
   - Supports authentication and secure communication

2. **OpenRPC Interface**: Provides a JSON-RPC 2.0 interface following the OpenRPC specification
   - Located in `pkg/processmanager/interfaces/openrpc`
   - Enables programmatic access via HTTP or WebSockets
   - Includes auto-generated client libraries
   - Provides a standardized API schema

## Usage

### Starting the Process Manager

```bash
./processmanager -socket /tmp/processmanager.sock -secret mysecretkey
```

### Using the Command-line Client

```bash
# Start a process
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey start -name myprocess -command "echo hello world" -log

# List all processes
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey list -format json

# Get process status
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey status -name myprocess -format json

# Stop a process
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey stop -name myprocess

# Restart a process
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey restart -name myprocess

# Delete a process
./pmclient -socket /tmp/processmanager.sock -secret mysecretkey delete -name myprocess
```

### Using the Telnet Interface

You can connect to the Process Manager using a telnet client:

```bash
telnet /tmp/processmanager.sock
```

After connecting, you need to authenticate with the secret key:

```
mysecretkey
```

Once authenticated, you can send heroscript commands:

```
!!process.start name:'myprocess' command:'echo hello world' log:true
!!process.list format:json
!!process.status name:'myprocess' format:json
!!process.stop name:'myprocess'
!!process.restart name:'myprocess'
!!process.delete name:'myprocess'
```

## Heroscript Commands

The Process Manager supports the following heroscript commands:

### process.start

Starts a new process.

```
!!process.start name:'processname' command:'command which can be multiline' log:true deadline:30 cron:'0 0 * * *' jobid:'e42'
```

Parameters:
- `name`: Name of the process (required)
- `command`: Command to run (required)
- `log`: Enable logging (optional, default: false)
- `deadline`: Deadline in seconds (optional)
- `cron`: Cron schedule (optional)
- `jobid`: Job ID (optional)

### process.list

Lists all processes.

```
!!process.list format:json
```

Parameters:
- `format`: Output format (optional, values: json or empty for text)

### process.delete

Deletes a process.

```
!!process.delete name:'processname'
```

Parameters:
- `name`: Name of the process (required)

### process.status

Gets the status of a process.

```
!!process.status name:'processname' format:json
```

Parameters:
- `name`: Name of the process (required)
- `format`: Output format (optional, values: json or empty for text)

### process.restart

Restarts a process.

```
!!process.restart name:'processname'
```

Parameters:
- `name`: Name of the process (required)

### process.stop

Stops a process.

```
!!process.stop name:'processname'
```

Parameters:
- `name`: Name of the process (required)
