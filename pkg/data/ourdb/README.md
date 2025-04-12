# OurDB

OurDB is a simple key-value database implementation that provides:

- Efficient key-value storage with history tracking
- Data integrity verification using CRC32
- Support for multiple backend files
- Lookup table for fast data retrieval

## Overview

The database consists of three main components:

1. **DB Interface** - Provides the public API for database operations
2. **Lookup Table** - Maps keys to data locations for efficient retrieval
3. **Backend Storage** - Handles the actual data storage and file management

## Features

- **Key-Value Storage**: Store and retrieve binary data using numeric keys
- **History Tracking**: Maintain a linked list of previous values for each key
- **Data Integrity**: Verify data integrity using CRC32 checksums
- **Multiple Backends**: Support for multiple storage files to handle large datasets
- **Incremental Mode**: Automatically assign IDs for new records

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/freeflowuniverse/herolauncher/pkg/ourdb"
)

func main() {
    // Create a new database
    config := ourdb.DefaultConfig()
    config.Path = "/path/to/database"
    
    db, err := ourdb.New(config)
    if err != nil {
        log.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    // Store data
    data := []byte("Hello, World!")
    id := uint32(1)
    _, err = db.Set(ourdb.OurDBSetArgs{
        ID:   &id,
        Data: data,
    })
    if err != nil {
        log.Fatalf("Failed to store data: %v", err)
    }
    
    // Retrieve data
    retrievedData, err := db.Get(id)
    if err != nil {
        log.Fatalf("Failed to retrieve data: %v", err)
    }
    
    fmt.Printf("Retrieved data: %s\n", string(retrievedData))
}
```

### Using the Client

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/freeflowuniverse/herolauncher/pkg/ourdb"
)

func main() {
    // Create a new client
    client, err := ourdb.NewClient("/path/to/database")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // Add data with auto-generated ID
    data := []byte("Hello, World!")
    id, err := client.Add(data)
    if err != nil {
        log.Fatalf("Failed to add data: %v", err)
    }
    
    fmt.Printf("Data stored with ID: %d\n", id)
    
    // Retrieve data
    retrievedData, err := client.Get(id)
    if err != nil {
        log.Fatalf("Failed to retrieve data: %v", err)
    }
    
    fmt.Printf("Retrieved data: %s\n", string(retrievedData))
    
    // Store data with specific ID
    err = client.Set(2, []byte("Another value"))
    if err != nil {
        log.Fatalf("Failed to set data: %v", err)
    }
    
    // Get history of a value
    history, err := client.GetHistory(id, 5)
    if err != nil {
        log.Fatalf("Failed to get history: %v", err)
    }
    
    fmt.Printf("History count: %d\n", len(history))
    
    // Delete data
    err = client.Delete(id)
    if err != nil {
        log.Fatalf("Failed to delete data: %v", err)
    }
}
```

## Configuration Options

- **RecordNrMax**: Maximum number of records (default: 16777215)
- **RecordSizeMax**: Maximum size of a record in bytes (default: 4KB)
- **FileSize**: Maximum size of a database file (default: 500MB)
- **IncrementalMode**: Automatically assign IDs for new records (default: true)
- **Reset**: Reset the database on initialization (default: false)

## Notes

This is a Go port of the original V implementation from the herolib repository.
