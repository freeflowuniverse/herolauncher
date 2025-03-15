package vfs9p

import (
	"context"
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/proto"
)

// VFS9P adapts a herolauncher VFS to the go9p.Srv interface
type VFS9P struct {
	vfsImpl vfs.VFSImplementation
	fidMap  sync.Map // Maps fids to open file info
	nextQid uint64
	mu      sync.Mutex
}

// fidInfo tracks information about an open file descriptor
type fidInfo struct {
	path     string
	entry    vfs.FSEntry
	openMode proto.Mode
	offset   uint64
	user     string
}

// Conn9P implements the go9p.Conn interface
type Conn9P struct {
	tags map[uint16]context.CancelFunc
	mu   sync.Mutex
}

// TagContext implements go9p.Conn.TagContext
func (c *Conn9P) TagContext(tag uint16) context.Context {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	ctx, cancel := context.WithCancel(context.Background())
	c.tags[tag] = cancel
	return ctx
}

// DropContext implements go9p.Conn.DropContext
func (c *Conn9P) DropContext(tag uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if cancel, ok := c.tags[tag]; ok {
		cancel()
		delete(c.tags, tag)
	}
}

// NewVFS9P creates a new VFS9P adapter
func NewVFS9P(vfsImpl vfs.VFSImplementation) *VFS9P {
	return &VFS9P{
		vfsImpl: vfsImpl,
		nextQid: 1,
	}
}

// NewConn implements go9p.Srv.NewConn
func (v *VFS9P) NewConn() go9p.Conn {
	return &Conn9P{
		tags: make(map[uint16]context.CancelFunc),
	}
}

// Version implements go9p.Srv.Version
func (v *VFS9P) Version(conn go9p.Conn, t *proto.TRVersion) (proto.FCall, error) {
	if t.Msize < 4096 {
		t.Msize = 4096
	}
	reply := *t
	reply.Type = proto.Rversion
	reply.Version = "9P2000"
	return &reply, nil
}

// Auth implements go9p.Srv.Auth
func (v *VFS9P) Auth(conn go9p.Conn, t *proto.TAuth) (proto.FCall, error) {
	// For simplicity, we'll not implement authentication
	return &proto.RError{
		Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
		Ename:  "Authentication not required",
	}, nil
}

// Attach implements go9p.Srv.Attach
func (v *VFS9P) Attach(conn go9p.Conn, t *proto.TAttach) (proto.FCall, error) {
	// Get the root directory
	root, err := v.vfsImpl.RootGet()
	if err != nil {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  err.Error(),
		}, nil
	}
	
	// Create a new fid for the root directory
	v.fidMap.Store(t.Fid, &fidInfo{
		path:  "/",
		entry: root,
		user:  t.Uname,
	})
	
	// Create a Qid for the root directory
	qid := v.entryToQid(root)
	
	return &proto.RAttach{
		Header: proto.Header{Type: proto.Rattach, Tag: t.Tag},
		Qid:    qid,
	}, nil
}

// Walk implements go9p.Srv.Walk
func (v *VFS9P) Walk(conn go9p.Conn, t *proto.TWalk) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// If no path elements, just clone the fid
	if t.Nwname == 0 {
		v.fidMap.Store(t.Newfid, &fidInfo{
			path:  fid.path,
			entry: fid.entry,
			user:  fid.user,
		})
		return &proto.RWalk{
			Header: proto.Header{Type: proto.Rwalk, Tag: t.Tag},
			Nwqid:  0,
			Wqid:   []proto.Qid{},
		}, nil
	}
	
	// Walk the path elements
	qids := make([]proto.Qid, 0, t.Nwname)
	currentPath := fid.path
	currentEntry := fid.entry
	
	for i := 0; i < int(t.Nwname); i++ {
		name := t.Wname[i]
		
		// Handle special cases
		if name == "." {
			qids = append(qids, v.entryToQid(currentEntry))
			continue
		}
		
		if name == ".." {
			if currentPath == "/" {
				qids = append(qids, v.entryToQid(currentEntry))
				continue
			}
			
			parentPath := vfs.PathDir(currentPath)
			parent, err := v.vfsImpl.Get(parentPath)
			if err != nil {
				// If we've walked some elements successfully, return those
				if len(qids) > 0 {
					return &proto.RWalk{
						Header: proto.Header{Type: proto.Rwalk, Tag: t.Tag},
						Nwqid:  uint16(len(qids)),
						Wqid:   qids,
					}, nil
				}
				
				return &proto.RError{
					Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
					Ename:  err.Error(),
				}, nil
			}
			
			currentPath = parentPath
			currentEntry = parent
			qids = append(qids, v.entryToQid(parent))
			continue
		}
		
		// Regular path element
		nextPath := vfs.JoinPath(currentPath, name)
		next, err := v.vfsImpl.Get(nextPath)
		if err != nil {
			// If we've walked some elements successfully, return those
			if len(qids) > 0 {
				break
			}
			
			return &proto.RError{
				Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
				Ename:  err.Error(),
			}, nil
		}
		
		currentPath = nextPath
		currentEntry = next
		qids = append(qids, v.entryToQid(next))
	}
	
	// Store the new fid
	v.fidMap.Store(t.Newfid, &fidInfo{
		path:  currentPath,
		entry: currentEntry,
		user:  fid.user,
	})
	
	return &proto.RWalk{
		Header: proto.Header{Type: proto.Rwalk, Tag: t.Tag},
		Nwqid:  uint16(len(qids)),
		Wqid:   qids,
	}, nil
}

