package vfsdb

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs"
)

func TestDatabaseVFS(t *testing.T) {
	// Create a temporary directory for the test databases
	tempDir, err := os.MkdirTemp("", "vfsdb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_vfs")

	// Create a new DatabaseVFS instance
	fs, err := NewFromPath(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DatabaseVFS: %v", err)
	}

	// Test root directory
	root, err := fs.RootGet()
	if err != nil {
		t.Fatalf("Failed to get root directory: %v", err)
	}

	if root.GetMetadata().FileType != vfs.FileTypeDirectory {
		t.Errorf("Root is not a directory")
	}

	// Test directory creation
	t.Run("DirectoryOperations", func(t *testing.T) {
		// Create a directory
		dir, err := fs.DirCreate("/testdir")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		if dir.GetMetadata().Name != "testdir" {
			t.Errorf("Directory name mismatch: got %s, want testdir", dir.GetMetadata().Name)
		}

		// Check if directory exists
		if !fs.Exists("/testdir") {
			t.Errorf("Directory should exist")
		}

		// Create a subdirectory
		subdir, err := fs.DirCreate("/testdir/subdir")
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		if subdir.GetMetadata().Name != "subdir" {
			t.Errorf("Subdirectory name mismatch: got %s, want subdir", subdir.GetMetadata().Name)
		}

		// List directory contents
		entries, err := fs.DirList("/testdir")
		if err != nil {
			t.Fatalf("Failed to list directory: %v", err)
		}

		if len(entries) != 1 {
			t.Errorf("Directory should have 1 entry, got %d", len(entries))
		}

		// Rename directory
		renamed, err := fs.Rename("/testdir/subdir", "/testdir/renamed")
		if err != nil {
			t.Fatalf("Failed to rename directory: %v", err)
		}

		if renamed.GetMetadata().Name != "renamed" {
			t.Errorf("Renamed directory name mismatch: got %s, want renamed", renamed.GetMetadata().Name)
		}

		// Move directory
		_, err = fs.DirCreate("/otherdir")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		moved, err := fs.Move("/testdir/renamed", "/otherdir/moved")
		if err != nil {
			t.Fatalf("Failed to move directory: %v", err)
		}

		if moved.GetMetadata().Name != "moved" {
			t.Errorf("Moved directory name mismatch: got %s, want moved", moved.GetMetadata().Name)
		}

		if !fs.Exists("/otherdir/moved") {
			t.Errorf("Moved directory should exist at new location")
		}

		if fs.Exists("/testdir/renamed") {
			t.Errorf("Directory should not exist at old location after move")
		}

		// Delete directory
		err = fs.DirDelete("/otherdir/moved")
		if err != nil {
			t.Fatalf("Failed to delete directory: %v", err)
		}

		if fs.Exists("/otherdir/moved") {
			t.Errorf("Directory should not exist after deletion")
		}
	})

	// Test file operations
	t.Run("FileOperations", func(t *testing.T) {
		// Create a file
		file, err := fs.FileCreate("/testfile.txt")
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		if file.GetMetadata().Name != "testfile.txt" {
			t.Errorf("File name mismatch: got %s, want testfile.txt", file.GetMetadata().Name)
		}

		// Write to file
		testData := []byte("Hello, world!")
		err = fs.FileWrite("/testfile.txt", testData)
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}

		// Read from file
		data, err := fs.FileRead("/testfile.txt")
		if err != nil {
			t.Fatalf("Failed to read from file: %v", err)
		}

		if !bytes.Equal(data, testData) {
			t.Errorf("File content mismatch: got %s, want %s", data, testData)
		}

		// Append to file
		appendData := []byte(" More data.")
		err = fs.FileConcatenate("/testfile.txt", appendData)
		if err != nil {
			t.Fatalf("Failed to append to file: %v", err)
		}

		// Read again to verify append
		data, err = fs.FileRead("/testfile.txt")
		if err != nil {
			t.Fatalf("Failed to read from file after append: %v", err)
		}

		expectedData := append(testData, appendData...)
		if !bytes.Equal(data, expectedData) {
			t.Errorf("File content mismatch after append: got %s, want %s", data, expectedData)
		}

		// Copy file
		_, err = fs.Copy("/testfile.txt", "/testfile_copy.txt")
		if err != nil {
			t.Fatalf("Failed to copy file: %v", err)
		}

		// Read copied file
		copyData, err := fs.FileRead("/testfile_copy.txt")
		if err != nil {
			t.Fatalf("Failed to read from copied file: %v", err)
		}

		if !bytes.Equal(copyData, expectedData) {
			t.Errorf("Copied file content mismatch: got %s, want %s", copyData, expectedData)
		}

		// Delete file
		err = fs.FileDelete("/testfile.txt")
		if err != nil {
			t.Fatalf("Failed to delete file: %v", err)
		}

		if fs.Exists("/testfile.txt") {
			t.Errorf("File should not exist after deletion")
		}
	})

	// Test symlink operations
	t.Run("SymlinkOperations", func(t *testing.T) {
		// Create a target file
		_, err := fs.FileCreate("/target.txt")
		if err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}

		err = fs.FileWrite("/target.txt", []byte("Target file content"))
		if err != nil {
			t.Fatalf("Failed to write to target file: %v", err)
		}

		// Create a symlink
		symlink, err := fs.LinkCreate("/target.txt", "/link.txt")
		if err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		if symlink.GetMetadata().Name != "link.txt" {
			t.Errorf("Symlink name mismatch: got %s, want link.txt", symlink.GetMetadata().Name)
		}

		// Read symlink target
		target, err := fs.LinkRead("/link.txt")
		if err != nil {
			t.Fatalf("Failed to read symlink target: %v", err)
		}

		if target != "/target.txt" {
			t.Errorf("Symlink target mismatch: got %s, want /target.txt", target)
		}

		// Delete symlink
		err = fs.LinkDelete("/link.txt")
		if err != nil {
			t.Fatalf("Failed to delete symlink: %v", err)
		}

		if fs.Exists("/link.txt") {
			t.Errorf("Symlink should not exist after deletion")
		}
	})

	// Test path operations
	t.Run("PathOperations", func(t *testing.T) {
		// Create nested directories
		_, err := fs.DirCreate("/a")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		_, err = fs.DirCreate("/a/b")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		_, err = fs.DirCreate("/a/b/c")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Create a file in the nested directory
		file, err := fs.FileCreate("/a/b/c/file.txt")
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Get path for the file
		path, err := fs.GetPath(file)
		if err != nil {
			t.Fatalf("Failed to get path for file: %v", err)
		}

		if path != "/a/b/c/file.txt" {
			t.Errorf("Path mismatch: got %s, want /a/b/c/file.txt", path)
		}

		// Test non-existent path
		_, err = fs.Get("/nonexistent")
		if err == nil {
			t.Errorf("Expected error for non-existent path")
		}
	})

	// Test encoding and decoding
	t.Run("EncodingDecoding", func(t *testing.T) {
		// Create a directory
		dir, err := fs.DirCreate("/encode_test")
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Create a file
		file, err := fs.FileCreate("/encode_test/file.txt")
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Write to file
		err = fs.FileWrite("/encode_test/file.txt", []byte("Test encoding"))
		if err != nil {
			t.Fatalf("Failed to write to file: %v", err)
		}

		// Create a symlink
		_, err = fs.LinkCreate("/encode_test/file.txt", "/encode_test/link.txt")
		if err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Get entries
		entries, err := fs.DirList("/encode_test")
		if err != nil {
			t.Fatalf("Failed to list directory: %v", err)
		}

		if len(entries) != 2 {
			t.Errorf("Directory should have 2 entries, got %d", len(entries))
		}

		// Test directory encoding/decoding
		dirEntry, ok := dir.(*DirectoryEntry)
		if !ok {
			t.Fatalf("Failed to cast to DirectoryEntry")
		}

		dirData, err := encodeDirectory(dirEntry)
		if err != nil {
			t.Fatalf("Failed to encode directory: %v", err)
		}

		decodedDir, err := decodeDirectory(dirData, fs)
		if err != nil {
			t.Fatalf("Failed to decode directory: %v", err)
		}

		if decodedDir.metadata.ID != dirEntry.metadata.ID {
			t.Errorf("Directory ID mismatch after encoding/decoding")
		}

		// Test file encoding/decoding
		fileEntry, ok := file.(*FileEntry)
		if !ok {
			t.Fatalf("Failed to cast to FileEntry")
		}

		fileData, err := encodeFile(fileEntry)
		if err != nil {
			t.Fatalf("Failed to encode file: %v", err)
		}

		decodedFile, err := decodeFile(fileData, fs)
		if err != nil {
			t.Fatalf("Failed to decode file: %v", err)
		}

		if decodedFile.metadata.ID != fileEntry.metadata.ID {
			t.Errorf("File ID mismatch after encoding/decoding")
		}
	})
}
