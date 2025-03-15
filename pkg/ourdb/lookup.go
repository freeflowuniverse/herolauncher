package ourdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const (
	dataFileName        = "data"
	incrementalFileName = ".inc"
)

// LookupConfig contains configuration for the lookup table
type LookupConfig struct {
	Size            uint32
	KeySize         uint8
	LookupPath      string
	IncrementalMode bool
}

// LookupTable manages the mapping between IDs and data locations
type LookupTable struct {
	KeySize     uint8
	LookupPath  string
	Data        []byte
	Incremental *uint32
}

// NewLookup creates a new lookup table
func NewLookup(config LookupConfig) (*LookupTable, error) {
	// Verify keysize is valid
	if config.KeySize != 2 && config.KeySize != 3 && config.KeySize != 4 && config.KeySize != 6 {
		return nil, errors.New("keysize must be 2, 3, 4 or 6")
	}

	var incremental *uint32
	if config.IncrementalMode {
		inc := getIncrementalInfo(config)
		incremental = &inc
	}

	if config.LookupPath != "" {
		if _, err := os.Stat(config.LookupPath); os.IsNotExist(err) {
			if err := os.MkdirAll(config.LookupPath, 0755); err != nil {
				return nil, err
			}
		}

		// For disk-based lookup, create empty file if it doesn't exist
		dataPath := filepath.Join(config.LookupPath, dataFileName)
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			data := make([]byte, config.Size*uint32(config.KeySize))
			if err := ioutil.WriteFile(dataPath, data, 0644); err != nil {
				return nil, err
			}
		}

		return &LookupTable{
			Data:        []byte{},
			KeySize:     config.KeySize,
			LookupPath:  config.LookupPath,
			Incremental: incremental,
		}, nil
	}

	return &LookupTable{
		Data:        make([]byte, config.Size*uint32(config.KeySize)),
		KeySize:     config.KeySize,
		LookupPath:  "",
		Incremental: incremental,
	}, nil
}

// getIncrementalInfo gets the next incremental ID value
func getIncrementalInfo(config LookupConfig) uint32 {
	if !config.IncrementalMode {
		return 0
	}

	if config.LookupPath != "" {
		incPath := filepath.Join(config.LookupPath, incrementalFileName)
		if _, err := os.Stat(incPath); os.IsNotExist(err) {
			// Create a separate file for storing the incremental value
			if err := ioutil.WriteFile(incPath, []byte("1"), 0644); err != nil {
				panic(fmt.Sprintf("failed to write .inc file: %v", err))
			}
		}

		incBytes, err := ioutil.ReadFile(incPath)
		if err != nil {
			panic(fmt.Sprintf("failed to read .inc file: %v", err))
		}

		incremental, err := strconv.ParseUint(string(incBytes), 10, 32)
		if err != nil {
			panic(fmt.Sprintf("failed to parse incremental value: %v", err))
		}

		return uint32(incremental)
	}

	return 1
}

// Get retrieves a location from the lookup table
func (lut *LookupTable) Get(x uint32) (Location, error) {
	entrySize := lut.KeySize
	if lut.LookupPath != "" {
		// Check file size first
		dataPath := filepath.Join(lut.LookupPath, dataFileName)
		fileInfo, err := os.Stat(dataPath)
		if err != nil {
			return Location{}, err
		}
		fileSize := fileInfo.Size()
		startPos := x * uint32(entrySize)

		if startPos+uint32(entrySize) > uint32(fileSize) {
			return Location{}, fmt.Errorf("invalid read for get in lut: %s: %d would exceed file size %d",
				lut.LookupPath, startPos+uint32(entrySize), fileSize)
		}

		// Read directly from file for disk-based lookup
		file, err := os.Open(dataPath)
		if err != nil {
			return Location{}, err
		}
		defer file.Close()

		data := make([]byte, entrySize)
		bytesRead, err := file.ReadAt(data, int64(startPos))
		if err != nil {
			return Location{}, err
		}
		if bytesRead < int(entrySize) {
			return Location{}, fmt.Errorf("incomplete read: expected %d bytes but got %d", entrySize, bytesRead)
		}
		return lut.LocationNew(data)
	}

	if x*uint32(entrySize) >= uint32(len(lut.Data)) {
		return Location{}, errors.New("index out of bounds")
	}

	start := x * uint32(entrySize)
	return lut.LocationNew(lut.Data[start : start+uint32(entrySize)])
}

