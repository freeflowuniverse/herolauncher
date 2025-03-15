package vfs

// BaseEntry provides a common implementation for all filesystem entries
type BaseEntry struct {
	metadata *Metadata
}

// GetMetadata returns the metadata for this entry
func (e *BaseEntry) GetMetadata() *Metadata {
	return e.metadata
}

// IsDir returns true if the entry is a directory
func (e *BaseEntry) IsDir() bool {
	return e.metadata.IsDir()
}

// IsFile returns true if the entry is a file
func (e *BaseEntry) IsFile() bool {
	return e.metadata.IsFile()
}

// IsSymlink returns true if the entry is a symlink
func (e *BaseEntry) IsSymlink() bool {
	return e.metadata.IsSymlink()
}

// FileEntry represents a file in the filesystem
type FileEntry struct {
	BaseEntry
	parentID uint32
	vfs      VFSImplementation
}

// NewFileEntry creates a new file entry
func NewFileEntry(metadata *Metadata, parentID uint32, vfs VFSImplementation) *FileEntry {
	return &FileEntry{
		BaseEntry: BaseEntry{metadata: metadata},
		parentID:  parentID,
		vfs:       vfs,
	}
}

// Open opens the file and returns a ReadWriteSeeker
func (f *FileEntry) Open() (ReadWriteSeeker, error) {
	// This would typically be implemented by the specific VFS implementation
	return nil, ErrNotImplemented
}

// Read reads the entire file content
func (f *FileEntry) Read() ([]byte, error) {
	path, err := f.vfs.GetPath(f)
	if err != nil {
		return nil, err
	}
	return f.vfs.FileRead(path)
}

// Write writes data to the file
func (f *FileEntry) Write(data []byte) error {
	path, err := f.vfs.GetPath(f)
	if err != nil {
		return err
	}
	return f.vfs.FileWrite(path, data)
}

// Append appends data to the file
func (f *FileEntry) Append(data []byte) error {
	path, err := f.vfs.GetPath(f)
	if err != nil {
		return err
	}
	return f.vfs.FileConcatenate(path, data)
}

// DirectoryEntry represents a directory in the filesystem
type DirectoryEntry struct {
	BaseEntry
	parentID uint32
	children []uint32 // IDs of child entries
	vfs      VFSImplementation
}

// NewDirectoryEntry creates a new directory entry
func NewDirectoryEntry(metadata *Metadata, parentID uint32, vfs VFSImplementation) *DirectoryEntry {
	return &DirectoryEntry{
		BaseEntry: BaseEntry{metadata: metadata},
		parentID:  parentID,
		children:  []uint32{},
		vfs:       vfs,
	}
}

// List returns the entries in this directory
func (d *DirectoryEntry) List() ([]FSEntry, error) {
	path, err := d.vfs.GetPath(d)
	if err != nil {
		return nil, err
	}
	return d.vfs.DirList(path)
}

// Create creates a new file or directory in this directory
func (d *DirectoryEntry) Create(name string, isDir bool) (FSEntry, error) {
	path, err := d.vfs.GetPath(d)
	if err != nil {
		return nil, err
	}
	
	fullPath := JoinPath(path, name)
	
	if isDir {
		return d.vfs.DirCreate(fullPath)
	}
	return d.vfs.FileCreate(fullPath)
}

// Remove removes an entry from this directory
func (d *DirectoryEntry) Remove(name string) error {
	path, err := d.vfs.GetPath(d)
	if err != nil {
		return err
	}
	
	fullPath := JoinPath(path, name)
	return d.vfs.Delete(fullPath)
}

// SymlinkEntry represents a symbolic link in the filesystem
type SymlinkEntry struct {
	BaseEntry
	target   string
	parentID uint32
	vfs      VFSImplementation
}

// NewSymlinkEntry creates a new symlink entry
func NewSymlinkEntry(metadata *Metadata, target string, parentID uint32, vfs VFSImplementation) *SymlinkEntry {
	return &SymlinkEntry{
		BaseEntry: BaseEntry{metadata: metadata},
		target:    target,
		parentID:  parentID,
		vfs:       vfs,
	}
}

// Target returns the target path of this symlink
func (s *SymlinkEntry) Target() (string, error) {
	path, err := s.vfs.GetPath(s)
	if err != nil {
		return "", err
	}
	return s.vfs.LinkRead(path)
}

// UpdateTarget updates the target path of this symlink
func (s *SymlinkEntry) UpdateTarget(newTarget string) error {
	// This would typically be implemented by the specific VFS implementation
	s.target = newTarget
	s.metadata.SetModified()
	return nil
}
