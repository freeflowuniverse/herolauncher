# VFS WebDAV

This package provides a WebDAV server implementation that uses the VFS (Virtual File System) interface from the herolauncher project as its backend. It allows you to expose any VFS implementation as a WebDAV server, making it accessible through standard WebDAV clients.

## Features

- Full WebDAV server implementation using the [go-webdav](https://github.com/emersion/go-webdav) library
- Support for all standard WebDAV operations (GET, PUT, DELETE, MKCOL, COPY, MOVE, etc.)
- Compatible with any VFS implementation from the herolauncher project
- Support for nested VFS implementations

## Usage

### Basic Usage

```go
package main

import (
    "log"

    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
    "github.com/freeflowuniverse/herolauncher/pkg/vfswebdav"
)

func main() {
    // Create a local VFS implementation
    localVFS, err := vfslocal.New("/path/to/directory")
    if err != nil {
        log.Fatalf("Error creating local VFS: %v", err)
    }

    // Create a WebDAV server with the VFS implementation
    server := vfswebdav.NewServer(localVFS, ":8080")

    // Start the server
    log.Println("Starting WebDAV server on :8080")
    log.Fatal(server.Start())
}
```

### Using Nested VFS

```go
package main

import (
    "log"

    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsnested"
    "github.com/freeflowuniverse/herolauncher/pkg/vfswebdav"
)

func main() {
    // Create a nested VFS
    nestedVFS := vfsnested.New()

    // Create local VFS implementations
    docsVFS, err := vfslocal.New("/path/to/documents")
    if err != nil {
        log.Fatalf("Error creating documents VFS: %v", err)
    }

    photosVFS, err := vfslocal.New("/path/to/photos")
    if err != nil {
        log.Fatalf("Error creating photos VFS: %v", err)
    }

    // Add the local VFS implementations to the nested VFS
    nestedVFS.AddVFS("/docs", docsVFS)
    nestedVFS.AddVFS("/photos", photosVFS)

    // Create a WebDAV server with the nested VFS implementation
    server := vfswebdav.NewServer(nestedVFS, ":8080")

    // Start the server
    log.Println("Starting WebDAV server on :8080")
    log.Fatal(server.Start())
}
```

## Command-Line Tool

The package includes a command-line tool that can be used to start a WebDAV server with a local or nested VFS implementation.

### Building the Command-Line Tool

```bash
cd pkg/vfswebdav/cmd
go build -o webdav-server
```

### Usage

```bash
# Serve a single directory
./webdav-server -root /path/to/directory

# Serve multiple directories using nested VFS
./webdav-server -nested -dirs "/docs=/path/to/documents,/photos=/path/to/photos"

# Specify a custom address
./webdav-server -addr :8081 -root /path/to/directory
```

## Connecting to the WebDAV Server

You can connect to the WebDAV server using any WebDAV client. Here are some examples:

### macOS Finder

1. In Finder, click on "Go" > "Connect to Server..."
2. Enter the WebDAV URL: `http://localhost:8080`
3. Click "Connect"

### Windows Explorer

1. Right-click on "This PC" and select "Map network drive..."
2. Enter the WebDAV URL: `http://localhost:8080`
3. Click "Finish"

### Command Line (using curl)

```bash
# List files
curl -X PROPFIND http://localhost:8080/ -H "Depth: 1"

# Upload a file
curl -T file.txt http://localhost:8080/file.txt

# Download a file
curl http://localhost:8080/file.txt -o file.txt

# Create a directory
curl -X MKCOL http://localhost:8080/newdir
```

## License

Same as the herolauncher project.
