package vfsdb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// Version byte for the encoding format
const encodingVersion byte = 1

// encodeMetadata encodes the common metadata structure
func encodeMetadata(metadata *vfs.Metadata, buf *bytes.Buffer) error {
	// Write metadata fields
	binary.Write(buf, binary.LittleEndian, metadata.ID)
	
	// Write name as length-prefixed string
	nameBytes := []byte(metadata.Name)
	binary.Write(buf, binary.LittleEndian, uint16(len(nameBytes)))
	buf.Write(nameBytes)
	
	// Write file type
	buf.WriteByte(byte(metadata.FileType))
	
	// Write size
	binary.Write(buf, binary.LittleEndian, metadata.Size)
	
	// Write timestamps
	binary.Write(buf, binary.LittleEndian, metadata.CreatedAt)
	binary.Write(buf, binary.LittleEndian, metadata.ModifiedAt)
	binary.Write(buf, binary.LittleEndian, metadata.AccessedAt)
	
	// Write mode
	binary.Write(buf, binary.LittleEndian, metadata.Mode)
	
	// Write owner as length-prefixed string
	ownerBytes := []byte(metadata.Owner)
	binary.Write(buf, binary.LittleEndian, uint16(len(ownerBytes)))
	buf.Write(ownerBytes)
	
	// Write group as length-prefixed string
	groupBytes := []byte(metadata.Group)
	binary.Write(buf, binary.LittleEndian, uint16(len(groupBytes)))
	buf.Write(groupBytes)
	
	return nil
}

// encodeDirectory encodes a Directory to binary format
func encodeDirectory(dir *DirectoryEntry) ([]byte, error) {
	buf := new(bytes.Buffer)
	
	// Write version byte
	buf.WriteByte(encodingVersion)
	
	// Write type byte
	buf.WriteByte(byte(vfs.FileTypeDirectory))
	
	// Encode metadata
	if err := encodeMetadata(dir.metadata, buf); err != nil {
		return nil, err
	}
	
	// Encode parent ID
	binary.Write(buf, binary.LittleEndian, dir.parentID)
	
	// Encode children IDs
	binary.Write(buf, binary.LittleEndian, uint16(len(dir.children)))
	for _, childID := range dir.children {
		binary.Write(buf, binary.LittleEndian, childID)
	}
	
	return buf.Bytes(), nil
}

// encodeFile encodes a File to binary format
func encodeFile(file *FileEntry) ([]byte, error) {
	buf := new(bytes.Buffer)
	
	// Write version byte
	buf.WriteByte(encodingVersion)
	
	// Write type byte
	buf.WriteByte(byte(vfs.FileTypeFile))
	
	// Encode metadata
	if err := encodeMetadata(file.metadata, buf); err != nil {
		return nil, err
	}
	
	// Encode parent ID
	binary.Write(buf, binary.LittleEndian, file.parentID)
	
	// Encode chunk IDs
	binary.Write(buf, binary.LittleEndian, uint16(len(file.chunkIDs)))
	for _, chunkID := range file.chunkIDs {
		binary.Write(buf, binary.LittleEndian, chunkID)
	}
	
	return buf.Bytes(), nil
}

// encodeSymlink encodes a Symlink to binary format
func encodeSymlink(symlink *SymlinkEntry) ([]byte, error) {
	buf := new(bytes.Buffer)
	
	// Write version byte
	buf.WriteByte(encodingVersion)
	
	// Write type byte
	buf.WriteByte(byte(vfs.FileTypeSymlink))
	
	// Encode metadata
	if err := encodeMetadata(symlink.metadata, buf); err != nil {
		return nil, err
	}
	
	// Encode parent ID
	binary.Write(buf, binary.LittleEndian, symlink.parentID)
	
	// Encode target path as length-prefixed string
	targetBytes := []byte(symlink.target)
	binary.Write(buf, binary.LittleEndian, uint16(len(targetBytes)))
	buf.Write(targetBytes)
	
	return buf.Bytes(), nil
}

// decodeEntryType decodes the entry type from the binary data
func decodeEntryType(data []byte) (vfs.FileType, error) {
	if len(data) < 2 {
		return vfs.FileTypeUnknown, errors.New("corrupt metadata bytes")
	}
	
	// Check version
	if data[0] != encodingVersion {
		return vfs.FileTypeUnknown, errors.New("unsupported encoding version")
	}
	
	return vfs.FileType(data[1]), nil
}

