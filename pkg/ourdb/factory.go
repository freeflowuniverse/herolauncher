package ourdb

import (
	"os"
)

const mbyte = 1000000

// OurDBConfig contains configuration options for creating a new database
type OurDBConfig struct {
	RecordNrMax    uint32
	RecordSizeMax  uint32
	FileSize       uint32
	Path           string
	IncrementalMode bool
	Reset          bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() OurDBConfig {
	return OurDBConfig{
		RecordNrMax:    16777216 - 1,    // max size of records
		RecordSizeMax:  1024 * 4,        // max size in bytes of a record, is 4 KB default
		FileSize:       500 * (1 << 20), // 500MB
		IncrementalMode: true,
	}
}

// New creates a new database with the given configuration
func New(config OurDBConfig) (*OurDB, error) {
	// Determine appropriate keysize based on configuration
	var keysize uint8 = 4

	if config.RecordNrMax < 65536 {
		keysize = 2
	} else if config.RecordNrMax < 16777216 {
		keysize = 3
	} else {
		keysize = 4
	}

	if float64(config.RecordSizeMax*config.RecordNrMax)/2 > mbyte*10 {
		keysize = 6 // will use multiple files
	}

	// Create lookup table
	l, err := NewLookup(LookupConfig{
		Size:            config.RecordNrMax,
		KeySize:         keysize,
		IncrementalMode: config.IncrementalMode,
	})
	if err != nil {
		return nil, err
	}

	// Reset database if requested
	if config.Reset {
		os.RemoveAll(config.Path)
	}

	// Create database directory
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		return nil, err
	}

	// Create database instance
	db := &OurDB{
		path:            config.Path,
		lookup:          l,
		fileSize:        config.FileSize,
		incrementalMode: config.IncrementalMode,
	}

	// Load existing data if available
	if err := db.Load(); err != nil {
		return nil, err
	}

	return db, nil
}
