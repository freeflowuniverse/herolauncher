package main

import (
	"fmt"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql"
)

func main() {
	// Create a new PostgreSQL builder with default settings
	builder := postgresql.NewBuilder()

	// Optionally customize the builder
	// builder.WithPostgresURL("https://github.com/postgres/postgres/archive/refs/tags/REL_17_4.tar.gz")
	// builder.WithInstallPrefix("/opt/postgresql")

	// Build PostgreSQL
	if err := builder.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "Error building PostgreSQL: %v\n", err)
		os.Exit(1) // Ensure we exit with non-zero status on error
	}

	fmt.Println("PostgreSQL build completed successfully!")
}