// decodeMetadata decodes the common metadata structure
func decodeMetadata(data []byte, offset int) (*vfs.Metadata, int, error) {
	if len(data) < offset+4 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	
	metadata := &vfs.Metadata{}
	
	// Read ID
	metadata.ID = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// Read name length
	if len(data) < offset+2 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	nameLen := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	// Read name
	if len(data) < offset+int(nameLen) {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.Name = string(data[offset : offset+int(nameLen)])
	offset += int(nameLen)
	
	// Read file type
	if len(data) < offset+1 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.FileType = vfs.FileType(data[offset])
	offset++
	
	// Read size
	if len(data) < offset+8 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.Size = binary.LittleEndian.Uint64(data[offset:])
	offset += 8
	
	// Read timestamps
	if len(data) < offset+24 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.CreatedAt = int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	metadata.ModifiedAt = int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	metadata.AccessedAt = int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	
	// Read mode
	if len(data) < offset+4 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.Mode = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// Read owner length
	if len(data) < offset+2 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	ownerLen := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	// Read owner
	if len(data) < offset+int(ownerLen) {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.Owner = string(data[offset : offset+int(ownerLen)])
	offset += int(ownerLen)
	
	// Read group length
	if len(data) < offset+2 {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	groupLen := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	// Read group
	if len(data) < offset+int(groupLen) {
		return nil, 0, errors.New("corrupt metadata bytes")
	}
	metadata.Group = string(data[offset : offset+int(groupLen)])
	offset += int(groupLen)
	
	return metadata, offset, nil
}

// decodeDirectory decodes a binary format back to Directory
func decodeDirectory(data []byte, fs *DatabaseVFS) (*DirectoryEntry, error) {
	if len(data) < 2 {
		return nil, errors.New("corrupt directory data")
	}
	
	// Check version
	if data[0] != encodingVersion {
		return nil, errors.New("unsupported encoding version")
	}
	
	// Check type
	if vfs.FileType(data[1]) != vfs.FileTypeDirectory {
		return nil, errors.New("invalid type byte for directory")
	}
	
	// Decode metadata
	metadata, offset, err := decodeMetadata(data, 2)
	if err != nil {
		return nil, err
	}
	
	// Decode parent ID
	if len(data) < offset+4 {
		return nil, errors.New("corrupt directory data")
	}
	parentID := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// Decode children IDs
	if len(data) < offset+2 {
		return nil, errors.New("corrupt directory data")
	}
	childrenCount := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	children := make([]uint32, 0, childrenCount)
	for i := 0; i < int(childrenCount); i++ {
		if len(data) < offset+4 {
			return nil, errors.New("corrupt directory data")
		}
		childID := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		children = append(children, childID)
	}
	
	return &DirectoryEntry{
		metadata: metadata,
		parentID: parentID,
		children: children,
		vfs:      fs,
	}, nil
}

// decodeFile decodes a binary format back to File
func decodeFile(data []byte, fs *DatabaseVFS) (*FileEntry, error) {
	if len(data) < 2 {
		return nil, errors.New("corrupt file data")
	}
	
	// Check version
	if data[0] != encodingVersion {
		return nil, errors.New("unsupported encoding version")
	}
	
	// Check type
	if vfs.FileType(data[1]) != vfs.FileTypeFile {
		return nil, errors.New("invalid type byte for file")
	}
	
	// Decode metadata
	metadata, offset, err := decodeMetadata(data, 2)
	if err != nil {
		return nil, err
	}
	
	// Decode parent ID
	if len(data) < offset+4 {
		return nil, errors.New("corrupt file data")
	}
	parentID := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// Decode chunk IDs
	if len(data) < offset+2 {
		return nil, errors.New("corrupt file data")
	}
	chunkCount := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	chunkIDs := make([]uint32, 0, chunkCount)
	for i := 0; i < int(chunkCount); i++ {
		if len(data) < offset+4 {
			return nil, errors.New("corrupt file data")
		}
		chunkID := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		chunkIDs = append(chunkIDs, chunkID)
	}
	
	return &FileEntry{
		metadata: metadata,
		parentID: parentID,
		chunkIDs: chunkIDs,
		vfs:      fs,
	}, nil
}

// decodeSymlink decodes a binary format back to Symlink
func decodeSymlink(data []byte, fs *DatabaseVFS) (*SymlinkEntry, error) {
	if len(data) < 2 {
		return nil, errors.New("corrupt symlink data")
	}
	
	// Check version
	if data[0] != encodingVersion {
		return nil, errors.New("unsupported encoding version")
	}
	
	// Check type
	if vfs.FileType(data[1]) != vfs.FileTypeSymlink {
		return nil, errors.New("invalid type byte for symlink")
	}
	
	// Decode metadata
	metadata, offset, err := decodeMetadata(data, 2)
	if err != nil {
		return nil, err
	}
	
	// Decode parent ID
	if len(data) < offset+4 {
		return nil, errors.New("corrupt symlink data")
	}
	parentID := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	
	// Decode target path
	if len(data) < offset+2 {
		return nil, errors.New("corrupt symlink data")
	}
	targetLen := binary.LittleEndian.Uint16(data[offset:])
	offset += 2
	
	if len(data) < offset+int(targetLen) {
		return nil, errors.New("corrupt symlink data")
	}
	target := string(data[offset : offset+int(targetLen)])
	
	return &SymlinkEntry{
		metadata: metadata,
		parentID: parentID,
		target:   target,
		vfs:      fs,
	}, nil
}
