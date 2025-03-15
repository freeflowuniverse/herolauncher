package vfsdb

import (
	"errors"
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// getEntry retrieves a filesystem entry at the specified path
func (fs *DatabaseVFS) getEntry(path string) (FSEntry, error) {
	if path == "/" || path == "" || path == "." {
		return fs.rootGetAsDir()
	}
	
	parts := vfs.SplitPath(path)
	
	root, err := fs.rootGetAsDir()
	if err != nil {
		return nil, err
	}
	
	var current = root
	
	for i, part := range parts {
		found := false
		
		for _, childID := range current.children {
			entry, err := fs.LoadEntry(childID)
			if err != nil {
				return nil, fmt.Errorf("failed to load child entry: %w", err)
			}
			
			if entry.GetMetadata().Name == part {
				found = true
				
				if i == len(parts)-1 {
					// Last part, return the entry
					return entry, nil
				}
				
				// Not the last part, must be a directory
				if dir, ok := entry.(*DirectoryEntry); ok {
					current = dir
					break
				} else {
					return nil, fmt.Errorf("path component %s is not a directory", part)
				}
			}
		}
		
		if !found {
			return nil, vfs.ErrNotFound
		}
	}
	
	return nil, vfs.ErrNotFound
}

// getDirectory retrieves a directory at the specified path
func (fs *DatabaseVFS) getDirectory(path string) (*DirectoryEntry, error) {
	entry, err := fs.getEntry(path)
	if err != nil {
		return nil, err
	}
	
	dir, ok := entry.(*DirectoryEntry)
	if !ok {
		return nil, vfs.ErrNotDirectory
	}
	
	return dir, nil
}

// directoryMkdir creates a new directory in the parent directory
func (fs *DatabaseVFS) directoryMkdir(parent *DirectoryEntry, name string) (*DirectoryEntry, error) {
	// Check if directory already exists
	for _, childID := range parent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == name {
			return nil, vfs.ErrAlreadyExists
		}
	}
	
	// Create new directory
	newDir := &DirectoryEntry{
		metadata: &vfs.Metadata{
			ID:          fs.GetNextID(),
			Name:        name,
			FileType:    vfs.FileTypeDirectory,
			CreatedAt:   time.Now().Unix(),
			ModifiedAt:  time.Now().Unix(),
			AccessedAt:  time.Now().Unix(),
			Mode:        0755,
			Owner:       "user",
			Group:       "user",
		},
		parentID: parent.metadata.ID,
		children: []uint32{},
		vfs:      fs,
	}
	
	// Save new directory
	if err := fs.SaveEntry(newDir); err != nil {
		return nil, err
	}
	
	// Update parent directory
	parent.children = append(parent.children, newDir.metadata.ID)
	parent.metadata.SetModified()
	
	if err := fs.SaveEntry(parent); err != nil {
		return nil, err
	}
	
	return newDir, nil
}

// directoryTouch creates a new empty file in the parent directory
func (fs *DatabaseVFS) directoryTouch(parent *DirectoryEntry, name string) (*FileEntry, error) {
	// Check if file already exists
	for _, childID := range parent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == name {
			return nil, vfs.ErrAlreadyExists
		}
	}
	
	// Create new file
	file := &FileEntry{
		metadata: &vfs.Metadata{
			ID:          fs.GetNextID(),
			Name:        name,
			FileType:    vfs.FileTypeFile,
			Size:        0,
			CreatedAt:   time.Now().Unix(),
			ModifiedAt:  time.Now().Unix(),
			AccessedAt:  time.Now().Unix(),
			Mode:        0644,
			Owner:       "user",
			Group:       "user",
		},
		parentID: parent.metadata.ID,
		chunkIDs: []uint32{},
		vfs:      fs,
	}
	
	// Save new file
	if err := fs.SaveEntry(file); err != nil {
		return nil, err
	}
	
	// Update parent directory
	parent.children = append(parent.children, file.metadata.ID)
	parent.metadata.SetModified()
	
	if err := fs.SaveEntry(parent); err != nil {
		return nil, err
	}
	
	return file, nil
}

