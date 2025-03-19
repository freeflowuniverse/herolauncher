package vfsadapter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"golang.org/x/net/webdav"
)

// VFSAdapter adapts a VFS implementation to the webdav.FileSystem interface
type VFSAdapter struct {
	vfsImpl vfs.VFSImplementation
}

// NewVFSAdapter creates a new VFS adapter for WebDAV
func NewVFSAdapter(vfsImpl vfs.VFSImplementation) *VFSAdapter {
	return &VFSAdapter{
		vfsImpl: vfsImpl,
	}
}

// Mkdir creates a directory
func (a *VFSAdapter) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	name = normalizePath(name)

	// Check if the directory already exists
	if a.vfsImpl.Exists(name) {
		// If it exists and is a directory, just return success
		entry, err := a.vfsImpl.Get(name)
		if err == nil && entry.IsDir() {
			return nil
		}
		return os.ErrExist
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(name)
	if dir != "/" && !a.vfsImpl.Exists(dir) {
		_, err := a.vfsImpl.DirCreate(dir)
		if err != nil {
			return err
		}
	}

	// Create the directory
	_, err := a.vfsImpl.DirCreate(name)
	return err
}

// OpenFile opens a file or directory
func (a *VFSAdapter) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = normalizePath(name)

	// Check if the file or directory exists
	exists := a.vfsImpl.Exists(name)

	// Handle directory creation
	if flag&os.O_CREATE != 0 && !exists && strings.HasSuffix(name, "/") {
		err := a.Mkdir(ctx, name, perm)
		if err != nil {
			return nil, err
		}
		return &vfsFile{
			name:    name,
			vfsImpl: a.vfsImpl,
			isDir:   true,
		}, nil
	}

	// Handle file operations
	if flag&os.O_CREATE != 0 {
		if exists {
			// File exists and we want to create it
			if flag&os.O_EXCL != 0 {
				// If O_EXCL is set, fail if the file exists
				return nil, os.ErrExist
			}
			// Otherwise, truncate the file if O_TRUNC is set
			if flag&os.O_TRUNC != 0 {
				err := a.vfsImpl.Delete(name)
				if err != nil {
					return nil, err
				}
				exists = false
			}
		}

		// Create the file if it doesn't exist
		if !exists {
			// For regular files, create an empty file
			err := a.vfsImpl.FileWrite(name, []byte{})
			if err != nil {
				return nil, err
			}
		}
	}

	// Open the file or directory
	// Always return a file handle, even if the file doesn't exist yet
	// This is important for PUT operations where the file is created during the Write operation
	if exists {
		entry, err := a.vfsImpl.Get(name)
		if err != nil {
			return nil, err
		}

		return &vfsFile{
			name:    name,
			vfsImpl: a.vfsImpl,
			isDir:   entry.IsDir(),
		}, nil
	} else {
		// For non-existent files, return a file handle anyway
		// The file will be created when Write is called
		return &vfsFile{
			name:    name,
			vfsImpl: a.vfsImpl,
			isDir:   false,
		}, nil
	}
}

// RemoveAll removes a file or directory and any children it contains
func (a *VFSAdapter) RemoveAll(ctx context.Context, name string) error {
	name = normalizePath(name)
	return a.vfsImpl.Delete(name)
}

// Rename renames a file or directory
func (a *VFSAdapter) Rename(ctx context.Context, oldName, newName string) error {
	oldName = normalizePath(oldName)
	newName = normalizePath(newName)

	// Check if the source exists
	if !a.vfsImpl.Exists(oldName) {
		return os.ErrNotExist
	}

	// Check if the destination exists
	if a.vfsImpl.Exists(newName) {
		// If destination exists, delete it first
		err := a.vfsImpl.Delete(newName)
		if err != nil {
			return err
		}
	}

	// Get the source entry
	entry, err := a.vfsImpl.Get(oldName)
	if err != nil {
		return err
	}

	// Handle directory rename
	if entry.IsDir() {
		// Create the destination directory
		_, err = a.vfsImpl.DirCreate(newName)
		if err != nil {
			return err
		}

		// Copy all files from the source to the destination
		entries, err := a.vfsImpl.DirList(oldName)
		if err != nil {
			return err
		}

		for _, childEntry := range entries {
			childName := childEntry.GetMetadata().Name
			oldChildPath := filepath.Join(oldName, childName)
			newChildPath := filepath.Join(newName, childName)

			// Recursively rename child items
			err = a.Rename(ctx, oldChildPath, newChildPath)
			if err != nil {
				return err
			}
		}

		// Delete the source directory
		return a.vfsImpl.Delete(oldName)
	}

	// Handle file rename
	data, err := a.vfsImpl.FileRead(oldName)
	if err != nil {
		return err
	}

	// Write to the destination
	err = a.vfsImpl.FileWrite(newName, data)
	if err != nil {
		return err
	}

	// Delete the source
	return a.vfsImpl.Delete(oldName)
}

// Stat returns file info for a file or directory
func (a *VFSAdapter) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = normalizePath(name)

	// Handle root directory
	if name == "/" {
		return &vfsFileInfo{
			name:    "/",
			size:    0,
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// Check if the file or directory exists
	if !a.vfsImpl.Exists(name) {
		return nil, os.ErrNotExist
	}

	// Get the entry
	entry, err := a.vfsImpl.Get(name)
	if err != nil {
		return nil, err
	}

	// Get metadata
	metadata := entry.GetMetadata()

	// Create file info
	return &vfsFileInfo{
		name:    filepath.Base(name),
		size:    int64(metadata.Size),
		mode:    os.FileMode(metadata.Mode),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}

// normalizePath ensures the path is properly formatted for VFS operations
func normalizePath(path string) string {
	// Remove trailing slashes except for root
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	
	// Ensure path starts with a slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	return path
}
