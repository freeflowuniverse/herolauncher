package main

import (
	"fmt"
	"log"
	"os/user"
	"path"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

// getCurrentUsername gets the current user's username for permission checks
func getCurrentUsername() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("WARNING: Could not get current user, using 'nobody': %v", err)
		return "nobody"
	}
	return currentUser.Username
}

// createVFSDBFile returns a function that creates a file in the vfsdb backend
func createVFSDBFile(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.File, error) {
	return func(fsys *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.File, error) {
		// Get the parent directory path
		parentDir, ok := parent.(*VFSDBDir)
		if !ok {
			return nil, fmt.Errorf("parent is not a VFSDBDir")
		}

		// Construct the full path for the new file
		filePath := path.Join(parentDir.path, name)
		log.Printf("Creating file %s with permissions %o", filePath, perm)

		// Create the file in the vfsdb backend
		entry, err := vfsImpl.FileCreate(filePath)
		if err != nil {
			log.Printf("Error creating file %s: %v", filePath, err)
			return nil, err
		}

		// Get the file metadata
		metadata := entry.GetMetadata()

		// Set the file permissions and ownership
		metadata.Mode = perm
		metadata.Owner = user
		metadata.Group = user

		// Create a stat for the new file
		stat := &proto.Stat{
			Type:  0,
			Dev:   0,
			Qid:   proto.Qid{Type: 0, Version: 0, Path: uint64(metadata.ID)},
			Mode:  perm,
			Atime: uint32(metadata.AccessedAt),
			Mtime: uint32(metadata.ModifiedAt),
			Name:  name,
			Uid:   user,
			Gid:   user,
			Muid:  user,
		}

		// Create and return a new VFSDBFile
		return NewVFSDBFile(stat, vfsImpl, filePath), nil
	}
}

// createVFSDBDir returns a function that creates a directory in the vfsdb backend
func createVFSDBDir(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.Dir, error) {
	return func(fsys *fs.FS, parent fs.Dir, user, name string, perm uint32, mode uint8) (fs.Dir, error) {
		// Get the parent directory path
		parentDir, ok := parent.(*VFSDBDir)
		if !ok {
			return nil, fmt.Errorf("parent is not a VFSDBDir")
		}

		// Construct the full path for the new directory
		dirPath := path.Join(parentDir.path, name)
		log.Printf("Creating directory %s with permissions %o", dirPath, perm)

		// Create the directory in the vfsdb backend
		entry, err := vfsImpl.DirCreate(dirPath)
		if err != nil {
			log.Printf("Error creating directory %s: %v", dirPath, err)
			return nil, err
		}

		// Get the directory metadata
		metadata := entry.GetMetadata()

		// Set the directory permissions and ownership
		metadata.Mode = perm | 0x80000000 // Set directory bit
		metadata.Owner = user
		metadata.Group = user

		// Create a stat for the new directory
		stat := &proto.Stat{
			Type:  0,
			Dev:   0,
			Qid:   proto.Qid{Type: proto.QTDIR, Version: 0, Path: uint64(metadata.ID)},
			Mode:  perm | proto.DMDIR,
			Atime: uint32(metadata.AccessedAt),
			Mtime: uint32(metadata.ModifiedAt),
			Name:  name,
			Uid:   user,
			Gid:   user,
			Muid:  user,
		}

		// Create and return a new VFSDBDir
		return NewVFSDBDir(stat, vfsImpl, dirPath), nil
	}
}

