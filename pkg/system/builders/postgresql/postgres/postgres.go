package postgres

import (
	"fmt"
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

// PatchPostmasterC patches the postmaster.c file to allow running as root
func (b *PostgresBuilder) PatchPostmasterC(baseDir string) error {
	fmt.Println("Patching to allow root...")

	// Look for the postmaster.c file in the expected location
	file := filepath.Join(baseDir, b.PatchFile)

	// If the file doesn't exist, try to find it
	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("File not found in the expected location, searching for it...")

		// Search for postmaster.c
		var postmasterPath string
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "postmaster.c" {
				postmasterPath = path
				return filepath.SkipAll
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to search for postmaster.c: %w", err)
		}

		if postmasterPath == "" {
			return fmt.Errorf("could not find postmaster.c in the extracted directory")
		}

		fmt.Printf("Found postmaster.c at: %s\n", postmasterPath)
		file = postmasterPath
	}

	// Read the file
	input, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Patch the file
	modified := strings.Replace(string(input),
		"if (geteuid() == 0)",
		"if (false /* patched to allow root */)",
		1)

	if err := os.WriteFile(file, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Println("Successfully patched postmaster.c")
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
	// Check if PostgreSQL is already installed
	binPath := filepath.Join(b.InstallPrefix, "bin", "postgres")
	if _, err := os.Stat(binPath); err == nil {
		fmt.Printf("✅ PostgreSQL already installed at %s, skipping build\n", b.InstallPrefix)
		return nil
	}

	// Check if install directory exists but is incomplete/corrupt
	if _, err := os.Stat(b.InstallPrefix); err == nil {
		fmt.Printf("Found incomplete installation at %s, removing it to start fresh\n", b.InstallPrefix)
		if err := os.RemoveAll(b.InstallPrefix); err != nil {
			return fmt.Errorf("failed to clean incomplete installation: %w", err)
		}
	}

	// Download PostgreSQL source
	if err := b.DownloadPostgres(); err != nil {
		return err
	}

	// Extract the source code
	srcDir, err := b.ExtractTarGz()
	if err != nil {
		return err
	}

	// Patch to allow running as root
	if err := b.PatchPostmasterC(srcDir); err != nil {
		return err
	}

	// Build PostgreSQL
	if err := b.BuildPostgres(srcDir); err != nil {
		// Clean up on build failure
		fmt.Printf("Build failed, cleaning up installation directory %s\n", b.InstallPrefix)
		cleanErr := os.RemoveAll(b.InstallPrefix)
		if cleanErr != nil {
			fmt.Printf("Warning: Failed to clean up installation directory: %v\n", cleanErr)
		}
		return err
	}

	// Final cleanup
	if err := b.CleanInstall(); err != nil {
		return err
	}

	fmt.Println("✅ Done! PostgreSQL installed in:", b.InstallPrefix)
	return nil
}
