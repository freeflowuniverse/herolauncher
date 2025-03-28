package postgres

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadPostgres downloads the PostgreSQL source code if it doesn't already exist
func (b *PostgresBuilder) DownloadPostgres() error {
	// Check if the file already exists
	if _, err := os.Stat(b.PostgresTar); err == nil {
		fmt.Printf("PostgreSQL source already downloaded at %s, skipping download\n", b.PostgresTar)
		return nil
	}

	fmt.Println("Downloading PostgreSQL source...")
	return downloadFile(b.PostgresURL, b.PostgresTar)
}

// downloadFile downloads a file from url to destination path
func downloadFile(url, dst string) error {
	// Create the file
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s when downloading %s", resp.Status, url)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", dst, err)
	}

	return nil
}
