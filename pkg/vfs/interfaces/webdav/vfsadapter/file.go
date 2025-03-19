package vfsadapter

import (
	"io"
	"os"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// vfsFile implements the webdav.File interface for a VFS file or directory
type vfsFile struct {
	name     string
	vfsImpl  vfs.VFSImplementation
	isDir    bool
	dirEnts  []os.FileInfo
	dirIndex int
	data     []byte
	offset   int64
}

// Close closes the file
func (f *vfsFile) Close() error {
	// Reset any internal state
	f.dirEnts = nil
	f.dirIndex = 0
	f.data = nil
	f.offset = 0
	return nil
}

// Read reads from the file
func (f *vfsFile) Read(p []byte) (int, error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Lazy load file data if not already loaded
	if f.data == nil {
		var err error
		f.data, err = f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
	}

	// Check if we've reached the end of the file
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}

	// Calculate how many bytes to read
	n := copy(p, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

// Seek sets the offset for the next Read or Write
func (f *vfsFile) Seek(offset int64, whence int) (int64, error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// Lazy load file data if not already loaded
	if f.data == nil {
		var err error
		f.data, err = f.vfsImpl.FileRead(f.name)
		if err != nil {
			return 0, err
		}
	}

	// Calculate the new offset
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = f.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(f.data)) + offset
	default:
		return 0, os.ErrInvalid
	}

	// Check if the new offset is valid
	if newOffset < 0 {
		return 0, os.ErrInvalid
	}

	// Set the new offset
	f.offset = newOffset
	return f.offset, nil
}

// Write writes to the file
func (f *vfsFile) Write(p []byte) (int, error) {
	if f.isDir {
		return 0, os.ErrInvalid
	}

	// For WebDAV PUT operations, we often just want to replace the entire file
	// So if we're writing at offset 0, just replace the entire file
	if f.offset == 0 {
		// Write the data directly to the VFS
		err := f.vfsImpl.FileWrite(f.name, p)
		if err != nil {
			return 0, err
		}
		
		// Update our internal state
		f.data = p
		f.offset = int64(len(p))
		return len(p), nil
	}

	// For other cases, handle partial writes
	// Lazy load file data if not already loaded
	if f.data == nil {
		var err error
		f.data, err = f.vfsImpl.FileRead(f.name)
		if err != nil && !os.IsNotExist(err) {
			return 0, err
		}
		// If the file doesn't exist, create an empty one
		if os.IsNotExist(err) {
			f.data = []byte{}
		}
	}

	// If we're writing at the end of the file, just append
	if f.offset == int64(len(f.data)) {
		f.data = append(f.data, p...)
		n := len(p)
		f.offset += int64(n)

		// Write the data to the VFS
		err := f.vfsImpl.FileWrite(f.name, f.data)
		if err != nil {
			return 0, err
		}

		return n, nil
	}

	// Otherwise, we need to insert the data at the current offset
	newData := make([]byte, f.offset+int64(len(p)))
	copy(newData, f.data[:f.offset])
	copy(newData[f.offset:], p)

	// If there's data after the current offset, append it
	if f.offset+int64(len(p)) < int64(len(f.data)) {
		newData = append(newData, f.data[f.offset+int64(len(p)):]...)
	}

	f.data = newData
	n := len(p)
	f.offset += int64(n)

	// Write the data to the VFS
	err := f.vfsImpl.FileWrite(f.name, f.data)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Readdir reads the contents of the directory
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
				mode:    os.FileMode(metadata.Mode),
				modTime: time.Unix(metadata.ModifiedAt, 0),
				isDir:   entry.IsDir(),
			})
		}
	}

	// If count <= 0, return all entries
	if count <= 0 {
		result := f.dirEnts[f.dirIndex:]
		f.dirIndex = len(f.dirEnts)
		return result, nil
	}

	// Otherwise, return at most count entries
	if f.dirIndex >= len(f.dirEnts) {
		return nil, io.EOF
	}

	end := f.dirIndex + count
	if end > len(f.dirEnts) {
		end = len(f.dirEnts)
	}

	result := f.dirEnts[f.dirIndex:end]
	f.dirIndex = end

	return result, nil
}

// Stat returns file info for this file
func (f *vfsFile) Stat() (os.FileInfo, error) {
	// Handle root directory
	if f.name == "/" {
		return &vfsFileInfo{
			name:    "/",
			size:    0,
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// Get the entry
	entry, err := f.vfsImpl.Get(f.name)
	if err != nil {
		return nil, err
	}

	// Get metadata
	metadata := entry.GetMetadata()

	// Create file info
	return &vfsFileInfo{
		name:    metadata.Name,
		size:    int64(metadata.Size),
		mode:    os.FileMode(metadata.Mode),
		modTime: time.Unix(metadata.ModifiedAt, 0),
		isDir:   entry.IsDir(),
	}, nil
}
