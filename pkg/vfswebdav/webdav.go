package vfswebdav

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// CreateOptions holds options for the Create operation
type CreateOptions struct {
	Overwrite bool
}

// RemoveAllOptions holds options for the RemoveAll operation
type RemoveAllOptions struct {
	Recursive bool
}

// FileSystem implements the webdav.FileSystem interface using a vfs.VFSImplementation
type FileSystem struct {
	vfsImpl vfs.VFSImplementation
}

// NewFileSystem creates a new WebDAV filesystem backed by the given VFS implementation
func NewFileSystem(vfsImpl vfs.VFSImplementation) *FileSystem {
	return &FileSystem{
		vfsImpl: vfsImpl,
	}
}

// Open opens the file at the specified path for reading
func (fs *FileSystem) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	name = normalizePath(name)
	
	// Check if the file exists
	if !fs.vfsImpl.Exists(name) {
		return nil, webdav.NewHTTPError(http.StatusNotFound, vfs.ErrNotFound)
	}

	// Use the VFS implementation directly to read the file
	data, err := fs.vfsImpl.FileRead(name)
	if err != nil {
		return nil, webdav.NewHTTPError(http.StatusInternalServerError, err)
	}

	return ioutil.NopCloser(strings.NewReader(string(data))), nil
}

// Stat returns information about the file or directory at the specified path
func (fs *FileSystem) Stat(ctx context.Context, name string) (*webdav.FileInfo, error) {
	name = normalizePath(name)
	
	entry, err := fs.vfsImpl.Get(name)
	if err != nil {
		if err == vfs.ErrNotFound {
			return nil, webdav.NewHTTPError(http.StatusNotFound, err)
		}
		return nil, err
	}

	return fs.entryToFileInfo(entry, name)
}

// ReadDir returns a list of directory entries
func (fs *FileSystem) ReadDir(ctx context.Context, name string, recursive bool) ([]webdav.FileInfo, error) {
	name = normalizePath(name)
	
	// Try to list the directory directly using the VFS implementation
	entries, err := fs.vfsImpl.DirList(name)
	if err != nil {
		// If DirList fails, try to get the directory and list it
		entry, err := fs.vfsImpl.Get(name)
		if err != nil {
			if err == vfs.ErrNotFound {
				return nil, webdav.NewHTTPError(http.StatusNotFound, err)
			}
			return nil, err
		}

		if !entry.IsDir() {
			return nil, webdav.NewHTTPError(http.StatusBadRequest, vfs.ErrNotDirectory)
		}

		// Try to create the directory if it doesn't exist
		if !fs.vfsImpl.Exists(name) {
			_, err = fs.vfsImpl.DirCreate(name)
			if err != nil {
				return nil, err
			}
		}

		// Try listing again after creating the directory
		entries, err = fs.vfsImpl.DirList(name)
		if err != nil {
			return nil, err
		}
	}

	result := make([]webdav.FileInfo, 0, len(entries))
	for _, child := range entries {
		childPath, err := fs.vfsImpl.GetPath(child)
		if err != nil {
			return nil, err
		}

		fileInfo, err := fs.entryToFileInfo(child, childPath)
		if err != nil {
			return nil, err
		}
		result = append(result, *fileInfo)

		// If recursive and this is a directory, add its children
		if recursive && child.IsDir() {
			childDir, ok := child.(vfs.Directory)
			if !ok {
				continue
			}

			childEntries, err := childDir.List()
			if err != nil {
				return nil, err
			}

			for _, grandchild := range childEntries {
				grandchildPath, err := fs.vfsImpl.GetPath(grandchild)
				if err != nil {
					return nil, err
				}

				grandchildInfo, err := fs.entryToFileInfo(grandchild, grandchildPath)
				if err != nil {
					return nil, err
				}
				result = append(result, *grandchildInfo)
			}
		}
	}

	return result, nil
}

