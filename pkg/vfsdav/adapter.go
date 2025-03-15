package vfsdav

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

// Adapter implements the webdav.FileSystem interface using a vfs.VFSImplementation backend
type Adapter struct {
	vfs vfs.VFSImplementation
}

// New creates a new WebDAV adapter for the given VFS implementation
func New(vfs vfs.VFSImplementation) *Adapter {
	return &Adapter{
		vfs: vfs,
	}
}

// Open implements webdav.FileSystem.Open
func (a *Adapter) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	name = normalizePath(name)
	
	entry, err := a.vfs.Get(name)
	if err != nil {
		return nil, mapError(err)
	}

	if entry.IsDir() {
		return nil, webdav.NewHTTPError(403, errors.New("cannot open directory for reading"))
	}

	data, err := a.vfs.FileRead(name)
	if err != nil {
		return nil, mapError(err)
	}

	return io.NopCloser(strings.NewReader(string(data))), nil
}

// Stat implements webdav.FileSystem.Stat
func (a *Adapter) Stat(ctx context.Context, name string) (*webdav.FileInfo, error) {
	name = normalizePath(name)
	
	entry, err := a.vfs.Get(name)
	if err != nil {
		return nil, mapError(err)
	}

	return a.entryToFileInfo(entry, name)
}

// ReadDir implements webdav.FileSystem.ReadDir
func (a *Adapter) ReadDir(ctx context.Context, name string, recursive bool) ([]webdav.FileInfo, error) {
	name = normalizePath(name)
	
	entry, err := a.vfs.Get(name)
	if err != nil {
		return nil, mapError(err)
	}

	if !entry.IsDir() {
		return nil, webdav.NewHTTPError(403, errors.New("not a directory"))
	}

	entries, err := a.vfs.DirList(name)
	if err != nil {
		return nil, mapError(err)
	}

	fileInfos := make([]webdav.FileInfo, 0, len(entries))
	for _, e := range entries {
		entryPath, err := a.vfs.GetPath(e)
		if err != nil {
			return nil, mapError(err)
		}

		fi, err := a.entryToFileInfo(e, entryPath)
		if err != nil {
			return nil, err
		}
		fileInfos = append(fileInfos, *fi)

		if recursive && e.IsDir() {
			subEntries, err := a.ReadDir(ctx, entryPath, recursive)
			if err != nil {
				return nil, err
			}
			fileInfos = append(fileInfos, subEntries...)
		}
	}

	return fileInfos, nil
}

// Create implements webdav.FileSystem.Create
func (a *Adapter) Create(ctx context.Context, name string, body io.ReadCloser) (fileInfo *webdav.FileInfo, created bool, err error) {
	name = normalizePath(name)
	
	// Check if file exists to determine if we're creating or updating
	exists := a.vfs.Exists(name)

	// In v0.6.0 we don't have conditional operations

	// Read all data from the body
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, false, err
	}
	defer body.Close()

	var entry vfs.FSEntry
	if exists {
		// Update existing file
		err = a.vfs.FileWrite(name, data)
		if err != nil {
			return nil, false, mapError(err)
		}
		entry, err = a.vfs.Get(name)
	} else {
		// Create parent directories if they don't exist
		dir := path.Dir(name)
		if dir != "/" && dir != "." && !a.vfs.Exists(dir) {
			_, err = a.vfs.DirCreate(dir)
			if err != nil {
				return nil, false, mapError(err)
			}
		}

		// Create new file
		entry, err = a.vfs.FileCreate(name)
		if err != nil {
			return nil, false, mapError(err)
		}
		err = a.vfs.FileWrite(name, data)
	}

	if err != nil {
		return nil, false, mapError(err)
	}

	fi, err := a.entryToFileInfo(entry, name)
	if err != nil {
		return nil, false, err
	}

	return fi, !exists, nil
}

// RemoveAll implements webdav.FileSystem.RemoveAll
func (a *Adapter) RemoveAll(ctx context.Context, name string) error {
	name = normalizePath(name)
	
	if !a.vfs.Exists(name) {
		return webdav.NewHTTPError(404, errors.New("not found"))
	}

	entry, err := a.vfs.Get(name)
	if err != nil {
		return mapError(err)
	}

	// In v0.6.0 we don't have conditional operations

	if entry.IsDir() {
		err = a.vfs.DirDelete(name)
	} else {
		err = a.vfs.FileDelete(name)
	}

	return mapError(err)
}

