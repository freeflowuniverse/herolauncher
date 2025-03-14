# DocTree Package

The DocTree package provides functionality for managing collections of markdown pages and files. It uses Redis to store metadata about the collections, pages, and files.

## Features

- Organize markdown pages and files into collections
- Retrieve markdown pages and convert them to HTML
- Include content from other pages using a simple include directive
- Cross-collection includes
- File URL generation for static file serving
- Path management for pages and files

## Usage

### Creating a DocTree

```go
import "github.com/freeflowuniverse/herolauncher/pkg/doctree"

// Create a new DocTree with a path and name
dt, err := doctree.New("/path/to/collection", "My Collection")
if err != nil {
    log.Fatalf("Failed to create DocTree: %v", err)
}
```

### Getting Collection Information

```go
// Get information about the collection
info := dt.Info()
fmt.Printf("Collection Name: %s\n", info["name"])
fmt.Printf("Collection Path: %s\n", info["path"])
```

### Working with Pages

```go
// Get a page by name
content, err := dt.PageGet("page-name")
if err != nil {
    log.Fatalf("Failed to get page: %v", err)
}
fmt.Println(content)

// Get a page as HTML
html, err := dt.PageGetHtml("page-name")
if err != nil {
    log.Fatalf("Failed to get page as HTML: %v", err)
}
fmt.Println(html)

// Get the path of a page
path, err := dt.PageGetPath("page-name")
if err != nil {
    log.Fatalf("Failed to get page path: %v", err)
}
fmt.Printf("Page path: %s\n", path)
```

### Working with Files

```go
// Get the URL for a file
url, err := dt.FileGetUrl("image.png")
if err != nil {
    log.Fatalf("Failed to get file URL: %v", err)
}
fmt.Printf("File URL: %s\n", url)
```

### Rescanning a Collection

```go
// Rescan the collection to update Redis metadata
err = dt.Scan()
if err != nil {
    log.Fatalf("Failed to rescan collection: %v", err)
}
```

## Include Directive

You can include content from other pages using the include directive:

```markdown
# My Page

This is my page content.

!!include name:'other-page'
```

This will include the content of 'other-page' at that location.

You can also include content from other collections:

```markdown
# My Page

This is my page content.

!!include name:'other-collection:other-page'
```

## Implementation Details

- All page and file names are "namefixed" (lowercase, non-ASCII characters removed, special characters replaced with underscores)
- Metadata is stored in Redis using hsets with the key format `collections:$name`
- Each hkey in the hset is a namefixed filename, and the value is the relative path in the collection
- The package uses a global Redis client to store metadata, rather than starting its own Redis server

## Example

See the [example](./example/example.go) for a complete demonstration of how to use the DocTree package.
