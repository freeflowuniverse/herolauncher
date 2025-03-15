// Package vfsnested provides a virtual filesystem implementation that can contain multiple nested VFS instances
package vfsnested

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// NestedVFS represents a VFS that can contain multiple nested VFS instances
type NestedVFS struct {
	vfsMap map[string]vfs.VFSImplementation // Map of path prefixes to VFS implementations
	mu     sync.RWMutex                     // Mutex for thread safety
}

// New creates a new NestedVFS instance
func New() *NestedVFS {
	return &NestedVFS{
		vfsMap: make(map[string]vfs.VFSImplementation),
	}
}

// AddVFS adds a new VFS implementation at the specified path prefix
func (n *NestedVFS) AddVFS(prefix string, impl vfs.VFSImplementation) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.vfsMap[prefix]; exists {
		return fmt.Errorf("VFS already exists at prefix: %s", prefix)
	}
	n.vfsMap[prefix] = impl
	return nil
}

// findVFS finds the appropriate VFS implementation for a given path
func (n *NestedVFS) findVFS(path string) (vfs.VFSImplementation, string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if path == "" || path == "/" {
		return n, "/", nil
	}

	// Sort prefixes by length (longest first) to match most specific path
	var prefixes []string
	for prefix := range n.vfsMap {
		prefixes = append(prefixes, prefix)
	}

	// Sort by length in descending order
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			relativePath := path[len(prefix):]
			if !strings.HasPrefix(relativePath, "/") {
				relativePath = "/" + relativePath
			}
			return n.vfsMap[prefix], relativePath, nil
		}
	}

	return nil, "", fmt.Errorf("no VFS found for path: %s", path)
}

// Implementation of VFSImplementation interface

// RootGet returns the root filesystem entry
func (n *NestedVFS) RootGet() (vfs.FSEntry, error) {
	// Return a special root entry that represents the nested VFS
	return &RootEntry{
		metadata: &vfs.Metadata{
			ID:         0,
			Name:       "",
			FileType:   vfs.FileTypeDirectory,
			Size:       0,
			CreatedAt:  0,
			ModifiedAt: 0,
			AccessedAt: 0,
		},
	}, nil
}

// Delete deletes a filesystem entry
func (n *NestedVFS) Delete(path string) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.Delete(relPath)
}

// LinkDelete deletes a symlink
func (n *NestedVFS) LinkDelete(path string) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.LinkDelete(relPath)
}

// FileCreate creates a new file
func (n *NestedVFS) FileCreate(path string) (vfs.FSEntry, error) {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return nil, err
	}

	subEntry, err := impl.FileCreate(relPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == impl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(subEntry, prefix), nil
}

// FileRead reads the content of a file
func (n *NestedVFS) FileRead(path string) ([]byte, error) {
	// Special handling for macOS resource fork files (._*)
	if strings.HasPrefix(path, "/._") || strings.Contains(path, "/._") {
		// Return empty data for resource fork files
		return []byte{}, nil
	}

	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return nil, err
	}
	return impl.FileRead(relPath)
}

// FileWrite writes data to a file
func (n *NestedVFS) FileWrite(path string, data []byte) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.FileWrite(relPath, data)
}

// FileConcatenate appends data to a file
func (n *NestedVFS) FileConcatenate(path string, data []byte) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.FileConcatenate(relPath, data)
}

// FileDelete deletes a file
func (n *NestedVFS) FileDelete(path string) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.FileDelete(relPath)
}

// DirCreate creates a new directory
func (n *NestedVFS) DirCreate(path string) (vfs.FSEntry, error) {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return nil, err
	}

	subEntry, err := impl.DirCreate(relPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == impl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(subEntry, prefix), nil
}

// DirList lists the entries in a directory
func (n *NestedVFS) DirList(path string) ([]vfs.FSEntry, error) {
	// Special case for root directory
	if path == "" || path == "/" {
		var entries []vfs.FSEntry

		n.mu.RLock()
		for prefix, impl := range n.vfsMap {
			// Skip errors when getting root entry
			_, err := impl.RootGet()
			if err != nil {
				continue
			}

			entries = append(entries, &MountEntry{
				metadata: &vfs.Metadata{
					ID:         0,
					Name:       prefix,
					FileType:   vfs.FileTypeDirectory,
					Size:       0,
					CreatedAt:  0,
					ModifiedAt: 0,
					AccessedAt: 0,
				},
				impl: impl,
			})
		}
		n.mu.RUnlock()

		return entries, nil
	}

	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return nil, err
	}

	subEntries, err := impl.DirList(relPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == impl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	// Convert all entries to nested entries
	var entries []vfs.FSEntry
	for _, subEntry := range subEntries {
		entries = append(entries, n.nesterEntry(subEntry, prefix))
	}

	return entries, nil
}

// DirDelete deletes a directory
func (n *NestedVFS) DirDelete(path string) error {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return err
	}
	return impl.DirDelete(relPath)
}

// Exists checks if a path exists
func (n *NestedVFS) Exists(path string) bool {
	// Root always exists
	if path == "" || path == "/" {
		return true
	}

	// Special handling for macOS resource fork files (._*)
	if strings.HasPrefix(path, "/._") || strings.Contains(path, "/._") {
		return true // Pretend these files exist for WebDAV Class 2 compatibility
	}

	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return false
	}
	return impl.Exists(relPath)
}