// Open implements go9p.Srv.Open
func (v *VFS9P) Open(conn go9p.Conn, t *proto.TOpen) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Check if already open
	if fid.openMode != proto.None {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "File already open",
		}, nil
	}
	
	// Check permissions (simplified)
	// In a real implementation, you'd check user permissions here
	
	// Update the fid info
	fid.openMode = t.Mode
	fid.offset = 0
	
	return &proto.ROpen{
		Header: proto.Header{Type: proto.Ropen, Tag: t.Tag},
		Qid:    v.entryToQid(fid.entry),
		Iounit: 8192,
	}, nil
}

// Read implements go9p.Srv.Read
func (v *VFS9P) Read(conn go9p.Conn, t *proto.TRead) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Check if file is open for reading
	if fid.openMode&3 != proto.Oread && fid.openMode&3 != proto.Ordwr {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "File not open for reading",
		}, nil
	}
	
	// Handle directory reads
	if fid.entry.IsDir() {
		entries, err := v.vfsImpl.DirList(fid.path)
		if err != nil {
			return &proto.RError{
				Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
				Ename:  err.Error(),
			}, nil
		}
		
		// Convert entries to 9p stats
		stats := make([]proto.Stat, 0, len(entries))
		for _, entry := range entries {
			stats = append(stats, v.entryToStat(entry))
		}
		
		// Serialize stats
		data := make([]byte, 0)
		for _, stat := range stats {
			data = append(data, stat.Compose()...)
		}
		
		// Handle offset and count
		if t.Offset >= uint64(len(data)) {
			return &proto.RRead{
				Header: proto.Header{Type: proto.Rread, Tag: t.Tag},
				Count:  0,
				Data:   []byte{},
			}, nil
		}
		
		end := t.Offset + uint64(t.Count)
		if end > uint64(len(data)) {
			end = uint64(len(data))
		}
		
		return &proto.RRead{
			Header: proto.Header{Type: proto.Rread, Tag: t.Tag},
			Count:  uint32(end - t.Offset),
			Data:   data[t.Offset:end],
		}, nil
	}
	
	// Handle file reads
	if fid.entry.IsFile() {
		data, err := v.vfsImpl.FileRead(fid.path)
		if err != nil {
			return &proto.RError{
				Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
				Ename:  err.Error(),
			}, nil
		}
		
		// Handle offset and count
		if t.Offset >= uint64(len(data)) {
			return &proto.RRead{
				Header: proto.Header{Type: proto.Rread, Tag: t.Tag},
				Count:  0,
				Data:   []byte{},
			}, nil
		}
		
		end := t.Offset + uint64(t.Count)
		if end > uint64(len(data)) {
			end = uint64(len(data))
		}
		
		return &proto.RRead{
			Header: proto.Header{Type: proto.Rread, Tag: t.Tag},
			Count:  uint32(end - t.Offset),
			Data:   data[t.Offset:end],
		}, nil
	}
	
	return &proto.RError{
		Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
		Ename:  "Not a file or directory",
	}, nil
}

// Write implements go9p.Srv.Write
func (v *VFS9P) Write(conn go9p.Conn, t *proto.TWrite) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Check if file is open for writing
	if fid.openMode&3 != proto.Owrite && fid.openMode&3 != proto.Ordwr {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "File not open for writing",
		}, nil
	}
	
	// Handle file writes
	if fid.entry.IsFile() {
		// Read existing data
		data, err := v.vfsImpl.FileRead(fid.path)
		if err != nil {
			return &proto.RError{
				Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
				Ename:  err.Error(),
			}, nil
		}
		
		// Extend data if needed
		if t.Offset+uint64(len(t.Data)) > uint64(len(data)) {
			newData := make([]byte, t.Offset+uint64(len(t.Data)))
			copy(newData, data)
			data = newData
		}
		
		// Write data at offset
		copy(data[t.Offset:], t.Data)
		
		// Save the file
		err = v.vfsImpl.FileWrite(fid.path, data)
		if err != nil {
			return &proto.RError{
				Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
				Ename:  err.Error(),
			}, nil
		}
		
		return &proto.RWrite{
			Header: proto.Header{Type: proto.Rwrite, Tag: t.Tag},
			Count:  uint32(len(t.Data)),
		}, nil
	}
	
	return &proto.RError{
		Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
		Ename:  "Not a file",
	}, nil
}

