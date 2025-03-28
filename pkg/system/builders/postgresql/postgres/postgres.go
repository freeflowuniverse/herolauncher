package postgres

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

// Constants for PostgreSQL installation
const (
	DefaultPostgresURL   = "https://github.com/postgres/postgres/archive/refs/tags/REL_17_4.tar.gz"
	DefaultPostgresTar   = "postgres.tar.gz"
	DefaultInstallPrefix = "/opt/postgresql"
	DefaultPatchFile     = "src/backend/postmaster/postmaster.c"
)

// PostgresBuilder represents a PostgreSQL builder
type PostgresBuilder struct {
	PostgresURL   string
	PostgresTar   string
	InstallPrefix string
	PatchFile     string
}

// NewPostgresBuilder creates a new PostgreSQL builder with default values
func NewPostgresBuilder() *PostgresBuilder {
	return &PostgresBuilder{
		PostgresURL:   DefaultPostgresURL,
		PostgresTar:   DefaultPostgresTar,
		InstallPrefix: DefaultInstallPrefix,
		PatchFile:     DefaultPatchFile,
	}
}

// WithPostgresURL sets the PostgreSQL download URL
func (b *PostgresBuilder) WithPostgresURL(url string) *PostgresBuilder {
	b.PostgresURL = url
	return b
}

// WithInstallPrefix sets the installation prefix
func (b *PostgresBuilder) WithInstallPrefix(prefix string) *PostgresBuilder {
	b.InstallPrefix = prefix
	return b
}

// run executes a command with the given arguments
func (b *PostgresBuilder) run(cmd string, args ...string) error {
	fmt.Println("Running:", cmd, strings.Join(args, " "))
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// DownloadPostgres downloads the PostgreSQL source code
func (b *PostgresBuilder) DownloadPostgres() error {
	fmt.Println("Downloading PostgreSQL source...")
	out, err := os.Create(b.PostgresTar)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	resp, err := http.Get(b.PostgresURL)
	if err != nil {
		return fmt.Errorf("failed to download PostgreSQL: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

// ExtractTarGz extracts the tar.gz file and returns the top directory
func (b *PostgresBuilder) ExtractTarGz() (string, error) {
	fmt.Println("Extracting...")

	// Create a temporary directory to extract to
	tempDir, err := os.MkdirTemp("", "postgres-extract-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp dir when function returns

	// Extract the archive using archiver
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

	// Move the contents to the current directory
	err = moveContents(topDirPath, ".")
	if err != nil {
		return "", fmt.Errorf("failed to move contents from temp directory: %w", err)
	}

	fmt.Println("Extracted to directory:", topDir)
	return topDir, nil
}

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

// PatchPostmasterC patches the postmaster.c file to allow running as root
func (b *PostgresBuilder) PatchPostmasterC(baseDir string) error {
	fmt.Println("Patching to allow root...")
	file := filepath.Join(baseDir, b.PatchFile)
	input, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	modified := strings.Replace(string(input),
		"if (geteuid() == 0)",
		"if (false /* patched to allow root */)",
		1)

	if err := os.WriteFile(file, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

// BuildPostgres builds PostgreSQL
func (b *PostgresBuilder) BuildPostgres(sourceDir string) error {
	fmt.Println("Building PostgreSQL...")
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(currentDir)

	if err := os.Chdir(sourceDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Add --without-icu to disable ICU dependency
	if err := b.run("/usr/bin/bash", "configure", "--prefix="+b.InstallPrefix, "--without-icu"); err != nil {
		return fmt.Errorf("failed to configure PostgreSQL: %w", err)
	}

	if err := b.run("make", "-j4"); err != nil {
		return fmt.Errorf("failed to build PostgreSQL: %w", err)
	}

	if err := b.run("make", "install"); err != nil {
		return fmt.Errorf("failed to install PostgreSQL: %w", err)
	}

	return nil
}

// CleanInstall cleans the installation directory
func (b *PostgresBuilder) CleanInstall() error {
	fmt.Println("Cleaning install dir...")
	keepDirs := []string{"bin", "lib", "share"}
	entries, err := os.ReadDir(b.InstallPrefix)
	if err != nil {
		return fmt.Errorf("failed to read install directory: %w", err)
	}

	for _, entry := range entries {
		keep := false
		for _, d := range keepDirs {
			if entry.Name() == d {
				keep = true
				break
			}
		}
		if !keep {
			if err := os.RemoveAll(filepath.Join(b.InstallPrefix, entry.Name())); err != nil {
				return fmt.Errorf("failed to remove directory: %w", err)
			}
		}
	}
	return nil
}

// Build builds PostgreSQL
func (b *PostgresBuilder) Build() error {
	if err := b.DownloadPostgres(); err != nil {
		return err
	}

	srcDir, err := b.ExtractTarGz()
	if err != nil {
		return err
	}

	if err := b.PatchPostmasterC(srcDir); err != nil {
		return err
	}

	if err := b.BuildPostgres(srcDir); err != nil {
		return err
	}

	if err := b.CleanInstall(); err != nil {
		return err
	}

	fmt.Println("✅ Done! PostgreSQL installed in:", b.InstallPrefix)
	return nil
}