// Get returns the filesystem entry at the specified path
func (n *NestedVFS) Get(path string) (vfs.FSEntry, error) {
	if path == "" || path == "/" {
		return n.RootGet()
	}

	// Special handling for macOS resource fork files (._*)
	if strings.HasPrefix(path, "/._") || strings.Contains(path, "/._") {
		// Extract the filename from the path
		parts := strings.Split(path, "/")
		filename := parts[len(parts)-1]

		// Create a dummy resource fork entry
		return &ResourceForkEntry{
			metadata: &vfs.Metadata{
				ID:         0,
				Name:       filename,
				FileType:   vfs.FileTypeFile,
				Size:       0,
				CreatedAt:  time.Now().Unix(),
				ModifiedAt: time.Now().Unix(),
				AccessedAt: time.Now().Unix(),
			},
			path: path,
		}, nil
	}

	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return nil, err
	}

	// Convert entry of nested VFS to entry of nester
	subEntry, err := impl.Get(relPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == impl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(subEntry, prefix), nil
}

// nesterEntry converts an FSEntry from a sub VFS to an FSEntry for the nester VFS
// by prefixing the nested VFS's path onto the FSEntry's path
func (n *NestedVFS) nesterEntry(entry vfs.FSEntry, prefix string) vfs.FSEntry {
	return &NestedEntry{
		original: entry,
		prefix:   prefix,
	}
}

// Rename renames a filesystem entry
func (n *NestedVFS) Rename(oldPath, newPath string) (vfs.FSEntry, error) {
	oldImpl, oldRelPath, err := n.findVFS(oldPath)
	if err != nil {
		return nil, err
	}

	newImpl, newRelPath, err := n.findVFS(newPath)
	if err != nil {
		return nil, err
	}

	if oldImpl != newImpl {
		return nil, errors.New("cannot rename across different VFS implementations")
	}

	renamedFile, err := oldImpl.Rename(oldRelPath, newRelPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == oldImpl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(renamedFile, prefix), nil
}

// Copy copies a filesystem entry
func (n *NestedVFS) Copy(srcPath, dstPath string) (vfs.FSEntry, error) {
	srcImpl, srcRelPath, err := n.findVFS(srcPath)
	if err != nil {
		return nil, err
	}

	dstImpl, dstRelPath, err := n.findVFS(dstPath)
	if err != nil {
		return nil, err
	}

	if srcImpl == dstImpl {
		copiedFile, err := srcImpl.Copy(srcRelPath, dstRelPath)
		if err != nil {
			return nil, err
		}

		// Find the prefix for this VFS implementation
		var prefix string
		n.mu.RLock()
		for p, v := range n.vfsMap {
			if v == srcImpl {
				prefix = p
				break
			}
		}
		n.mu.RUnlock()

		return n.nesterEntry(copiedFile, prefix), nil
	}

	// Copy across different VFS implementations
	data, err := srcImpl.FileRead(srcRelPath)
	if err != nil {
		return nil, err
	}

	newFile, err := dstImpl.FileCreate(dstRelPath)
	if err != nil {
		return nil, err
	}

	err = dstImpl.FileWrite(dstRelPath, data)
	if err != nil {
		return nil, err
	}

	// Find the prefix for the destination VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == dstImpl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(newFile, prefix), nil
}

// Move moves a filesystem entry
func (n *NestedVFS) Move(srcPath, dstPath string) (vfs.FSEntry, error) {
	srcImpl, srcRelPath, err := n.findVFS(srcPath)
	if err != nil {
		return nil, err
	}

	dstImpl, dstRelPath, err := n.findVFS(dstPath)
	if err != nil {
		return nil, err
	}

	if srcImpl != dstImpl {
		return nil, errors.New("cannot move across different VFS implementations")
	}

	movedFile, err := srcImpl.Move(srcRelPath, dstRelPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == srcImpl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(movedFile, prefix), nil
}

// LinkCreate creates a new symlink
func (n *NestedVFS) LinkCreate(targetPath, linkPath string) (vfs.FSEntry, error) {
	impl, relPath, err := n.findVFS(linkPath)
	if err != nil {
		return nil, err
	}

	linkEntry, err := impl.LinkCreate(targetPath, relPath)
	if err != nil {
		return nil, err
	}

	// Find the prefix for this VFS implementation
	var prefix string
	n.mu.RLock()
	for p, v := range n.vfsMap {
		if v == impl {
			prefix = p
			break
		}
	}
	n.mu.RUnlock()

	return n.nesterEntry(linkEntry, prefix), nil
}

// LinkRead reads the target of a symlink
func (n *NestedVFS) LinkRead(path string) (string, error) {
	impl, relPath, err := n.findVFS(path)
	if err != nil {
		return "", err
	}
	return impl.LinkRead(relPath)
}

// Destroy cleans up resources
func (n *NestedVFS) Destroy() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var lastErr error
	for _, impl := range n.vfsMap {
		if err := impl.Destroy(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetPath returns the path for the given entry
func (n *NestedVFS) GetPath(entry vfs.FSEntry) (string, error) {
	switch e := entry.(type) {
	case *RootEntry:
		return "/", nil
	case *MountEntry:
		return "/" + strings.TrimLeft(e.metadata.Name, "/"), nil
	case *NestedEntry:
		// Get the path from the original entry's metadata
		originalMeta := e.original.GetMetadata()
		originalName := originalMeta.Name
		
		// For root entries, just return the prefix
		if originalName == "" || originalName == "/" {
			return e.prefix, nil
		}
		
		return e.prefix + "/" + strings.TrimLeft(originalName, "/"), nil
	case *ResourceForkEntry:
		return e.path, nil
	default:
		return "", errors.New("unknown entry type")
	}
}
