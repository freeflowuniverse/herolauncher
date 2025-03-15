// Package vfslocal provides a local filesystem implementation of the VFS interface
package vfslocal

import (
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// FileEntry represents a file in the local filesystem
type FileEntry struct {
	metadata *vfs.Metadata
	path     string
}

// GetMetadata returns the metadata for the entry
func (e *FileEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// GetPath returns the path for the entry
func (e *FileEntry) GetPath() string {
	return e.path
}

// IsDir returns true if the entry is a directory
func (e *FileEntry) IsDir() bool {
	return false
}

// IsFile returns true if the entry is a file
func (e *FileEntry) IsFile() bool {
	return true
}

// IsSymlink returns true if the entry is a symlink
func (e *FileEntry) IsSymlink() bool {
	return false
}

// DirectoryEntry represents a directory in the local filesystem
type DirectoryEntry struct {
	metadata *vfs.Metadata
	path     string
}

// GetMetadata returns the metadata for the entry
func (e *DirectoryEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// GetPath returns the path for the entry
func (e *DirectoryEntry) GetPath() string {
	return e.path
}

// IsDir returns true if the entry is a directory
func (e *DirectoryEntry) IsDir() bool {
	return true
}

// IsFile returns true if the entry is a file
func (e *DirectoryEntry) IsFile() bool {
	return false
}

// IsSymlink returns true if the entry is a symlink
func (e *DirectoryEntry) IsSymlink() bool {
	return false
}

// SymlinkEntry represents a symlink in the local filesystem
type SymlinkEntry struct {
	metadata *vfs.Metadata
	path     string
	target   string
}

// GetMetadata returns the metadata for the entry
func (e *SymlinkEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// GetPath returns the path for the entry
func (e *SymlinkEntry) GetPath() string {
	return e.path
}

// IsDir returns true if the entry is a directory
func (e *SymlinkEntry) IsDir() bool {
	return false
}

// IsFile returns true if the entry is a file
func (e *SymlinkEntry) IsFile() bool {
	return false
}

// IsSymlink returns true if the entry is a symlink
func (e *SymlinkEntry) IsSymlink() bool {
	return true
}

// GetTarget returns the target of the symlink
func (e *SymlinkEntry) GetTarget() string {
	return e.target
}
