# 9p2000 VFS Adapter

This package provides a 9p2000 server implementation that uses the vfsdb backend from the herolauncher package. It allows you to expose a vfsdb virtual filesystem via the 9p protocol, making it accessible to 9p clients.

## Overview

The 9p2000 adapter implements the necessary interfaces from the go9p package to serve a 9p filesystem. It maps 9p operations to the corresponding vfsdb operations, allowing clients to interact with the vfsdb filesystem using the 9p protocol.

## Features

- Implements the full 9p2000 protocol
- Uses vfsdb as the backend storage
- Supports file and directory operations
- Handles file permissions and ownership

## Usage

To use this package, you can run the main program directly:

```bash
go run . --listen 0.0.0.0:9999 --db ./vfsdb
```

Or build and run the binary:

```bash
go build -o 9p2000
./9p2000 --listen 0.0.0.0:9999 --db ./vfsdb
```

### Command Line Options

- `--listen`: The address and port to listen on (default: "0.0.0.0:9999")
- `--db`: The path to the vfsdb database (default: "./vfsdb")

## Connecting to the Server

You can connect to the 9p server using any 9p client. For example, on Plan 9 or with plan9port:

```bash
9fs tcp!localhost!9999
```

Or mount it on Linux:

```bash
mount -t 9p -o trans=tcp,port=9999 localhost /mnt/9p
```

## Implementation Details

The package consists of three main components:

1. **VFSDBFile**: Implements the `fs.File` interface for vfsdb files
2. **VFSDBDir**: Implements the `fs.Dir` interface for vfsdb directories
3. **Main program**: Sets up the 9p server and connects it to the vfsdb backend

The implementation follows the same pattern as the ramfs example from the go9p package, but uses vfsdb as the backend instead of an in-memory filesystem.
