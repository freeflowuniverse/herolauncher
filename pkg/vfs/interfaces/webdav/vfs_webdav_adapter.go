package vfswebdav

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"golang.org/x/net/webdav"
)

// VFSWebDAVAdapter adapts a VFS implementation to the golang.org/x/net/webdav FileSystem interface
type VFSWebDAVAdapter struct {
	vfsImpl vfs.VFSImplementation
}

// NewVFSWebDAVAdapter creates a new adapter for the given VFS implementation
func NewVFSWebDAVAdapter(vfsImpl vfs.VFSImplementation) *VFSWebDAVAdapter {
	return &VFSWebDAVAdapter{
		vfsImpl: vfsImpl,
	}
}

// Mkdir creates a directory
func (a *VFSWebDAVAdapter) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	name = normalizeWebDAVPath(name)

	// Check if the directory already exists
	if a.vfsImpl.Exists(name) {
		entry, err := a.vfsImpl.Get(name)
		if err != nil {
			return err
		}

		if entry.IsDir() {
			// Directory already exists, which is fine for WebDAV
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

	_, err := a.vfsImpl.DirCreate(name)
	return err
}

// OpenFile opens a file with the given flags
func (a *VFSWebDAVAdapter) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = normalizeWebDAVPath(name)

	// Handle directory creation for O_CREATE flag
	if flag&os.O_CREATE != 0 {
		if !a.vfsImpl.Exists(filepath.Dir(name)) {
			err := a.Mkdir(ctx, filepath.Dir(name), perm)
			if err != nil {
				return nil, err
			}
		}
	}

	// Check if the path exists
	exists := a.vfsImpl.Exists(name)

	// Handle different file open modes
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
			// Check if this is a directory path
			if strings.HasSuffix(name, "/") {
				err := a.Mkdir(ctx, name, perm)
				if err != nil {
					return nil, err
				}
				return &vfsWebDAVFile{
					name:    name,
					vfsImpl: a.vfsImpl,
					isDir:   true,
				}, nil
			}

			// For regular files, create an empty file
			err := a.vfsImpl.FileWrite(name, []byte{})
			if err != nil {
				return nil, err
			}
		}
	}

	// Check if the path exists after potential creation
	if !a.vfsImpl.Exists(name) {
		return nil, os.ErrNotExist
	}

	// Get file info
	entry, err := a.vfsImpl.Get(name)
	if err != nil {
		return nil, err
	}

	// Create and return a vfsFile
	return &vfsWebDAVFile{
		name:    name,
		vfsImpl: a.vfsImpl,
		isDir:   entry.IsDir(),
	}, nil
}

// RemoveAll removes a file or directory
func (a *VFSWebDAVAdapter) RemoveAll(ctx context.Context, name string) error {
	name = normalizeWebDAVPath(name)

	if !a.vfsImpl.Exists(name) {
		return os.ErrNotExist
	}

	return a.vfsImpl.Delete(name)
}

// Rename renames a file or directory
func (a *VFSWebDAVAdapter) Rename(ctx context.Context, oldName, newName string) error {
	oldName = normalizeWebDAVPath(oldName)
	newName = normalizeWebDAVPath(newName)

	if !a.vfsImpl.Exists(oldName) {
		return os.ErrNotExist
	}

	_, err := a.vfsImpl.Move(oldName, newName)
	return err
}

// Stat returns file info
func (a *VFSWebDAVAdapter) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = normalizeWebDAVPath(name)

	if !a.vfsImpl.Exists(name) {
		return nil, os.ErrNotExist
	}

	entry, err := a.vfsImpl.Get(name)
	if err != nil {
		return nil, err
	}

	metadata := entry.GetMetadata()
	return &vfsWebDAVFileInfo{
		name:    filepath.Base(name),
		size:    int64(metadata.Size),
		mode:    os.FileMode(metadata.Mode),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}

// vfsWebDAVFile implements the webdav.File interface
type vfsWebDAVFile struct {
	name     string
	vfsImpl  vfs.VFSImplementation
	isDir    bool
	dirEnts  []os.FileInfo
	readPos  int64
	writePos int64
	content  []byte
}

