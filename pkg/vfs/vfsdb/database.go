package vfsdb

import (
	"errors"
	"github.com/freeflowuniverse/herolauncher/pkg/ourdb"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"sync"
)

// Database defines the interface for database operations
type Database interface {
	// Get retrieves data by ID
	Get(id uint32) ([]byte, error)
	
	// Set stores data and returns an ID
	Set(data []byte) (uint32, error)
	
	// Update updates data at an existing ID
	Update(id uint32, data []byte) error
	
	// Delete removes data by ID
	Delete(id uint32) error
}

// OurDBAdapter adapts ourdb.OurDB to the Database interface
type OurDBAdapter struct {
	db *ourdb.OurDB
}

// Update updates data at an existing ID
func (a *OurDBAdapter) Update(id uint32, data []byte) error {
	// Since ourdb doesn't have an Update method, we'll use Set with the ID
	_, err := a.db.Set(ourdb.OurDBSetArgs{ID: &id, Data: data})
	return err
}

// NewOurDBAdapter creates a new OurDBAdapter
func NewOurDBAdapter(db *ourdb.OurDB) *OurDBAdapter {
	return &OurDBAdapter{db: db}
}

// Get retrieves data by ID
func (a *OurDBAdapter) Get(id uint32) ([]byte, error) {
	return a.db.Get(id)
}

// Set stores data and returns an ID
func (a *OurDBAdapter) Set(data []byte) (uint32, error) {
	return a.db.Set(ourdb.OurDBSetArgs{Data: data})
}

// Delete removes data by ID
func (a *OurDBAdapter) Delete(id uint32) error {
	return a.db.Delete(id)
}

// DatabaseVFS implements the vfs.VFSImplementation interface using a database backend
type DatabaseVFS struct {
	rootID     uint32
	blockSize  uint32
	dbData     Database
	dbMetadata Database
	nextID     uint32
	idTable    map[uint32]uint32
	mu         sync.RWMutex
}

// New creates a new DatabaseVFS instance
func New(dataDB, metadataDB Database) (*DatabaseVFS, error) {
	fs := &DatabaseVFS{
		rootID:     1,
		blockSize:  4 * 1024, // 4KB blocks
		dbData:     dataDB,
		dbMetadata: metadataDB,
		nextID:     1,
		idTable:    make(map[uint32]uint32),
	}
	
	return fs, nil
}

// GetNextID returns the next available ID
func (fs *DatabaseVFS) GetNextID() uint32 {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	id := fs.nextID
	fs.nextID++
	return id
}

// SaveEntry saves an entry to the database
func (fs *DatabaseVFS) SaveEntry(entry FSEntry) error {
	var data []byte
	var err error
	
	data, err = entry.encode()
	if err != nil {
		return err
	}
	
	id := entry.GetMetadata().ID
	
	fs.mu.RLock()
	dbID, exists := fs.idTable[id]
	fs.mu.RUnlock()
	
	if exists {
		// Update existing entry
		if err := fs.dbMetadata.Update(dbID, data); err != nil {
			return err
		}
	} else {
		// Create new entry
		dbID, err = fs.dbMetadata.Set(data)
		if err != nil {
			return err
		}
		
		fs.mu.Lock()
		fs.idTable[id] = dbID
		fs.mu.Unlock()
	}
	
	return nil
}

// LoadEntry loads an entry from the database by ID
func (fs *DatabaseVFS) LoadEntry(vfsID uint32) (FSEntry, error) {
	fs.mu.RLock()
	dbID, ok := fs.idTable[vfsID]
	fs.mu.RUnlock()
	
	if !ok {
		return nil, vfs.ErrNotFound
	}
	
	metadata, err := fs.dbMetadata.Get(dbID)
	if err != nil {
		return nil, err
	}
	
	entryType, err := decodeEntryType(metadata)
	if err != nil {
		return nil, err
	}
	
	switch entryType {
	case vfs.FileTypeDirectory:
		return decodeDirectory(metadata, fs)
	case vfs.FileTypeFile:
		return decodeFile(metadata, fs)
	case vfs.FileTypeSymlink:
		return decodeSymlink(metadata, fs)
	default:
		return nil, errors.New("unknown entry type")
	}
}
