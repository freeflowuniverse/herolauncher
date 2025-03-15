package vfs

import (
	"io"
)

// FileType represents the type of a filesystem entry
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeFile
	FileTypeDirectory
	FileTypeSymlink
)

// String returns a string representation of the FileType
func (ft FileType) String() string {
	switch ft {
	case FileTypeFile:
		return "file"
	case FileTypeDirectory:
		return "directory"
	case FileTypeSymlink:
		return "symlink"
	default:
		return "unknown"
	}
}

// FSEntry represents a filesystem entry (file, directory, or symlink)
type FSEntry interface {
	// GetMetadata returns the metadata for this entry
	GetMetadata() *Metadata
	
	// IsDir returns true if the entry is a directory
	IsDir() bool
	
	// IsFile returns true if the entry is a file
	IsFile() bool
	
	// IsSymlink returns true if the entry is a symlink
	IsSymlink() bool
}

// VFSImplementation defines the interface for a virtual filesystem
type VFSImplementation interface {
	// Root operations
	RootGet() (FSEntry, error)
	
	// File operations
	FileCreate(path string) (FSEntry, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, data []byte) error
	FileConcatenate(path string, data []byte) error
	FileDelete(path string) error
	
	// Directory operations
	DirCreate(path string) (FSEntry, error)
	DirList(path string) ([]FSEntry, error)
	DirDelete(path string) error
	
	// Symlink operations
	LinkCreate(targetPath, linkPath string) (FSEntry, error)
	LinkRead(path string) (string, error)
	LinkDelete(path string) error
	
	// Common operations
	Exists(path string) bool
	Get(path string) (FSEntry, error)
	Rename(oldPath, newPath string) (FSEntry, error)
	Copy(srcPath, dstPath string) (FSEntry, error)
	Move(srcPath, dstPath string) (FSEntry, error)
	Delete(path string) error
	Destroy() error
	
	// Path operations
	GetPath(entry FSEntry) (string, error)
}

// ReadWriteSeeker combines io.Reader, io.Writer, and io.Seeker interfaces
type ReadWriteSeeker interface {
	io.Reader
	io.Writer
	io.Seeker
}

// File represents a file in the virtual filesystem
type File interface {
	FSEntry
	
	// Open opens the file and returns a ReadWriteSeeker
	Open() (ReadWriteSeeker, error)
	
	// Read reads the entire file content
	Read() ([]byte, error)
	
	// Write writes data to the file
	Write(data []byte) error
	
	// Append appends data to the file
	Append(data []byte) error
}

// Directory represents a directory in the virtual filesystem
type Directory interface {
	FSEntry
	
	// List returns the entries in this directory
	List() ([]FSEntry, error)
	
	// Create creates a new file or directory in this directory
	Create(name string, isDir bool) (FSEntry, error)
	
	// Remove removes an entry from this directory
	Remove(name string) error
}

// Symlink represents a symbolic link in the virtual filesystem
type Symlink interface {
	FSEntry
	
	// Target returns the target path of this symlink
	Target() (string, error)
	
	// UpdateTarget updates the target path of this symlink
	UpdateTarget(newTarget string) error
}
