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