// Mkdir implements webdav.FileSystem.Mkdir
func (a *Adapter) Mkdir(ctx context.Context, name string) error {
	name = normalizePath(name)
	
	if a.vfs.Exists(name) {
		entry, err := a.vfs.Get(name)
		if err != nil {
			return mapError(err)
		}
		if entry.IsDir() {
			return webdav.NewHTTPError(405, errors.New("directory already exists"))
		}
		return webdav.NewHTTPError(405, errors.New("file exists at this path"))
	}

	// Create parent directories if they don't exist
	dir := path.Dir(name)
	if dir != "/" && dir != "." && !a.vfs.Exists(dir) {
		_, err := a.vfs.DirCreate(dir)
		if err != nil {
			return mapError(err)
		}
	}

	_, err := a.vfs.DirCreate(name)
	return mapError(err)
}

// Copy implements webdav.FileSystem.Copy
func (a *Adapter) Copy(ctx context.Context, name, dest string, options *webdav.CopyOptions) (created bool, err error) {
	name = normalizePath(name)
	dest = normalizePath(dest)
	
	if !a.vfs.Exists(name) {
		return false, webdav.NewHTTPError(404, errors.New("source not found"))
	}

	destExists := a.vfs.Exists(dest)
	if destExists && options != nil && options.NoOverwrite {
		return false, os.ErrExist
	}

	// Create parent directories if they don't exist
	destDir := path.Dir(dest)
	if destDir != "/" && destDir != "." && !a.vfs.Exists(destDir) {
		_, err := a.vfs.DirCreate(destDir)
		if err != nil {
			return false, mapError(err)
		}
	}

	_, err = a.vfs.Copy(name, dest)
	if err != nil {
		return false, mapError(err)
	}

	return !destExists, nil
}

// Move implements webdav.FileSystem.Move
func (a *Adapter) Move(ctx context.Context, name, dest string, options *webdav.MoveOptions) (created bool, err error) {
	name = normalizePath(name)
	dest = normalizePath(dest)
	
	if !a.vfs.Exists(name) {
		return false, webdav.NewHTTPError(404, errors.New("source not found"))
	}

	destExists := a.vfs.Exists(dest)
	if destExists && options != nil && options.NoOverwrite {
		return false, os.ErrExist
	}

	// Create parent directories if they don't exist
	destDir := path.Dir(dest)
	if destDir != "/" && destDir != "." && !a.vfs.Exists(destDir) {
		_, err := a.vfs.DirCreate(destDir)
		if err != nil {
			return false, mapError(err)
		}
	}

	_, err = a.vfs.Move(name, dest)
	if err != nil {
		return false, mapError(err)
	}

	return !destExists, nil
}

// Helper functions

// entryToFileInfo converts a vfs.FSEntry to a webdav.FileInfo
func (a *Adapter) entryToFileInfo(entry vfs.FSEntry, entryPath string) (*webdav.FileInfo, error) {
	metadata := entry.GetMetadata()
	
	// Ensure path starts with a slash for WebDAV
	if !strings.HasPrefix(entryPath, "/") {
		entryPath = "/" + entryPath
	}

	return &webdav.FileInfo{
		Path:     entryPath,
		Size:     int64(metadata.Size),
		ModTime:  time.Unix(metadata.ModifiedAt, 0),
		IsDir:    entry.IsDir(),
		MIMEType: getMIMEType(entryPath, entry),
		ETag:     generateETag(metadata),
	}, nil
}

// normalizePath ensures the path is properly formatted for VFS
func normalizePath(name string) string {
	// Ensure path starts with a slash
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	
	// Remove trailing slash except for root
	if name != "/" && strings.HasSuffix(name, "/") {
		name = name[:len(name)-1]
	}
	
	return name
}

// mapError maps VFS errors to WebDAV errors
func mapError(err error) error {
	if err == nil {
		return nil
	}

	// TODO: Add more specific error mappings based on the VFS error types
	if os.IsNotExist(err) {
		return webdav.NewHTTPError(404, err)
	}
	if os.IsPermission(err) {
		return webdav.NewHTTPError(403, err)
	}
	if os.IsExist(err) {
		return webdav.NewHTTPError(409, err)
	}

	return webdav.NewHTTPError(500, err)
}

// getMIMEType returns the MIME type for a file
func getMIMEType(filePath string, entry vfs.FSEntry) string {
	if entry.IsDir() {
		return "directory"
	}
	
	// Simple MIME type detection based on extension
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// generateETag generates an ETag for a file based on its metadata
func generateETag(metadata *vfs.Metadata) string {
	// Simple ETag generation based on size and modification time
	return fmt.Sprintf("\"%d-%d-%d\"", metadata.ID, metadata.Size, metadata.ModifiedAt)
}
