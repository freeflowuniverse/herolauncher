# VFS9P - 9P Server for Herolauncher VFS

VFS9P is an adapter that implements the 9P2000 protocol server interface using the herolauncher VFS package as a backend. This allows you to expose a database-backed virtual filesystem via the 9P protocol, making it accessible to clients that support 9P mounts.

## Overview

The VFS9P adapter bridges the gap between:

1. **herolauncher/pkg/vfs**: A virtual filesystem with database storage
2. **knusbaum/go9p**: A Go implementation of the 9P2000 protocol

By using this adapter, you can leverage the benefits of both systems:
- The database storage and flexible API of herolauncher VFS
- The network accessibility of the 9P protocol

## Usage

### Basic Usage

```go
import (
    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
    "github.com/freeflowuniverse/herolauncher/pkg/vfs9p"
    "github.com/knusbaum/go9p"
)

// Create a VFS instance
vfsImpl, err := vfsdb.NewFromPath("/path/to/database")
if err != nil {
    // Handle error
}

// Create VFS9P adapter
adapter := vfs9p.NewVFS9P(vfsImpl)

// Serve 9p on port 5640
err = go9p.Serve("0.0.0.0:5640", adapter)
if err != nil {
    // Handle error
}
```

### Command Line Tool

The package includes a simple command-line tool that creates a VFS instance and serves it over 9P:

```bash
# Build the command-line tool
cd pkg/vfs9p/cmd
go build -o vfs9p-server

# Run the server with default settings (database at ./vfs9p.db, listening on 0.0.0.0:5640)
./vfs9p-server

# Run with custom settings
./vfs9p-server -db /path/to/database -addr localhost:9999
```

## Mounting the 9P Filesystem

Once your server is running, you can mount it on various systems:

### Linux

```bash
# Install 9p support
sudo modprobe 9p
sudo modprobe 9pnet
sudo modprobe 9pnet_tcp

# Mount the filesystem
sudo mkdir -p /mnt/vfs9p
sudo mount -t 9p -o trans=tcp,port=5640,version=9p2000 localhost /mnt/vfs9p
```

### Plan 9

```
mount -c tcp!localhost!5640 /n/vfs9p
```

### macOS (with plan9port)

```bash
9pfuse localhost:5640 /mnt/vfs9p
```

## Implementation Details

The VFS9P adapter implements the `go9p.Srv` interface, which defines methods for handling various 9P operations:

- **Version**: Negotiates protocol version and message size
- **Auth**: Handles authentication (simplified in this implementation)
- **Attach**: Connects a client to the root of the filesystem
- **Walk**: Navigates the filesystem hierarchy
- **Open**: Opens files and directories
- **Read**: Reads file content or directory entries
- **Write**: Writes to files
- **Create**: Creates new files and directories
- **Remove**: Deletes files and directories
- **Stat**: Retrieves file metadata
- **Wstat**: Updates file metadata (not fully implemented)

The adapter maps these 9P operations to corresponding VFS operations, handling file descriptors, permissions, and metadata conversion.

## Limitations

- Authentication is not fully implemented (returns "Authentication not required")
- Wstat (metadata modification) is not fully implemented
- Symlinks are not fully tested
- No caching mechanism for improved performance

## Future Improvements

- Implement full authentication support
- Complete Wstat implementation for metadata changes
- Add caching for better performance
- Improve error handling and logging
- Add unit tests
