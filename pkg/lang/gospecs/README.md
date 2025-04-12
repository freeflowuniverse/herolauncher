# Go Specs Processor

This package provides functionality to analyze Go code and extract public API specifications.

The GoProcessor extracts:
- Public structs and their fields
- Public interfaces and their methods
- Public standalone functions
- Public methods

It preserves documentation comments and follows Go's convention of using capitalization to determine if an element is public.

## Usage

```go
import "github.com/freeflowuniverse/herolauncher/pkg/lang/gospecs"

processor := gospecs.NewGoProcessor()
specs, err := processor.GetSpec("/path/to/go/project")
if err != nil {
    // Handle error
}

fmt.Println(specs) // Will contain all public Go declarations
```

The output will include the public types and functions from all Go files in the specified path, preserving their documentation comments but excluding implementation details.