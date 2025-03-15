package main

import (
	"fmt"
	"path"
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

// VFSDBDir implements the fs.Dir interface for vfsdb
type VFSDBDir struct {
	fs.BaseNode
	vfsImpl vfs.VFSImplementation
	path    string
	mu      sync.RWMutex
}

// NewVFSDBDir creates a new VFSDBDir
func NewVFSDBDir(s *proto.Stat, vfsImpl vfs.VFSImplementation, path string) *VFSDBDir {
	return &VFSDBDir{
		BaseNode: fs.BaseNode{
			FStat: *s,
		},
		vfsImpl: vfsImpl,
		path:    path,
	}
}

// Children implements fs.Dir.Children
func (d *VFSDBDir) Children() map[string]fs.FSNode {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// List the directory contents
	entries, err := d.vfsImpl.DirList(d.path)
	if err != nil {
		// If there's an error, return an empty map
		return map[string]fs.FSNode{}
	}

	// Create a map of children
	children := make(map[string]fs.FSNode)
	for _, entry := range entries {
		metadata := entry.GetMetadata()
		name := metadata.Name
		entryPath := path.Join(d.path, name)

		// Create a stat for the entry
		stat := proto.Stat{
			Type:   0,
			Dev:    0,
			Qid: proto.Qid{
				Qtype: uint8(metadata.Mode >> 24),
				Vers:  0,
				Uid:   uint64(metadata.ID),
			},
			Mode:   metadata.Mode,
			Atime:  uint32(metadata.AccessedAt),
			Mtime:  uint32(metadata.ModifiedAt),
			Length: metadata.Size,
			Name:   name,
			Uid:    metadata.Owner,
			Gid:    metadata.Group,
			Muid:   metadata.Owner,
		}

		if entry.IsDir() {
			// Create a directory node
			children[name] = NewVFSDBDir(&stat, d.vfsImpl, entryPath)
		} else {
			// Create a file node
			children[name] = NewVFSDBFile(&stat, d.vfsImpl, entryPath)
		}
	}

	return children
}

// AddChild implements fs.ModDir.AddChild
func (d *VFSDBDir) AddChild(n fs.FSNode) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get the name of the child
	name := n.Stat().Name
	childPath := path.Join(d.path, name)

	// Check if the child already exists
	if d.vfsImpl.Exists(childPath) {
		return fmt.Errorf("%s already exists", childPath)
	}

	// If the child is a directory, create it in vfsdb
	if n.Stat().Mode&proto.DMDIR > 0 {
		_, err := d.vfsImpl.DirCreate(childPath)
		if err != nil {
			return fmt.Errorf("failed to create directory in vfsdb: %w", err)
		}
	}

	// Set the parent of the child
	n.SetParent(d)
	return nil
}

// DeleteChild implements fs.ModDir.DeleteChild
func (d *VFSDBDir) DeleteChild(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get the full path of the child
	childPath := path.Join(d.path, name)

	// Check if the child exists
	if !d.vfsImpl.Exists(childPath) {
		return fmt.Errorf("%s does not exist", childPath)
	}

	// Get the child entry
	entry, err := d.vfsImpl.Get(childPath)
	if err != nil {
		return fmt.Errorf("failed to get child entry: %w", err)
	}

	// Delete the child based on its type
	if entry.IsDir() {
		err = d.vfsImpl.DirDelete(childPath)
	} else {
		err = d.vfsImpl.FileDelete(childPath)
	}

	if err != nil {
		return fmt.Errorf("failed to delete child: %w", err)
	}

	return nil
}

// createVFSDBDir returns a function that creates a VFSDBDir
func createVFSDBDir(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.Dir, error) {
	return func(fsys *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.Dir, error) {
		// Get the full path for the directory
		parentPath := getFullPath(parent)
		dirPath := path.Join(parentPath, name)

		// Create a new directory stat
		stat := fsys.NewStat(name, user, user, perm|proto.DMDIR)
		
		// Create a new VFSDBDir
		dir := NewVFSDBDir(stat, vfsImpl, dirPath)
		
		// Add the directory to the parent directory
		modParent, ok := parent.(fs.ModDir)
		if !ok {
			return nil, fmt.Errorf("%s does not support modification", fs.FullPath(parent))
		}
		
		err := modParent.AddChild(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to add directory to parent directory: %w", err)
		}
		
		// Create the directory in the vfsdb
		_, err = vfsImpl.DirCreate(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory in vfsdb: %w", err)
		}
		
		return dir, nil
	}
}

// removeVFSDBFile returns a function that removes a file or directory
func removeVFSDBFile(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, f fs.FSNode) error {
	return func(fsys *fs.FS, f fs.FSNode) error {
		// Get the full path of the file or directory
		nodePath := fs.FullPath(f)
		
		// Check if the node exists
		if !vfsImpl.Exists(nodePath) {
			return fmt.Errorf("%s does not exist", nodePath)
		}
		
		// Get the parent directory
		parent, ok := f.Parent().(fs.ModDir)
		if !ok {
			return fmt.Errorf("%s does not support modification", fs.FullPath(f.Parent()))
		}
		
		// Delete the child from the parent directory
		err := parent.DeleteChild(f.Stat().Name)
		if err != nil {
			return fmt.Errorf("failed to delete child from parent directory: %w", err)
		}
		
		// Delete the node from vfsdb
		err = vfsImpl.Delete(nodePath)
		if err != nil {
			return fmt.Errorf("failed to delete node from vfsdb: %w", err)
		}
		
		return nil
	}
}
