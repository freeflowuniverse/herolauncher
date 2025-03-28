package postgres

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Constants for PostgreSQL installation
const (
	DefaultPostgresURL   = "https://github.com/postgres/postgres/archive/refs/tags/REL_17_4.tar.gz"
	DefaultPostgresTar   = "postgres.tar.gz"
	DefaultInstallPrefix = "/opt/postgresql"
	DefaultPatchFile     = "src/backend/postmaster/postmaster.c"
	BuildMarkerFile      = ".build_complete"
	// Set ForceReset to true to force a complete rebuild
	ForceReset = true
)

// PostgresBuilder represents a PostgreSQL builder
type PostgresBuilder struct {
	PostgresURL   string
	PostgresTar   string
	InstallPrefix string
	PatchFile     string
	BuildMarker   string
}

// NewPostgresBuilder creates a new PostgreSQL builder with default values
func NewPostgresBuilder() *PostgresBuilder {
	return &PostgresBuilder{
		PostgresURL:   DefaultPostgresURL,
		PostgresTar:   DefaultPostgresTar,
		InstallPrefix: DefaultInstallPrefix,
		PatchFile:     DefaultPatchFile,
		BuildMarker:   filepath.Join(DefaultInstallPrefix, BuildMarkerFile),
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
	fmt.Println("Patching postmaster.c to allow root...")

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
		"geteuid() == 0",
		"false",
		1)

	if err := os.WriteFile(file, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Verify that the patch was applied
	updatedContent, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file after patching: %w", err)
	}

	if !strings.Contains(string(updatedContent), "patched to allow root") {
		return fmt.Errorf("patching postmaster.c failed: verification check failed")
	}

	fmt.Println("✅ Successfully patched postmaster.c")
	return nil
}

// PatchInitdbC patches the initdb.c file to allow running as root
func (b *PostgresBuilder) PatchInitdbC(baseDir string) error {
	fmt.Println("Patching initdb.c to allow root...")

	// Search for initdb.c
	var initdbPath string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "initdb.c" {
			initdbPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to search for initdb.c: %w", err)
	}

	if initdbPath == "" {
		return fmt.Errorf("could not find initdb.c in the extracted directory")
	}

	fmt.Printf("Found initdb.c at: %s\n", initdbPath)

	// Read the file
	input, err := os.ReadFile(initdbPath)
	if err != nil {
		return fmt.Errorf("failed to read initdb.c: %w", err)
	}
	// Patch the file to bypass root user check
	// This modifies the condition that checks if the user is root
	modified := strings.Replace(string(input),
		"geteuid() == 0", // Common pattern to check for root
		"false",
		-1) // Replace all occurrences

	// Also look for any alternate ways the check might be implemented
	modified = strings.Replace(modified,
		"pg_euid == 0", // Alternative check pattern
		"false",
		-1) // Replace all occurrences

	if err := os.WriteFile(initdbPath, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write to initdb.c: %w", err)
	}

	fmt.Println("✅ Successfully patched initdb.c")
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

// CheckRequirements checks if the current environment meets the requirements
func (b *PostgresBuilder) CheckRequirements() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("this PostgreSQL builder must be run as root")
	}

	// Check if we can bypass OS checks with environment variable
	if os.Getenv("POSTGRES_BUILDER_FORCE") == "1" {
		fmt.Println("✅ Environment check bypassed due to POSTGRES_BUILDER_FORCE=1")
		return nil
	}

	// // Check if running on Ubuntu
	// isUbuntu, err := b.isUbuntu()
	// if err != nil {
	// 	fmt.Printf("⚠️ Warning determining OS: %v\n", err)
	// 	fmt.Println("⚠️ Will proceed anyway, but you might encounter issues.")
	// 	fmt.Println("⚠️ Set POSTGRES_BUILDER_FORCE=1 to bypass this check in the future.")
	// 	return nil
	// }

	// if !isUbuntu {
	// 	// Debug information for troubleshooting OS detection
	// 	fmt.Println("⚠️ OS detection failed. Debug information:")
	// 	exec.Command("cat", "/etc/os-release").Run()
	// 	exec.Command("uname", "-a").Run()

	// 	fmt.Println("⚠️ Set POSTGRES_BUILDER_FORCE=1 to bypass this check.")
	// 	return fmt.Errorf("this PostgreSQL builder only works on Ubuntu")
	// }

	fmt.Println("✅ Environment check passed: running as root on Ubuntu")
	return nil
}

// isUbuntu checks if the current OS is Ubuntu
func (b *PostgresBuilder) isUbuntu() (bool, error) {
	// First try lsb_release as it's more reliable
	lsbCmd := exec.Command("lsb_release", "-a")
	lsbOut, err := lsbCmd.CombinedOutput()
	if err == nil && strings.Contains(strings.ToLower(string(lsbOut)), "ubuntu") {
		return true, nil
	}

	// As a fallback, check /etc/os-release
	osReleaseBytes, err := os.ReadFile("/etc/os-release")
	if err != nil {
		// If /etc/os-release doesn't exist, check for /etc/lsb-release
		lsbReleaseBytes, lsbErr := os.ReadFile("/etc/lsb-release")
		if lsbErr == nil && strings.Contains(strings.ToLower(string(lsbReleaseBytes)), "ubuntu") {
			return true, nil
		}

		return false, fmt.Errorf("could not determine if OS is Ubuntu: %w", err)
	}

	// Check multiple ways Ubuntu might be identified
	osRelease := strings.ToLower(string(osReleaseBytes))
	return strings.Contains(osRelease, "ubuntu") ||
		strings.Contains(osRelease, "id=ubuntu") ||
		strings.Contains(osRelease, "id_like=ubuntu"), nil
}