// removeVFSDBFile returns a function that removes a file or directory from the vfsdb backend
func removeVFSDBFile(vfsImpl vfs.VFSImplementation) func(fs *fs.FS, f fs.FSNode) error {
	return func(fsys *fs.FS, node fs.FSNode) error {
		var nodePath string
		
		// Get the path based on the node type
		switch n := node.(type) {
		case *VFSDBFile:
			nodePath = n.path
		case *VFSDBDir:
			nodePath = n.path
		default:
			return fmt.Errorf("node is not a VFSDBFile or VFSDBDir")
		}
		
		log.Printf("DEBUG: Removing node at path '%s'", nodePath)
		
		// Get the entry to check its metadata and permissions
		entry, err := vfsImpl.Get(nodePath)
		if err != nil {
			log.Printf("DEBUG: Error getting entry for '%s': %v", nodePath, err)
			return err
		}
		
		// Get the current username for permission checks
		currentUsername := getCurrentUsername()
		log.Printf("DEBUG: Current username: %s", currentUsername)
		
		// Get the metadata for permission checks
		metadata := entry.GetMetadata()
		log.Printf("DEBUG: Entry metadata - Mode: %o (0x%x), Owner: %s, Group: %s", 
			metadata.Mode, metadata.Mode, metadata.Owner, metadata.Group)
		
		// Parse the permission bits
		owner_r := (metadata.Mode & 0400) != 0
		owner_w := (metadata.Mode & 0200) != 0
		owner_x := (metadata.Mode & 0100) != 0
		group_r := (metadata.Mode & 0040) != 0
		group_w := (metadata.Mode & 0020) != 0
		group_x := (metadata.Mode & 0010) != 0
		other_r := (metadata.Mode & 0004) != 0
		other_w := (metadata.Mode & 0002) != 0
		other_x := (metadata.Mode & 0001) != 0
		
		// Check if the current user is the owner or in the same group
		isOwner := currentUsername == metadata.Owner
		isInGroup := currentUsername == metadata.Group // Simple check, in real systems would use groups
		log.Printf("DEBUG: Permission relationship - Is owner: %v, Is in group: %v", isOwner, isInGroup)
		
		// Determine which permission bits are relevant based on the relationship
		var relevantPermission string
		var hasWritePermission bool
		if isOwner {
			relevantPermission = fmt.Sprintf("owner (r=%v,w=%v,x=%v)", owner_r, owner_w, owner_x)
			hasWritePermission = owner_w
		} else if isInGroup {
			relevantPermission = fmt.Sprintf("group (r=%v,w=%v,x=%v)", group_r, group_w, group_x)
			hasWritePermission = group_w
		} else {
			relevantPermission = fmt.Sprintf("other (r=%v,w=%v,x=%v)", other_r, other_w, other_x)
			hasWritePermission = other_w
		}
		log.Printf("DEBUG: Relevant permission bits for %s: %s", nodePath, relevantPermission)
		log.Printf("DEBUG: Has write permission: %v (Required for deletion)", hasWritePermission)
		
		// Log the parent directory permissions as well, since we need write permission on the parent
		parentPath := path.Dir(nodePath)
		parentEntry, err := vfsImpl.Get(parentPath)
		if err == nil {
			parentMetadata := parentEntry.GetMetadata()
			log.Printf("DEBUG: Parent directory %s metadata - Mode: %o (0x%x), Owner: %s, Group: %s", 
				parentPath, parentMetadata.Mode, parentMetadata.Mode, parentMetadata.Owner, parentMetadata.Group)
			
			// Check parent directory write permissions
			parent_owner_w := (parentMetadata.Mode & 0200) != 0
			parent_group_w := (parentMetadata.Mode & 0020) != 0
			parent_other_w := (parentMetadata.Mode & 0002) != 0
			
			parent_isOwner := currentUsername == parentMetadata.Owner
			parent_isInGroup := currentUsername == parentMetadata.Group
			
			var parentHasWritePermission bool
			if parent_isOwner {
				parentHasWritePermission = parent_owner_w
			} else if parent_isInGroup {
				parentHasWritePermission = parent_group_w
			} else {
				parentHasWritePermission = parent_other_w
			}
			
			log.Printf("DEBUG: Parent directory write permission: %v (Required for deletion)", parentHasWritePermission)
		}
		
		if entry.IsDir() {
			log.Printf("DEBUG: Attempting to delete directory %s", nodePath)
			err = vfsImpl.DirDelete(nodePath)
			if err != nil {
				log.Printf("DEBUG: Failed to delete directory %s: %v", nodePath, err)
				// Log the error details to help diagnose permission issues
				if strings.Contains(err.Error(), "permission") {
					log.Printf("DEBUG: Permission error detected when deleting directory %s", nodePath)
				}
			}
		} else {
			log.Printf("DEBUG: Attempting to delete file %s", nodePath)
			err = vfsImpl.FileDelete(nodePath)
			if err != nil {
				log.Printf("DEBUG: Failed to delete file %s: %v", nodePath, err)
				// Log the error details to help diagnose permission issues
				if strings.Contains(err.Error(), "permission") {
					log.Printf("DEBUG: Permission error detected when deleting file %s", nodePath)
				}
			}
		}
		
		return err
	}
}
