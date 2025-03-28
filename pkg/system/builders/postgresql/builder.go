package postgresql

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/dependencies"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/gosp"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/postgres"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/verification"
)

// Constants for PostgreSQL installation
const (
	DefaultInstallPrefix = "/opt/postgresql"
)

// Builder represents a PostgreSQL builder
type Builder struct {
	InstallPrefix     string
	PostgresBuilder   *postgres.PostgresBuilder
	GoSPBuilder       *gosp.GoSPBuilder
	DependencyManager *dependencies.DependencyManager
	Verifier          *verification.Verifier
}

// NewBuilder creates a new PostgreSQL builder with default values
func NewBuilder() *Builder {
	installPrefix := DefaultInstallPrefix

	return &Builder{
		InstallPrefix:     installPrefix,
		PostgresBuilder:   postgres.NewPostgresBuilder().WithInstallPrefix(installPrefix),
		GoSPBuilder:       gosp.NewGoSPBuilder(installPrefix),
		DependencyManager: dependencies.NewDependencyManager("bison", "flex", "libreadline-dev"),
		Verifier:          verification.NewVerifier(installPrefix),
	}
}

// WithInstallPrefix sets the installation prefix
func (b *Builder) WithInstallPrefix(prefix string) *Builder {
	b.InstallPrefix = prefix
	b.PostgresBuilder.WithInstallPrefix(prefix)
	b.GoSPBuilder = gosp.NewGoSPBuilder(prefix)
	return b
}

// WithPostgresURL sets the PostgreSQL download URL
// RunPostgresInScreen starts PostgreSQL in a screen session
func (b *Builder) RunPostgresInScreen() error {
	return b.PostgresBuilder.RunPostgresInScreen()
}

// CheckPostgresUser checks if PostgreSQL can be run as postgres user
func (b *Builder) CheckPostgresUser() error {
	return b.PostgresBuilder.CheckPostgresUser()
}

func (b *Builder) WithPostgresURL(url string) *Builder {
	b.PostgresBuilder.WithPostgresURL(url)
	return b
}

// WithDependencies sets the dependencies to install
func (b *Builder) WithDependencies(deps ...string) *Builder {
	b.DependencyManager.WithDependencies(deps...)
	return b
}

// Build builds PostgreSQL
func (b *Builder) Build() error {
	fmt.Println("=== Starting PostgreSQL Build ===")

	// Install dependencies
	fmt.Println("Installing dependencies...")
	if err := b.DependencyManager.Install(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Build PostgreSQL
	if err := b.PostgresBuilder.Build(); err != nil {
		return fmt.Errorf("failed to build PostgreSQL: %w", err)
	}

	// Ensure Go is installed first to get its path
	goInstaller := postgres.NewGoInstaller()
	goPath, err := goInstaller.InstallGo()
	if err != nil {
		return fmt.Errorf("failed to ensure Go is installed: %w", err)
	}
	fmt.Printf("Using Go executable from: %s\n", goPath)
	
	// Pass the Go path explicitly to the GoSPBuilder
	b.GoSPBuilder.WithGoPath(goPath)
	
	// For the Go stored procedure, we'll create and execute a shell script directly
	// to ensure all environment variables are properly set
	fmt.Println("Building Go stored procedure via shell script...")
	
	tempDir, err := os.MkdirTemp("", "gosp-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create the Go source file in the temp directory
	libPath := filepath.Join(tempDir, "gosp.go")
	libSrc := `
package main
import "C"
import "fmt"

//export helloworld
func helloworld() {
	fmt.Println("Hello from Go stored procedure!")
}

func main() {}
`
	if err := os.WriteFile(libPath, []byte(libSrc), 0644); err != nil {
		return fmt.Errorf("failed to write Go source file: %w", err)
	}
	
	// Create a shell script to build the Go stored procedure
	buildScript := filepath.Join(tempDir, "build.sh")
	buildScriptContent := fmt.Sprintf(`#!/bin/sh
set -e

# Set environment variables
export GOROOT=/usr/local/go
export GOPATH=/root/go
export PATH=/usr/local/go/bin:$PATH

echo "Current directory: $(pwd)"
echo "Go source file: %s"
echo "Output file: %s/lib/libgosp.so"

# Create output directory
mkdir -p %s/lib

# Run the build command
echo "Running: go build -buildmode=c-shared -o %s/lib/libgosp.so %s"
go build -buildmode=c-shared -o %s/lib/libgosp.so %s

echo "Go stored procedure built successfully!"
`,
		libPath, b.InstallPrefix, b.InstallPrefix, b.InstallPrefix, libPath, b.InstallPrefix, libPath)
	
	if err := os.WriteFile(buildScript, []byte(buildScriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write build script: %w", err)
	}
	
	// Execute the build script
	cmd := exec.Command("/bin/sh", buildScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("Executing build script:", buildScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run build script: %w", err)
	}

	// Verify the installation
	fmt.Println("Verifying installation...")
	success, err := b.Verifier.Verify()
	if err != nil {
		fmt.Printf("Warning: Verification had issues: %v\n", err)
	}

	if success {
		fmt.Println("✅ Done! PostgreSQL installed and verified in:", b.InstallPrefix)
	} else {
		fmt.Println("⚠️ Done with warnings! PostgreSQL installed in:", b.InstallPrefix)
	}

	return nil
}
