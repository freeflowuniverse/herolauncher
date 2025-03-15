package ourdb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBasicOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ourdb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new database
	config := DefaultConfig()
	config.Path = tempDir
	
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
	
	// Test data
	testData := []byte("Hello, OurDB!")
	
	// Store data with auto-generated ID
	id, err := db.Set(OurDBSetArgs{
		Data: testData,
	})
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}
	
	// Retrieve data
	retrievedData, err := db.Get(id)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	
	// Verify data
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data doesn't match original: got %s, want %s", 
			string(retrievedData), string(testData))
	}
	
	// Test client interface with incremental mode (default)
	clientTest(t, tempDir, true)
	
	// Test client interface with incremental mode disabled
	clientTest(t, filepath.Join(tempDir, "non_incremental"), false)
}

func clientTest(t *testing.T, dbPath string, incremental bool) {
	// Create a new client with specified incremental mode
	clientPath := filepath.Join(dbPath, "client_test")
	config := DefaultConfig()
	config.IncrementalMode = incremental
	client, err := NewClientWithConfig(clientPath, config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	
	testData := []byte("Client Test Data")
	var id uint32
	
	if incremental {
		// In incremental mode, add data with auto-generated ID
		var err error
		id, err = client.Add(testData)
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
	} else {
		// In non-incremental mode, set data with specific ID
		id = 1
		err = client.Set(id, testData)
		if err != nil {
			t.Fatalf("Failed to set data with ID %d: %v", id, err)
		}
	}
	
	// Retrieve data
	retrievedData, err := client.Get(id)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	
	// Verify data
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved client data doesn't match original: got %s, want %s", 
			string(retrievedData), string(testData))
	}
	
	// Test setting data with specific ID (only if incremental mode is disabled)
	if !incremental {
		specificID := uint32(100)
		specificData := []byte("Specific ID Data")
		err = client.Set(specificID, specificData)
		if err != nil {
			t.Fatalf("Failed to set data with specific ID: %v", err)
		}
		
		// Retrieve and verify specific ID data
		retrievedSpecific, err := client.Get(specificID)
		if err != nil {
			t.Fatalf("Failed to retrieve specific ID data: %v", err)
		}
		
		if string(retrievedSpecific) != string(specificData) {
			t.Errorf("Retrieved specific ID data doesn't match: got %s, want %s", 
				string(retrievedSpecific), string(specificData))
		}
	} else {
		// In incremental mode, test that setting a specific ID fails as expected
		specificID := uint32(100)
		specificData := []byte("Specific ID Data")
		err = client.Set(specificID, specificData)
		if err == nil {
			t.Errorf("Setting specific ID in incremental mode should fail but succeeded")
		}
	}
}
