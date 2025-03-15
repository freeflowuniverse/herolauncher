package vfsnested

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedVFS(t *testing.T) {
	// Create test directories
	testDir := filepath.Join(os.TempDir(), "test_nested_vfs")
	vfs1Dir := filepath.Join(testDir, "vfs1")
	vfs2Dir := filepath.Join(testDir, "vfs2")
	vfs3Dir := filepath.Join(testDir, "vfs3")

	// Clean up after test
	defer os.RemoveAll(testDir)

	// Create directories
	require.NoError(t, os.MkdirAll(vfs1Dir, 0755))
	require.NoError(t, os.MkdirAll(vfs2Dir, 0755))
	require.NoError(t, os.MkdirAll(vfs3Dir, 0755))

	// Create VFS instances
	vfs1, err := vfslocal.New(vfs1Dir)
	require.NoError(t, err)

	vfs2, err := vfslocal.New(vfs2Dir)
	require.NoError(t, err)

	vfs3, err := vfslocal.New(vfs3Dir)
	require.NoError(t, err)

	// Create nested VFS
	nestedVFS := New()

	// Add VFS instances at different paths
	require.NoError(t, nestedVFS.AddVFS("/data", vfs1))
	require.NoError(t, nestedVFS.AddVFS("/config", vfs2))
	require.NoError(t, nestedVFS.AddVFS("/data/backup", vfs3))

	t.Run("FileOperations", func(t *testing.T) {
		// Create and write to files in different VFS instances
		_, err := nestedVFS.FileCreate("/data/test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/data/test.txt", []byte("Hello from VFS1")))

		_, err = nestedVFS.FileCreate("/config/settings.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/config/settings.txt", []byte("Hello from VFS2")))

		_, err = nestedVFS.FileCreate("/data/backup/archive.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/data/backup/archive.txt", []byte("Hello from VFS3")))

		// Read and verify file contents
		data1, err := nestedVFS.FileRead("/data/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from VFS1", string(data1))

		data2, err := nestedVFS.FileRead("/config/settings.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from VFS2", string(data2))

		data3, err := nestedVFS.FileRead("/data/backup/archive.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from VFS3", string(data3))
	})

	t.Run("DirectoryOperations", func(t *testing.T) {
		// List root directory
		rootEntries, err := nestedVFS.DirList("/")
		require.NoError(t, err)
		// There should be at least 2 entries: /data and /config
		assert.GreaterOrEqual(t, len(rootEntries), 2)

		// Create and list directories
		_, err = nestedVFS.DirCreate("/data/subdir")
		require.NoError(t, err)
		_, err = nestedVFS.FileCreate("/data/subdir/file.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/data/subdir/file.txt", []byte("Nested file content")))

		dataEntries, err := nestedVFS.DirList("/data")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(dataEntries), 2) // test.txt and subdir (and possibly backup)

		// Verify file content
		fileData, err := nestedVFS.FileRead("/data/subdir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Nested file content", string(fileData))
	})

	t.Run("CrossVFSOperations", func(t *testing.T) {
		// Copy file between different VFS instances
		_, err := nestedVFS.Copy("/data/test.txt", "/config/test_copy.txt")
		require.NoError(t, err)

		copyData, err := nestedVFS.FileRead("/config/test_copy.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from VFS1", string(copyData))
	})

	t.Run("PathOperations", func(t *testing.T) {
		// Test path resolution
		entry, err := nestedVFS.Get("/data/test.txt")
		require.NoError(t, err)
		path, err := nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/data/test.txt", path)

		entry, err = nestedVFS.Get("/config/settings.txt")
		require.NoError(t, err)
		path, err = nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/config/settings.txt", path)

		entry, err = nestedVFS.Get("/data/backup/archive.txt")
		require.NoError(t, err)
		path, err = nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/data/backup/archive.txt", path)
	})

	t.Run("ExistsCheck", func(t *testing.T) {
		assert.True(t, nestedVFS.Exists("/"))
		assert.True(t, nestedVFS.Exists("/data"))
		assert.True(t, nestedVFS.Exists("/data/test.txt"))
		assert.True(t, nestedVFS.Exists("/config"))
		assert.True(t, nestedVFS.Exists("/config/settings.txt"))
		assert.True(t, nestedVFS.Exists("/data/backup"))
		assert.True(t, nestedVFS.Exists("/data/backup/archive.txt"))
		assert.False(t, nestedVFS.Exists("/nonexistent"))
	})

	// Clean up
	require.NoError(t, nestedVFS.Destroy())
}

func TestNestedVFSEdgeCases(t *testing.T) {
	// Create test directories
	testDir := filepath.Join(os.TempDir(), "test_nested_vfs_edge")
	vfs1Dir := filepath.Join(testDir, "vfs1")
	vfs2Dir := filepath.Join(testDir, "vfs2")

	// Clean up after test
	defer os.RemoveAll(testDir)

	// Create directories
	require.NoError(t, os.MkdirAll(vfs1Dir, 0755))
	require.NoError(t, os.MkdirAll(vfs2Dir, 0755))

	// Create VFS instances
	vfs1, err := vfslocal.New(vfs1Dir)
	require.NoError(t, err)

	vfs2, err := vfslocal.New(vfs2Dir)
	require.NoError(t, err)

	// Create nested VFS
	nestedVFS := New()

	// Add VFS instances at different paths
	require.NoError(t, nestedVFS.AddVFS("/data", vfs1))
	require.NoError(t, nestedVFS.AddVFS("/config", vfs2))

	t.Run("DuplicateMount", func(t *testing.T) {
		// Try to add a VFS at an existing path
		err := nestedVFS.AddVFS("/data", vfs2)
		assert.Error(t, err)
	})

	t.Run("ResourceForkFiles", func(t *testing.T) {
		// Test special handling for macOS resource fork files
		assert.True(t, nestedVFS.Exists("/._resource"))
		assert.True(t, nestedVFS.Exists("/data/._resource"))

		// Get a resource fork file
		entry, err := nestedVFS.Get("/data/._resource")
		require.NoError(t, err)
		assert.True(t, entry.IsFile())
		assert.False(t, entry.IsDir())
		assert.False(t, entry.IsSymlink())
		path, err := nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/data/._resource", path)

		// Read a resource fork file (should return empty data)
		data, err := nestedVFS.FileRead("/data/._resource")
		require.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("CrossVFSRename", func(t *testing.T) {
		// Create a test file
		_, err := nestedVFS.FileCreate("/data/rename_test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/data/rename_test.txt", []byte("Test content")))

		// Try to rename across different VFS implementations
		_, err = nestedVFS.Rename("/data/rename_test.txt", "/config/renamed.txt")
		assert.Error(t, err)
	})

	t.Run("CrossVFSMove", func(t *testing.T) {
		// Create a test file
		_, err := nestedVFS.FileCreate("/data/move_test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/data/move_test.txt", []byte("Test content")))

		// Try to move across different VFS implementations
		_, err = nestedVFS.Move("/data/move_test.txt", "/config/moved.txt")
		assert.Error(t, err)
	})

	t.Run("NoVFSForPath", func(t *testing.T) {
		// Try to access a path that doesn't match any VFS
		_, err := nestedVFS.Get("/nonexistent/path")
		assert.Error(t, err)

		_, err = nestedVFS.FileCreate("/nonexistent/file.txt")
		assert.Error(t, err)
	})

	// Clean up
	require.NoError(t, nestedVFS.Destroy())
}
