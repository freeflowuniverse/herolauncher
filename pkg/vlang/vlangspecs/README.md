# VlangSpecs

VlangSpecs is a Go package (in the `vlang` package) that extracts public structs, enums, and methods from V language files to generate API specifications. It recursively walks through a directory structure, finds all `.v` files, and extracts their public interfaces without implementation code.

## Features

- Recursively scans directories for V language files
- Extracts public structs with their fields
- Extracts public enums with their values
- Extracts public methods with their signatures (without implementation)
- Preserves documentation comments
- Skips generated files (files with names ending in `_.v` or starting with `_`)

## Installation

```bash
go get github.com/freeflowuniverse/herolauncher/pkg/vlang
```

## Usage

The primary function exposed by this package is `GetSpec`. Here's how to use it:

```go
package main

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/vlang"
)

func main() {
	// Create a new VlangProcessor
	processor := vlang.NewVlangProcessor()

	// Specify the path to your V language files
	vPath := "/Users/despiegk/code/github/freeflowuniverse/herolib/lib/circles/core"

	// Get the specification
	spec, err := processor.GetSpec(vPath)
	if err != nil {
		log.Fatalf("Error processing V files: %v", err)
	}

	// Use the generated specification
	fmt.Println(spec)
	
	// Alternatively, save it to a file
	// err = os.WriteFile("vlang_spec.v", []byte(spec), 0644)
	// if err != nil {
	//     log.Fatalf("Error writing specification to file: %v", err)
	// }
}
```

## Example Output

The output will be a V language file containing all public structs, enums, and methods without implementation code. For example:

```v
// From file: /path/to/your/vlang/project/user.v
// User represents a system user
pub struct User {
	id string
	name string
	email string
	is_active bool
}

// CreateUser creates a new user in the system
pub fn (u &User) CreateUser() {}

// From file: /path/to/your/vlang/project/role.v
pub enum Role {
	admin
	user
	guest
}
```

## Notes

- The package only extracts public declarations (those with the `pub` keyword)
- Method implementations are replaced with empty bodies (`{}`)
- The package preserves the original documentation comments
- Generated files (ending with `_.v` or starting with `_`) are skipped
