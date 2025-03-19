package vfsadapter

import (
	"os"
	"time"
)

// vfsFileInfo implements the os.FileInfo interface for a VFS file or directory
type vfsFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// Name returns the base name of the file
func (fi *vfsFileInfo) Name() string {
	return fi.name
}

// Size returns the size of the file in bytes
func (fi *vfsFileInfo) Size() int64 {
	return fi.size
}

// Mode returns the file mode bits
func (fi *vfsFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime returns the modification time
func (fi *vfsFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir returns whether the file is a directory
func (fi *vfsFileInfo) IsDir() bool {
	return fi.isDir
}

// Sys returns the underlying data source (always nil for our implementation)
func (fi *vfsFileInfo) Sys() interface{} {
	return nil
}
