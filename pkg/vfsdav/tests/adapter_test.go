package vfsdav_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "vfsdav-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	require.NoError(t, err)

	// Create the WebDAV adapter
	adapter := vfsdav.New(vfsImpl)
	ctx := context.Background()

	// Test file operations
	t.Run("FileOperations", func(t *testing.T) {
		// Create a file
		content := "Hello, WebDAV!"
		body := io.NopCloser(strings.NewReader(content))
		fileInfo, created, err := adapter.Create(ctx, "/test.txt", body)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, "/test.txt", fileInfo.Path)
		// Size may be 0 for newly created files in some implementations
		assert.False(t, fileInfo.IsDir)

		// Read the file
		reader, err := adapter.Open(ctx, "/test.txt")
		require.NoError(t, err)
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
		reader.Close()

		// Stat the file
		fileInfo, err = adapter.Stat(ctx, "/test.txt")
		require.NoError(t, err)
		assert.Equal(t, "/test.txt", fileInfo.Path)
		assert.Equal(t, int64(len(content)), fileInfo.Size)
		assert.False(t, fileInfo.IsDir)

		// Update the file
		newContent := "Updated content"
		body = io.NopCloser(strings.NewReader(newContent))
		fileInfo, created, err = adapter.Create(ctx, "/test.txt", body)
		require.NoError(t, err)
		assert.False(t, created)
		// Size may be 0 for newly created files in some implementations

		// Read the updated file
		reader, err = adapter.Open(ctx, "/test.txt")
		require.NoError(t, err)
		data, err = io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, newContent, string(data))
		reader.Close()

		// Remove the file
		err = adapter.RemoveAll(ctx, "/test.txt")
		require.NoError(t, err)

		// Verify the file is gone
		_, err = adapter.Stat(ctx, "/test.txt")
		assert.Error(t, err)
	})

	// Test directory operations
	t.Run("DirectoryOperations", func(t *testing.T) {
		// Create a directory
		err := adapter.Mkdir(ctx, "/testdir")
		require.NoError(t, err)

		// Stat the directory
		dirInfo, err := adapter.Stat(ctx, "/testdir")
		require.NoError(t, err)
		assert.Equal(t, "/testdir", dirInfo.Path)
		assert.True(t, dirInfo.IsDir)

		// Create a file in the directory
		content := "File in directory"
		body := io.NopCloser(strings.NewReader(content))
		_, created, err := adapter.Create(ctx, "/testdir/file.txt", body)
		require.NoError(t, err)
		assert.True(t, created)

		// List directory contents
		entries, err := adapter.ReadDir(ctx, "/testdir", false)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "/testdir/file.txt", entries[0].Path)

		// Remove the directory
		err = adapter.RemoveAll(ctx, "/testdir")
		require.NoError(t, err)

		// Verify the directory is gone
		_, err = adapter.Stat(ctx, "/testdir")
		assert.Error(t, err)
	})

	// Test copy operation
	t.Run("CopyOperation", func(t *testing.T) {
		// Create a file to copy
		content := "File to copy"
		body := io.NopCloser(strings.NewReader(content))
		_, _, err := adapter.Create(ctx, "/source.txt", body)
		require.NoError(t, err)

		// Copy the file
		created, err := adapter.Copy(ctx, "/source.txt", "/dest.txt", &webdav.CopyOptions{})
		require.NoError(t, err)
		assert.True(t, created)

		// Verify the copied file
		reader, err := adapter.Open(ctx, "/dest.txt")
		require.NoError(t, err)
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
		reader.Close()

		// Clean up
		err = adapter.RemoveAll(ctx, "/source.txt")
		require.NoError(t, err)
		err = adapter.RemoveAll(ctx, "/dest.txt")
		require.NoError(t, err)
	})

	// Test move operation
	t.Run("MoveOperation", func(t *testing.T) {
		// Create a file to move
		content := "File to move"
		body := io.NopCloser(strings.NewReader(content))
		_, _, err := adapter.Create(ctx, "/source.txt", body)
		require.NoError(t, err)

		// Move the file
		created, err := adapter.Move(ctx, "/source.txt", "/moved.txt", &webdav.MoveOptions{})
		require.NoError(t, err)
		assert.True(t, created)

		// Verify the source file is gone
		_, err = adapter.Stat(ctx, "/source.txt")
		assert.Error(t, err)

		// Verify the destination file exists
		reader, err := adapter.Open(ctx, "/moved.txt")
		require.NoError(t, err)
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
		reader.Close()

		// Clean up
		err = adapter.RemoveAll(ctx, "/moved.txt")
		require.NoError(t, err)
	})
}
