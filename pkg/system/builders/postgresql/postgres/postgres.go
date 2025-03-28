package postgres

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	file, err := os.Open(b.PostgresTar)
	if err != nil {
		return "", fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var topDir string
	var firstDir string

	// First pass: find the top directory
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip pax_global_header
		if header.Name == "pax_global_header" {
			continue
		}

		// Find the first real directory
		if header.Typeflag == tar.TypeDir && firstDir == "" {
			firstDir = header.Name
			// Remove trailing slash if present
			firstDir = strings.TrimSuffix(firstDir, "/")
		}

		// If we have a file path with directories
		if strings.Contains(header.Name, "/") {
			parts := strings.SplitN(header.Name, "/", 2)
			if topDir == "" && parts[0] != "" {
				topDir = parts[0]
				break
			}
		}
	}

	// If we didn't find a top directory but found a first directory, use that
	if topDir == "" && firstDir != "" {
		topDir = firstDir
	}

	// Reset the file for the second pass
	file.Seek(0, 0)
	gzr.Close()
	gzr, err = gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()
	tr = tar.NewReader(gzr)

	// Second pass: extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip pax_global_header
		if header.Name == "pax_global_header" {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(header.Name, 0755); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create parent directories if they don't exist
			dir := filepath.Dir(header.Name)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return "", fmt.Errorf("failed to create parent directory: %w", err)
				}
			}

			f, err := os.Create(header.Name)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", fmt.Errorf("failed to write to file: %w", err)
			}
			f.Close()
		}
	}

	if topDir == "" {
		return "", fmt.Errorf("failed to find top directory in archive")
	}

	fmt.Println("Extracted to directory:", topDir)
	return topDir, nil
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
