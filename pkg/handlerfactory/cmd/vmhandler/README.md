# VM Handler Example

This example demonstrates how to use the HandlerFactory with a VM handler to process heroscript commands.

## Overview

The VM handler example shows how to:

1. Create a handler that processes VM-related actions
2. Register the handler with the HandlerFactory
3. Start a telnet server that uses the HandlerFactory to process commands
4. Connect to the telnet server and send heroscript commands

## Running the Example

To run the example:

```bash
cd /Users/despiegk/code/github/freeflowuniverse/herolauncher/handlerfactory/cmd/vmhandler
go run main.go
```

This will start a telnet server on:
- Unix socket: `/tmp/vmhandler.sock`
- TCP: `localhost:8024`

## Connecting to the Server

### Using Unix Socket

```bash
nc -U /tmp/vmhandler.sock
```

### Using TCP

```bash
telnet localhost 8024
```

## Authentication

When you connect, you'll need to authenticate with the secret:

```
!!auth secret:1234
```

## Available Commands

Once authenticated, you can use the following commands:

```
!!vm.define name:'test_vm' cpu:4 memory:'8GB' storage:'100GB'
!!vm.start name:'test_vm'
!!vm.stop name:'test_vm'
!!vm.disk_add name:'test_vm' size:'50GB' type:'SSD'
!!vm.list
!!vm.status name:'test_vm'
!!vm.delete name:'test_vm' force:true
```

## Example Session

Here's an example session:

```
$ telnet localhost 8024
Connected to localhost.
Escape character is '^]'.
 ** Welcome: you are not authenticated, provide secret.
!!auth secret:1234

 ** Welcome: you are authenticated.
!!vm.define name:'test_vm' cpu:4 memory:'8GB' storage:'100GB'
VM 'test_vm' defined successfully with 4 CPU, 8GB memory, and 100GB storage

!!vm.start name:'test_vm'
VM 'test_vm' started successfully

!!vm.disk_add name:'test_vm' size:'50GB' type:'SSD'
Added 50GB SSD disk to VM 'test_vm'

!!vm.status name:'test_vm'
VM 'test_vm' status:
- Status: running
- CPU: 4
- Memory: 8GB
- Storage: 100GB
- Attached disks:
  1. 50GB SSD

!!vm.list
Defined VMs:
- test_vm (running): 4 CPU, 8GB memory, 100GB storage
  Attached disks:
  1. 50GB SSD

!!vm.stop name:'test_vm'
VM 'test_vm' stopped successfully

!!vm.delete name:'test_vm'
VM 'test_vm' deleted successfully

!!quit
Goodbye!
Connection closed by foreign host.
```

## Other Commands

- `!!help`, `h`, or `?` - Show help
- `!!interactive` or `!!i` - Toggle interactive mode (with colors)
- `!!quit`, `!!exit`, or `q` - Disconnect from server

## How It Works

1. The `main.go` file creates a HandlerFactory and registers the VM handler
2. It starts a telnet server that uses the HandlerFactory to process commands
3. When a client connects and sends a heroscript command, the server:
   - Parses the command to determine the actor and action
   - Calls the appropriate method on the VM handler
   - Returns the result to the client

## Extending the Example

You can extend this example by:

1. Adding more methods to the VM handler
2. Creating new handlers for different actors
3. Registering multiple handlers with the HandlerFactory
