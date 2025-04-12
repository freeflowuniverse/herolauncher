package dedupestor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func setupTest(t *testing.T) {
	// Ensure test directories exist and are clean
	testDirs := []string{
		"/tmp/dedupestor_test",
		"/tmp/dedupestor_test_size",
		"/tmp/dedupestor_test_exists",
		"/tmp/dedupestor_test_multiple",
		"/tmp/dedupestor_test_refs",
	}

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); err == nil {
			err := os.RemoveAll(dir)
			if err != nil {
				t.Fatalf("Failed to remove test directory %s: %v", dir, err)
			}
		}
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}
}

func TestBasicOperations(t *testing.T) {
	setupTest(t)

	ds, err := New(NewArgs{
		Path:  "/tmp/dedupestor_test",
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()

	// Test storing and retrieving data
	value1 := []byte("test data 1")
	ref1 := Reference{Owner: 1, ID: 1}
	id1, err := ds.Store(value1, ref1)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	retrieved1, err := ds.Get(id1)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	if !bytes.Equal(retrieved1, value1) {
		t.Fatalf("Retrieved data doesn't match stored data")
	}

	// Test deduplication with different reference
	ref2 := Reference{Owner: 1, ID: 2}
	id2, err := ds.Store(value1, ref2)
	if err != nil {
		t.Fatalf("Failed to store data with second reference: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("Expected same ID for duplicate data, got %d and %d", id1, id2)
	}

	// Test different data gets different ID
	value2 := []byte("test data 2")
	ref3 := Reference{Owner: 1, ID: 3}
	id3, err := ds.Store(value2, ref3)
	if err != nil {
		t.Fatalf("Failed to store different data: %v", err)
	}
	if id1 == id3 {
		t.Fatalf("Expected different IDs for different data, got %d for both", id1)
	}

	retrieved2, err := ds.Get(id3)
	if err != nil {
		t.Fatalf("Failed to retrieve second data: %v", err)
	}
	if !bytes.Equal(retrieved2, value2) {
		t.Fatalf("Retrieved data doesn't match second stored data")
	}
}

func TestSizeLimit(t *testing.T) {
	setupTest(t)

	ds, err := New(NewArgs{
		Path:  "/tmp/dedupestor_test_size",
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()

	// Test data under size limit (1KB)
	smallData := make([]byte, 1024)
	for i := range smallData {
		smallData[i] = byte(i % 256)
	}
	ref := Reference{Owner: 1, ID: 1}
	smallID, err := ds.Store(smallData, ref)
	if err != nil {
		t.Fatalf("Failed to store small data: %v", err)
	}

	retrieved, err := ds.Get(smallID)
	if err != nil {
		t.Fatalf("Failed to retrieve small data: %v", err)
	}
	if !bytes.Equal(retrieved, smallData) {
		t.Fatalf("Retrieved data doesn't match stored small data")
	}

	// Test data over size limit (2MB)
	largeData := make([]byte, 2*1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	_, err = ds.Store(largeData, ref)
	if err == nil {
		t.Fatalf("Expected error for data exceeding size limit")
	}
}

func TestExists(t *testing.T) {
	setupTest(t)

	ds, err := New(NewArgs{
		Path:  "/tmp/dedupestor_test_exists",
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()

	value := []byte("test data")
	ref := Reference{Owner: 1, ID: 1}
	id, err := ds.Store(value, ref)
	if err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	if !ds.IDExists(id) {
		t.Fatalf("IDExists returned false for existing ID")
	}
	if ds.IDExists(99) {
		t.Fatalf("IDExists returned true for non-existent ID")
	}

	// Calculate hash to test HashExists
	data, err := ds.Get(id)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}
	hash := sha256Sum(data)

	if !ds.HashExists(hash) {
		t.Fatalf("HashExists returned false for existing hash")
	}
	if ds.HashExists("nonexistenthash") {
		t.Fatalf("HashExists returned true for non-existent hash")
	}
}

func TestMultipleOperations(t *testing.T) {
	setupTest(t)

	ds, err := New(NewArgs{
		Path:  "/tmp/dedupestor_test_multiple",
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()

	// Store multiple values
	values := [][]byte{}
	ids := []uint32{}

	for i := 0; i < 5; i++ {
		value := []byte("test data " + string(rune('0'+i)))
		values = append(values, value)
		ref := Reference{Owner: 1, ID: uint32(i)}
		id, err := ds.Store(value, ref)
		if err != nil {
			t.Fatalf("Failed to store data %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	// Verify all values can be retrieved
	for i, id := range ids {
		retrieved, err := ds.Get(id)
		if err != nil {
			t.Fatalf("Failed to retrieve data %d: %v", i, err)
		}
		if !bytes.Equal(retrieved, values[i]) {
			t.Fatalf("Retrieved data %d doesn't match stored data", i)
		}
	}

	// Test deduplication by storing same values again
	for i, value := range values {
		ref := Reference{Owner: 2, ID: uint32(i)}
		id, err := ds.Store(value, ref)
		if err != nil {
			t.Fatalf("Failed to store duplicate data %d: %v", i, err)
		}
		if id != ids[i] {
			t.Fatalf("Expected same ID for duplicate data %d, got %d and %d", i, ids[i], id)
		}
	}
}

func TestReferences(t *testing.T) {
	setupTest(t)

	ds, err := New(NewArgs{
		Path:  "/tmp/dedupestor_test_refs",
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()

	// Store same data with different references
	value := []byte("test data")
	ref1 := Reference{Owner: 1, ID: 1}
	ref2 := Reference{Owner: 1, ID: 2}
	ref3 := Reference{Owner: 2, ID: 1}

	// Store with first reference
	id, err := ds.Store(value, ref1)
	if err != nil {
		t.Fatalf("Failed to store data with first reference: %v", err)
	}

	// Store same data with second reference
	id2, err := ds.Store(value, ref2)
	if err != nil {
		t.Fatalf("Failed to store data with second reference: %v", err)
	}
	if id != id2 {
		t.Fatalf("Expected same ID for same data, got %d and %d", id, id2)
	}

	// Store same data with third reference
	id3, err := ds.Store(value, ref3)
	if err != nil {
		t.Fatalf("Failed to store data with third reference: %v", err)
	}
	if id != id3 {
		t.Fatalf("Expected same ID for same data, got %d and %d", id, id3)
	}

	// Delete first reference - data should still exist
	err = ds.Delete(id, ref1)
	if err != nil {
		t.Fatalf("Failed to delete first reference: %v", err)
	}
	if !ds.IDExists(id) {
		t.Fatalf("Data should still exist after deleting first reference")
	}

	// Delete second reference - data should still exist
	err = ds.Delete(id, ref2)
	if err != nil {
		t.Fatalf("Failed to delete second reference: %v", err)
	}
	if !ds.IDExists(id) {
		t.Fatalf("Data should still exist after deleting second reference")
	}

	// Delete last reference - data should be gone
	err = ds.Delete(id, ref3)
	if err != nil {
		t.Fatalf("Failed to delete third reference: %v", err)
	}
	if ds.IDExists(id) {
		t.Fatalf("Data should be deleted after removing all references")
	}

	// Verify data is actually deleted by trying to get it
	_, err = ds.Get(id)
	if err == nil {
		t.Fatalf("Expected error getting deleted data")
	}
}

func TestMetadataConversion(t *testing.T) {
	// Test Reference conversion
	ref := Reference{
		Owner: 12345,
		ID:    67890,
	}

	bytes := ref.ToBytes()
	recovered := BytesToReference(bytes)

	if ref.Owner != recovered.Owner || ref.ID != recovered.ID {
		t.Fatalf("Reference conversion failed: original %+v, recovered %+v", ref, recovered)
	}

	// Test Metadata conversion
	metadata := Metadata{
		ID:         42,
		References: []Reference{},
	}

	ref1 := Reference{Owner: 1, ID: 100}
	ref2 := Reference{Owner: 2, ID: 200}

	metadata, err := metadata.AddReference(ref1)
	if err != nil {
		t.Fatalf("Failed to add reference: %v", err)
	}
	metadata, err = metadata.AddReference(ref2)
	if err != nil {
		t.Fatalf("Failed to add reference: %v", err)
	}

	bytes = metadata.ToBytes()
	recovered2 := BytesToMetadata(bytes)

	if metadata.ID != recovered2.ID || len(metadata.References) != len(recovered2.References) {
		t.Fatalf("Metadata conversion failed: original %+v, recovered %+v", metadata, recovered2)
	}

	for i, ref := range metadata.References {
		if ref.Owner != recovered2.References[i].Owner || ref.ID != recovered2.References[i].ID {
			t.Fatalf("Reference in metadata conversion failed at index %d", i)
		}
	}
}

func TestAddRemoveReference(t *testing.T) {
	metadata := Metadata{
		ID:         1,
		References: []Reference{},
	}

	ref1 := Reference{Owner: 1, ID: 100}
	ref2 := Reference{Owner: 2, ID: 200}

	// Add first reference
	metadata, err := metadata.AddReference(ref1)
	if err != nil {
		t.Fatalf("Failed to add first reference: %v", err)
	}
	if len(metadata.References) != 1 {
		t.Fatalf("Expected 1 reference after adding first, got %d", len(metadata.References))
	}
	if metadata.References[0].Owner != ref1.Owner || metadata.References[0].ID != ref1.ID {
		t.Fatalf("First reference not added correctly")
	}

	// Add second reference
	metadata, err = metadata.AddReference(ref2)
	if err != nil {
		t.Fatalf("Failed to add second reference: %v", err)
	}
	if len(metadata.References) != 2 {
		t.Fatalf("Expected 2 references after adding second, got %d", len(metadata.References))
	}

	// Try adding duplicate reference
	metadata, err = metadata.AddReference(ref1)
	if err != nil {
		t.Fatalf("Failed to add duplicate reference: %v", err)
	}
	if len(metadata.References) != 2 {
		t.Fatalf("Expected 2 references after adding duplicate, got %d", len(metadata.References))
	}

	// Remove first reference
	metadata, err = metadata.RemoveReference(ref1)
	if err != nil {
		t.Fatalf("Failed to remove first reference: %v", err)
	}
	if len(metadata.References) != 1 {
		t.Fatalf("Expected 1 reference after removing first, got %d", len(metadata.References))
	}
	if metadata.References[0].Owner != ref2.Owner || metadata.References[0].ID != ref2.ID {
		t.Fatalf("Wrong reference removed")
	}

	// Remove non-existent reference
	metadata, err = metadata.RemoveReference(Reference{Owner: 999, ID: 999})
	if err != nil {
		t.Fatalf("Failed to remove non-existent reference: %v", err)
	}
	if len(metadata.References) != 1 {
		t.Fatalf("Expected 1 reference after removing non-existent, got %d", len(metadata.References))
	}

	// Remove last reference
	metadata, err = metadata.RemoveReference(ref2)
	if err != nil {
		t.Fatalf("Failed to remove last reference: %v", err)
	}
	if len(metadata.References) != 0 {
		t.Fatalf("Expected 0 references after removing last, got %d", len(metadata.References))
	}
}

func TestEmptyMetadataBytes(t *testing.T) {
	empty := BytesToMetadata([]byte{})
	if empty.ID != 0 || len(empty.References) != 0 {
		t.Fatalf("Expected empty metadata, got %+v", empty)
	}
}

func TestDeduplicationSize(t *testing.T) {
	testDir := "/tmp/dedupestor_test_dedup_size"
	
	// Clean up test directory
	if _, err := os.Stat(testDir); err == nil {
		os.RemoveAll(testDir)
	}
	os.MkdirAll(testDir, 0755)
	
	// Create a new dedupe store
	ds, err := New(NewArgs{
		Path:  testDir,
		Reset: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dedupe store: %v", err)
	}
	defer ds.Close()
	
	// Store a large piece of data (100KB)
	largeData := make([]byte, 100*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	// Store the data with first reference
	ref1 := Reference{Owner: 1, ID: 1}
	id1, err := ds.Store(largeData, ref1)
	if err != nil {
		t.Fatalf("Failed to store data with first reference: %v", err)
	}
	
	// Get the size of the data directory after first store
	dataDir := testDir + "/data"
	sizeAfterFirst, err := getDirSize(dataDir)
	if err != nil {
		t.Fatalf("Failed to get directory size: %v", err)
	}
	t.Logf("Size after first store: %d bytes", sizeAfterFirst)
	
	// Store the same data with different references multiple times
	for i := 2; i <= 10; i++ {
		ref := Reference{Owner: uint16(i), ID: uint32(i)}
		id, err := ds.Store(largeData, ref)
		if err != nil {
			t.Fatalf("Failed to store data with reference %d: %v", i, err)
		}
		
		// Verify we get the same ID (deduplication is working)
		if id != id1 {
			t.Fatalf("Expected same ID for duplicate data, got %d and %d", id1, id)
		}
	}
	
	// Get the size after storing the same data multiple times
	sizeAfterMultiple, err := getDirSize(dataDir)
	if err != nil {
		t.Fatalf("Failed to get directory size: %v", err)
	}
	t.Logf("Size after storing same data 10 times: %d bytes", sizeAfterMultiple)
	
	// The size should be approximately the same (allowing for metadata overhead)
	// We'll check that it hasn't grown significantly (less than 10% increase)
	if sizeAfterMultiple > sizeAfterFirst*110/100 {
		t.Fatalf("Directory size grew significantly after storing duplicate data: %d -> %d bytes", 
			sizeAfterFirst, sizeAfterMultiple)
	}
	
	// Now store different data
	differentData := make([]byte, 100*1024)
	for i := range differentData {
		differentData[i] = byte((i + 128) % 256) // Different pattern
	}
	
	ref11 := Reference{Owner: 11, ID: 11}
	_, err = ds.Store(differentData, ref11)
	if err != nil {
		t.Fatalf("Failed to store different data: %v", err)
	}
	
	// Get the size after storing different data
	sizeAfterDifferent, err := getDirSize(dataDir)
	if err != nil {
		t.Fatalf("Failed to get directory size: %v", err)
	}
	t.Logf("Size after storing different data: %d bytes", sizeAfterDifferent)
	
	// The size should have increased significantly
	if sizeAfterDifferent <= sizeAfterMultiple*110/100 {
		t.Fatalf("Directory size didn't grow as expected after storing different data: %d -> %d bytes", 
			sizeAfterMultiple, sizeAfterDifferent)
	}
}

// getDirSize returns the total size of all files in a directory in bytes
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
