# HeroLib Wiki Server

A modern Markdown wiki server that renders Markdown files with a clean UI, built with Go Fiber and Goldmark. This wiki server is designed to display documentation for the HeroLib project and other related repositories.

## Features

- Clean, responsive design
- Server-side rendering with Go Fiber
- Markdown rendering with proper syntax highlighting
- On-page anchor navigation for headings
- Configurable sidebar with custom sections via JSON configuration
- Hierarchical navigation structure that preserves directory paths
- VFS (Virtual File System) support for flexible content storage

## Usage

### Running the Wiki Server

```bash
go run main.go <content_path> [config_path] [port]
```

Parameters:
- `content_path`: Path to the directory containing markdown files (required)
- `config_path`: Path to the JSON configuration file (optional)
- `port`: Port to run the server on (default: 3002)

### Configuration File

The wiki server now loads its sidebar structure from a JSON configuration file instead of generating it dynamically. The configuration file should have the following structure:

```json
{
  "Sidebar": [
    {
      "Title": "Section Title",
      "Items": [
        {
          "Title": "Item Title",
          "Href": "/path/to/file",
          "IsDir": false
        },
        {
          "Title": "Directory",
          "Href": "/path/to/directory",
          "IsDir": true,
          "Children": [
            {
              "Title": "Nested Item",
              "Href": "/path/to/directory/nested_file",
              "IsDir": false
            }
          ]
        }
      ]
    }
  ],
  "Title": "Wiki Title"
}
```

### Path Structure

The sidebar paths in the configuration must include the top-level directory name to ensure proper navigation. For example:

- Correct: `/core/concepts/name_registry`
- Incorrect: `/concepts/name_registry`

### Generating Configuration

A script like `serve_wiki.sh` in the HeroLib manual directory can be used to generate the configuration file based on the directory structure. This script should:

1. Create a temporary configuration file
2. Scan the content directory and build a hierarchical structure
3. Ensure paths include the top-level directory name
4. Write the configuration to the temporary file
5. Pass the configuration file path to the wiki server

## Project Structure

```
wiki/
├── main.go                # Main application file
├── templates/             # HTML templates
│   └── layout.html        # Main layout template
```

## Technologies Used

- **Go Fiber v2.52.6**: Fast, flexible web framework for Go
- **Goldmark v1.6.0**: Markdown parser and compiler for Go
- **VFS**: Virtual File System for content access
