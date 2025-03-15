package vfsdav_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "vfsdav-server-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	require.NoError(t, err)

	// Create a sample file in the temp directory
	sampleContent := "Hello, WebDAV Server Test!"
	sampleFile := "sample.txt"
	err = os.WriteFile(tempDir+"/"+sampleFile, []byte(sampleContent), 0644)
	require.NoError(t, err)

	// Create and start the WebDAV server
	addr := "localhost:8081" // Use a different port than the examples
	server := vfsdav.NewServer(vfsImpl, addr)

	// Start the server in a goroutine
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	// Create an HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Test GET request to read a file
	t.Run("GetFile", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/%s", addr, sampleFile)
		resp, err := client.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, sampleContent, string(data))
	})

	// Test PUT request to create a file
	t.Run("PutFile", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/test-put.txt", addr)
		content := "This is a test PUT request"
		req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(content))
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify the file was created
		getResp, err := client.Get(url)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		data, err := io.ReadAll(getResp.Body)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	// Test MKCOL request to create a directory
	t.Run("MakeCollection", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/test-dir", addr)
		req, err := http.NewRequest("MKCOL", url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify the directory was created by putting a file in it
		putURL := fmt.Sprintf("http://%s/test-dir/file.txt", addr)
		content := "File in directory"
		putReq, err := http.NewRequest(http.MethodPut, putURL, strings.NewReader(content))
		require.NoError(t, err)

		putResp, err := client.Do(putReq)
		require.NoError(t, err)
		defer putResp.Body.Close()

		assert.Equal(t, http.StatusCreated, putResp.StatusCode)
	})

	// Test PROPFIND request to list directory contents
	t.Run("PropFind", func(t *testing.T) {
		url := fmt.Sprintf("http://%s/", addr)
		req, err := http.NewRequest("PROPFIND", url, nil)
		require.NoError(t, err)
		req.Header.Add("Depth", "1")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusMultiStatus, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		
		// Just check that we got some XML response with our files
		responseStr := string(data)
		assert.Contains(t, responseStr, "<?xml")
		assert.Contains(t, responseStr, "sample.txt")
		assert.Contains(t, responseStr, "test-put.txt")
		assert.Contains(t, responseStr, "test-dir")
	})

	// Test DELETE request
	t.Run("DeleteFile", func(t *testing.T) {
		// First create a file to delete
		url := fmt.Sprintf("http://%s/to-delete.txt", addr)
		content := "This file will be deleted"
		req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(content))
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Now delete it
		delReq, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		delResp, err := client.Do(delReq)
		require.NoError(t, err)
		defer delResp.Body.Close()

		assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

		// Verify it's gone
		getResp, err := client.Get(url)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}
