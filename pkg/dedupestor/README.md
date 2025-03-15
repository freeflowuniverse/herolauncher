# Dedupestor

Dedupestor is a Go package that provides a key-value store with deduplication based on content hashing. It allows for efficient storage of data by ensuring that duplicate content is stored only once, while maintaining references to the original data.

## Features

- Content-based deduplication using SHA-256 hashing
- Reference tracking to maintain data integrity
- Automatic cleanup when all references to data are removed
- Size limits to prevent excessive memory usage
- Persistent storage using the ourdb and radixtree packages

## Usage

```go
import (
    "github.com/freeflowuniverse/herolauncher/pkg/dedupestor"
)

// Create a new dedupe store
ds, err := dedupestor.New(dedupestor.NewArgs{
    Path:  "/path/to/store",
    Reset: false, // Set to true to reset existing data
})
if err != nil {
    // Handle error
}
defer ds.Close()

// Store data with a reference
data := []byte("example data")
ref := dedupestor.Reference{Owner: 1, ID: 1}
id, err := ds.Store(data, ref)
if err != nil {
    // Handle error
}

// Retrieve data by ID
retrievedData, err := ds.Get(id)
if err != nil {
    // Handle error
}

// Check if data exists
exists := ds.IDExists(id)

// Delete a reference to data
err = ds.Delete(id, ref)
if err != nil {
    // Handle error
}
```

## How It Works

1. When data is stored, a SHA-256 hash is calculated for the content
2. If the hash already exists in the store, a new reference is added to the existing data
3. If the hash doesn't exist, the data is stored and a new reference is created
4. When a reference is deleted, it's removed from the metadata
5. When the last reference to data is deleted, the data itself is removed from storage

## Dependencies

- [ourdb](../ourdb): For persistent storage of the actual data
- [radixtree](../radixtree): For efficient storage and retrieval of hash-to-ID mappings
