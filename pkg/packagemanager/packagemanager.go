package packagemanager

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// PackageManager handles package installation across different platforms
type PackageManager struct {
	platform string
}

// NewPackageManager creates a new package manager instance
func NewPackageManager() *PackageManager {
	return &PackageManager{
		platform: runtime.GOOS,
	}
}

// InstallPackage installs a package using the appropriate package manager
func (pm *PackageManager) InstallPackage(packageName string) (string, error) {
	var cmd *exec.Cmd

	switch pm.platform {
	case "darwin":
		// macOS - use Homebrew
		cmd = exec.Command("brew", "install", packageName)
	case "linux":
		// Linux - use apt
		cmd = exec.Command("apt", "install", "-y", packageName)
	case "windows":
		// Windows - use scoop
		cmd = exec.Command("scoop", "install", packageName)
	default:
		return "", fmt.Errorf("unsupported platform: %s", pm.platform)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// UninstallPackage removes a package
func (pm *PackageManager) UninstallPackage(packageName string) (string, error) {
	var cmd *exec.Cmd

	switch pm.platform {
	case "darwin":
		cmd = exec.Command("brew", "uninstall", packageName)
	case "linux":
		cmd = exec.Command("apt", "remove", "-y", packageName)
	case "windows":
		cmd = exec.Command("scoop", "uninstall", packageName)
	default:
		return "", fmt.Errorf("unsupported platform: %s", pm.platform)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ListInstalledPackages returns a list of installed packages
func (pm *PackageManager) ListInstalledPackages() ([]string, error) {
	var cmd *exec.Cmd

	switch pm.platform {
	case "darwin":
		cmd = exec.Command("brew", "list")
	case "linux":
		cmd = exec.Command("apt", "list", "--installed")
	case "windows":
		cmd = exec.Command("scoop", "list")
	default:
		return nil, fmt.Errorf("unsupported platform: %s", pm.platform)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Parse output into a list of packages
	packages := strings.Split(string(output), "\n")
	var result []string
	for _, pkg := range packages {
		if pkg = strings.TrimSpace(pkg); pkg != "" {
			result = append(result, pkg)
		}
	}

	return result, nil
}

// SearchPackage searches for a package
func (pm *PackageManager) SearchPackage(query string) ([]string, error) {
	var cmd *exec.Cmd

	switch pm.platform {
	case "darwin":
		cmd = exec.Command("brew", "search", query)
	case "linux":
		cmd = exec.Command("apt", "search", query)
	case "windows":
		cmd = exec.Command("scoop", "search", query)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", pm.platform)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// Parse output into a list of packages
	packages := strings.Split(string(output), "\n")
	var result []string
	for _, pkg := range packages {
		if pkg = strings.TrimSpace(pkg); pkg != "" {
			result = append(result, pkg)
		}
	}

	return result, nil
}
