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
	
	// Ensure Go is installed before proceeding
	goInstaller := postgres.NewGoInstaller()
	goExePath, err := goInstaller.InstallGo()
	if err != nil {
		return fmt.Errorf("failed to ensure Go is installed: %w", err)
	}
	
	fmt.Printf("Using Go executable from: %s\n", goExePath)
	
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
	
	// Use our run helper that includes proper environment variables
	if err := b.run(goExePath, "build", "-buildmode=c-shared", "-o", filepath.Join(b.InstallPrefix, "lib", "libgosp.so"), libPath); err != nil {
		return fmt.Errorf("failed to build Go stored procedure: %w", err)
	}

	fmt.Println("✅ Go stored procedure built successfully!")
	return nil
}
