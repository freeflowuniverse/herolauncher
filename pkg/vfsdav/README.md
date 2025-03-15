# VFS WebDAV Adapter

This package provides a WebDAV server implementation that uses the herolauncher VFS (Virtual File System) as a backend. It allows you to serve any VFS implementation over WebDAV, making it accessible to any WebDAV client.

## Overview

The `vfsdav` package implements the WebDAV protocol as defined in [RFC 4918](https://tools.ietf.org/html/rfc4918) using the [go-webdav](https://github.com/emersion/go-webdav) library. It provides an adapter that maps the VFS interface to the WebDAV interface.

## Features

- Implements the full WebDAV protocol
- Support for all VFS operations (read, write, create, delete, etc.)
- Automatic MIME type detection
- ETag generation for caching
- Support for conditional requests (If-Match, If-None-Match)
- Command-line tool for quickly serving directories over WebDAV

## Usage

### Basic Usage

```go
package main

import (
    "log"

    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
    "github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
)

func main() {
    // Create a VFS implementation (e.g., local filesystem)
    vfsImpl, err := vfslocal.New("/path/to/directory")
    if err != nil {
        log.Fatalf("Failed to create VFS: %v", err)
    }

    // Create and start the WebDAV server
    server := vfsdav.NewServer(vfsImpl, "localhost:8080")
    
    log.Println("WebDAV server started at http://localhost:8080")
    err = server.ListenAndServe()
    if err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

### Custom HTTP Server

If you want to use the WebDAV handler with a custom HTTP server:

```go
package main

import (
    "log"
    "net/http"

    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
    "github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
)

func main() {
    // Create a VFS implementation
    vfsImpl, err := vfslocal.New("/path/to/directory")
    if err != nil {
        log.Fatalf("Failed to create VFS: %v", err)
    }

    // Create the WebDAV server
    server := vfsdav.NewServer(vfsImpl, "")
    
    // Mount the WebDAV handler on a specific path
    http.Handle("/webdav/", http.StripPrefix("/webdav", server.Handler()))
    
    // Add other handlers
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    // Start the HTTP server
    log.Println("Server started at http://localhost:8080")
    err = http.ListenAndServe("localhost:8080", nil)
    if err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

## WebDAV Clients

You can connect to the WebDAV server using various clients:

- **Windows**: Map a network drive to `http://localhost:8080`
- **macOS**: In Finder, use "Connect to Server" (⌘K) and enter `http://localhost:8080`
- **Linux**: Use `davfs2` to mount the WebDAV server
- **Web browsers**: Most modern browsers can browse WebDAV servers directly

## Command-Line Tool

The package includes a command-line tool that allows you to quickly serve a directory over WebDAV:

```bash
# Start a WebDAV server serving the current directory
go run ./pkg/vfsdav/cmd/main.go --dir .

# Start a WebDAV server on a specific host and port
go run ./pkg/vfsdav/cmd/main.go --host 0.0.0.0 --port 9090 --dir /path/to/directory

# Start a WebDAV server with a temporary directory (for testing)
go run ./pkg/vfsdav/cmd/main.go
```

## Examples

The package includes several examples to demonstrate how to use the WebDAV server:

### Basic Example

The basic example shows how to create a simple WebDAV server serving a temporary directory:

```bash
go run ./pkg/vfsdav/examples/goclient/main.go
```

### WebDAV Client Example

This example demonstrates how to use the WebDAV client API to interact with the WebDAV server:

```bash
go run ./pkg/vfsdav/examples/goclient/client/main.go
```

### Rclone Example

This example shows how to use the popular `rclone` tool to interact with the WebDAV server:

```bash
# Requires rclone to be installed
go run ./pkg/vfsdav/examples/rclone/main.go
```

## Testing

The package includes comprehensive tests to ensure that all functionality works correctly:

```bash
# Run all tests
go test ./pkg/vfsdav/tests

# Run the verification script to test all components
./pkg/vfsdav/scripts/works.sh
```

## Implementation Details

The adapter maps VFS operations to WebDAV operations as follows:

- `Open` → `FileRead`
- `Stat` → `Get`
- `ReadDir` → `DirList`
- `Create` → `FileCreate` + `FileWrite`
- `RemoveAll` → `FileDelete` or `DirDelete`
- `Mkdir` → `DirCreate`
- `Copy` → `Copy`
- `Move` → `Move`

## Limitations

- Symlinks are not fully supported in the WebDAV protocol
- Some advanced WebDAV features like property storage are not implemented
- Performance may be limited by the underlying VFS implementation
