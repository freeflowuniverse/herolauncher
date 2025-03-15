// Package vfslocal provides a local filesystem implementation of the VFS interface
package vfslocal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// LocalVFS implements the VFSImplementation interface using the local filesystem
type LocalVFS struct {
	rootPath string
	mu       sync.RWMutex // Mutex for thread safety
}

// New creates a new LocalVFS instance with the given root path
func New(rootPath string) (*LocalVFS, error) {
	// Ensure the root path exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("root path does not exist: %s", rootPath)
	}

	return &LocalVFS{
		rootPath: rootPath,
	}, nil
}

// getAbsPath converts a VFS path to an absolute filesystem path
func (l *LocalVFS) getAbsPath(path string) string {
	// Normalize path
	if path == "" {
		path = "/"
	}
	
	// Remove leading slash for joining with root path
	cleanPath := strings.TrimPrefix(path, "/")
	
	// Join with root path
	return filepath.Join(l.rootPath, cleanPath)
}

// getRelPath converts an absolute filesystem path to a VFS path
func (l *LocalVFS) getRelPath(absPath string) string {
	// Get relative path from root
	relPath, err := filepath.Rel(l.rootPath, absPath)
	if err != nil {
		return "/" // Default to root if there's an error
	}
	
	// Convert to forward slashes and ensure leading slash
	relPath = filepath.ToSlash(relPath)
	if relPath == "." {
		return "/"
	}
	
	return "/" + relPath
}

// getMetadataFromFileInfo creates a Metadata struct from os.FileInfo
func (l *LocalVFS) getMetadataFromFileInfo(info os.FileInfo, path string) *vfs.Metadata {
	fileType := vfs.FileTypeFile
	if info.IsDir() {
		fileType = vfs.FileTypeDirectory
	} else if info.Mode()&os.ModeSymlink != 0 {
		fileType = vfs.FileTypeSymlink
	}
	
	return &vfs.Metadata{
		ID:         0, // Local filesystem doesn't use IDs
		Name:       info.Name(),
		FileType:   fileType,
		Size:       uint64(info.Size()),
		CreatedAt:  info.ModTime().Unix(), // Use ModTime for CreatedAt as os.FileInfo doesn't provide creation time
		ModifiedAt: info.ModTime().Unix(),
		AccessedAt: time.Now().Unix(), // Use current time as AccessedAt
	}
}

// createFSEntry creates an FSEntry from a file path
func (l *LocalVFS) createFSEntry(absPath string) (vfs.FSEntry, error) {
	info, err := os.Lstat(absPath)
	if err != nil {
		return nil, err
	}
	
	relPath := l.getRelPath(absPath)
	metadata := l.getMetadataFromFileInfo(info, relPath)
	
	if info.IsDir() {
		return &DirectoryEntry{
			metadata: metadata,
			path:     relPath,
		}, nil
	} else if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(absPath)
		if err != nil {
			return nil, err
		}
		
		return &SymlinkEntry{
			metadata: metadata,
			path:     relPath,
			target:   target,
		}, nil
	} else {
		return &FileEntry{
			metadata: metadata,
			path:     relPath,
		}, nil
	}
}

// Implementation of VFSImplementation interface

// RootGet returns the root filesystem entry
func (l *LocalVFS) RootGet() (vfs.FSEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	return l.createFSEntry(l.rootPath)
}

// Delete deletes a filesystem entry
func (l *LocalVFS) Delete(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	return os.RemoveAll(absPath)
}

// LinkDelete deletes a symlink
func (l *LocalVFS) LinkDelete(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Check if it's a symlink
	info, err := os.Lstat(absPath)
	if err != nil {
		return err
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("path is not a symlink: %s", path)
	}
	
	return os.Remove(absPath)
}

// FileCreate creates a new file
func (l *LocalVFS) FileCreate(path string) (vfs.FSEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Create parent directories if they don't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	// Create the file
	file, err := os.Create(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	return l.createFSEntry(absPath)
}

// FileRead reads the content of a file
func (l *LocalVFS) FileRead(path string) ([]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	absPath := l.getAbsPath(path)
	return os.ReadFile(absPath)
}

// FileWrite writes data to a file
func (l *LocalVFS) FileWrite(path string, data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Create parent directories if they don't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(absPath, data, 0644)
}

// FileConcatenate appends data to a file
func (l *LocalVFS) FileConcatenate(path string, data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Open the file for appending
	file, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.Write(data)
	return err
}

// FileDelete deletes a file
func (l *LocalVFS) FileDelete(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Check if it's a file
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}
	
	return os.Remove(absPath)
}

