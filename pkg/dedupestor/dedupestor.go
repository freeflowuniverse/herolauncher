// Package dedupestor provides a key-value store with deduplication based on content hashing
package dedupestor

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/ourdb"
	"github.com/freeflowuniverse/herolauncher/pkg/radixtree"
)

// MaxValueSize is the maximum allowed size for values (1MB)
const MaxValueSize = 1024 * 1024

// DedupeStore provides a key-value store with deduplication based on content hashing
type DedupeStore struct {
	Radix *radixtree.RadixTree // For storing hash -> id mappings
	Data  *ourdb.OurDB         // For storing the actual data
}

// NewArgs contains arguments for creating a new DedupeStore
type NewArgs struct {
	Path  string // Base path for the store
	Reset bool   // Whether to reset existing data
}

// New creates a new deduplication store
func New(args NewArgs) (*DedupeStore, error) {
	// Create the radixtree for hash -> id mapping
	rt, err := radixtree.New(radixtree.NewArgs{
		Path:  filepath.Join(args.Path, "radixtree"),
		Reset: args.Reset,
	})
	if err != nil {
		return nil, err
	}

	// Create the ourdb for actual data storage
	config := ourdb.DefaultConfig()
	config.Path = filepath.Join(args.Path, "data")
	config.RecordSizeMax = MaxValueSize
	config.IncrementalMode = true // We want auto-incrementing IDs
	config.Reset = args.Reset

	db, err := ourdb.New(config)
	if err != nil {
		return nil, err
	}

	return &DedupeStore{
		Radix: rt,
		Data:  db,
	}, nil
}

// Store stores data with its reference and returns its id
// If the data already exists (same hash), returns the existing id without storing again
// appends reference to the radix tree entry of the hash to track references
func (ds *DedupeStore) Store(data []byte, ref Reference) (uint32, error) {
	// Check size limit
	if len(data) > MaxValueSize {
		return 0, errors.New("value size exceeds maximum allowed size of 1MB")
	}

	// Calculate SHA-256 hash of the value (using SHA-256 instead of blake2b for Go compatibility)
	hash := sha256Sum(data)

	// Check if this hash already exists
	metadataBytes, err := ds.Radix.Get(hash)
	if err == nil {
		// Value already exists, add new ref & return the id
		metadata := BytesToMetadata(metadataBytes)
		metadata, err = metadata.AddReference(ref)
		if err != nil {
			return 0, err
		}

		err = ds.Radix.Update(hash, metadata.ToBytes())
		if err != nil {
			return 0, err
		}

		return metadata.ID, nil
	}

	// Store the actual data in ourdb
	id, err := ds.Data.Set(ourdb.OurDBSetArgs{
		Data: data,
	})
	if err != nil {
		return 0, err
	}

	metadata := Metadata{
		ID:         id,
		References: []Reference{ref},
	}

	// Store the mapping of hash -> id in radixtree
	err = ds.Radix.Set(hash, metadata.ToBytes())
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Get retrieves a value by its ID
func (ds *DedupeStore) Get(id uint32) ([]byte, error) {
	return ds.Data.Get(id)
}

// GetFromHash retrieves a value by its hash
func (ds *DedupeStore) GetFromHash(hash string) ([]byte, error) {
	// Get the ID from radixtree
	metadataBytes, err := ds.Radix.Get(hash)
	if err != nil {
		return nil, err
	}

	// Convert bytes back to metadata
	metadata := BytesToMetadata(metadataBytes)

	// Get the actual data from ourdb
	return ds.Data.Get(metadata.ID)
}

// IDExists checks if a value with the given ID exists
func (ds *DedupeStore) IDExists(id uint32) bool {
	_, err := ds.Data.Get(id)
	return err == nil
}

// HashExists checks if a value with the given hash exists
func (ds *DedupeStore) HashExists(hash string) bool {
	_, err := ds.Radix.Get(hash)
	return err == nil
}

// Delete removes a reference from the hash entry
// If it's the last reference, removes the hash entry and its data
func (ds *DedupeStore) Delete(id uint32, ref Reference) error {
	// Get the data to calculate its hash
	data, err := ds.Data.Get(id)
	if err != nil {
		return err
	}

	// Calculate hash of the value
	hash := sha256Sum(data)

	// Get the current entry from radixtree
	metadataBytes, err := ds.Radix.Get(hash)
	if err != nil {
		return err
	}

	metadata := BytesToMetadata(metadataBytes)
	metadata, err = metadata.RemoveReference(ref)
	if err != nil {
		return err
	}

	if len(metadata.References) == 0 {
		// Delete from radixtree
		err = ds.Radix.Delete(hash)
		if err != nil {
			return err
		}

		// Delete from data db
		return ds.Data.Delete(id)
	}

	// Update hash metadata
	return ds.Radix.Update(hash, metadata.ToBytes())
}

// Close closes the dedupe store
func (ds *DedupeStore) Close() error {
	err1 := ds.Radix.Close()
	err2 := ds.Data.Close()

	if err1 != nil {
		return err1
	}
	return err2
}

// Helper function to calculate SHA-256 hash and return as hex string
func sha256Sum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