// directoryRm removes a file or directory from the parent directory
func (fs *DatabaseVFS) directoryRm(parent *DirectoryEntry, name string) error {
	var entryToRemove FSEntry
	var entryID uint32
	var entryIndex int
	
	// Find the entry to remove
	for i, childID := range parent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == name {
			entryToRemove = entry
			entryID = childID
			entryIndex = i
			break
		}
	}
	
	if entryToRemove == nil {
		return vfs.ErrNotFound
	}
	
	// Check if it's a non-empty directory
	if dir, ok := entryToRemove.(*DirectoryEntry); ok {
		if len(dir.children) > 0 {
			return vfs.ErrNotEmpty
		}
	}
	
	// If it's a file, delete its chunks
	if file, ok := entryToRemove.(*FileEntry); ok {
		for _, chunkID := range file.chunkIDs {
			if err := fs.dbData.Delete(chunkID); err != nil {
				return fmt.Errorf("failed to delete file chunk: %w", err)
			}
		}
	}
	
	// Delete the entry from metadata DB
	fs.mu.RLock()
	dbID, ok := fs.idTable[entryID]
	fs.mu.RUnlock()
	
	if ok {
		if err := fs.dbMetadata.Delete(dbID); err != nil {
			return fmt.Errorf("failed to delete entry metadata: %w", err)
		}
	}
	
	// Remove the entry from the parent's children list
	parent.children = append(parent.children[:entryIndex], parent.children[entryIndex+1:]...)
	parent.metadata.SetModified()
	
	// Save the updated parent
	if err := fs.SaveEntry(parent); err != nil {
		return fmt.Errorf("failed to save updated parent directory: %w", err)
	}
	
	// Remove the entry from the ID table
	fs.mu.Lock()
	delete(fs.idTable, entryID)
	fs.mu.Unlock()
	
	return nil
}

// directoryRename renames a file or directory in the parent directory
func (fs *DatabaseVFS) directoryRename(parent *DirectoryEntry, srcName, dstName string) (FSEntry, error) {
	var entryToRename FSEntry
	
	// Find the entry to rename
	for _, childID := range parent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == srcName {
			entryToRename = entry
			break
		}
	}
	
	if entryToRename == nil {
		return nil, vfs.ErrNotFound
	}
	
	// Check if destination name already exists
	for _, childID := range parent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == dstName {
			return nil, vfs.ErrAlreadyExists
		}
	}
	
	// Update the entry's name
	entryToRename.GetMetadata().Name = dstName
	entryToRename.GetMetadata().SetModified()
	
	// Save the updated entry
	if err := fs.SaveEntry(entryToRename); err != nil {
		return nil, fmt.Errorf("failed to save renamed entry: %w", err)
	}
	
	return entryToRename, nil
}

// directoryCopy copies a file or directory to a new location
func (fs *DatabaseVFS) directoryCopy(srcParent *DirectoryEntry, srcName, dstName string, dstParent *DirectoryEntry) (FSEntry, error) {
	var srcEntry FSEntry
	
	// Find the source entry
	for _, childID := range srcParent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == srcName {
			srcEntry = entry
			break
		}
	}
	
	if srcEntry == nil {
		return nil, vfs.ErrNotFound
	}
	
	// Check if destination name already exists
	for _, childID := range dstParent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == dstName {
			return nil, vfs.ErrAlreadyExists
		}
	}
	
	var newEntry FSEntry
	
	// Create a copy based on the entry type
	switch e := srcEntry.(type) {
	case *DirectoryEntry:
		// Create a new directory
		newDir := &DirectoryEntry{
			metadata: &vfs.Metadata{
				ID:          fs.GetNextID(),
				Name:        dstName,
				FileType:    vfs.FileTypeDirectory,
				CreatedAt:   time.Now().Unix(),
				ModifiedAt:  time.Now().Unix(),
				AccessedAt:  time.Now().Unix(),
				Mode:        e.metadata.Mode,
				Owner:       e.metadata.Owner,
				Group:       e.metadata.Group,
			},
			parentID: dstParent.metadata.ID,
			children: []uint32{},
			vfs:      fs,
		}
		
		// Save the new directory
		if err := fs.SaveEntry(newDir); err != nil {
			return nil, err
		}
		
		// Recursively copy children
		for _, childID := range e.children {
			child, err := fs.LoadEntry(childID)
			if err != nil {
				return nil, fmt.Errorf("failed to load child entry: %w", err)
			}
			
			// Create a path for the child
			childPath, err := fs.GetPath(child)
			if err != nil {
				return nil, fmt.Errorf("failed to get child path: %w", err)
			}
			
			// Create a destination path for the child
			dstPath, err := fs.GetPath(newDir)
			if err != nil {
				return nil, fmt.Errorf("failed to get destination path: %w", err)
			}
			dstPath = vfs.JoinPath(dstPath, child.GetMetadata().Name)
			
			// Copy the child
			_, err = fs.Copy(childPath, dstPath)
			if err != nil {
				return nil, fmt.Errorf("failed to copy child: %w", err)
			}
		}
		
		newEntry = newDir
		
	case *FileEntry:
		// Create a new file
		newFile := &FileEntry{
			metadata: &vfs.Metadata{
				ID:          fs.GetNextID(),
				Name:        dstName,
				FileType:    vfs.FileTypeFile,
				Size:        e.metadata.Size,
				CreatedAt:   time.Now().Unix(),
				ModifiedAt:  time.Now().Unix(),
				AccessedAt:  time.Now().Unix(),
				Mode:        e.metadata.Mode,
				Owner:       e.metadata.Owner,
				Group:       e.metadata.Group,
			},
			parentID: dstParent.metadata.ID,
			chunkIDs: []uint32{},
			vfs:      fs,
		}
		
		// Copy chunks
		for _, chunkID := range e.chunkIDs {
			chunkData, err := fs.dbData.Get(chunkID)
			if err != nil {
				return nil, fmt.Errorf("failed to get chunk data: %w", err)
			}
			
			newChunkID, err := fs.dbData.Set(chunkData)
			if err != nil {
				return nil, fmt.Errorf("failed to save chunk data: %w", err)
			}
			
			newFile.chunkIDs = append(newFile.chunkIDs, newChunkID)
		}
		
		// Save the new file
		if err := fs.SaveEntry(newFile); err != nil {
			return nil, err
		}
		
		newEntry = newFile
		
	case *SymlinkEntry:
		// Create a new symlink
		newSymlink := &SymlinkEntry{
			metadata: &vfs.Metadata{
				ID:          fs.GetNextID(),
				Name:        dstName,
				FileType:    vfs.FileTypeSymlink,
				CreatedAt:   time.Now().Unix(),
				ModifiedAt:  time.Now().Unix(),
				AccessedAt:  time.Now().Unix(),
				Mode:        e.metadata.Mode,
				Owner:       e.metadata.Owner,
				Group:       e.metadata.Group,
			},
			target:   e.target,
			parentID: dstParent.metadata.ID,
			vfs:      fs,
		}
		
		// Save the new symlink
		if err := fs.SaveEntry(newSymlink); err != nil {
			return nil, err
		}
		
		newEntry = newSymlink
		
	default:
		return nil, errors.New("unknown entry type")
	}
	
	// Update the destination parent directory
	dstParent.children = append(dstParent.children, newEntry.GetMetadata().ID)
	dstParent.metadata.SetModified()
	
	if err := fs.SaveEntry(dstParent); err != nil {
		return nil, fmt.Errorf("failed to save destination parent: %w", err)
	}
	
	return newEntry, nil
}

