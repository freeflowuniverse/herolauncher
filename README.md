# HeroLauncher

[![Go Tests](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-tests.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-tests.yml)
[![Go Lint](https://github.com/freeflowuniverse/herolauncher/actions/workflows/lint.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/lint.yml)
[![Build](https://github.com/freeflowuniverse/herolauncher/actions/workflows/build.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/build.yml)


HeroLauncher is a comprehensive launcher application written in V language with multiple modules:

- **Installer Module**: Handles installation of dependencies and components
- **Web Server Module**: Provides a web UI, Swagger UI, and OpenAPI REST interface (v3.1.0)
- **IPFS Server Module**: Manages IPFS functionality

## Features

- Web server with modern UI
- OpenAPI v3.1.0 REST interfaces
- Swagger UI for API documentation
- Command execution with job tracking
- Package management (apt, brew, scoop)
- IPFS integration

## Installation

### Prerequisites

- [V language](https://vlang.io/) installed
- For IPFS functionality: [IPFS](https://ipfs.io/) installed

### Building from Source

```bash
v .  # Build the project
```

## Usage

### Running HeroLauncher

```bash
# Run with default settings
./herolauncher

# Run with web server on a specific port
./herolauncher -w -p 9090

# Enable IPFS server
./herolauncher -i

# Run in installer mode
./herolauncher --install

# Show help
./herolauncher -h
```

### Command Line Options

- `-w, --web`: Enable web server (default: true)
- `-p, --port`: Web server port (default: 9001)
- `--host`: Web server host (default: localhost)
- `-i, --ipfs`: Enable IPFS server
- `--install`: Run in installer mode
- `-h, --help`: Show help message

## API Documentation

When the web server is running, you can access the Swagger UI at:

```
http://localhost:9001/swagger
```

The OpenAPI specification is available at:

```
http://localhost:9001/openapi.json
```

## Project Structure

```
/
├── modules/
│   ├── installer/       # Installer module
│   ├── webserver/       # Web server module
│   │   ├── endpoints/
│   │   │   ├── executor/       # Command execution endpoint
│   │   │   └── packagemanager/ # Package management endpoint
│   └── ipfs/           # IPFS server module
├── main.v              # Main application entry point
└── v.mod               # V module definition
```

## Development

### Running Tests

```bash
# Run all tests
./test.sh

# Run tests with debug output
./test.sh --debug
```

The test script will run all Go tests in the project and display a summary of the results at the end. You can exclude specific packages by uncommenting them in the `EXCLUDED_MODULES` array in the test.sh file.

### Continuous Integration and Deployment

This project uses GitHub Actions for CI/CD:

- **Go Tests**: Runs all tests using the test.sh script on every push and pull request
- **Go Lint**: Performs linting using golangci-lint to ensure code quality
- **Build**: Builds the application for multiple platforms (Linux Intel/ARM, macOS Intel/ARM, Windows) and makes the binaries available as artifacts

### Downloading Binaries from CI

The Build workflow creates binaries for multiple platforms and makes them available as artifacts. To download the binaries:

1. Go to the [Actions](https://github.com/freeflowuniverse/herolauncher/actions) tab in the repository
2. Click on the latest successful Build workflow run
3. Scroll down to the Artifacts section
4. Download the artifact for your platform:
   - `herolauncher-linux-amd64.tar.gz` for Linux (Intel)
   - `herolauncher-linux-arm64.tar.gz` for Linux (ARM)
   - `herolauncher-darwin-amd64.tar.gz` for macOS (Intel)
   - `herolauncher-darwin-arm64.tar.gz` for macOS (ARM)
   - `herolauncher-windows-amd64.zip` for Windows
5. Extract the archive to get the binaries
6. The archive contains the following executables:
   - `pmclient-[platform]`: Process Manager client
   - `telnettest-[platform]`: Telnet test utility
   - `webdavclient-[platform]`: WebDAV client
   - `webdavserver-[platform]`: WebDAV server
7. Run the desired executable from the command line

To create a new release:

```bash
# Tag a new version
git tag v1.0.0

# Push the tag to trigger the release workflow
git push origin v1.0.0
```

#### Docker

A Docker image is automatically built and pushed to Docker Hub on each push to main/master and on tag releases. To use the Docker image:

```bash
# Pull the latest image
docker pull username/herolauncher:latest

# Run the container
docker run -p 9001:9001 username/herolauncher:latest
```

Replace `username` with the actual Docker Hub username configured in the repository secrets.

## License

MIT
