// Package dedupestor provides a key-value store with deduplication based on content hashing
package dedupestor

import (
	"encoding/binary"
)

// Metadata represents a stored value with its ID and references
type Metadata struct {
	ID         uint32      // ID of the stored data in the database
	References []Reference // List of references to this data
}

// Reference represents a reference to stored data
type Reference struct {
	Owner uint16 // Owner identifier
	ID    uint32 // Reference identifier
}

// ToBytes converts Metadata to bytes for storage
func (m Metadata) ToBytes() []byte {
	// Calculate size: 4 bytes for ID + 6 bytes per reference
	size := 4 + (len(m.References) * 6)
	result := make([]byte, size)

	// Write ID (4 bytes)
	binary.LittleEndian.PutUint32(result[0:4], m.ID)

	// Write references (6 bytes each)
	offset := 4
	for _, ref := range m.References {
		refBytes := ref.ToBytes()
		copy(result[offset:offset+6], refBytes)
		offset += 6
	}

	return result
}

// BytesToMetadata converts bytes back to Metadata
func BytesToMetadata(b []byte) Metadata {
	if len(b) < 4 {
		return Metadata{
			ID:         0,
			References: []Reference{},
		}
	}

	id := binary.LittleEndian.Uint32(b[0:4])
	refs := []Reference{}

	// Parse references (each reference is 6 bytes)
	for i := 4; i < len(b); i += 6 {
		if i+6 <= len(b) {
			refs = append(refs, BytesToReference(b[i:i+6]))
		}
	}

	return Metadata{
		ID:         id,
		References: refs,
	}
}

// AddReference adds a new reference if it doesn't already exist
func (m Metadata) AddReference(ref Reference) (Metadata, error) {
	// Check if reference already exists
	for _, existing := range m.References {
		if existing.Owner == ref.Owner && existing.ID == ref.ID {
			return m, nil
		}
	}

	// Add the new reference
	newRefs := append(m.References, ref)
	return Metadata{
		ID:         m.ID,
		References: newRefs,
	}, nil
}

// RemoveReference removes a reference if it exists
func (m Metadata) RemoveReference(ref Reference) (Metadata, error) {
	newRefs := []Reference{}
	for _, existing := range m.References {
		if existing.Owner != ref.Owner || existing.ID != ref.ID {
			newRefs = append(newRefs, existing)
		}
	}

	return Metadata{
		ID:         m.ID,
		References: newRefs,
	}, nil
}

// ToBytes converts Reference to bytes
func (r Reference) ToBytes() []byte {
	result := make([]byte, 6)
	
	// Write owner (2 bytes)
	binary.LittleEndian.PutUint16(result[0:2], r.Owner)
	
	// Write ID (4 bytes)
	binary.LittleEndian.PutUint32(result[2:6], r.ID)
	
	return result
}

// BytesToReference converts bytes to Reference
func BytesToReference(b []byte) Reference {
	if len(b) < 6 {
		return Reference{}
	}
	
	owner := binary.LittleEndian.Uint16(b[0:2])
	id := binary.LittleEndian.Uint32(b[2:6])
	
	return Reference{
		Owner: owner,
		ID:    id,
	}
}
