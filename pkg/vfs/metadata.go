package vfs

import (
	"time"
)

// Metadata represents common metadata for files and directories
type Metadata struct {
	ID          uint32   // Unique identifier
	Name        string   // Name of the file or directory
	FileType    FileType // Type of the entry (file, directory, symlink)
	Size        uint64   // Size in bytes
	CreatedAt   int64    // Unix timestamp of creation time
	ModifiedAt  int64    // Unix timestamp of last modification time
	AccessedAt  int64    // Unix timestamp of last access time
	Mode        uint32   // File permissions
	Owner       string   // Owner of the file
	Group       string   // Group of the file
}

// NewMetadata creates a new Metadata instance with default values
func NewMetadata(id uint32, name string, fileType FileType) *Metadata {
	now := time.Now().Unix()
	return &Metadata{
		ID:          id,
		Name:        name,
		FileType:    fileType,
		Size:        0,
		CreatedAt:   now,
		ModifiedAt:  now,
		AccessedAt:  now,
		Mode:        0644, // Default file permissions
		Owner:       "user",
		Group:       "user",
	}
}

// Created returns the creation time as a time.Time
func (m *Metadata) Created() time.Time {
	return time.Unix(m.CreatedAt, 0)
}

// Modified returns the modification time as a time.Time
func (m *Metadata) Modified() time.Time {
	return time.Unix(m.ModifiedAt, 0)
}

// Accessed returns the access time as a time.Time
func (m *Metadata) Accessed() time.Time {
	return time.Unix(m.AccessedAt, 0)
}

// SetModified updates the modification time to the current time
func (m *Metadata) SetModified() {
	m.ModifiedAt = time.Now().Unix()
}

// SetAccessed updates the access time to the current time
func (m *Metadata) SetAccessed() {
	m.AccessedAt = time.Now().Unix()
}

// IsDir returns true if this entry is a directory
func (m *Metadata) IsDir() bool {
	return m.FileType == FileTypeDirectory
}

// IsFile returns true if this entry is a file
func (m *Metadata) IsFile() bool {
	return m.FileType == FileTypeFile
}

// IsSymlink returns true if this entry is a symlink
func (m *Metadata) IsSymlink() bool {
	return m.FileType == FileTypeSymlink
}
