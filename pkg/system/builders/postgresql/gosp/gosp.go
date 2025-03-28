package gosp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// run executes a command with the given arguments
func (b *GoSPBuilder) run(cmd string, args ...string) error {
	fmt.Println("Running:", cmd, args)
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// Build builds a Go stored procedure
func (b *GoSPBuilder) Build() error {
	fmt.Println("Building Go stored procedure...")
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

	if err := b.run("go", "build", "-buildmode=c-shared", "-o", filepath.Join(b.InstallPrefix, "lib", "libgosp.so"), libPath); err != nil {
		return fmt.Errorf("failed to build Go stored procedure: %w", err)
	}

	fmt.Println("✅ Go stored procedure built successfully!")
	return nil
}