// directoryMove moves a file or directory to a new location
func (fs *DatabaseVFS) directoryMove(srcParent *DirectoryEntry, srcName, dstName string, dstParent *DirectoryEntry) (FSEntry, error) {
	var srcEntry FSEntry
	var srcEntryIndex int
	
	// Find the source entry
	for i, childID := range srcParent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == srcName {
			srcEntry = entry
			srcEntryIndex = i
			break
		}
	}
	
	if srcEntry == nil {
		return nil, vfs.ErrNotFound
	}
	
	// Check if destination name already exists
	for _, childID := range dstParent.children {
		entry, err := fs.LoadEntry(childID)
		if err != nil {
			return nil, fmt.Errorf("failed to load child entry: %w", err)
		}
		
		if entry.GetMetadata().Name == dstName {
			return nil, vfs.ErrAlreadyExists
		}
	}
	
	// Update the entry's name and parent
	srcEntry.GetMetadata().Name = dstName
	srcEntry.GetMetadata().SetModified()
	
	// Update the entry's parent ID based on its type
	switch e := srcEntry.(type) {
	case *DirectoryEntry:
		e.parentID = dstParent.metadata.ID
	case *FileEntry:
		e.parentID = dstParent.metadata.ID
	case *SymlinkEntry:
		e.parentID = dstParent.metadata.ID
	default:
		return nil, errors.New("unknown entry type")
	}
	
	// Save the updated entry
	if err := fs.SaveEntry(srcEntry); err != nil {
		return nil, fmt.Errorf("failed to save moved entry: %w", err)
	}
	
	// Remove the entry from the source parent's children list
	srcParent.children = append(srcParent.children[:srcEntryIndex], srcParent.children[srcEntryIndex+1:]...)
	srcParent.metadata.SetModified()
	
	// Save the updated source parent
	if err := fs.SaveEntry(srcParent); err != nil {
		return nil, fmt.Errorf("failed to save source parent: %w", err)
	}
	
	// Add the entry to the destination parent's children list
	dstParent.children = append(dstParent.children, srcEntry.GetMetadata().ID)
	dstParent.metadata.SetModified()
	
	// Save the updated destination parent
	if err := fs.SaveEntry(dstParent); err != nil {
		return nil, fmt.Errorf("failed to save destination parent: %w", err)
	}
	
	return srcEntry, nil
}