// FindLastEntry scans the lookup table to find the highest ID with a non-zero entry
func (lut *LookupTable) FindLastEntry() (uint32, error) {
	var lastID uint32 = 0
	entrySize := lut.KeySize

	if lut.LookupPath != "" {
		// For disk-based lookup, read the file in chunks
		dataPath := filepath.Join(lut.LookupPath, dataFileName)
		file, err := os.Open(dataPath)
		if err != nil {
			return 0, err
		}
		defer file.Close()

		fileInfo, err := os.Stat(dataPath)
		if err != nil {
			return 0, err
		}
		fileSize := fileInfo.Size()

		buffer := make([]byte, entrySize)
		var pos uint32 = 0

		for {
			if int64(pos)*int64(entrySize) >= fileSize {
				break
			}

			bytesRead, err := file.Read(buffer)
			if err != nil || bytesRead < int(entrySize) {
				break
			}

			location, err := lut.LocationNew(buffer)
			if err == nil && (location.Position != 0 || location.FileNr != 0) {
				lastID = pos
			}
			pos++
		}
	} else {
		// For memory-based lookup
		for i := uint32(0); i < uint32(len(lut.Data)/int(entrySize)); i++ {
			location, err := lut.Get(i)
			if err != nil {
				continue
			}
			if location.Position != 0 || location.FileNr != 0 {
				lastID = i
			}
		}
	}

	return lastID, nil
}

// GetNextID returns the next available ID for incremental mode
func (lut *LookupTable) GetNextID() (uint32, error) {
	if lut.Incremental == nil {
		return 0, errors.New("lookup table not in incremental mode")
	}

	var tableSize uint32
	if lut.LookupPath != "" {
		dataPath := filepath.Join(lut.LookupPath, dataFileName)
		fileInfo, err := os.Stat(dataPath)
		if err != nil {
			return 0, err
		}
		tableSize = uint32(fileInfo.Size())
	} else {
		tableSize = uint32(len(lut.Data))
	}

	if (*lut.Incremental)*uint32(lut.KeySize) >= tableSize {
		return 0, errors.New("lookup table is full")
	}

	return *lut.Incremental, nil
}

// IncrementIndex increments the index for the next insertion
func (lut *LookupTable) IncrementIndex() error {
	if lut.Incremental == nil {
		return errors.New("lookup table not in incremental mode")
	}

	*lut.Incremental++
	if lut.LookupPath != "" {
		incPath := filepath.Join(lut.LookupPath, incrementalFileName)
		return ioutil.WriteFile(incPath, []byte(strconv.FormatUint(uint64(*lut.Incremental), 10)), 0644)
	}
	return nil
}

// Set updates a location in the lookup table
func (lut *LookupTable) Set(x uint32, location Location) error {
	entrySize := lut.KeySize

	// Handle incremental mode
	if lut.Incremental != nil {
		if x == *lut.Incremental {
			if err := lut.IncrementIndex(); err != nil {
				return err
			}
		}

		if x > *lut.Incremental {
			return errors.New("cannot set id for insertions when incremental mode is enabled")
		}
	}

	// Convert location to bytes
	locationBytes, err := location.ToLookupBytes(lut.KeySize)
	if err != nil {
		return err
	}

	if lut.LookupPath != "" {
		// For disk-based lookup, write directly to file
		dataPath := filepath.Join(lut.LookupPath, dataFileName)
		file, err := os.OpenFile(dataPath, os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		startPos := x * uint32(entrySize)
		if _, err := file.WriteAt(locationBytes, int64(startPos)); err != nil {
			return err
		}
	} else {
		// For memory-based lookup
		startPos := x * uint32(entrySize)
		if startPos+uint32(entrySize) > uint32(len(lut.Data)) {
			return errors.New("index out of bounds")
		}

		copy(lut.Data[startPos:startPos+uint32(entrySize)], locationBytes)
	}

	return nil
}

// Delete removes an entry from the lookup table
func (lut *LookupTable) Delete(x uint32) error {
	// Create an empty location
	emptyLocation := Location{}
	return lut.Set(x, emptyLocation)
}

// GetDataFilePath returns the path to the data file
func (lut *LookupTable) GetDataFilePath() (string, error) {
	if lut.LookupPath == "" {
		return "", errors.New("lookup table is not disk-based")
	}
	return filepath.Join(lut.LookupPath, dataFileName), nil
}

// GetIncFilePath returns the path to the incremental file
func (lut *LookupTable) GetIncFilePath() (string, error) {
	if lut.LookupPath == "" {
		return "", errors.New("lookup table is not disk-based")
	}
	return filepath.Join(lut.LookupPath, incrementalFileName), nil
}

// ExportSparse exports the lookup table to a file in sparse format
func (lut *LookupTable) ExportSparse(path string) error {
	// Implementation would be similar to the V version
	// For now, this is a placeholder
	return errors.New("export sparse not implemented yet")
}

// ImportSparse imports the lookup table from a file in sparse format
func (lut *LookupTable) ImportSparse(path string) error {
	// Implementation would be similar to the V version
	// For now, this is a placeholder
	return errors.New("import sparse not implemented yet")
}
