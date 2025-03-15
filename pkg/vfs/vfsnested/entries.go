// Package vfsnested provides a virtual filesystem implementation that can contain multiple nested VFS instances
package vfsnested

import (
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// RootEntry represents the root directory of the nested VFS
type RootEntry struct {
	metadata *vfs.Metadata
}

// GetMetadata returns the metadata for the entry
func (e *RootEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// Path returns the path for the entry
func (e *RootEntry) Path() string {
	return "/"
}

// IsDir returns true if the entry is a directory
func (e *RootEntry) IsDir() bool {
	return e.metadata.FileType == vfs.FileTypeDirectory
}

// IsFile returns true if the entry is a file
func (e *RootEntry) IsFile() bool {
	return e.metadata.FileType == vfs.FileTypeFile
}

// IsSymlink returns true if the entry is a symlink
func (e *RootEntry) IsSymlink() bool {
	return e.metadata.FileType == vfs.FileTypeSymlink
}

// MountEntry represents a mount point in the nested VFS
type MountEntry struct {
	metadata *vfs.Metadata
	impl     vfs.VFSImplementation
}

// GetMetadata returns the metadata for the entry
func (e *MountEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// Path returns the path for the entry
func (e *MountEntry) Path() string {
	return "/" + strings.TrimLeft(e.metadata.Name, "/")
}

// IsDir returns true if the entry is a directory
func (e *MountEntry) IsDir() bool {
	return e.metadata.FileType == vfs.FileTypeDirectory
}

// IsFile returns true if the entry is a file
func (e *MountEntry) IsFile() bool {
	return e.metadata.FileType == vfs.FileTypeFile
}

// IsSymlink returns true if the entry is a symlink
func (e *MountEntry) IsSymlink() bool {
	return e.metadata.FileType == vfs.FileTypeSymlink
}

// NestedEntry wraps an FSEntry from a sub VFS and prefixes its path
type NestedEntry struct {
	original vfs.FSEntry
	prefix   string
}

// GetMetadata returns the metadata for the entry
func (e *NestedEntry) GetMetadata() *vfs.Metadata {
	return e.original.GetMetadata()
}

// Path returns the path for the entry
func (e *NestedEntry) Path() string {
	// We need to get the path from the original entry's metadata
	originalMeta := e.original.GetMetadata()
	originalName := originalMeta.Name
	
	// For root entries, just return the prefix
	if originalName == "" || originalName == "/" {
		return e.prefix
	}
	
	return e.prefix + "/" + strings.TrimLeft(originalName, "/")
}

// IsDir returns true if the entry is a directory
func (e *NestedEntry) IsDir() bool {
	return e.original.IsDir()
}

// IsFile returns true if the entry is a file
func (e *NestedEntry) IsFile() bool {
	return e.original.IsFile()
}

// IsSymlink returns true if the entry is a symlink
func (e *NestedEntry) IsSymlink() bool {
	return e.original.IsSymlink()
}

// ResourceForkEntry represents a macOS resource fork file (._*)
type ResourceForkEntry struct {
	metadata *vfs.Metadata
	path     string
}

// GetMetadata returns the metadata for the entry
func (e *ResourceForkEntry) GetMetadata() *vfs.Metadata {
	return e.metadata
}

// Path returns the path for the entry
func (e *ResourceForkEntry) Path() string {
	return e.path
}

// IsDir returns true if the entry is a directory
func (e *ResourceForkEntry) IsDir() bool {
	return false
}

// IsFile returns true if the entry is a file
func (e *ResourceForkEntry) IsFile() bool {
	return true
}

// IsSymlink returns true if the entry is a symlink
func (e *ResourceForkEntry) IsSymlink() bool {
	return false
}
