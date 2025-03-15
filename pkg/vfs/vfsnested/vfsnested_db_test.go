package vfsnested

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedVFSWithDB(t *testing.T) {
	// Create test directories and database paths
	testDir := filepath.Join(os.TempDir(), "test_nested_vfs_db")
	vfs1Dir := filepath.Join(testDir, "vfs1")
	dbPath := filepath.Join(testDir, "vfs_db")

	// Clean up after test
	defer os.RemoveAll(testDir)

	// Create directories
	require.NoError(t, os.MkdirAll(vfs1Dir, 0755))

	// Create VFS instances
	vfs1, err := vfslocal.New(vfs1Dir)
	require.NoError(t, err)

	vfs2, err := vfsdb.NewFromPath(dbPath)
	require.NoError(t, err)
	defer vfs2.Destroy()

	// Create nested VFS
	nestedVFS := New()

	// Add VFS instances at different paths
	require.NoError(t, nestedVFS.AddVFS("/local", vfs1))
	require.NoError(t, nestedVFS.AddVFS("/db", vfs2))

	t.Run("MixedFileOperations", func(t *testing.T) {
		// Create and write to files in different VFS instances
		_, err := nestedVFS.FileCreate("/local/test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/local/test.txt", []byte("Hello from local VFS")))

		_, err = nestedVFS.FileCreate("/db/settings.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db/settings.txt", []byte("Hello from DB VFS")))

		// Read from files in different VFS instances
		data1, err := nestedVFS.FileRead("/local/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from local VFS", string(data1))

		data2, err := nestedVFS.FileRead("/db/settings.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from DB VFS", string(data2))
	})

	t.Run("MixedDirectoryOperations", func(t *testing.T) {
		// List root directory
		rootEntries, err := nestedVFS.DirList("/")
		require.NoError(t, err)
		// There should be at least 2 entries: /local and /db
		assert.GreaterOrEqual(t, len(rootEntries), 2)

		// Create and list directories in both VFS types
		_, err = nestedVFS.DirCreate("/local/subdir")
		require.NoError(t, err)
		_, err = nestedVFS.FileCreate("/local/subdir/file.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/local/subdir/file.txt", []byte("Local nested file content")))

		_, err = nestedVFS.DirCreate("/db/subdir")
		require.NoError(t, err)
		_, err = nestedVFS.FileCreate("/db/subdir/file.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db/subdir/file.txt", []byte("DB nested file content")))

		// List directories in both VFS types
		localEntries, err := nestedVFS.DirList("/local")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(localEntries), 2) // test.txt and subdir

		dbEntries, err := nestedVFS.DirList("/db")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(dbEntries), 2) // settings.txt and subdir

		// Verify file content
		localData, err := nestedVFS.FileRead("/local/subdir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Local nested file content", string(localData))

		dbData, err := nestedVFS.FileRead("/db/subdir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "DB nested file content", string(dbData))
	})

	t.Run("CrossVFSOperations", func(t *testing.T) {
		// Test copying between different VFS types
		_, err := nestedVFS.Copy("/local/test.txt", "/db/test_copy.txt")
		require.NoError(t, err)

		copyData, err := nestedVFS.FileRead("/db/test_copy.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from local VFS", string(copyData))

		// Test copying from DB to local
		_, err = nestedVFS.Copy("/db/settings.txt", "/local/settings_copy.txt")
		require.NoError(t, err)

		copyData2, err := nestedVFS.FileRead("/local/settings_copy.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from DB VFS", string(copyData2))
	})

	t.Run("PathOperations", func(t *testing.T) {
		// Test path resolution for both VFS types
		entry, err := nestedVFS.Get("/local/test.txt")
		require.NoError(t, err)
		path, err := nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/local/test.txt", path)

		entry, err = nestedVFS.Get("/db/settings.txt")
		require.NoError(t, err)
		path, err = nestedVFS.GetPath(entry)
		require.NoError(t, err)
		assert.Equal(t, "/db/settings.txt", path)
	})

	t.Run("ExistsCheck", func(t *testing.T) {
		assert.True(t, nestedVFS.Exists("/"))
		assert.True(t, nestedVFS.Exists("/local"))
		assert.True(t, nestedVFS.Exists("/local/test.txt"))
		assert.True(t, nestedVFS.Exists("/db"))
		assert.True(t, nestedVFS.Exists("/db/settings.txt"))
		assert.False(t, nestedVFS.Exists("/nonexistent"))
	})

	// Clean up
	require.NoError(t, nestedVFS.Destroy())
}

