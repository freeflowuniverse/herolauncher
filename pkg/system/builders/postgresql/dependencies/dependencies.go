package dependencies

import (
	"fmt"
	"os/exec"
	"strings"
)

// DependencyManager handles the installation of dependencies
type DependencyManager struct {
	Dependencies []string
}

// NewDependencyManager creates a new dependency manager
func NewDependencyManager(dependencies ...string) *DependencyManager {
	return &DependencyManager{
		Dependencies: dependencies,
	}
}

// WithDependencies sets the dependencies to install
func (d *DependencyManager) WithDependencies(dependencies ...string) *DependencyManager {
	d.Dependencies = dependencies
	return d
}

// Install installs the dependencies
func (d *DependencyManager) Install() error {
	if len(d.Dependencies) == 0 {
		fmt.Println("No dependencies to install")
		return nil
	}

	fmt.Printf("Installing dependencies: %s\n", strings.Join(d.Dependencies, ", "))

	// Update package lists
	updateCmd := exec.Command("apt-get", "update")
	updateCmd.Stdout = nil
	updateCmd.Stderr = nil
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("failed to update package lists: %w", err)
	}

	// Install dependencies
	args := append([]string{"install", "-y"}, d.Dependencies...)
	installCmd := exec.Command("apt-get", args...)
	installCmd.Stdout = nil
	installCmd.Stderr = nil
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Println("✅ Dependencies installed successfully")
	return nil
}
