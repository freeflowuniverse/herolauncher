package gosp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/postgres"
)

// Constants for Go stored procedure
const (
	DefaultGoSharedLibDir = "go_sp"
)

// GoSPBuilder represents a Go stored procedure builder
type GoSPBuilder struct {
	GoSharedLibDir string
	InstallPrefix  string
	GoPath         string // Path to Go executable
}

// NewGoSPBuilder creates a new Go stored procedure builder
func NewGoSPBuilder(installPrefix string) *GoSPBuilder {
	return &GoSPBuilder{
		GoSharedLibDir: DefaultGoSharedLibDir,
		InstallPrefix:  installPrefix,
	}
}

// WithGoSharedLibDir sets the Go shared library directory
func (b *GoSPBuilder) WithGoSharedLibDir(dir string) *GoSPBuilder {
	b.GoSharedLibDir = dir
	return b
}

// WithGoPath sets the path to the Go executable
func (b *GoSPBuilder) WithGoPath(path string) *GoSPBuilder {
	b.GoPath = path
	return b
}

// run executes a command with the given arguments and environment variables
func (b *GoSPBuilder) run(cmd string, args ...string) error {
	fmt.Println("Running:", cmd, args)
	c := exec.Command(cmd, args...)
	// Set environment variables
	c.Env = append(os.Environ(), 
		"GOROOT=/usr/local/go",
		"GOPATH=/root/go", 
		"PATH=/usr/local/go/bin:" + os.Getenv("PATH"))
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// Build builds a Go stored procedure
func (b *GoSPBuilder) Build() error {
	fmt.Println("Building Go stored procedure...")
	
	// Use the explicitly provided Go path if available
	var goExePath string
	if b.GoPath != "" {
		goExePath = b.GoPath
		fmt.Printf("Using explicitly provided Go executable: %s\n", goExePath)
	} else {
		// Fallback to ensuring Go is installed via the installer
		goInstaller := postgres.NewGoInstaller()
		var err error
		goExePath, err = goInstaller.InstallGo()
		if err != nil {
			return fmt.Errorf("failed to ensure Go is installed: %w", err)
		}
		fmt.Printf("Using detected Go executable from: %s\n", goExePath)
	}
	
	if err := os.MkdirAll(b.GoSharedLibDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	libPath := filepath.Join(b.GoSharedLibDir, "gosp.go")
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
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Use the full path to Go rather than relying on PATH
	fmt.Println("Running Go build with full path:", goExePath)
	
	// Show debug information
	fmt.Println("Environment variables that will be set:")
	fmt.Println("  GOROOT=/usr/local/go")
	fmt.Println("  GOPATH=/root/go")
	fmt.Println("  PATH=/usr/local/go/bin:" + os.Getenv("PATH"))
	
	// Verify that the Go executable exists before using it
	if _, err := os.Stat(goExePath); err != nil {
		return fmt.Errorf("Go executable not found at %s: %w", goExePath, err)
	}
	
	// Create the output directory if it doesn't exist
	outputDir := filepath.Join(b.InstallPrefix, "lib")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}
	
	// Prepare output path
	outputPath := filepath.Join(outputDir, "libgosp.so")
	
	// Instead of relying on environment variables, create a wrapper shell script
	// that sets all required environment variables and then calls the Go executable
	tempDir, err := os.MkdirTemp("", "go-build-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up when done
	
	goRoot := filepath.Dir(filepath.Dir(goExePath)) // /usr/local/go
	wrapperScript := filepath.Join(tempDir, "go-wrapper.sh")
	wrapperContent := fmt.Sprintf(`#!/bin/sh
# Go wrapper script created by GoSPBuilder
export GOROOT=%s
export GOPATH=/root/go
export PATH=%s:$PATH

echo "=== Go environment variables ==="
echo "GOROOT=$GOROOT"
echo "GOPATH=$GOPATH"
echo "PATH=$PATH"

echo "=== Running Go command ==="
echo "%s $@"
exec %s "$@"
`, 
		goRoot, 
		filepath.Dir(goExePath),
		goExePath,
		goExePath)
	
	// Write the wrapper script
	if err := os.WriteFile(wrapperScript, []byte(wrapperContent), 0755); err != nil {
		return fmt.Errorf("failed to write wrapper script: %w", err)
	}
	
	fmt.Printf("Created wrapper script at %s\n", wrapperScript)
	
	// Use the wrapper script to build the Go shared library
	cmd := exec.Command(wrapperScript, "build", "-buildmode=c-shared", "-o", outputPath, libPath)
	cmd.Dir = filepath.Dir(libPath) // Set working directory to where the source file is
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Printf("Executing Go build via wrapper script\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Go stored procedure: %w", err)
	}

	fmt.Println("✅ Go stored procedure built successfully!")
	return nil
}