func TestNestedVFSMultipleDBInstances(t *testing.T) {
	// Create test directories and database paths
	testDir := filepath.Join(os.TempDir(), "test_nested_vfs_multi_db")
	db1Path := filepath.Join(testDir, "vfs_db1")
	db2Path := filepath.Join(testDir, "vfs_db2")

	// Clean up after test
	defer os.RemoveAll(testDir)

	// Create VFS instances
	vfs1, err := vfsdb.NewFromPath(db1Path)
	require.NoError(t, err)
	defer vfs1.Destroy()

	vfs2, err := vfsdb.NewFromPath(db2Path)
	require.NoError(t, err)
	defer vfs2.Destroy()

	// Create nested VFS
	nestedVFS := New()

	// Add VFS instances at different paths
	require.NoError(t, nestedVFS.AddVFS("/db1", vfs1))
	require.NoError(t, nestedVFS.AddVFS("/db2", vfs2))

	t.Run("BasicOperations", func(t *testing.T) {
		// Create and write to files in different DB VFS instances
		_, err := nestedVFS.FileCreate("/db1/test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db1/test.txt", []byte("Hello from DB1")))

		_, err = nestedVFS.FileCreate("/db2/test.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db2/test.txt", []byte("Hello from DB2")))

		// Read from files in different DB VFS instances
		data1, err := nestedVFS.FileRead("/db1/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from DB1", string(data1))

		data2, err := nestedVFS.FileRead("/db2/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from DB2", string(data2))
	})

	t.Run("NestedDirectories", func(t *testing.T) {
		// Create nested directories in both DB VFS instances
		_, err := nestedVFS.DirCreate("/db1/dir1")
		require.NoError(t, err)
		_, err = nestedVFS.DirCreate("/db1/dir1/subdir")
		require.NoError(t, err)
		
		_, err = nestedVFS.DirCreate("/db2/dir2")
		require.NoError(t, err)
		_, err = nestedVFS.DirCreate("/db2/dir2/subdir")
		require.NoError(t, err)

		// Create files in nested directories
		_, err = nestedVFS.FileCreate("/db1/dir1/subdir/file.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db1/dir1/subdir/file.txt", []byte("Nested file in DB1")))

		_, err = nestedVFS.FileCreate("/db2/dir2/subdir/file.txt")
		require.NoError(t, err)
		require.NoError(t, nestedVFS.FileWrite("/db2/dir2/subdir/file.txt", []byte("Nested file in DB2")))

		// Verify file content
		data1, err := nestedVFS.FileRead("/db1/dir1/subdir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Nested file in DB1", string(data1))

		data2, err := nestedVFS.FileRead("/db2/dir2/subdir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Nested file in DB2", string(data2))
	})

	t.Run("CrossDBOperations", func(t *testing.T) {
		// Test copying between different DB VFS instances
		_, err := nestedVFS.Copy("/db1/test.txt", "/db2/test_copy_from_db1.txt")
		require.NoError(t, err)

		copyData, err := nestedVFS.FileRead("/db2/test_copy_from_db1.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello from DB1", string(copyData))

		// Test copying nested files
		_, err = nestedVFS.Copy("/db1/dir1/subdir/file.txt", "/db2/dir2/copied_nested_file.txt")
		require.NoError(t, err)

		copyData2, err := nestedVFS.FileRead("/db2/dir2/copied_nested_file.txt")
		require.NoError(t, err)
		assert.Equal(t, "Nested file in DB1", string(copyData2))
	})

	// Clean up
	require.NoError(t, nestedVFS.Destroy())
}
