package main

import (
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

// Mock implementations for testing

// mockFS implements fs.FS interface for testing
type mockFS struct {
	newStatFunc func(name, uid, gid string, mode uint32) *proto.Stat
	rootFunc func() fs.Dir
}

func (m *mockFS) NewStat(name, uid, gid string, mode uint32) *proto.Stat {
	if m.newStatFunc != nil {
		return m.newStatFunc(name, uid, gid, mode)
	}
	return &proto.Stat{
		Name: name,
		Uid:  uid,
		Gid:  gid,
		Mode: mode,
	}
}

func (m *mockFS) Server() interface{} {
	return nil
}

func (m *mockFS) Root() fs.Dir {
	if m.rootFunc != nil {
		return m.rootFunc()
	}
	return nil
}

// Implement the remaining fs.FS interface methods
func (m *mockFS) Attach() (fs.FSNode, error) {
	return nil, nil
}

func (m *mockFS) Auth(user, aname string) (fs.FSNode, error) {
	return nil, nil
}

func (m *mockFS) Begin() error {
	return nil
}

func (m *mockFS) End() error {
	return nil
}

func (m *mockFS) CreateFile(parent fs.Dir, name, user string, perms uint32) (fs.File, error) {
	return nil, nil
}

func (m *mockFS) CreateDir(parent fs.Dir, name, user string, perms uint32) (fs.Dir, error) {
	return nil, nil
}

func (m *mockFS) CanOpen(user string, f fs.File, mode uint8) bool {
	return true
}

type mockDir struct {
	addChildFunc func(n interface{}) error
	statFunc     func() proto.Stat
	writeStatFunc func(*proto.Stat) error
	childrenFunc func() map[string]fs.FSNode
	parentFunc func() fs.Dir
	setParentFunc func(fs.Dir)
}

func (m *mockDir) AddChild(n interface{}) error {
	if m.addChildFunc != nil {
		return m.addChildFunc(n)
	}
	return nil
}

func (m *mockDir) Stat() proto.Stat {
	if m.statFunc != nil {
		return m.statFunc()
	}
	return proto.Stat{}
}

func (m *mockDir) Children() map[string]fs.FSNode {
	if m.childrenFunc != nil {
		return m.childrenFunc()
	}
	return make(map[string]fs.FSNode)
}

func (m *mockDir) Parent() fs.Dir {
	if m.parentFunc != nil {
		return m.parentFunc()
	}
	return nil
}

func (m *mockDir) SetParent(p fs.Dir) {
	if m.setParentFunc != nil {
		m.setParentFunc(p)
	}
}

func (m *mockDir) WriteStat(stat *proto.Stat) error {
	if m.writeStatFunc != nil {
		return m.writeStatFunc(stat)
	}
	return nil
}

// mockModDir implements fs.ModDir interface for testing
type mockModDir struct {
	addChildFunc    func(n interface{}) error
	deleteChildFunc func(name string) error
	statFunc        func() proto.Stat
	writeStatFunc   func(*proto.Stat) error
	childrenFunc    func() map[string]fs.FSNode
	parentFunc      func() fs.Dir
	setParentFunc   func(fs.Dir)
}

func (m *mockModDir) AddChild(n interface{}) error {
	if m.addChildFunc != nil {
		return m.addChildFunc(n)
	}
	return nil
}

func (m *mockModDir) DeleteChild(name string) error {
	if m.deleteChildFunc != nil {
		return m.deleteChildFunc(name)
	}
	return nil
}

func (m *mockModDir) Stat() proto.Stat {
	if m.statFunc != nil {
		return m.statFunc()
	}
	return proto.Stat{}
}

func (m *mockModDir) Children() map[string]fs.FSNode {
	if m.childrenFunc != nil {
		return m.childrenFunc()
	}
	return make(map[string]fs.FSNode)
}

func (m *mockModDir) Parent() fs.Dir {
	if m.parentFunc != nil {
		return m.parentFunc()
	}
	return nil
}

func (m *mockModDir) SetParent(p fs.Dir) {
	if m.setParentFunc != nil {
		m.setParentFunc(p)
	}
}

func (m *mockModDir) WriteStat(stat *proto.Stat) error {
	if m.writeStatFunc != nil {
		return m.writeStatFunc(stat)
	}
	return nil
}
