# WebDAV Server Implementation

This document describes the WebDAV server implementation for HeroLauncher.

## Overview

The WebDAV server provides a way to access and manage files through the WebDAV protocol, which allows for remote file management over HTTP/HTTPS. This implementation uses the Go standard library's WebDAV package from `golang.org/x/net/webdav`.

The server supports both HTTP and HTTPS connections, basic authentication, and includes comprehensive debug logging for troubleshooting.

## Implementation Details

The WebDAV server is implemented in the `pkg/webdavserver` package. The server can be configured with various options including:

- Host and port to listen on
- Base path for the WebDAV endpoint
- File system path to serve files from
- Read and write timeouts
- Debug mode for verbose logging
- Basic authentication with username/password
- HTTPS support with TLS certificate and key files

## Usage

### Starting the WebDAV Server

To start the WebDAV server, use the `cmd/webdavserver/main.go` command:

```bash
go run cmd/webdavserver/main.go [options]
```

Available options:

- `-host`: Host address to bind to (default: "0.0.0.0")
- `-port`: Port to listen on (default: 9999)
- `-base-path`: Base URL path for WebDAV (default: "/")
- `-fs`: File system path to serve (default: system temp directory + "/herolauncher")
- `-debug`: Enable debug mode with verbose logging (default: false)
- `-auth`: Enable basic authentication (default: false)
- `-username`: Username for basic authentication (default: "admin")
- `-password`: Password for basic authentication (default: "1234")
- `-https`: Enable HTTPS (default: false)
- `-cert`: Path to TLS certificate file (optional if auto-generation is enabled)
- `-key`: Path to TLS key file (optional if auto-generation is enabled)
- `-auto-gen-certs`: Auto-generate certificates if they don't exist (default: true)
- `-cert-validity`: Validity period in days for auto-generated certificates (default: 365)
- `-cert-org`: Organization name for auto-generated certificates (default: "HeroLauncher WebDAV Server")

### Connecting to WebDAV from macOS

A bash script is provided to easily connect to the WebDAV server from macOS:

```bash
./scripts/open_webdav_osx.sh [options]
```

Available options:

- `-h, --host`: WebDAV server hostname (default: "localhost")
- `-p, --port`: WebDAV server port (default: 9999)
- `-path, --path-prefix`: Path prefix for WebDAV URL (default: "")
- `-s, --https`: Use HTTPS instead of HTTP (default: false)
- `-u, --username`: Username for authentication
- `-pw, --password`: Password for authentication
- `--help`: Show help message

## API

### Server Configuration

```go
// Config holds the configuration for the WebDAV server
type Config struct {
    Host                string
    Port                int
    BasePath            string
    FileSystem          string
    ReadTimeout         time.Duration
    WriteTimeout        time.Duration
    DebugMode           bool
    UseAuth             bool
    Username            string
    Password            string
    UseHTTPS            bool
    CertFile            string
    KeyFile             string
    AutoGenerateCerts   bool
    CertValidityDays    int
    CertOrganization    string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config
```

### Server Methods

```go
// NewServer creates a new WebDAV server
func NewServer(config Config) (*Server, error)

// Start starts the WebDAV server
func (s *Server) Start() error

// Stop stops the WebDAV server
func (s *Server) Stop() error
```

## Integration with HeroLauncher

The WebDAV server can be integrated with the main HeroLauncher application by adding it to the server initialization in `cmd/server/main.go`.

## Directory Structure

The WebDAV server uses the following directory structure:

```
<parent-of-fs>/
  ├── <fs-dir>/       # WebDAV files served to clients (specified by -fs)
  └── certificates/   # TLS certificates for HTTPS
```

Where certificates are stored in a `certificates` directory next to the filesystem directory specified with the `-fs` parameter.

## Security Considerations

- Basic authentication is supported but disabled by default
- HTTPS is supported but disabled by default
- The server can automatically generate self-signed certificates if needed
- For production use, always enable authentication and HTTPS
- Use strong passwords and properly signed certificates for production
- Be careful about which directories you expose through WebDAV
- Consider implementing IP-based access restrictions for additional security

## Debugging

When troubleshooting WebDAV connections, the debug mode can be enabled with the `-debug` flag. This will provide detailed logging of:

- All incoming requests
- Request headers
- Client information
- Authentication attempts
- WebDAV operations

Debug logs are prefixed with `[WebDAV DEBUG]` for easy filtering.

## Examples

### Starting a secure WebDAV server with auto-generated certificates

```bash
go run cmd/webdavserver/main.go -auth -username myuser -password mypass -https -fs /path/to/files -debug
```

### Starting a secure WebDAV server with existing certificates

```bash
go run cmd/webdavserver/main.go -auth -username myuser -password mypass -https -cert /path/to/cert.pem -key /path/to/key.pem -fs /path/to/files -debug -auto-gen-certs=false
```

### Connecting from macOS with authentication

```bash
./scripts/open_webdav_osx.sh -s -u myuser -pw mypass
```

## References

- [WebDAV Protocol (RFC 4918)](https://tools.ietf.org/html/rfc4918)
- [Go WebDAV Package](https://pkg.go.dev/golang.org/x/net/webdav)
- [TLS in Go](https://pkg.go.dev/crypto/tls)
