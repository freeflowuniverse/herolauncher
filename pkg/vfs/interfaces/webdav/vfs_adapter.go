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

// VFSAdapter adapts a VFS implementation to the golang.org/x/net/webdav FileSystem interface
type VFSAdapter struct {
	vfsImpl vfs.VFSImplementation
}

// NewVFSAdapter creates a new adapter for the given VFS implementation
func NewVFSAdapter(vfsImpl vfs.VFSImplementation) *VFSAdapter {
	return &VFSAdapter{
		vfsImpl: vfsImpl,
	}
}

// Mkdir creates a directory
func (a *VFSAdapter) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	name = normalizeAdapterPath(name)

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

// Define O_DIRECTORY constant since it's not available in the os package
const O_DIRECTORY = 0x10000 // Value doesn't matter as we only use it for flag checks

// OpenFile opens a file with the given flags
func (a *VFSAdapter) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = normalizeAdapterPath(name)

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
			// For directories, use Mkdir
			if flag&O_DIRECTORY != 0 {
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
	return &vfsFile{
		name:    name,
		vfsImpl: a.vfsImpl,
		isDir:   entry.IsDir(),
	}, nil
}

// RemoveAll removes a file or directory
func (a *VFSAdapter) RemoveAll(ctx context.Context, name string) error {
	name = normalizeAdapterPath(name)

	if !a.vfsImpl.Exists(name) {
		return os.ErrNotExist
	}

	return a.vfsImpl.Delete(name)
}

// Stat returns file info
func (a *VFSAdapter) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = normalizeAdapterPath(name)

	if !a.vfsImpl.Exists(name) {
		return nil, os.ErrNotExist
	}

	entry, err := a.vfsImpl.Get(name)
	if err != nil {
		return nil, err
	}

	metadata := entry.GetMetadata()
	return &vfsFileInfo{
		name:    filepath.Base(name),
		size:    int64(metadata.Size),
		mode:    getFileMode(entry),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}

// vfsFile implements the webdav.File interface
type vfsFile struct {
	name     string
	vfsImpl  vfs.VFSImplementation
	isDir    bool
	dirEnts  []os.FileInfo
	readPos  int64
	writePos int64
	content  []byte
}

// Close closes the file
func (f *vfsFile) Close() error {
	// Reset all file state
	f.content = nil
	f.readPos = 0
	f.writePos = 0
	return nil
}

// Read reads from the file
func (f *vfsFile) Read(p []byte) (n int, err error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Check if file exists
	if !f.vfsImpl.Exists(f.name) {
		return 0, os.ErrNotExist
	}

	// Always read directly from the filesystem
	data, err := f.vfsImpl.FileRead(f.name)
	if err != nil {
		return 0, err
	}

	// Check if we're at the end of the file
	if f.readPos >= int64(len(data)) {
		return 0, io.EOF
	}

	// Calculate how many bytes we can read
	availableBytes := int64(len(data)) - f.readPos
	bytesToCopy := int64(len(p))
	if availableBytes < bytesToCopy {
		bytesToCopy = availableBytes
	}

	// Copy the data to the output buffer
	n = copy(p, data[f.readPos:f.readPos+bytesToCopy])
	f.readPos += int64(n)

	return n, nil
}

// Seek sets the offset for the next Read or Write
func (f *vfsFile) Seek(offset int64, whence int) (int64, error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Get the file size to calculate positions correctly
	fileSize := int64(0)
	if f.vfsImpl.Exists(f.name) {
		data, err := f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
		fileSize = int64(len(data))
	}

	// Calculate the new position based on whence
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = f.readPos + offset
	case io.SeekEnd:
		newPos = fileSize + offset
	default:
		return 0, os.ErrInvalid
	}

	// Ensure the position is valid
	if newPos < 0 {
		return 0, os.ErrInvalid
	}

	// Update both read and write positions
	f.readPos = newPos
	f.writePos = newPos
	return newPos, nil
}

// Readdir reads directory entries
func (f *vfsFile) Readdir(count int) ([]os.FileInfo, error) {
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
			f.dirEnts = append(f.dirEnts, &vfsFileInfo{
				name:    metadata.Name,
				size:    int64(metadata.Size),
				mode:    getFileMode(entry),
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
func (f *vfsFile) Stat() (os.FileInfo, error) {
	entry, err := f.vfsImpl.Get(f.name)
	if err != nil {
		return nil, err
	}

	metadata := entry.GetMetadata()
	return &vfsFileInfo{
		name:    filepath.Base(f.name),
		size:    int64(metadata.Size),
		mode:    getFileMode(entry),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}

// Write writes to the file
func (f *vfsFile) Write(p []byte) (n int, err error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// For handling all files, especially binary files
	// We'll use a direct file I/O approach without any buffering

	// If this is a new file or we're at position 0, simply write directly
	if f.writePos == 0 {
		// Write directly to the filesystem
		err = f.vfsImpl.FileWrite(f.name, p)
		if err != nil {
			return 0, err
		}

		// Update position
		n = len(p)
		f.writePos = int64(n)
		return n, nil
	}

	// For appends or random writes, we need to handle them carefully
	data, err := f.vfsImpl.FileRead(f.name)
	if err != nil {
		// If file doesn't exist, treat it as empty
		if os.IsNotExist(err) {
			data = make([]byte, 0)
		} else {
			return 0, err
		}
	}

	// Ensure we have enough space
	if f.writePos+int64(len(p)) > int64(len(data)) {
		newData := make([]byte, f.writePos+int64(len(p)))
		copy(newData, data)
		data = newData
	}

	// Copy the data at the current position
	n = copy(data[f.writePos:], p)
	f.writePos += int64(n)

	// Write the entire file back to disk
	err = f.vfsImpl.FileWrite(f.name, data)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// vfsFileInfo implements the os.FileInfo interface
type vfsFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *vfsFileInfo) Name() string       { return fi.name }
func (fi *vfsFileInfo) Size() int64        { return fi.size }
func (fi *vfsFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *vfsFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *vfsFileInfo) IsDir() bool        { return fi.isDir }
func (fi *vfsFileInfo) Sys() interface{}   { return nil }

// Helper functions

// normalizeAdapterPath ensures the path is properly formatted for VFS operations
func normalizeAdapterPath(path string) string {
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

// getFileMode returns an appropriate file mode for the entry
func getFileMode(entry vfs.FSEntry) os.FileMode {
	if entry.IsDir() {
		return os.ModeDir | 0755
	}
	return 0644
}
