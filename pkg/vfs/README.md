# VFS - Virtual File System

The VFS package provides a flexible and extensible virtual file system implementation for the herolauncher project. It allows for different backend implementations while maintaining a consistent interface for file and directory operations.

## Overview

The VFS package consists of two main components:

1. **Core VFS Interface**: Defines the interface for file system operations and common data structures.
2. **VFS_DB Implementation**: A database-backed implementation of the VFS interface using the `ourdb` package.

## Core VFS Interface

The core VFS interface defines the following components:

- `VFSImplementation`: Interface for file system operations
- `FSEntry`: Interface for file system entries (files, directories, symlinks)
- `Metadata`: Common metadata for file system entries

### Key Features

- File operations: create, read, write, append, delete
- Directory operations: create, list, delete
- Symlink operations: create, read, delete
- Path operations: exists, get, rename, copy, move

## VFS_DB Implementation

The VFS_DB implementation provides a database-backed virtual file system using the `ourdb` package. It includes:

- `DatabaseVFS`: Implementation of the `VFSImplementation` interface
- Database models for directories, files, and symlinks
- Encoding/decoding functions for storing entries in the database

### Key Features

- Persistent storage of file system entries and data
- Efficient binary encoding of metadata and content
- Concurrent access support with mutex locking

## Usage

### Creating a VFS Instance

```go
import (
    "github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
)

// Create a new database-backed VFS
fs, err := vfsdb.New("/path/to/database")
if err != nil {
    // Handle error
}
```

### File Operations

```go
// Create a file
file, err := fs.FileCreate("/path/to/file.txt")
if err != nil {
    // Handle error
}

// Write to a file
err = fs.FileWrite("/path/to/file.txt", []byte("Hello, world!"))
if err != nil {
    // Handle error
}

// Read from a file
data, err := fs.FileRead("/path/to/file.txt")
if err != nil {
    // Handle error
}

// Append to a file
err = fs.FileConcatenate("/path/to/file.txt", []byte(" More data."))
if err != nil {
    // Handle error
}

// Delete a file
err = fs.FileDelete("/path/to/file.txt")
if err != nil {
    // Handle error
}
```

### Directory Operations

```go
// Create a directory
dir, err := fs.DirCreate("/path/to/directory")
if err != nil {
    // Handle error
}

// List directory contents
entries, err := fs.DirList("/path/to/directory")
if err != nil {
    // Handle error
}

// Delete a directory
err = fs.DirDelete("/path/to/directory")
if err != nil {
    // Handle error
}
```

### Symlink Operations

```go
// Create a symlink
symlink, err := fs.LinkCreate("/target/path", "/link/path")
if err != nil {
    // Handle error
}

// Read a symlink target
target, err := fs.LinkRead("/link/path")
if err != nil {
    // Handle error
}

// Delete a symlink
err = fs.LinkDelete("/link/path")
if err != nil {
    // Handle error
}
```

### Path Operations

```go
// Check if a path exists
exists := fs.Exists("/path/to/check")

// Get an entry at a path
entry, err := fs.Get("/path/to/entry")
if err != nil {
    // Handle error
}

// Rename an entry
renamed, err := fs.Rename("/old/path", "/new/path")
if err != nil {
    // Handle error
}

// Copy an entry
copied, err := fs.Copy("/source/path", "/destination/path")
if err != nil {
    // Handle error
}

// Move an entry
moved, err := fs.Move("/source/path", "/destination/path")
if err != nil {
    // Handle error
}
```

## Testing

The package includes comprehensive tests for both the core VFS interface and the VFS_DB implementation. Run the tests with:

```bash
go test ./pkg/vfs/...
```

## Implementation Details

### Metadata

Each file system entry has associated metadata including:

- ID: Unique identifier
- Name: Entry name
- FileType: Type of entry (file, directory, symlink)
- Size: Size in bytes (for files)
- Timestamps: Created, modified, accessed
- Mode: File permissions
- Owner and Group: Ownership information

### Database Storage

The VFS_DB implementation uses two databases:

1. **Metadata Database**: Stores entry metadata and structure
2. **Data Database**: Stores file content in chunks

Entries are encoded to binary format for efficient storage and retrieval.

## Error Handling

The VFS package defines common errors for file system operations:

- `ErrNotFound`: Entry not found
- `ErrAlreadyExists`: Entry already exists
- `ErrNotDirectory`: Entry is not a directory
- `ErrNotFile`: Entry is not a file
- `ErrNotSymlink`: Entry is not a symlink
- `ErrNotEmpty`: Directory is not empty

## Concurrency

The VFS_DB implementation uses mutex locking to ensure thread-safe operations on the file system.
