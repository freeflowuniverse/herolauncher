# HeroLauncher

[![Go Tests](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-tests.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-tests.yml)
[![Go CI](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-ci.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/go-ci.yml)
[![Docker](https://github.com/freeflowuniverse/herolauncher/actions/workflows/docker.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/docker.yml)
[![Release](https://github.com/freeflowuniverse/herolauncher/actions/workflows/release.yml/badge.svg)](https://github.com/freeflowuniverse/herolauncher/actions/workflows/release.yml)

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
- **Go CI**: Performs linting, builds with multiple Go versions, and generates code coverage reports
- **Docker**: Builds and pushes Docker images on pushes to main/master and on tag releases
- **Release**: Automatically creates GitHub releases with binaries for multiple platforms when a tag is pushed

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