// Close closes the file
func (f *vfsWebDAVFile) Close() error {
	// If we've written content, save it
	if f.writePos > 0 && !f.isDir {
		return f.vfsImpl.FileWrite(f.name, f.content)
	}
	return nil
}

// Read reads from the file
func (f *vfsWebDAVFile) Read(p []byte) (n int, err error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Lazy load content if not already loaded
	if f.content == nil {
		data, err := f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
		f.content = data
	}

	if f.readPos >= int64(len(f.content)) {
		return 0, io.EOF
	}

	n = copy(p, f.content[f.readPos:])
	f.readPos += int64(n)
	return n, nil
}

// Seek sets the offset for the next Read or Write
func (f *vfsWebDAVFile) Seek(offset int64, whence int) (int64, error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Lazy load content if not already loaded
	if f.content == nil {
		data, err := f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
		f.content = data
	}

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = f.readPos + offset
	case io.SeekEnd:
		newPos = int64(len(f.content)) + offset
	default:
		return 0, os.ErrInvalid
	}

	if newPos < 0 {
		return 0, os.ErrInvalid
	}

	f.readPos = newPos
	return newPos, nil
}

// Readdir reads directory entries
func (f *vfsWebDAVFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, os.ErrInvalid
	}

	// Lazy load directory entries if not already loaded
	if f.dirEnts == nil {
		entries, err := f.vfsImpl.DirList(f.name)
		if err != nil {
			return nil, err
		}

		f.dirEnts = make([]os.FileInfo, 0, len(entries))
		for _, entry := range entries {
			metadata := entry.GetMetadata()
			f.dirEnts = append(f.dirEnts, &vfsWebDAVFileInfo{
				name:    metadata.Name,
				size:    int64(metadata.Size),
				mode:    os.FileMode(metadata.Mode),
				modTime: time.Unix(metadata.ModifiedAt, 0),
				isDir:   entry.IsDir(),
			})
		}
	}

	if count <= 0 {
		result := f.dirEnts
		f.dirEnts = nil
		return result, nil
	}

	if len(f.dirEnts) == 0 {
		return nil, io.EOF
	}

	n := count
	if n > len(f.dirEnts) {
		n = len(f.dirEnts)
	}

	result := f.dirEnts[:n]
	f.dirEnts = f.dirEnts[n:]
	return result, nil
}

// Stat returns file info
func (f *vfsWebDAVFile) Stat() (os.FileInfo, error) {
	entry, err := f.vfsImpl.Get(f.name)
	if err != nil {
		return nil, err
	}

	metadata := entry.GetMetadata()
	return &vfsWebDAVFileInfo{
		name:    filepath.Base(f.name),
		size:    int64(metadata.Size),
		mode:    os.FileMode(metadata.Mode),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}

// Write writes to the file
func (f *vfsWebDAVFile) Write(p []byte) (n int, err error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Lazy load content if not already loaded and we're not creating a new file
	if f.content == nil && f.vfsImpl.Exists(f.name) {
		data, err := f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
		f.content = data
	}

	// Initialize content if it's nil
	if f.content == nil {
		f.content = make([]byte, 0)
	}

	// Ensure content is large enough
	if f.writePos+int64(len(p)) > int64(len(f.content)) {
		newContent := make([]byte, f.writePos+int64(len(p)))
		copy(newContent, f.content)
		f.content = newContent
	}

	// Write data
	n = copy(f.content[f.writePos:], p)
	f.writePos += int64(n)
	return n, nil
}

// vfsWebDAVFileInfo implements the os.FileInfo interface
type vfsWebDAVFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *vfsWebDAVFileInfo) Name() string       { return fi.name }
func (fi *vfsWebDAVFileInfo) Size() int64        { return fi.size }
func (fi *vfsWebDAVFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *vfsWebDAVFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *vfsWebDAVFileInfo) IsDir() bool        { return fi.isDir }
func (fi *vfsWebDAVFileInfo) Sys() interface{}   { return nil }

// Helper functions

// normalizeWebDAVPath ensures the path is properly formatted for VFS operations
func normalizeWebDAVPath(path string) string {
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
