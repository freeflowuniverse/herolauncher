package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WebDAVClient struct {
	URL      string
	Username string
	Password string
	Client   *http.Client
}

// NewClient creates a new WebDAV client
func NewClient(url, username, password string) *WebDAVClient {
	return &WebDAVClient{
		URL:      strings.TrimSuffix(url, "/"),
		Username: username,
		Password: password,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListFiles lists files and directories at the specified path
func (c *WebDAVClient) ListFiles(path string) error {
	req, err := http.NewRequest("PROPFIND", c.URL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set depth to 1 to list immediate children
	req.Header.Add("Depth", "1")

	// Add basic authentication if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Connected to WebDAV server at %s\n\n", c.URL)
	fmt.Printf("Directory listing for: %s\n", path)
	fmt.Println("----------------------------------------")

	// Parse the XML response to extract file information
	if strings.Contains(string(body), "<D:multistatus") {
		// Simple parsing of the XML response to extract file names and properties
		responses := strings.Split(string(body), "<D:response>")
		
		// Skip the first split result as it's the XML header
		for i, response := range responses {
			if i == 0 {
				continue
			}
			
			// Extract the href (path)
			hrefStart := strings.Index(response, "<D:href>")
			hrefEnd := strings.Index(response, "</D:href>")
			if hrefStart == -1 || hrefEnd == -1 {
				continue
			}
			
			href := response[hrefStart+8:hrefEnd]
			
			// Skip the current directory entry
			if href == path || href+"/" == path {
				continue
			}
			
			// Extract file name from href
			name := filepath.Base(href)
			if name == "." || name == "" {
				name = href
			}
			
			// Check if it's a directory
			isDir := strings.Contains(response, "<D:collection")
			
			// Extract last modified time
			lastModified := "Unknown"
			lmStart := strings.Index(response, "<D:getlastmodified>")
			lmEnd := strings.Index(response, "</D:getlastmodified>")
			if lmStart != -1 && lmEnd != -1 {
				lastModified = response[lmStart+19:lmEnd]
			}
			
			// Extract content length for files
			size := "-"
			if !isDir {
				clStart := strings.Index(response, "<D:getcontentlength>")
				clEnd := strings.Index(response, "</D:getcontentlength>")
				if clStart != -1 && clEnd != -1 {
					size = response[clStart+20:clEnd] + " bytes"
				}
			}
			
			// Format and print the entry
			fileType := "File"
			if isDir {
				fileType = "Directory"
				name += "/"
			}
			
			fmt.Printf("%-12s %-30s %-20s %s\n", fileType, name, size, lastModified)
		}
	} else {
		// Fallback to displaying the raw response
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Response:\n%s\n", string(body))
	}

	fmt.Println("\nUse -action upload to upload files or -action mkdir to create directories")
	return nil
}

// UploadFile uploads a file to the WebDAV server
func (c *WebDAVClient) UploadFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", c.URL+remotePath, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic authentication if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Printf("File uploaded successfully: %s -> %s\n", localPath, remotePath)
	return nil
}

// DownloadFile downloads a file from the WebDAV server
func (c *WebDAVClient) DownloadFile(remotePath, localPath string) error {
	req, err := http.NewRequest("GET", c.URL+remotePath, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic authentication if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the response body to the file
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("File downloaded successfully: %s -> %s\n", remotePath, localPath)
	return nil
}

// CreateDirectory creates a directory on the WebDAV server
func (c *WebDAVClient) CreateDirectory(path string) error {
	req, err := http.NewRequest("MKCOL", c.URL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic authentication if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Printf("Directory created successfully: %s\n", path)
	return nil
}

// DeleteFile deletes a file or directory from the WebDAV server
func (c *WebDAVClient) DeleteFile(path string) error {
	req, err := http.NewRequest("DELETE", c.URL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic authentication if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Printf("File or directory deleted successfully: %s\n", path)
	return nil
}

// RunComprehensiveTest runs a comprehensive test of all WebDAV operations
func (c *WebDAVClient) RunComprehensiveTest() error {
	testDir := "/webdav_test_suite"
	testSubDir := testDir + "/subdir"
	testFile1 := testDir + "/test1.txt"
	testFile2 := testSubDir + "/test2.txt"
	tempFile1 := "/tmp/webdav_test1.txt"
	tempFile2 := "/tmp/webdav_test2.txt"
	downloadFile := "/tmp/webdav_downloaded.txt"

	// Create test files
	if err := os.WriteFile(tempFile1, []byte("This is test file 1 for WebDAV test suite."), 0644); err != nil {
		return fmt.Errorf("failed to create test file 1: %w", err)
	}
	if err := os.WriteFile(tempFile2, []byte("This is test file 2 for WebDAV test suite."), 0644); err != nil {
		return fmt.Errorf("failed to create test file 2: %w", err)
	}

	fmt.Println("\n=== WebDAV Comprehensive Test Suite ===")
	fmt.Println("\n1. Initial directory listing")
	fmt.Println("----------------------------")
	if err := c.ListFiles("/"); err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	fmt.Println("\n2. Creating test directories")
	fmt.Println("----------------------------")
	if err := c.CreateDirectory(testDir); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	if err := c.CreateDirectory(testSubDir); err != nil {
		return fmt.Errorf("failed to create test subdirectory: %w", err)
	}

	fmt.Println("\n3. Uploading test files")
	fmt.Println("----------------------------")
	if err := c.UploadFile(tempFile1, testFile1); err != nil {
		return fmt.Errorf("failed to upload test file 1: %w", err)
	}
	if err := c.UploadFile(tempFile2, testFile2); err != nil {
		return fmt.Errorf("failed to upload test file 2: %w", err)
	}

	fmt.Println("\n4. Listing test directory")
	fmt.Println("----------------------------")
	if err := c.ListFiles(testDir); err != nil {
		return fmt.Errorf("failed to list test directory: %w", err)
	}

	fmt.Println("\n5. Listing test subdirectory")
	fmt.Println("----------------------------")
	if err := c.ListFiles(testSubDir); err != nil {
		return fmt.Errorf("failed to list test subdirectory: %w", err)
	}

	fmt.Println("\n6. Downloading a test file")
	fmt.Println("----------------------------")
	if err := c.DownloadFile(testFile1, downloadFile); err != nil {
		return fmt.Errorf("failed to download test file: %w", err)
	}
	
	// Verify downloaded file
	content, err := os.ReadFile(downloadFile)
	if err != nil {
		return fmt.Errorf("failed to read downloaded file: %w", err)
	}
	fmt.Printf("Downloaded file content: %s\n", string(content))

	fmt.Println("\n7. Deleting files")
	fmt.Println("----------------------------")
	if err := c.DeleteFile(testFile1); err != nil {
		return fmt.Errorf("failed to delete test file 1: %w", err)
	}
	if err := c.DeleteFile(testFile2); err != nil {
		return fmt.Errorf("failed to delete test file 2: %w", err)
	}

	fmt.Println("\n8. Listing after file deletion")
	fmt.Println("----------------------------")
	if err := c.ListFiles(testDir); err != nil {
		return fmt.Errorf("failed to list test directory after file deletion: %w", err)
	}

	fmt.Println("\n9. Deleting directories")
	fmt.Println("----------------------------")
	if err := c.DeleteFile(testSubDir); err != nil {
		return fmt.Errorf("failed to delete test subdirectory: %w", err)
	}
	if err := c.DeleteFile(testDir); err != nil {
		return fmt.Errorf("failed to delete test directory: %w", err)
	}

	fmt.Println("\n10. Final directory listing")
	fmt.Println("----------------------------")
	if err := c.ListFiles("/"); err != nil {
		return fmt.Errorf("failed to list files after cleanup: %w", err)
	}

	fmt.Println("\n=== Test Suite Completed Successfully ===")
	return nil
}

func main() {
	url := flag.String("url", "http://localhost:9999", "WebDAV server URL")
	username := flag.String("username", "", "Username for basic authentication")
	password := flag.String("password", "", "Password for basic authentication")
	action := flag.String("action", "test", "Action to perform: test, list, upload, download, mkdir, delete")
	path := flag.String("path", "/", "Path on the WebDAV server")
	localFile := flag.String("local", "", "Local file path for upload/download")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	if *debug {
		log.Printf("Connecting to WebDAV server at %s", *url)
	}

	// Create WebDAV client
	client := NewClient(*url, *username, *password)

	// Create a test file if we're uploading and no local file is specified
	if *action == "upload" && *localFile == "" {
		tempFile := "/tmp/webdav_test_file.txt"
		if err := os.WriteFile(tempFile, []byte("This is a test file for WebDAV upload."), 0644); err != nil {
			log.Fatalf("Failed to create test file: %v", err)
		}
		*localFile = tempFile
		if *debug {
			log.Printf("Created test file at %s", *localFile)
		}
	}

	var err error
	switch *action {
	case "test":
		// Run comprehensive test suite
		err = client.RunComprehensiveTest()
	case "list":
		if *debug {
			log.Printf("Listing files at %s", *path)
		}
		err = client.ListFiles(*path)
	case "upload":
		if *localFile == "" {
			log.Fatalf("Local file path is required for upload")
		}
		if *debug {
			log.Printf("Uploading %s to %s", *localFile, *path)
		}
		remotePath := *path
		if remotePath == "/" {
			remotePath = "/" + filepath.Base(*localFile)
		}
		err = client.UploadFile(*localFile, remotePath)
	case "download":
		if *localFile == "" {
			log.Fatalf("Local file path is required for download")
		}
		if *debug {
			log.Printf("Downloading %s to %s", *path, *localFile)
		}
		err = client.DownloadFile(*path, *localFile)
	case "mkdir":
		if *debug {
			log.Printf("Creating directory %s", *path)
		}
		err = client.CreateDirectory(*path)
	case "delete":
		if *debug {
			log.Printf("Deleting %s", *path)
		}
		err = client.DeleteFile(*path)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
