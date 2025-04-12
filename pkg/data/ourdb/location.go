package ourdb

import (
	"errors"
	"fmt"
)

// Location represents a position in a database file
type Location struct {
	FileNr   uint16
	Position uint32
}

// LocationNew creates a new Location from bytes
func (lut *LookupTable) LocationNew(b_ []byte) (Location, error) {
	newLocation := Location{
		FileNr:   0,
		Position: 0,
	}

	// First verify keysize is valid
	if lut.KeySize != 2 && lut.KeySize != 3 && lut.KeySize != 4 && lut.KeySize != 6 {
		return newLocation, errors.New("keysize must be 2, 3, 4 or 6")
	}

	// Create padded b
	b := make([]byte, lut.KeySize)
	startIdx := int(lut.KeySize) - len(b_)
	if startIdx < 0 {
		return newLocation, errors.New("input bytes exceed keysize")
	}

	for i := 0; i < len(b_); i++ {
		b[startIdx+i] = b_[i]
	}

	switch lut.KeySize {
	case 2:
		// Only position, 2 bytes big endian
		newLocation.Position = uint32(b[0])<<8 | uint32(b[1])
		newLocation.FileNr = 0
	case 3:
		// Only position, 3 bytes big endian
		newLocation.Position = uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
		newLocation.FileNr = 0
	case 4:
		// Only position, 4 bytes big endian
		newLocation.Position = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		newLocation.FileNr = 0
	case 6:
		// 2 bytes file_nr + 4 bytes position, all big endian
		newLocation.FileNr = uint16(b[0])<<8 | uint16(b[1])
		newLocation.Position = uint32(b[2])<<24 | uint32(b[3])<<16 | uint32(b[4])<<8 | uint32(b[5])
	}

	// Verify limits based on keysize
	switch lut.KeySize {
	case 2:
		if newLocation.Position > 0xFFFF {
			return newLocation, errors.New("position exceeds max value for keysize=2 (max 65535)")
		}
		if newLocation.FileNr != 0 {
			return newLocation, errors.New("file_nr must be 0 for keysize=2")
		}
	case 3:
		if newLocation.Position > 0xFFFFFF {
			return newLocation, errors.New("position exceeds max value for keysize=3 (max 16777215)")
		}
		if newLocation.FileNr != 0 {
			return newLocation, errors.New("file_nr must be 0 for keysize=3")
		}
	case 4:
		if newLocation.FileNr != 0 {
			return newLocation, errors.New("file_nr must be 0 for keysize=4")
		}
	case 6:
		// For keysize 6: both file_nr and position can use their full range
		// No additional checks needed as u16 and u32 already enforce limits
	}

	return newLocation, nil
}

// ToBytes converts a Location to a 6-byte array
func (loc Location) ToBytes() ([]byte, error) {
	bytes := make([]byte, 6)

	// Put file_nr first (2 bytes)
	bytes[0] = byte(loc.FileNr >> 8)
	bytes[1] = byte(loc.FileNr)

	// Put position next (4 bytes)
	bytes[2] = byte(loc.Position >> 24)
	bytes[3] = byte(loc.Position >> 16)
	bytes[4] = byte(loc.Position >> 8)
	bytes[5] = byte(loc.Position)

	return bytes, nil
}

// ToLookupBytes converts a Location to bytes according to the keysize
func (loc Location) ToLookupBytes(keysize uint8) ([]byte, error) {
	bytes := make([]byte, keysize)

	switch keysize {
	case 2:
		if loc.Position > 0xFFFF {
			return nil, errors.New("position exceeds max value for keysize=2 (max 65535)")
		}
		if loc.FileNr != 0 {
			return nil, errors.New("file_nr must be 0 for keysize=2")
		}
		bytes[0] = byte(loc.Position >> 8)
		bytes[1] = byte(loc.Position)
	case 3:
		if loc.Position > 0xFFFFFF {
			return nil, errors.New("position exceeds max value for keysize=3 (max 16777215)")
		}
		if loc.FileNr != 0 {
			return nil, errors.New("file_nr must be 0 for keysize=3")
		}
		bytes[0] = byte(loc.Position >> 16)
		bytes[1] = byte(loc.Position >> 8)
		bytes[2] = byte(loc.Position)
	case 4:
		if loc.FileNr != 0 {
			return nil, errors.New("file_nr must be 0 for keysize=4")
		}
		bytes[0] = byte(loc.Position >> 24)
		bytes[1] = byte(loc.Position >> 16)
		bytes[2] = byte(loc.Position >> 8)
		bytes[3] = byte(loc.Position)
	case 6:
		bytes[0] = byte(loc.FileNr >> 8)
		bytes[1] = byte(loc.FileNr)
		bytes[2] = byte(loc.Position >> 24)
		bytes[3] = byte(loc.Position >> 16)
		bytes[4] = byte(loc.Position >> 8)
		bytes[5] = byte(loc.Position)
	default:
		return nil, fmt.Errorf("invalid keysize: %d", keysize)
	}

	return bytes, nil
}

// ToUint64 converts a Location to uint64, with file_nr as most significant (big endian)
func (loc Location) ToUint64() (uint64, error) {
	return (uint64(loc.FileNr) << 32) | uint64(loc.Position), nil
}
