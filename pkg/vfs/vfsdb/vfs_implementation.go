package vfsdb

import (
	"errors"
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// RootGet returns the root filesystem entry
func (fs *DatabaseVFS) RootGet() (vfs.FSEntry, error) {
	return fs.rootGetAsDir()
}

// rootGetAsDir returns the root directory
func (fs *DatabaseVFS) rootGetAsDir() (*DirectoryEntry, error) {
	// Try to load root directory from DB if it exists
	fs.mu.RLock()
	dbID, ok := fs.idTable[fs.rootID]
	fs.mu.RUnlock()

	if ok {
		data, err := fs.dbMetadata.Get(dbID)
		if err == nil {
			root, err := decodeDirectory(data, fs)
			if err == nil {
				return root, nil
			}
		}
	}

	// Create and save new root directory
	root := &DirectoryEntry{
		metadata: &vfs.Metadata{
			ID:          fs.GetNextID(),
			Name:        "",
			FileType:    vfs.FileTypeDirectory,
			Size:        0,
			CreatedAt:   time.Now().Unix(),
			ModifiedAt:  time.Now().Unix(),
			AccessedAt:  time.Now().Unix(),
			Mode:        0755,
			Owner:       "user",
			Group:       "user",
		},
		parentID: 0,
		children: []uint32{},
		vfs:      fs,
	}

	if err := fs.SaveEntry(root); err != nil {
		return nil, fmt.Errorf("failed to set root: %w", err)
	}

	fs.rootID = root.metadata.ID
	return root, nil
}

// FileCreate creates a new file at the specified path
func (fs *DatabaseVFS) FileCreate(path string) (vfs.FSEntry, error) {
	path = vfs.FixPath(path)
	
	// Get parent directory
	parentPath := vfs.PathDir(path)
	fileName := vfs.PathBase(path)
	
	parent, err := fs.getDirectory(parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent directory %s: %w", parentPath, err)
	}
	
	return fs.directoryTouch(parent, fileName)
}

// FileRead reads the content of a file at the specified path
func (fs *DatabaseVFS) FileRead(path string) ([]byte, error) {
	path = vfs.FixPath(path)
	
	entry, err := fs.getEntry(path)
	if err != nil {
		return nil, err
	}
	
	file, ok := entry.(*FileEntry)
	if !ok {
		return nil, vfs.ErrNotFile
	}
	
	var fileData []byte
	
	for _, id := range file.chunkIDs {
		chunkBytes, err := fs.dbData.Get(id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch file data: %w", err)
		}
		fileData = append(fileData, chunkBytes...)
	}
	
	file.metadata.SetAccessed()
	fs.SaveEntry(file)
	
	return fileData, nil
}

// FileWrite writes data to a file at the specified path
func (fs *DatabaseVFS) FileWrite(path string, data []byte) error {
	path = vfs.FixPath(path)
	
	entry, err := fs.getEntry(path)
	if err == nil {
		file, ok := entry.(*FileEntry)
		if !ok {
			return vfs.ErrNotFile
		}
		
		// Delete old chunks if they exist
		for _, id := range file.chunkIDs {
			if err := fs.dbData.Delete(id); err != nil {
				return fmt.Errorf("failed to delete old chunk: %w", err)
			}
		}
		
		// Clear chunk IDs
		file.chunkIDs = []uint32{}
		
		// Split data into chunks
		if len(data) > 0 {
			chunkSize := 64 * 1024 // 64KB chunks
			for i := 0; i < len(data); i += chunkSize {
				end := i + chunkSize
				if end > len(data) {
					end = len(data)
				}
				
				chunk := data[i:end]
				chunkID, err := fs.dbData.Set(chunk)
				if err != nil {
					return fmt.Errorf("failed to save file data chunk: %w", err)
				}
				
				file.chunkIDs = append(file.chunkIDs, chunkID)
			}
		}
		
		// Update file metadata
		file.metadata.Size = uint64(len(data))
		file.metadata.SetModified()
		
		return fs.SaveEntry(file)
	} else {
		// File doesn't exist, create it
		_, err := fs.FileCreate(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		
		return fs.FileWrite(path, data)
	}
}

// FileConcatenate appends data to a file at the specified path
func (fs *DatabaseVFS) FileConcatenate(path string, data []byte) error {
	if len(data) == 0 {
		return nil // Nothing to append
	}
	
	path = vfs.FixPath(path)
	
	entry, err := fs.getEntry(path)
	if err == nil {
		file, ok := entry.(*FileEntry)
		if !ok {
			return vfs.ErrNotFile
		}
		
		// Split new data into chunks
		chunkSize := 64 * 1024 // 64KB chunks
		for i := 0; i < len(data); i += chunkSize {
			end := i + chunkSize
			if end > len(data) {
				end = len(data)
			}
			
			chunk := data[i:end]
			chunkID, err := fs.dbData.Set(chunk)
			if err != nil {
				return fmt.Errorf("failed to save file data chunk: %w", err)
			}
			
			file.chunkIDs = append(file.chunkIDs, chunkID)
		}
		
		// Update file metadata
		file.metadata.Size += uint64(len(data))
		file.metadata.SetModified()
		
		return fs.SaveEntry(file)
	} else {
		// File doesn't exist, create it
		_, err := fs.FileCreate(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		
		return fs.FileWrite(path, data)
	}
}

// FileDelete deletes a file at the specified path
func (fs *DatabaseVFS) FileDelete(path string) error {
	return fs.Delete(path)
}

// DirCreate creates a new directory at the specified path
func (fs *DatabaseVFS) DirCreate(path string) (vfs.FSEntry, error) {
	path = vfs.FixPath(path)
	
	parentPath := vfs.PathDir(path)
	dirName := vfs.PathBase(path)
	
	parent, err := fs.getDirectory(parentPath)
	if err != nil {
		return nil, err
	}
	
	return fs.directoryMkdir(parent, dirName)
}

// DirList lists the entries in a directory at the specified path
func (fs *DatabaseVFS) DirList(path string) ([]vfs.FSEntry, error) {
	path = vfs.FixPath(path)
	
	dir, err := fs.getDirectory(path)
	if err != nil {
		return nil, err
	}
	
	entries := make([]vfs.FSEntry, 0, len(dir.children))
	
	for _, childID := range dir.children {
		child, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		entries = append(entries, child)
	}
	
	dir.metadata.SetAccessed()
	fs.SaveEntry(dir)
	
	return entries, nil
}

// DirDelete deletes a directory at the specified path
func (fs *DatabaseVFS) DirDelete(path string) error {
	return fs.Delete(path)
}

// LinkCreate creates a new symlink
func (fs *DatabaseVFS) LinkCreate(targetPath, linkPath string) (vfs.FSEntry, error) {
	linkPath = vfs.FixPath(linkPath)
	
	parentPath := vfs.PathDir(linkPath)
	linkName := vfs.PathBase(linkPath)
	
	parent, err := fs.getDirectory(parentPath)
	if err != nil {
		return nil, err
	}
	
	// Create new symlink
	symlink := &SymlinkEntry{
		metadata: &vfs.Metadata{
			ID:          fs.GetNextID(),
			Name:        linkName,
			FileType:    vfs.FileTypeSymlink,
			CreatedAt:   time.Now().Unix(),
			ModifiedAt:  time.Now().Unix(),
			AccessedAt:  time.Now().Unix(),
			Mode:        0777,
			Owner:       "user",
			Group:       "user",
		},
		target:   targetPath,
		parentID: parent.metadata.ID,
		vfs:      fs,
	}
	
	// Add symlink to parent directory
	parent.children = append(parent.children, symlink.metadata.ID)
	
	// Save both entries
	if err := fs.SaveEntry(symlink); err != nil {
		return nil, err
	}
	
	if err := fs.SaveEntry(parent); err != nil {
		return nil, err
	}
	
	return symlink, nil
}

// LinkRead reads the target of a symlink
func (fs *DatabaseVFS) LinkRead(path string) (string, error) {
	path = vfs.FixPath(path)
	
	entry, err := fs.getEntry(path)
	if err != nil {
		return "", err
	}
	
	symlink, ok := entry.(*SymlinkEntry)
	if !ok {
		return "", vfs.ErrNotSymlink
	}
	
	return symlink.GetTarget()
}

// LinkDelete deletes a symlink
func (fs *DatabaseVFS) LinkDelete(path string) error {
	return fs.Delete(path)
}

// Exists checks if a path exists
func (fs *DatabaseVFS) Exists(path string) bool {
	path = vfs.FixPath(path)
	
	if path == "/" {
		return true
	}
	
	_, err := fs.getEntry(path)
	return err == nil
}

// Get returns the filesystem entry at the specified path
func (fs *DatabaseVFS) Get(path string) (vfs.FSEntry, error) {
	path = vfs.FixPath(path)
	
	return fs.getEntry(path)
}

// Rename renames a filesystem entry
func (fs *DatabaseVFS) Rename(oldPath, newPath string) (vfs.FSEntry, error) {
	oldPath = vfs.FixPath(oldPath)
	newPath = vfs.FixPath(newPath)
	
	srcParentPath := vfs.PathDir(oldPath)
	srcName := vfs.PathBase(oldPath)
	dstName := vfs.PathBase(newPath)
	
	srcParent, err := fs.getDirectory(srcParentPath)
	if err != nil {
		return nil, err
	}
	
	return fs.directoryRename(srcParent, srcName, dstName)
}

// Copy copies a filesystem entry
func (fs *DatabaseVFS) Copy(srcPath, dstPath string) (vfs.FSEntry, error) {
	srcPath = vfs.FixPath(srcPath)
	dstPath = vfs.FixPath(dstPath)
	
	srcParentPath := vfs.PathDir(srcPath)
	dstParentPath := vfs.PathDir(dstPath)
	
	if !fs.Exists(srcParentPath) {
		return nil, fmt.Errorf("%s does not exist", srcParentPath)
	}
	
	if !fs.Exists(dstParentPath) {
		return nil, fmt.Errorf("%s does not exist", dstParentPath)
	}
	
	srcName := vfs.PathBase(srcPath)
	dstName := vfs.PathBase(dstPath)
	
	srcParent, err := fs.getDirectory(srcParentPath)
	if err != nil {
		return nil, err
	}
	
	dstParent, err := fs.getDirectory(dstParentPath)
	if err != nil {
		return nil, err
	}
	
	if srcParent.metadata.ID == dstParent.metadata.ID && srcName == dstName {
		return nil, errors.New("source and destination are the same")
	}
	
	return fs.directoryCopy(srcParent, srcName, dstName, dstParent)
}

// Move moves a filesystem entry
func (fs *DatabaseVFS) Move(srcPath, dstPath string) (vfs.FSEntry, error) {
	srcPath = vfs.FixPath(srcPath)
	dstPath = vfs.FixPath(dstPath)
	
	srcParentPath := vfs.PathDir(srcPath)
	dstParentPath := vfs.PathDir(dstPath)
	
	if !fs.Exists(srcParentPath) {
		return nil, fmt.Errorf("%s does not exist", srcParentPath)
	}
	
	if !fs.Exists(dstParentPath) {
		return nil, fmt.Errorf("%s does not exist", dstParentPath)
	}
	
	srcName := vfs.PathBase(srcPath)
	dstName := vfs.PathBase(dstPath)
	
	srcParent, err := fs.getDirectory(srcParentPath)
	if err != nil {
		return nil, err
	}
	
	dstParent, err := fs.getDirectory(dstParentPath)
	if err != nil {
		return nil, err
	}
	
	if srcParent.metadata.ID == dstParent.metadata.ID && srcName == dstName {
		return nil, errors.New("source and destination are the same")
	}
	
	return fs.directoryMove(srcParent, srcName, dstName, dstParent)
}

// Delete deletes a filesystem entry
func (fs *DatabaseVFS) Delete(path string) error {
	if path == "/" || path == "" || path == "." {
		return errors.New("cannot delete root")
	}
	
	path = vfs.FixPath(path)
	
	parentPath := vfs.PathDir(path)
	name := vfs.PathBase(path)
	
	parent, err := fs.getDirectory(parentPath)
	if err != nil {
		return err
	}
	
	return fs.directoryRm(parent, name)
}

// Destroy cleans up resources
func (fs *DatabaseVFS) Destroy() error {
	// Nothing to do as the database handles cleanup
	return nil
}

// GetPath returns the full path for an entry
func (fs *DatabaseVFS) GetPath(entry vfs.FSEntry) (string, error) {
	var fsEntry FSEntry
	var ok bool
	
	if fsEntry, ok = entry.(FSEntry); !ok {
		return "", errors.New("invalid entry type")
	}
	
	metadata := fsEntry.GetMetadata()
	
	// Special case for root
	if metadata.ID == fs.rootID {
		return "/", nil
	}
	
	var parentID uint32
	
	switch e := fsEntry.(type) {
	case *DirectoryEntry:
		parentID = e.parentID
	case *FileEntry:
		parentID = e.parentID
	case *SymlinkEntry:
		parentID = e.parentID
	default:
		return "", errors.New("unknown entry type")
	}
	
	if parentID == 0 {
		return "/" + metadata.Name, nil
	}
	
	parent, err := fs.LoadEntry(parentID)
	if err != nil {
		return "", err
	}
	
	parentPath, err := fs.GetPath(parent)
	if err != nil {
		return "", err
	}
	
	return vfs.JoinPath(parentPath, metadata.Name), nil
}
