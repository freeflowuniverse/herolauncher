package ourdb

import (
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
)

// calculateCRC computes CRC32 for the data
func calculateCRC(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// dbFileSelect opens the specified database file
func (db *OurDB) dbFileSelect(fileNr uint16) error {
	// Check file number limit
	if fileNr > 65535 {
		return errors.New("file_nr needs to be < 65536")
	}

	path := filepath.Join(db.path, fmt.Sprintf("%d.db", fileNr))

	// Always close the current file if it's open
	if db.file != nil {
		db.file.Close()
		db.file = nil
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := db.createNewDbFile(fileNr); err != nil {
			return err
		}
	}

	// Open the file fresh
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	db.file = file
	db.fileNr = fileNr
	return nil
}

// createNewDbFile creates a new database file
func (db *OurDB) createNewDbFile(fileNr uint16) error {
	newFilePath := filepath.Join(db.path, fmt.Sprintf("%d.db", fileNr))
	f, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write a single byte to make all positions start from 1
	_, err = f.Write([]byte{0})
	return err
}

// getFileNr returns the file number to use for the next write
func (db *OurDB) getFileNr() (uint16, error) {
	path := filepath.Join(db.path, fmt.Sprintf("%d.db", db.lastUsedFileNr))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := db.createNewDbFile(db.lastUsedFileNr); err != nil {
			return 0, err
		}
		return db.lastUsedFileNr, nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	if uint32(stat.Size()) >= db.fileSize {
		db.lastUsedFileNr++
		if err := db.createNewDbFile(db.lastUsedFileNr); err != nil {
			return 0, err
		}
	}

	return db.lastUsedFileNr, nil
}

// set_ stores data at position x
func (db *OurDB) set_(x uint32, oldLocation Location, data []byte) error {
	// Get file number to use
	fileNr, err := db.getFileNr()
	if err != nil {
		return err
	}

	// Select the file
	if err := db.dbFileSelect(fileNr); err != nil {
		return err
	}

	// Get current file position for lookup
	pos, err := db.file.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}
	newLocation := Location{
		FileNr:   fileNr,
		Position: uint32(pos),
	}

	// Calculate CRC of data
	crc := calculateCRC(data)

	// Create header (12 bytes total)
	header := make([]byte, headerSize)

	// Write size (2 bytes)
	size := uint16(len(data))
	header[0] = byte(size & 0xFF)
	header[1] = byte((size >> 8) & 0xFF)

	// Write CRC (4 bytes)
	header[2] = byte(crc & 0xFF)
	header[3] = byte((crc >> 8) & 0xFF)
	header[4] = byte((crc >> 16) & 0xFF)
	header[5] = byte((crc >> 24) & 0xFF)

	// Convert previous location to bytes and store in header
	prevBytes, err := oldLocation.ToBytes()
	if err != nil {
		return err
	}
	for i := 0; i < 6; i++ {
		header[6+i] = prevBytes[i]
	}

	// Write header
	if _, err := db.file.Write(header); err != nil {
		return err
	}

	// Write actual data
	if _, err := db.file.Write(data); err != nil {
		return err
	}

	if err := db.file.Sync(); err != nil {
		return err
	}

	// Update lookup table with new position
	return db.lookup.Set(x, newLocation)
}

// get_ retrieves data at specified location
func (db *OurDB) get_(location Location) ([]byte, error) {
	if err := db.dbFileSelect(location.FileNr); err != nil {
		return nil, err
	}

	if location.Position == 0 {
		return nil, fmt.Errorf("record not found, location: %+v", location)
	}

	// Read header
	header := make([]byte, headerSize)
	if _, err := db.file.ReadAt(header, int64(location.Position)); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Parse size (2 bytes)
	size := uint16(header[0]) | (uint16(header[1]) << 8)

	// Parse CRC (4 bytes)
	storedCRC := uint32(header[2]) | (uint32(header[3]) << 8) | (uint32(header[4]) << 16) | (uint32(header[5]) << 24)

	// Read data
	data := make([]byte, size)
	if _, err := db.file.ReadAt(data, int64(location.Position+headerSize)); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Verify CRC
	calculatedCRC := calculateCRC(data)
	if calculatedCRC != storedCRC {
		return nil, errors.New("CRC mismatch: data corruption detected")
	}

	return data, nil
}

// getPrevPos_ retrieves the previous position for a record
func (db *OurDB) getPrevPos_(location Location) (Location, error) {
	if location.Position == 0 {
		return Location{}, errors.New("record not found")
	}

	if err := db.dbFileSelect(location.FileNr); err != nil {
		return Location{}, err
	}

	// Skip size and CRC (6 bytes)
	prevBytes := make([]byte, 6)
	if _, err := db.file.ReadAt(prevBytes, int64(location.Position+6)); err != nil {
		return Location{}, fmt.Errorf("failed to read previous location bytes: %w", err)
	}

	return db.lookup.LocationNew(prevBytes)
}

// delete_ zeros out the record at specified location
func (db *OurDB) delete_(x uint32, location Location) error {
	if location.Position == 0 {
		return errors.New("record not found")
	}

	if err := db.dbFileSelect(location.FileNr); err != nil {
		return err
	}

	// Read size first
	sizeBytes := make([]byte, 2)
	if _, err := db.file.ReadAt(sizeBytes, int64(location.Position)); err != nil {
		return err
	}
	size := uint16(sizeBytes[0]) | (uint16(sizeBytes[1]) << 8)

	// Write zeros for the entire record (header + data)
	zeros := make([]byte, int(size)+headerSize)
	if _, err := db.file.WriteAt(zeros, int64(location.Position)); err != nil {
		return err
	}

	return nil
}

// close_ closes the database file
func (db *OurDB) close_() error {
	if db.file != nil {
		return db.file.Close()
	}
	return nil
}

// Condense removes empty records and updates positions
// This is a complex operation that creates a new file without the deleted records
func (db *OurDB) Condense() error {
	// This would be a complex implementation that would:
	// 1. Create a temporary file
	// 2. Copy all non-deleted records to the temp file
	// 3. Update all lookup entries to point to new locations
	// 4. Replace the original file with the temp file

	// For now, this is a placeholder for future implementation
	return errors.New("condense operation not implemented yet")
}
