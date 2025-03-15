package vfsdb

import (
	"fmt"
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/ourdb"
)

// NewFromPath creates a new DatabaseVFS instance from a path
func NewFromPath(dbPath string) (*DatabaseVFS, error) {
	// Open or create metadata database
	metadataConfig := ourdb.DefaultConfig()
	metadataConfig.Path = dbPath + "_metadata"
	metadataDB, err := ourdb.New(metadataConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}

	// Open or create data database
	dataConfig := ourdb.DefaultConfig()
	dataConfig.Path = dbPath + "_data"
	dataDB, err := ourdb.New(dataConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open data database: %w", err)
	}

	// Create adapter for metadata DB
	metadataAdapter := &OurDBAdapter{
		db: metadataDB,
	}

	// Create adapter for data DB
	dataAdapter := &OurDBAdapter{
		db: dataDB,
	}

	// Create VFS instance
	fs := &DatabaseVFS{
		dbMetadata: metadataAdapter,
		dbData:     dataAdapter,
		idTable:    make(map[uint32]uint32),
		nextID:     1,
		rootID:     0,
		mu:         sync.RWMutex{},
	}

	// Initialize the VFS by creating or loading the root directory
	_, err = fs.RootGet()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize root directory: %w", err)
	}

	return fs, nil
}

// NewFromDB creates a new DatabaseVFS instance from existing database instances
func NewFromDB(metadataDB, dataDB Database) (*DatabaseVFS, error) {
	// Create VFS instance
	fs := &DatabaseVFS{
		dbMetadata: metadataDB,
		dbData:     dataDB,
		idTable:    make(map[uint32]uint32),
		nextID:     1,
		rootID:     0,
		mu:         sync.RWMutex{},
	}

	// Initialize the VFS by creating or loading the root directory
	_, err := fs.RootGet()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize root directory: %w", err)
	}

	return fs, nil
}

// This function is now in database.go

// This function is now in database.go

// This function is now in database.go
