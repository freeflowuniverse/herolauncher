package vfs

import (
	"errors"
	"path/filepath"
	"strings"
)

// Common errors
var (
	ErrNotImplemented = errors.New("operation not implemented")
	ErrNotFound       = errors.New("entry not found")
	ErrAlreadyExists  = errors.New("entry already exists")
	ErrNotEmpty       = errors.New("directory not empty")
	ErrNotDirectory   = errors.New("not a directory")
	ErrNotFile        = errors.New("not a file")
	ErrNotSymlink     = errors.New("not a symlink")
	ErrInvalidPath    = errors.New("invalid path")
	ErrPermission     = errors.New("permission denied")
)

// JoinPath joins path elements with proper handling of leading/trailing slashes
func JoinPath(elem ...string) string {
	// Clean up each path element
	for i, e := range elem {
		elem[i] = strings.Trim(e, "/")
	}
	
	// Join with slashes
	result := "/" + strings.Join(elem, "/")
	
	// Remove double slashes
	for strings.Contains(result, "//") {
		result = strings.ReplaceAll(result, "//", "/")
	}
	
	return result
}

// SplitPath splits a path into its components
func SplitPath(path string) []string {
	// Trim leading and trailing slashes
	path = strings.Trim(path, "/")
	
	// Handle empty path
	if path == "" {
		return []string{}
	}
	
	// Split by slashes
	return strings.Split(path, "/")
}

// PathDir returns the directory part of a path
func PathDir(path string) string {
	// Special case for root
	if path == "/" {
		return "/"
	}
	
	// Clean the path
	path = filepath.Clean(path)
	
	// Get the directory
	dir := filepath.Dir(path)
	
	// Ensure leading slash
	if !strings.HasPrefix(dir, "/") {
		dir = "/" + dir
	}
	
	return dir
}

// PathBase returns the base name of a path
func PathBase(path string) string {
	// Special case for root
	if path == "/" {
		return ""
	}
	
	// Clean the path
	path = filepath.Clean(path)
	
	// Get the base name
	return filepath.Base(path)
}

// FixPath ensures a path has a leading slash and no trailing slash
func FixPath(path string) string {
	// Ensure leading slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Remove trailing slash unless it's the root
	if path != "/" && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	
	return path
}