// DirCreate creates a new directory
func (l *LocalVFS) DirCreate(path string) (vfs.FSEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Create the directory and any parent directories
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, err
	}
	
	return l.createFSEntry(absPath)
}

// DirList lists the entries in a directory
func (l *LocalVFS) DirList(path string) ([]vfs.FSEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	absPath := l.getAbsPath(path)
	
	// Check if it's a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}
	
	// Read directory entries
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}
	
	var result []vfs.FSEntry
	for _, entry := range entries {
		entryPath := filepath.Join(absPath, entry.Name())
		fsEntry, err := l.createFSEntry(entryPath)
		if err != nil {
			continue // Skip entries that can't be created
		}
		result = append(result, fsEntry)
	}
	
	return result, nil
}

// DirDelete deletes a directory
func (l *LocalVFS) DirDelete(path string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absPath := l.getAbsPath(path)
	
	// Check if it's a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	
	return os.RemoveAll(absPath)
}

// Exists checks if a path exists
func (l *LocalVFS) Exists(path string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	absPath := l.getAbsPath(path)
	_, err := os.Stat(absPath)
	return err == nil
}

// Get returns the filesystem entry at the specified path
func (l *LocalVFS) Get(path string) (vfs.FSEntry, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	absPath := l.getAbsPath(path)
	return l.createFSEntry(absPath)
}

// Rename renames a filesystem entry
func (l *LocalVFS) Rename(oldPath, newPath string) (vfs.FSEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absOldPath := l.getAbsPath(oldPath)
	absNewPath := l.getAbsPath(newPath)
	
	// Create parent directories if they don't exist
	dir := filepath.Dir(absNewPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	if err := os.Rename(absOldPath, absNewPath); err != nil {
		return nil, err
	}
	
	return l.createFSEntry(absNewPath)
}

// Copy copies a filesystem entry
func (l *LocalVFS) Copy(srcPath, dstPath string) (vfs.FSEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absSrcPath := l.getAbsPath(srcPath)
	absDstPath := l.getAbsPath(dstPath)
	
	// Check if source exists
	srcInfo, err := os.Stat(absSrcPath)
	if err != nil {
		return nil, err
	}
	
	// Create parent directories if they don't exist
	dir := filepath.Dir(absDstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	if srcInfo.IsDir() {
		// Copy directory recursively
		if err := copyDir(absSrcPath, absDstPath); err != nil {
			return nil, err
		}
	} else {
		// Copy file
		if err := copyFile(absSrcPath, absDstPath); err != nil {
			return nil, err
		}
	}
	
	return l.createFSEntry(absDstPath)
}

// Move moves a filesystem entry
func (l *LocalVFS) Move(srcPath, dstPath string) (vfs.FSEntry, error) {
	// Move is equivalent to rename in the local filesystem
	return l.Rename(srcPath, dstPath)
}

// LinkCreate creates a new symlink
func (l *LocalVFS) LinkCreate(targetPath, linkPath string) (vfs.FSEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	absLinkPath := l.getAbsPath(linkPath)
	
	// Create parent directories if they don't exist
	dir := filepath.Dir(absLinkPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	if err := os.Symlink(targetPath, absLinkPath); err != nil {
		return nil, err
	}
	
	return l.createFSEntry(absLinkPath)
}

// LinkRead reads the target of a symlink
func (l *LocalVFS) LinkRead(path string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	absPath := l.getAbsPath(path)
	
	// Check if it's a symlink
	info, err := os.Lstat(absPath)
	if err != nil {
		return "", err
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		return "", fmt.Errorf("path is not a symlink: %s", path)
	}
	
	return os.Readlink(absPath)
}

// Destroy cleans up resources
func (l *LocalVFS) Destroy() error {
	// Nothing to clean up for local filesystem
	return nil
}

// GetPath returns the path for the given entry
func (l *LocalVFS) GetPath(entry vfs.FSEntry) (string, error) {
	switch e := entry.(type) {
	case *FileEntry:
		return e.path, nil
	case *DirectoryEntry:
		return e.path, nil
	case *SymlinkEntry:
		return e.path, nil
	default:
		return "", errors.New("unknown entry type")
	}
}

// Helper functions

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Copy file mode
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}

// copyDir copies a directory recursively from src to dst
func copyDir(src, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	// Create the destination directory
	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			// Recursive copy for directories
			if err = copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Handle symlinks
			if entry.Type()&fs.ModeSymlink != 0 {
				// Read the symlink target
				target, err := os.Readlink(srcPath)
				if err != nil {
					return err
				}
				
				// Create a new symlink
				if err = os.Symlink(target, dstPath); err != nil {
					return err
				}
			} else {
				// Copy files
				if err = copyFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}