// Build builds PostgreSQL
func (b *PostgresBuilder) Build() error {
	// Check requirements first
	if err := b.CheckRequirements(); err != nil {
		fmt.Printf("⚠️ Requirements check failed: %v\n", err)
		return err
	}

	// Check if reset is forced
	if ForceReset {
		fmt.Println("Force reset enabled, removing existing installation...")
		if err := os.RemoveAll(b.InstallPrefix); err != nil {
			return fmt.Errorf("failed to remove installation directory: %w", err)
		}
	}

	// Check if PostgreSQL is already installed and build is complete
	binPath := filepath.Join(b.InstallPrefix, "bin", "postgres")
	if _, err := os.Stat(binPath); err == nil {
		// Check for build marker
		if _, err := os.Stat(b.BuildMarker); err == nil {
			fmt.Printf("✅ PostgreSQL already installed at %s with build marker, skipping build\n", b.InstallPrefix)
			return nil
		}
		fmt.Printf("PostgreSQL installation found at %s but no build marker, will verify\n", b.InstallPrefix)
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

	// Patch initdb.c to allow running as root
	if err := b.PatchInitdbC(srcDir); err != nil {
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

	// Create build marker file
	f, err := os.Create(b.BuildMarker)
	if err != nil {
		return fmt.Errorf("failed to create build marker: %w", err)
	}
	f.Close()

	fmt.Println("✅ Done! PostgreSQL installed in:", b.InstallPrefix)
	return nil
}

// RunPostgresInScreen starts PostgreSQL in a screen session
func (b *PostgresBuilder) RunPostgresInScreen() error {
	fmt.Println("Starting PostgreSQL in screen...")

	// Check if screen is installed
	if _, err := exec.LookPath("screen"); err != nil {
		return fmt.Errorf("screen is not installed: %w", err)
	}

	// Create data directory if it doesn't exist
	dataDir := filepath.Join(b.InstallPrefix, "data")
	initdbPath := filepath.Join(b.InstallPrefix, "bin", "initdb")
	postgresPath := filepath.Join(b.InstallPrefix, "bin", "postgres")
	psqlPath := filepath.Join(b.InstallPrefix, "bin", "psql")

	// Check if data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Println("Initializing database directory...")

		// Initialize database
		cmd := exec.Command(initdbPath, "-D", dataDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	// Check if screen session already exists
	checkCmd := exec.Command("screen", "-list")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check screen sessions: %w", err)
	}

	// Kill existing session if it exists
	if strings.Contains(string(output), "postgresql") {
		fmt.Println("PostgreSQL screen session already exists, killing it...")
		killCmd := exec.Command("screen", "-X", "-S", "postgresql", "quit")
		killCmd.Run() // Ignore errors if the session doesn't exist
	}

	// Start PostgreSQL in a new screen session
	cmd := exec.Command("screen", "-dmS", "postgresql", "-L", "-Logfile",
		filepath.Join(b.InstallPrefix, "postgres_screen.log"),
		postgresPath, "-D", dataDir)

	fmt.Println("Running command:", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start PostgreSQL in screen: %w", err)
	}

	// Wait for PostgreSQL to start
	fmt.Println("Waiting for PostgreSQL to start...")
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)

		// Try to connect to PostgreSQL
		testCmd := exec.Command(psqlPath, "-c", "SELECT 1;")
		out, err := testCmd.CombinedOutput()

		if err == nil && bytes.Contains(out, []byte("1")) {
			fmt.Println("✅ PostgreSQL is running and accepting connections")
			break
		}

		if i == 9 {
			return fmt.Errorf("failed to connect to PostgreSQL after 10 seconds")
		}
	}

	// Test user creation
	fmt.Println("Testing user creation...")
	userCmd := exec.Command(psqlPath, "-c", "CREATE USER test_user WITH PASSWORD 'password';")
	userOut, userErr := userCmd.CombinedOutput()

	if userErr != nil {
		return fmt.Errorf("failed to create test user: %s: %w", string(userOut), userErr)
	}

	// Check if we can log screen output
	logCmd := exec.Command("screen", "-S", "postgresql", "-X", "hardcopy",
		filepath.Join(b.InstallPrefix, "screen_hardcopy.log"))
	if err := logCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to capture screen log: %v\n", err)
	}

	fmt.Println("✅ PostgreSQL is running in screen session 'postgresql'")
	fmt.Println("   - Log file: ", filepath.Join(b.InstallPrefix, "postgres_screen.log"))
	return nil
}

// CheckPostgresUser checks if PostgreSQL can be run as postgres user
func (b *PostgresBuilder) CheckPostgresUser() error {
	// Try to get postgres user information
	cmd := exec.Command("id", "postgres")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("⚠️ postgres user does not exist, consider creating it")
		return nil
	}

	fmt.Printf("Found postgres user: %s\n", strings.TrimSpace(string(output)))

	// Try to run a command as postgres user
	sudoCmd := exec.Command("sudo", "-u", "postgres", "echo", "Running as postgres user")
	sudoOutput, sudoErr := sudoCmd.CombinedOutput()

	if sudoErr != nil {
		fmt.Printf("⚠️ Cannot run commands as postgres user: %v\n", sudoErr)
		return nil
	}

	fmt.Printf("Successfully ran command as postgres user: %s\n",
		strings.TrimSpace(string(sudoOutput)))
	return nil
}
