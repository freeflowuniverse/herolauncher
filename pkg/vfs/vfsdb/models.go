package vfsdb

import (
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// FSEntry represents any type of filesystem entry
type FSEntry interface {
	vfs.FSEntry
	encode() ([]byte, error)
}

// DirectoryEntry represents a directory in the virtual filesystem
type DirectoryEntry struct {
	metadata  *vfs.Metadata
	children  []uint32 // List of child entry IDs
	parentID  uint32   // ID of parent directory (0 for root)
	vfs       *DatabaseVFS
}

// encode encodes the directory to binary format
func (d *DirectoryEntry) encode() ([]byte, error) {
	return encodeDirectory(d)
}

// NewDirectoryEntry creates a new directory entry
func NewDirectoryEntry(metadata *vfs.Metadata, parentID uint32, vfs *DatabaseVFS) *DirectoryEntry {
	return &DirectoryEntry{
		metadata: metadata,
		children: []uint32{},
		parentID: parentID,
		vfs:      vfs,
	}
}

// GetMetadata returns the metadata for this entry
func (d *DirectoryEntry) GetMetadata() *vfs.Metadata {
	return d.metadata
}

// IsDir returns true if the entry is a directory
func (d *DirectoryEntry) IsDir() bool {
	return d.metadata.IsDir()
}

// IsFile returns true if the entry is a file
func (d *DirectoryEntry) IsFile() bool {
	return d.metadata.IsFile()
}

// IsSymlink returns true if the entry is a symlink
func (d *DirectoryEntry) IsSymlink() bool {
	return d.metadata.IsSymlink()
}

// FileEntry represents a file in the virtual filesystem
type FileEntry struct {
	metadata  *vfs.Metadata
	parentID  uint32   // ID of parent directory
	chunkIDs  []uint32 // List of data chunk IDs
	vfs       *DatabaseVFS
}

// encode encodes the file to binary format
func (f *FileEntry) encode() ([]byte, error) {
	return encodeFile(f)
}

// NewFileEntry creates a new file entry
func NewFileEntry(metadata *vfs.Metadata, parentID uint32, vfs *DatabaseVFS) *FileEntry {
	return &FileEntry{
		metadata: metadata,
		parentID: parentID,
		chunkIDs: []uint32{},
		vfs:      vfs,
	}
}

// GetMetadata returns the metadata for this entry
func (f *FileEntry) GetMetadata() *vfs.Metadata {
	return f.metadata
}

// IsDir returns true if the entry is a directory
func (f *FileEntry) IsDir() bool {
	return f.metadata.IsDir()
}

// IsFile returns true if the entry is a file
func (f *FileEntry) IsFile() bool {
	return f.metadata.IsFile()
}

// IsSymlink returns true if the entry is a symlink
func (f *FileEntry) IsSymlink() bool {
	return f.metadata.IsSymlink()
}

// SymlinkEntry represents a symbolic link in the virtual filesystem
type SymlinkEntry struct {
	metadata  *vfs.Metadata
	target    string // Path that this symlink points to
	parentID  uint32 // ID of parent directory
	vfs       *DatabaseVFS
}

// encode encodes the symlink to binary format
func (s *SymlinkEntry) encode() ([]byte, error) {
	return encodeSymlink(s)
}

// NewSymlinkEntry creates a new symlink entry
func NewSymlinkEntry(metadata *vfs.Metadata, target string, parentID uint32, vfs *DatabaseVFS) *SymlinkEntry {
	return &SymlinkEntry{
		metadata: metadata,
		target:   target,
		parentID: parentID,
		vfs:      vfs,
	}
}

// GetMetadata returns the metadata for this entry
func (s *SymlinkEntry) GetMetadata() *vfs.Metadata {
	return s.metadata
}

// IsDir returns true if the entry is a directory
func (s *SymlinkEntry) IsDir() bool {
	return s.metadata.IsDir()
}

// IsFile returns true if the entry is a file
func (s *SymlinkEntry) IsFile() bool {
	return s.metadata.IsFile()
}

// IsSymlink returns true if the entry is a symlink
func (s *SymlinkEntry) IsSymlink() bool {
	return s.metadata.IsSymlink()
}

// UpdateTarget updates the target path of this symlink
func (s *SymlinkEntry) UpdateTarget(newTarget string) error {
	s.target = newTarget
	s.metadata.SetModified()
	return s.vfs.SaveEntry(s)
}

// GetTarget returns the target path of this symlink
func (s *SymlinkEntry) GetTarget() (string, error) {
	s.metadata.SetAccessed()
	return s.target, nil
}
