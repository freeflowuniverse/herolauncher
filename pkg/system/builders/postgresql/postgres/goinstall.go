package postgres

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mholt/archiver/v3"
)

const (
	// DefaultGoVersion is the default Go version to install
	DefaultGoVersion = "1.22.2"
)

// GoInstaller handles Go installation checks and installation
type GoInstaller struct {
	Version string
}

// NewGoInstaller creates a new Go installer with the default version
func NewGoInstaller() *GoInstaller {
	return &GoInstaller{
		Version: DefaultGoVersion,
	}
}

// WithVersion sets the Go version to install
func (g *GoInstaller) WithVersion(version string) *GoInstaller {
	g.Version = version
	return g
}

// IsGoInstalled checks if Go is installed and available
func (g *GoInstaller) IsGoInstalled() bool {
	// Check if go command is available
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetGoVersion gets the installed Go version
func (g *GoInstaller) GetGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Go version: %w", err)
	}
	
	// Parse go version output (format: "go version go1.x.x ...")
	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected go version output format: %s", version)
	}
	
	// Return just the version number without the "go" prefix
	return strings.TrimPrefix(parts[2], "go"), nil
}

// InstallGo installs Go if it's not already installed and returns the path to the Go executable
func (g *GoInstaller) InstallGo() (string, error) {
	// First check if Go is available in PATH
	if path, err := exec.LookPath("go"); err == nil {
		// Test if it works
		cmd := exec.Command(path, "version")
		if output, err := cmd.Output(); err == nil {
			fmt.Printf("Found working Go in PATH: %s, version: %s\n", path, strings.TrimSpace(string(output)))
			return path, nil
		}
	}
	// Default Go installation location
	installDir := "/usr/local"
	goExePath := filepath.Join(installDir, "go", "bin", "go")
	
	// Check if Go is already installed by checking the binary directly
	if _, err := os.Stat(goExePath); err == nil {
		version, err := g.GetGoVersion()
		if err == nil {
			fmt.Printf("Go is already installed (version %s), skipping installation\n", version)
			return goExePath, nil
		}
	}
	
	// Also check if Go is available in PATH as a fallback
	if g.IsGoInstalled() {
		path, err := exec.LookPath("go")
		if err == nil {
			version, err := g.GetGoVersion()
			if err == nil {
				fmt.Printf("Go is already installed (version %s) at %s, skipping installation\n", version, path)
				return path, nil
			}
		}
	}
	
	fmt.Printf("Installing Go version %s...\n", g.Version)
	
	// Determine architecture and OS
	goOS := runtime.GOOS
	goArch := runtime.GOARCH
	
	// Construct download URL
	downloadURL := fmt.Sprintf("https://golang.org/dl/go%s.%s-%s.tar.gz", g.Version, goOS, goArch)
	
	// Create a temporary directory for download
	tempDir, err := os.MkdirTemp("", "go-install-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Download Go tarball
	tarballPath := filepath.Join(tempDir, "go.tar.gz")
	if err := downloadFile(downloadURL, tarballPath); err != nil {
		return "", fmt.Errorf("failed to download Go: %w", err)
	}
	
	// Install directory - typically /usr/local for Linux/macOS
	
	// Check if existing Go installation exists and remove it
	existingGoDir := filepath.Join(installDir, "go")
	if _, err := os.Stat(existingGoDir); err == nil {
		fmt.Printf("Removing existing Go installation at %s\n", existingGoDir)
		if err := os.RemoveAll(existingGoDir); err != nil {
			return "", fmt.Errorf("failed to remove existing Go installation: %w", err)
		}
	}
	
	// Extract tarball to install directory
	fmt.Printf("Extracting Go to %s\n", installDir)
	if err := extractTarGz(tarballPath, installDir); err != nil {
		return "", fmt.Errorf("failed to extract Go tarball: %w", err)
	}
	
	// Verify installation
	goExePath := filepath.Join(installDir, "go", "bin", "go")
	if _, err = os.Stat(goExePath); os.IsNotExist(err) {
		return "", fmt.Errorf("Go installation failed - go executable not found at %s", goExePath)
	}
	
	// Set up environment variables
	fmt.Println("Setting up Go environment variables...")
	
	// Update PATH in /etc/profile
	profilePath := "/etc/profile"
	profileContent, err := os.ReadFile(profilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read profile: %w", err)
	}
	
	// Add Go bin to PATH if not already there
	goBinPath := filepath.Join(installDir, "go", "bin")
	if !strings.Contains(string(profileContent), goBinPath) {
		newContent := string(profileContent) + fmt.Sprintf("\n# Added by PostgreSQL builder\nexport PATH=$PATH:%s\n", goBinPath)
		if err := os.WriteFile(profilePath, []byte(newContent), 0644); err != nil {
			return "", fmt.Errorf("failed to update profile: %w", err)
		}
	}
	
	fmt.Printf("✅ Go %s installed successfully!\n", g.Version)
	return goExePath, nil
}

// Helper function to extract tarball
func extractTarGz(src, dst string) error {
	return archiver.Unarchive(src, dst)
}
