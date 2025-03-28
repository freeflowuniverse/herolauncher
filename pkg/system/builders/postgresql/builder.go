package postgresql

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/dependencies"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/gosp"
	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql/postgres"
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
}

// NewBuilder creates a new PostgreSQL builder with default values
func NewBuilder() *Builder {
	installPrefix := DefaultInstallPrefix

	return &Builder{
		InstallPrefix:     installPrefix,
		PostgresBuilder:   postgres.NewPostgresBuilder().WithInstallPrefix(installPrefix),
		GoSPBuilder:       gosp.NewGoSPBuilder(installPrefix),
		DependencyManager: dependencies.NewDependencyManager("bison"),
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

	// Build Go stored procedure
	if err := b.GoSPBuilder.Build(); err != nil {
		return fmt.Errorf("failed to build Go stored procedure: %w", err)
	}

	fmt.Println("✅ Done! PostgreSQL installed in:", b.InstallPrefix)
	return nil
}
