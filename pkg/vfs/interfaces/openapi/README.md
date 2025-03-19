# VFS OpenAPI Server

A minimal OpenAPI server implementation for the VFS interface. This package provides HTTP endpoints to interact with any VFS implementation through a RESTful API.

## Overview

This package implements a simple HTTP server that exposes VFS operations through RESTful endpoints. It maps HTTP methods to appropriate VFS operations:

- `GET` - Read files, list directories, read symlinks
- `PUT` - Create or update files
- `POST` - Create directories, append to files, create symlinks, copy/move/rename operations
- `DELETE` - Delete files, directories, or symlinks

## API Endpoints

| Method | Endpoint | Description | Parameters |
|--------|----------|-------------|------------|
| GET | `/{path}` | Read file/list directory/read symlink | None |
| PUT | `/{path}` | Create or update file | File content in request body |
| POST | `/{path}?op=mkdir` | Create directory | None |
| POST | `/{path}?op=append` | Append to file | Content to append in request body |
| POST | `/{path}?op=symlink&target={target}` | Create symlink | `target`: Target path for symlink |
| POST | `/{path}?op=copy&src={src}` | Copy file/directory | `src`: Source path |
| POST | `/{path}?op=move&src={src}` | Move file/directory | `src`: Source path |
| POST | `/{path}?op=rename&old={old}` | Rename file/directory | `old`: Old path |
| DELETE | `/{path}` | Delete file/directory/symlink | None |

## Usage

```go
import (
    "github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/openapi"
    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
)

func main() {
    // Create a VFS implementation (e.g., vfsdb)
    vfsImpl, err := vfsdb.NewVFSDB("./vfsdb")
    if err != nil {
        panic(err)
    }
    defer vfsImpl.destroy()

    // Start the OpenAPI server on port 8080
    err = openapi.StartVFSOpenAPIServer(vfsImpl, 8080)
    if err != nil {
        panic(err)
    }
}
```

## Example Requests

### Read a file

```bash
curl -X GET http://localhost:8080/path/to/file
```

### Create a file

```bash
curl -X PUT http://localhost:8080/path/to/file -d "File content"
```

### Create a directory

```bash
curl -X POST http://localhost:8080/path/to/dir?op=mkdir
```

### Delete a file

```bash
curl -X DELETE http://localhost:8080/path/to/file
```

## Implementation Details

The server uses the standard Go `net/http` package to handle HTTP requests. It maps HTTP methods to the appropriate VFS operations defined in the `VFSImplementation` interface.