// Create implements go9p.Srv.Create
func (v *VFS9P) Create(conn go9p.Conn, t *proto.TCreate) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Check if parent is a directory
	if !fid.entry.IsDir() {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Not a directory",
		}, nil
	}
	
	// Create the new path
	newPath := vfs.JoinPath(fid.path, t.Name)
	
	var newEntry vfs.FSEntry
	var err error
	
	// Create directory or file
	if t.Perm&proto.DMDIR != 0 {
		newEntry, err = v.vfsImpl.DirCreate(newPath)
	} else {
		newEntry, err = v.vfsImpl.FileCreate(newPath)
	}
	
	if err != nil {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  err.Error(),
		}, nil
	}
	
	// Update the fid to point to the new file
	fid.path = newPath
	fid.entry = newEntry
	fid.openMode = proto.Mode(t.Mode)
	fid.offset = 0
	
	return &proto.RCreate{
		Header: proto.Header{Type: proto.Rcreate, Tag: t.Tag},
		Qid:    v.entryToQid(newEntry),
		Iounit: 8192,
	}, nil
}

// Clunk implements go9p.Srv.Clunk
func (v *VFS9P) Clunk(conn go9p.Conn, t *proto.TClunk) (proto.FCall, error) {
	v.fidMap.Delete(t.Fid)
	return &proto.RClunk{
		Header: proto.Header{Type: proto.Rclunk, Tag: t.Tag},
	}, nil
}

// Remove implements go9p.Srv.Remove
func (v *VFS9P) Remove(conn go9p.Conn, t *proto.TRemove) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Delete the entry
	err := v.vfsImpl.Delete(fid.path)
	
	// Clean up the fid
	v.fidMap.Delete(t.Fid)
	
	if err != nil {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  err.Error(),
		}, nil
	}
	
	return &proto.RRemove{
		Header: proto.Header{Type: proto.Rremove, Tag: t.Tag},
	}, nil
}

// Stat implements go9p.Srv.Stat
func (v *VFS9P) Stat(conn go9p.Conn, t *proto.TStat) (proto.FCall, error) {
	// Get the fid info
	fidValue, ok := v.fidMap.Load(t.Fid)
	if !ok {
		return &proto.RError{
			Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
			Ename:  "Invalid fid",
		}, nil
	}
	fid := fidValue.(*fidInfo)
	
	// Convert entry to stat
	stat := v.entryToStat(fid.entry)
	
	return &proto.RStat{
		Header: proto.Header{Type: proto.Rstat, Tag: t.Tag},
		Stat:   stat,
	}, nil
}

// Wstat implements go9p.Srv.Wstat
func (v *VFS9P) Wstat(conn go9p.Conn, t *proto.TWstat) (proto.FCall, error) {
	// This would be a complex implementation to handle metadata changes
	// For simplicity, we'll just return an error
	return &proto.RError{
		Header: proto.Header{Type: proto.Rerror, Tag: t.Tag},
		Ename:  "Wstat not implemented",
	}, nil
}

// entryToQid converts a VFS entry to a 9p Qid
func (v *VFS9P) entryToQid(entry vfs.FSEntry) proto.Qid {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	var qtype uint8
	if entry.IsDir() {
		qtype = uint8(proto.DMDIR >> 24)
	}
	
	qid := proto.Qid{
		Qtype: qtype,
		Vers:  0,
		Uid:   v.nextQid,
	}
	
	v.nextQid++
	return qid
}

// entryToStat converts a VFS entry to a 9p Stat
func (v *VFS9P) entryToStat(entry vfs.FSEntry) proto.Stat {
	metadata := entry.GetMetadata()
	
	var mode uint32 = metadata.Mode
	if entry.IsDir() {
		mode |= proto.DMDIR
	}
	
	return proto.Stat{
		Type:   0,
		Dev:    0,
		Qid:    v.entryToQid(entry),
		Mode:   mode,
		Atime:  uint32(metadata.AccessedAt),
		Mtime:  uint32(metadata.ModifiedAt),
		Length: metadata.Size,
		Name:   metadata.Name,
		Uid:    metadata.Owner,
		Gid:    metadata.Group,
		Muid:   metadata.Owner,
	}
}
