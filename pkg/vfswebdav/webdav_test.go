package vfswebdav

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test WebDAV server with a temporary directory
func setupTestServer(t *testing.T) (*httptest.Server, string) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "webdav-test-")
	require.NoError(t, err)

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	require.NoError(t, err)

	// Create a WebDAV filesystem
	fs := NewFileSystem(vfsImpl)

	// Create a WebDAV handler
	handler := &webdav.Handler{
		FileSystem: fs,
	}

	// Create a test server
	server := httptest.NewServer(handler)

	return server, tempDir
}

// createWebDAVClient creates a WebDAV client for the test server
func createWebDAVClient(serverURL string) (*webdav.Client, error) {
	return webdav.NewClient(http.DefaultClient, serverURL)
}

func TestWebDAVBasicOperations(t *testing.T) {
	// Setup test server
	server, tempDir := setupTestServer(t)
	defer server.Close()
	defer os.RemoveAll(tempDir)

	// Create a WebDAV client
	client, err := createWebDAVClient(server.URL)
	require.NoError(t, err)

	// Test context
	ctx := context.Background()

	// Test file content
	testContent := "Hello, WebDAV World!"

	t.Run("Create and Read File", func(t *testing.T) {
		// Create a test file
		filePath := "/test.txt"
		response, err := http.NewRequest("PUT", server.URL+filePath, strings.NewReader(testContent))
		require.NoError(t, err)
		
		resp, err := http.DefaultClient.Do(response)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		// Read response body for error details
		respBody, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			t.Logf("PUT request failed with status %d: %s", resp.StatusCode, string(respBody))
		}
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Read the file
		resp, err = http.Get(server.URL + filePath)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		// Read response body for error details
		respBody2, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Logf("GET request failed with status %d: %s", resp.StatusCode, string(respBody2))
		}
		// Reset the response body for further reading
		resp.Body = ioutil.NopCloser(strings.NewReader(string(respBody2)))
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content))

		// Verify the file exists on disk
		localPath := filepath.Join(tempDir, filePath)
		_, err = os.Stat(localPath)
		assert.NoError(t, err)
	})

	t.Run("Stat File", func(t *testing.T) {
		filePath := "/test.txt"
		info, err := client.Stat(ctx, filePath)
		require.NoError(t, err)
		assert.Equal(t, filePath, info.Path)
		assert.Equal(t, int64(len(testContent)), info.Size)
		assert.False(t, info.IsDir)
	})

	t.Run("Create and List Directory", func(t *testing.T) {
		// Create a directory
		dirPath := "/testdir"
		err := client.Mkdir(ctx, dirPath)
		require.NoError(t, err)

		// Create a file in the directory
		filePath := dirPath + "/file.txt"
		request, err := http.NewRequest("PUT", server.URL+filePath, strings.NewReader(testContent))
		require.NoError(t, err)
		
		resp, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// List the directory
		files, err := client.ReadDir(ctx, dirPath, false)
		require.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, filePath, files[0].Path)
	})

	t.Run("Remove File", func(t *testing.T) {
		filePath := "/test.txt"
		err := client.RemoveAll(ctx, filePath)
		require.NoError(t, err)

		// Verify the file doesn't exist
		_, err = client.Stat(ctx, filePath)
		assert.Error(t, err)
	})

	t.Run("Remove Directory", func(t *testing.T) {
		dirPath := "/testdir"
		err := client.RemoveAll(ctx, dirPath)
		require.NoError(t, err)

		// Verify the directory doesn't exist
		_, err = client.Stat(ctx, dirPath)
		assert.Error(t, err)
	})
}