// Create creates or updates a file
func (fs *FileSystem) Create(ctx context.Context, name string, body io.ReadCloser) (fileInfo *webdav.FileInfo, created bool, err error) {
	name = normalizePath(name)
	defer body.Close()
	
	// Read the entire body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, false, err
	}

	// Check if the file exists
	exists := fs.vfsImpl.Exists(name)
	
	var entry vfs.FSEntry
	if exists {
		// Check if it's a directory
		existingEntry, err := fs.vfsImpl.Get(name)
		if err == nil && existingEntry.IsDir() {
			return nil, false, webdav.NewHTTPError(http.StatusConflict, fmt.Errorf("cannot create file, path is a directory: %s", name))
		}

		// Update existing file
		err = fs.vfsImpl.FileWrite(name, data)
		if err != nil {
			return nil, false, webdav.NewHTTPError(http.StatusInternalServerError, err)
		}
		
		entry, err = fs.vfsImpl.Get(name)
		if err != nil {
			return nil, false, webdav.NewHTTPError(http.StatusInternalServerError, err)
		}
	} else {
		// Create parent directories if they don't exist
		dir := filepath.Dir(name)
		if dir != "/" && !fs.vfsImpl.Exists(dir) {
			_, err = fs.vfsImpl.DirCreate(dir)
			if err != nil {
				return nil, false, webdav.NewHTTPError(http.StatusConflict, 
					fmt.Errorf("failed to create parent directory: %v", err))
			}
		}
		
		// Create new file
		entry, err = fs.vfsImpl.FileCreate(name)
		if err != nil {
			return nil, false, webdav.NewHTTPError(http.StatusInternalServerError, 
				fmt.Errorf("failed to create file: %v", err))
		}
		
		// Write data to the new file
		err = fs.vfsImpl.FileWrite(name, data)
		if err != nil {
			return nil, false, webdav.NewHTTPError(http.StatusInternalServerError, 
				fmt.Errorf("failed to write to file: %v", err))
		}
		
		// Get the updated entry
		entry, err = fs.vfsImpl.Get(name)
		if err != nil {
			return nil, false, webdav.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	info, err := fs.entryToFileInfo(entry, name)
	if err != nil {
		return nil, false, err
	}

	return info, !exists, nil
}

// RemoveAll removes a file or directory
func (fs *FileSystem) RemoveAll(ctx context.Context, name string) error {
	name = normalizePath(name)
	
	// Check if the entry exists
	if !fs.vfsImpl.Exists(name) {
		return webdav.NewHTTPError(http.StatusNotFound, vfs.ErrNotFound)
	}

	return fs.vfsImpl.Delete(name)
}

// Mkdir creates a directory
func (fs *FileSystem) Mkdir(ctx context.Context, name string) error {
	name = normalizePath(name)
	
	// Check if the directory already exists
	if fs.vfsImpl.Exists(name) {
		entry, err := fs.vfsImpl.Get(name)
		if err != nil {
			return webdav.NewHTTPError(http.StatusInternalServerError, err)
		}
		
		if entry.IsDir() {
			// Directory already exists, which is fine for WebDAV
			return nil
		}
		return webdav.NewHTTPError(http.StatusConflict, 
			fmt.Errorf("cannot create directory, path exists as a file: %s", name))
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(name)
	if dir != "/" && !fs.vfsImpl.Exists(dir) {
		_, err := fs.vfsImpl.DirCreate(dir)
		if err != nil {
			return webdav.NewHTTPError(http.StatusConflict, 
				fmt.Errorf("failed to create parent directory: %v", err))
		}
	}

	_, err := fs.vfsImpl.DirCreate(name)
	if err != nil {
		return webdav.NewHTTPError(http.StatusInternalServerError, 
			fmt.Errorf("failed to create directory: %v", err))
	}

	return nil
}

// Copy copies a file or directory
func (fs *FileSystem) Copy(ctx context.Context, name, dest string, options *webdav.CopyOptions) (created bool, err error) {
	name = normalizePath(name)
	dest = normalizePath(dest)
	
	// Check if the source exists
	if !fs.vfsImpl.Exists(name) {
		return false, webdav.NewHTTPError(http.StatusNotFound, vfs.ErrNotFound)
	}

	// Check if the destination exists
	destExists := fs.vfsImpl.Exists(dest)
	if destExists && (options == nil || !options.NoOverwrite) {
		// Delete the destination if overwrite is allowed
		err = fs.vfsImpl.Delete(dest)
		if err != nil {
			return false, err
		}
		destExists = false
	} else if destExists {
		return false, os.ErrExist
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(dest)
	if destDir != "/" && !fs.vfsImpl.Exists(destDir) {
		_, err = fs.vfsImpl.DirCreate(destDir)
		if err != nil {
			return false, err
		}
	}

	// Perform the copy
	_, err = fs.vfsImpl.Copy(name, dest)
	if err != nil {
		return false, err
	}

	return !destExists, nil
}

// Move moves a file or directory
func (fs *FileSystem) Move(ctx context.Context, name, dest string, options *webdav.MoveOptions) (created bool, err error) {
	name = normalizePath(name)
	dest = normalizePath(dest)
	
	// Check if the source exists
	if !fs.vfsImpl.Exists(name) {
		return false, webdav.NewHTTPError(http.StatusNotFound, vfs.ErrNotFound)
	}

	// Check if the destination exists
	destExists := fs.vfsImpl.Exists(dest)
	if destExists && (options == nil || !options.NoOverwrite) {
		// Delete the destination if overwrite is allowed
		err = fs.vfsImpl.Delete(dest)
		if err != nil {
			return false, err
		}
		destExists = false
	} else if destExists {
		return false, os.ErrExist
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(dest)
	if destDir != "/" && !fs.vfsImpl.Exists(destDir) {
		_, err = fs.vfsImpl.DirCreate(destDir)
		if err != nil {
			return false, err
		}
	}

	// Perform the move
	_, err = fs.vfsImpl.Move(name, dest)
	if err != nil {
		return false, err
	}

	return !destExists, nil
}

// entryToFileInfo converts a vfs.FSEntry to a webdav.FileInfo
func (fs *FileSystem) entryToFileInfo(entry vfs.FSEntry, path string) (*webdav.FileInfo, error) {
	metadata := entry.GetMetadata()
	
	// Get the path for the entry
	entryPath, err := fs.vfsImpl.GetPath(entry)
	if err != nil {
		return nil, err
	}
	
	// Use the provided path if it's not empty
	if path == "" {
		path = entryPath
	}

	return &webdav.FileInfo{
		Path:    path,
		IsDir:   entry.IsDir(),
		Size:    int64(metadata.Size),
		ModTime: time.Unix(metadata.ModifiedAt, 0),
		// Use the file extension to determine MIME type
		MIMEType: getMIMEType(metadata.Name),
		// Generate a simple ETag based on the modification time and size
		ETag: generateETag(metadata),
	}, nil
}

// normalizePath ensures the path is properly formatted for VFS operations
func normalizePath(path string) string {
	// Ensure the path starts with a slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Remove trailing slash unless it's the root path
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	
	return path
}

// getMIMEType returns a MIME type based on the file extension
func getMIMEType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	default:
		return "application/octet-stream"
	}
}

// generateETag generates a simple ETag for a file
func generateETag(metadata *vfs.Metadata) string {
	return fmt.Sprintf("\"%d-%d-%d\"", metadata.ID, metadata.Size, metadata.ModifiedAt)
}
