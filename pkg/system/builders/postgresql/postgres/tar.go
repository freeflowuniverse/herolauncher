package postgres

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
)

// ExtractTarGz extracts the tar.gz file and returns the top directory
func (b *PostgresBuilder) ExtractTarGz() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Check if sources are already extracted
	srcDir := filepath.Join(cwd, "src")
	if _, err := os.Stat(srcDir); err == nil {
		fmt.Println("PostgreSQL source already extracted, skipping extraction")
		return cwd, nil
	}

	fmt.Println("Extracting...")
	fmt.Println("Current working directory:", cwd)

	// Check if the archive exists
	if _, err := os.Stat(b.PostgresTar); os.IsNotExist(err) {
		return "", fmt.Errorf("archive file %s does not exist", b.PostgresTar)
	}
	fmt.Println("Archive exists at:", b.PostgresTar)

	// Create a temporary directory to extract to
	tempDir, err := os.MkdirTemp("", "postgres-extract-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	fmt.Println("Created temp directory:", tempDir)
	defer os.RemoveAll(tempDir) // Clean up temp dir when function returns

	// Extract the archive using archiver
	fmt.Println("Extracting archive to:", tempDir)
	err = archiver.Unarchive(b.PostgresTar, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	// Find the top-level directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to read temp directory: %w", err)
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no files found in extracted archive")
	}

	// In most cases, a properly packaged tarball will extract to a single top directory
	topDir := entries[0].Name()
	topDirPath := filepath.Join(tempDir, topDir)
	fmt.Println("Top directory path:", topDirPath)

	// Verify the top directory exists
	if info, err := os.Stat(topDirPath); err != nil {
		return "", fmt.Errorf("top directory not found: %w", err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("top path is not a directory: %s", topDirPath)
	}

	// Create absolute path for the destination
	dstDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	fmt.Println("Destination directory (absolute):", dstDir)

	// Move the contents to the current directory
	fmt.Println("Moving contents from:", topDirPath, "to:", dstDir)
	err = moveContents(topDirPath, dstDir)
	if err != nil {
		return "", fmt.Errorf("failed to move contents from temp directory: %w", err)
	}

	fmt.Println("Extraction complete")
	return dstDir, nil
}