func TestWebDAVCopyMove(t *testing.T) {
	// Setup test server
	server, tempDir := setupTestServer(t)
	defer server.Close()
	defer os.RemoveAll(tempDir)

	// Create a WebDAV client
	client, err := createWebDAVClient(server.URL)
	require.NoError(t, err)

	// Test context
	ctx := context.Background()

	// Test file content
	testContent := "Hello, WebDAV World!"

	t.Run("Copy File", func(t *testing.T) {
		// Create a test file
		srcPath := "/source.txt"
		destPath := "/dest.txt"
		request, err := http.NewRequest("PUT", server.URL+srcPath, strings.NewReader(testContent))
		require.NoError(t, err)
		
		resp, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Copy the file
		err = client.Copy(ctx, srcPath, destPath, &webdav.CopyOptions{})
		require.NoError(t, err)

		// Verify both files exist
		_, err = client.Stat(ctx, srcPath)
		assert.NoError(t, err)
		_, err = client.Stat(ctx, destPath)
		assert.NoError(t, err)

		// Read the destination file
		resp, err = http.Get(server.URL + destPath)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("Move File", func(t *testing.T) {
		// Create a test file
		srcPath := "/move-source.txt"
		destPath := "/move-dest.txt"
		request, err := http.NewRequest("PUT", server.URL+srcPath, strings.NewReader(testContent))
		require.NoError(t, err)
		
		resp, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Move the file
		err = client.Move(ctx, srcPath, destPath, &webdav.MoveOptions{})
		require.NoError(t, err)

		// Verify source doesn't exist and destination does
		_, err = client.Stat(ctx, srcPath)
		assert.Error(t, err)
		_, err = client.Stat(ctx, destPath)
		assert.NoError(t, err)

		// Read the destination file
		resp, err = http.Get(server.URL + destPath)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("Copy Directory", func(t *testing.T) {
		// Create a directory with a file
		srcDir := "/src-dir"
		destDir := "/dest-dir"
		filePath := srcDir + "/file.txt"

		err := client.Mkdir(ctx, srcDir)
		require.NoError(t, err)
		
		request, err := http.NewRequest("PUT", server.URL+filePath, strings.NewReader(testContent))
		require.NoError(t, err)
		
		resp, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Copy the directory
		err = client.Copy(ctx, srcDir, destDir, &webdav.CopyOptions{})
		require.NoError(t, err)

		// Verify both directories exist
		_, err = client.Stat(ctx, srcDir)
		assert.NoError(t, err)
		_, err = client.Stat(ctx, destDir)
		assert.NoError(t, err)

		// Verify the file was copied
		_, err = client.Stat(ctx, destDir+"/file.txt")
		assert.NoError(t, err)
	})
}

func TestWebDAVNestedServer(t *testing.T) {
	// Create two temporary directories
	tempDir1, err := ioutil.TempDir("", "webdav-test-1-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir1)

	tempDir2, err := ioutil.TempDir("", "webdav-test-2-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir2)

	// Create test files in each directory
	content1 := "Content in directory 1"
	content2 := "Content in directory 2"
	
	err = ioutil.WriteFile(filepath.Join(tempDir1, "file1.txt"), []byte(content1), 0644)
	require.NoError(t, err)
	
	err = ioutil.WriteFile(filepath.Join(tempDir2, "file2.txt"), []byte(content2), 0644)
	require.NoError(t, err)

	// Start a WebDAV server with nested VFS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a nested WebDAV server
		if strings.HasPrefix(r.URL.Path, "/dir1/") {
			// Adjust the path to remove the prefix
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/dir1")
			
			// Create a local VFS for dir1
			vfsImpl, _ := vfslocal.New(tempDir1)
			fs := NewFileSystem(vfsImpl)
			handler := &webdav.Handler{FileSystem: fs}
			handler.ServeHTTP(w, r)
			
		} else if strings.HasPrefix(r.URL.Path, "/dir2/") {
			// Adjust the path to remove the prefix
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/dir2")
			
			// Create a local VFS for dir2
			vfsImpl, _ := vfslocal.New(tempDir2)
			fs := NewFileSystem(vfsImpl)
			handler := &webdav.Handler{FileSystem: fs}
			handler.ServeHTTP(w, r)
			
		} else {
			// Handle root or other paths
			fmt.Fprintf(w, "WebDAV Root")
		}
	}))
	defer server.Close()

	// We don't need the client for this test as we're directly testing the HTTP endpoints

	// Test accessing files in the nested directories
	t.Run("Access Files in Nested Directories", func(t *testing.T) {
		// Read file from dir1
		resp, err := http.Get(server.URL + "/dir1/file1.txt")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content1, string(content))

		// Read file from dir2
		resp, err = http.Get(server.URL + "/dir2/file2.txt")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		content, err = ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, content2, string(content))
	})
}
