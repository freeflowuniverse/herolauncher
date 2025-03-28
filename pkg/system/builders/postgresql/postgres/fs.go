package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// moveContents moves all contents from src directory to dst directory
func moveContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Handle existing destination
		if _, err := os.Stat(dstPath); err == nil {
			// If it exists, remove it first
			if err := os.RemoveAll(dstPath); err != nil {
				return fmt.Errorf("failed to remove existing path %s: %w", dstPath, err)
			}
		}

		// Move the file or directory
		if err := os.Rename(srcPath, dstPath); err != nil {
			// If rename fails (possibly due to cross-device link), try copy and delete
			if strings.Contains(err.Error(), "cross-device link") {
				if entry.IsDir() {
					if err := copyDir(srcPath, dstPath); err != nil {
						return err
					}
				} else {
					if err := copyFile(srcPath, dstPath); err != nil {
						return err
					}
				}
				os.RemoveAll(srcPath)
			} else {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
