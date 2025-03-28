package main

import (
	"fmt"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/system/builders/postgresql"
)

func main() {
	// Create a new PostgreSQL builder with default settings
	builder := postgresql.NewBuilder()

	// Build PostgreSQL
	if err := builder.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "Error building PostgreSQL: %v\n", err)
		os.Exit(1) // Ensure we exit with non-zero status on error
	}

	// Run PostgreSQL in screen
	if err := builder.PostgresBuilder.RunPostgresInScreen(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running PostgreSQL in screen: %v\n", err)
		os.Exit(1) // Ensure we exit with non-zero status on error
	}

	fmt.Println("PostgreSQL build completed successfully!")
}
