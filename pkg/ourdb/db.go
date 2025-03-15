// Package ourdb provides a simple key-value database implementation with history tracking
package ourdb

import (
	"errors"
	"os"
	"path/filepath"
)

// OurDB represents a binary database with variable-length records
type OurDB struct {
	lookup          *LookupTable
	path            string // Directory in which we will have the lookup db as well as all the backend
	incrementalMode bool
	fileSize        uint32
	file            *os.File
	fileNr          uint16 // The file which is open
	lastUsedFileNr  uint16
}

const headerSize = 12

// OurDBSetArgs contains the parameters for the Set method
type OurDBSetArgs struct {
	ID   *uint32
	Data []byte
}

// Set stores data at the specified key position
// The data is stored with a CRC32 checksum for integrity verification
// and maintains a linked list of previous values for history tracking
// Returns the ID used (either x if specified, or auto-incremented if x=0)
func (db *OurDB) Set(args OurDBSetArgs) (uint32, error) {
	if db.incrementalMode {
		// If ID points to an empty location, return an error
		// else, overwrite data
		if args.ID != nil {
			// This is an update
			location, err := db.lookup.Get(*args.ID)
			if err != nil {
				return 0, err
			}
			if location.Position == 0 {
				return 0, errors.New("cannot set id for insertions when incremental mode is enabled")
			}

			if err := db.set_(*args.ID, location, args.Data); err != nil {
				return 0, err
			}
			return *args.ID, nil
		}

		// This is an insert
		id, err := db.lookup.GetNextID()
		if err != nil {
			return 0, err
		}
		if err := db.set_(id, Location{}, args.Data); err != nil {
			return 0, err
		}
		return id, nil
	}

	// Using key-value mode
	if args.ID == nil {
		return 0, errors.New("id must be provided when incremental is disabled")
	}
	location, err := db.lookup.Get(*args.ID)
	if err != nil {
		return 0, err
	}
	if err := db.set_(*args.ID, location, args.Data); err != nil {
		return 0, err
	}
	return *args.ID, nil
}

// Get retrieves data stored at the specified key position
// Returns error if the key doesn't exist or data is corrupted
func (db *OurDB) Get(x uint32) ([]byte, error) {
	location, err := db.lookup.Get(x)
	if err != nil {
		return nil, err
	}
	return db.get_(location)
}

// GetHistory retrieves a list of previous values for the specified key
// depth parameter controls how many historical values to retrieve (max)
// Returns error if key doesn't exist or if there's an issue accessing the data
func (db *OurDB) GetHistory(x uint32, depth uint8) ([][]byte, error) {
	result := make([][]byte, 0)
	currentLocation, err := db.lookup.Get(x)
	if err != nil {
		return nil, err
	}

	// Traverse the history chain up to specified depth
	for i := uint8(0); i < depth; i++ {
		// Get current value
		data, err := db.get_(currentLocation)
		if err != nil {
			return nil, err
		}
		result = append(result, data)

		// Try to get previous location
		prevLocation, err := db.getPrevPos_(currentLocation)
		if err != nil {
			break
		}
		if prevLocation.Position == 0 {
			break
		}
		currentLocation = prevLocation
	}

	return result, nil
}

// Delete removes the data at the specified key position
// This operation zeros out the record but maintains the space in the file
// Use condense() to reclaim space from deleted records (happens in step after)
func (db *OurDB) Delete(x uint32) error {
	location, err := db.lookup.Get(x)
	if err != nil {
		return err
	}
	if err := db.delete_(x, location); err != nil {
		return err
	}
	return db.lookup.Delete(x)
}

// GetNextID returns the next id which will be used when storing
func (db *OurDB) GetNextID() (uint32, error) {
	if !db.incrementalMode {
		return 0, errors.New("incremental mode is not enabled")
	}
	return db.lookup.GetNextID()
}

// lookupDumpPath returns the path to the lookup dump file
func (db *OurDB) lookupDumpPath() string {
	return filepath.Join(db.path, "lookup_dump.db")
}

// Load metadata if exists
func (db *OurDB) Load() error {
	if _, err := os.Stat(db.lookupDumpPath()); err == nil {
		return db.lookup.ImportSparse(db.lookupDumpPath())
	}
	return nil
}

// Save ensures we have the metadata stored on disk
func (db *OurDB) Save() error {
	return db.lookup.ExportSparse(db.lookupDumpPath())
}

// Close closes the database file
func (db *OurDB) Close() error {
	if err := db.Save(); err != nil {
		return err
	}
	return db.close_()
}

// Destroy closes and removes the database
func (db *OurDB) Destroy() error {
	_ = db.Close()
	return os.RemoveAll(db.path)
}
