package main

import (
	"errors"
	"fmt"
	"path"
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

// VFSDBFile implements the fs.File interface for vfsdb
type VFSDBFile struct {
	fs.BaseFile
	vfsImpl vfs.VFSImplementation
	path    string
	fidMap  map[uint64]uint64 // Maps fids to offsets
	mu      sync.RWMutex
}

// NewVFSDBFile creates a new VFSDBFile
func NewVFSDBFile(s *proto.Stat, vfsImpl vfs.VFSImplementation, path string) *VFSDBFile {
	return &VFSDBFile{
		BaseFile: fs.BaseFile{},
		vfsImpl:  vfsImpl,
		path:     path,
		fidMap:   make(map[uint64]uint64),
	}
}

// Open implements fs.File.Open
func (f *VFSDBFile) Open(fid uint64, omode proto.Mode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if file exists, if not and omode is OWRITE or ORDWR, create it
	if !f.vfsImpl.Exists(f.path) && (omode&proto.Owrite > 0 || omode&proto.Ordwr > 0) {
		_, err := f.vfsImpl.FileCreate(f.path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	// Handle truncate flag
	if omode&proto.Otrunc > 0 {
		err := f.vfsImpl.FileWrite(f.path, []byte{})
		if err != nil {
			return fmt.Errorf("failed to truncate file: %w", err)
		}
	}

	// Initialize offset for this fid
	f.fidMap[fid] = 0
	return nil
}

// Read implements fs.File.Read
func (f *VFSDBFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Read the entire file
	data, err := f.vfsImpl.FileRead(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check if offset is beyond the file size
	if offset >= uint64(len(data)) {
		return []byte{}, nil
	}

	// Adjust count if it would read beyond the end of the file
	if offset+count > uint64(len(data)) {
		count = uint64(len(data)) - offset
	}

	// Return the requested portion of the file
	return data[offset : offset+count], nil
}

// Write implements fs.File.Write
func (f *VFSDBFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Read the entire file
	fileData, err := f.vfsImpl.FileRead(f.path)
	if err != nil && !isNotExist(err) {
		return 0, fmt.Errorf("failed to read file for writing: %w", err)
	}

	// If file doesn't exist or offset is beyond current size, extend the file
	if len(fileData) < int(offset+uint64(len(data))) {
		newData := make([]byte, offset+uint64(len(data)))
		copy(newData, fileData)
		fileData = newData
	}

	// Copy the new data at the specified offset
	copy(fileData[offset:], data)

	// Write the modified file back
	err = f.vfsImpl.FileWrite(f.path, fileData)
	if err != nil {
		return 0, fmt.Errorf("failed to write file: %w", err)
	}

	return uint32(len(data)), nil
}

// Close implements fs.File.Close
func (f *VFSDBFile) Close(fid uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Remove the fid from the map
	delete(f.fidMap, fid)
	return nil
}

// Stat implements fs.FSNode.Stat
func (f *VFSDBFile) Stat() proto.Stat {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Get the file metadata
	entry, err := f.vfsImpl.Get(f.path)
	if err != nil {
		// If file doesn't exist, return the base stat
		return f.BaseFile.Stat()
	}

	// Update the stat with the file metadata
	metadata := entry.GetMetadata()
	stat := f.BaseFile.Stat()
	stat.Length = metadata.Size
	stat.Atime = uint32(metadata.AccessedAt)
	stat.Mtime = uint32(metadata.ModifiedAt)
	stat.Mode = metadata.Mode
	stat.Name = path.Base(f.path)
	stat.Uid = metadata.Owner
	stat.Gid = metadata.Group

	return stat
}

// createVFSDBFile returns a function that creates a VFSDBFile
func createVFSDBFile(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.File, error) {
	return func(fsys *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.File, error) {
		// Get the full path for the file
		parentPath := getFullPath(parent)
		filePath := path.Join(parentPath, name)

		// Create a new file stat
		stat := fsys.NewStat(name, user, user, perm)
		
		// Create a new VFSDBFile
		file := NewVFSDBFile(stat, vfsImpl, filePath)
		
		// Add the file to the parent directory
		modParent, ok := parent.(fs.ModDir)
		if !ok {
			return nil, fmt.Errorf("%s does not support modification", fs.FullPath(parent))
		}
		
		err := modParent.AddChild(file)
		if err != nil {
			return nil, fmt.Errorf("failed to add file to parent directory: %w", err)
		}
		
		// Create the file in the vfsdb
		_, err = vfsImpl.FileCreate(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in vfsdb: %w", err)
		}
		
		return file, nil
	}
}

// Helper function to get the full path of a directory
func getFullPath(dir fs.Dir) string {
	if dir == nil {
		return "/"
	}
	
	fullPath := fs.FullPath(dir)
	if fullPath == "" {
		return "/"
	}
	
	return fullPath
}

// Helper function to check if an error is a "not exist" error
func isNotExist(err error) bool {
	return err != nil && (errors.Is(err, vfs.ErrNotFound) || err.Error() == "file does not exist")
}
