# Video Conference UI

A modern video conferencing UI package for the HeroLauncher platform. This package provides a web-based user interface for video conferencing using Fiber, Pug templates, and LiveKit for real-time communication.

## Features

- Modern, responsive UI for video conferencing
- Server-side rendering with Pug templates in Go
- Room creation and management
- Real-time video/audio communication using LiveKit
- Screen sharing capabilities
- Chat functionality between participants
- Participant management with video/audio controls
- Token-based authentication for secure room access
- End-to-end encryption (E2EE) support
- Adaptive video quality based on network conditions

## Architecture

The videoconf package is built with a clean architecture that separates concerns:

1. **Backend (Go)**: Handles room management, authentication, and API endpoints using Fiber
2. **Frontend (JavaScript)**: Manages real-time communication using LiveKit's client SDK
3. **Templates (Pug)**: Provides server-side rendering for the UI

## Getting Started

### Prerequisites

- Go 1.21 or higher
- LiveKit account or self-hosted LiveKit server

### Installation

```bash
cd /Users/timurgordon/code/github/freeflowuniverse/herolauncher/pkg/ui/videoconf
go mod tidy
```

### Configuration

The videoconf package can be configured through environment variables or programmatically:

```go
// Default configuration
config := videoconf.DefaultConfig()

// Custom configuration
config := videoconf.Config{
    Port:          9000,
    TemplatesPath: "./custom/templates",
    StaticPath:    "./custom/static",
}
```

### LiveKit Configuration

LiveKit credentials must be configured through environment variables. You can set them up in two ways:

### Option 1: Using the Setup Script

We've provided a setup script to help you configure your environment variables:

```bash
cd scripts
./setup_env.sh
```

The script will guide you through setting up your LiveKit credentials.

### Option 2: Manual Setup

Copy the `.env.example` file to create your own `.env` file:

```bash
cp .env.example .env
```

Then edit the `.env` file with your LiveKit credentials:

```
# LiveKit Configuration
LIVEKIT_URL=wss://your-livekit-instance.livekit.cloud
LIVEKIT_API_KEY=your_api_key_here
LIVEKIT_API_SECRET=your_api_secret_here

# Server Configuration
PORT=8096
```

### Required Environment Variables

These environment variables are required for the videoconf package to function properly:

- `LIVEKIT_URL`: The WebSocket URL for your LiveKit server
- `LIVEKIT_API_KEY`: Your LiveKit API key
- `LIVEKIT_API_SECRET`: Your LiveKit API secret
- `PORT`: The port on which the server will run (default: 8096)

## Usage

### As a Package

```go
import "github.com/freeflowuniverse/herolauncher/pkg/ui/videoconf"

func main() {
    // Create a new video conferencing UI server with default config
    config := videoconf.DefaultConfig()
    vc := videoconf.New(config)

    // Setup routes
    vc.SetupRoutes()

    // Start the server
    if err := vc.Start(); err != nil {
        log.Fatalf("Error starting video conferencing UI server: %v", err)
    }
}
```

### Accessing the UI

Once the server is running, you can access the video conferencing UI at:

- Home page: `http://localhost:<port>/`
- Room page: `http://localhost:<port>/rooms/<room-id>`

## API Endpoints

The videoconf package provides the following API endpoints:

- `GET /` - Home page with list of available rooms
- `GET /rooms/:roomId` - Join a specific room
- `POST /api/room` - Create a new room
- `GET /api/connection-details` - Get connection details for a room
- `POST /api/token` - Generate a token for joining a room

## Project Structure

```
pkg/ui/videoconf/
├── cmd/                  # Command-line tools
├── web/                  # Web assets
│   ├── static/           # Static assets
│   │   ├── css/          # CSS stylesheets
│   │   ├── js/           # JavaScript files
│   │   │   ├── main.js   # Main JavaScript for home page
│   │   │   ├── room.js   # Room functionality (LiveKit integration)
│   │   │   └── utils.js  # Utility functions
│   │   └── images/       # Image assets
│   └── templates/        # Pug templates
│       ├── home.pug      # Home page template
│       ├── layout.pug    # Layout template
│       └── room.pug      # Room page template
├── errors.go             # Custom error definitions
├── livekit.go            # LiveKit client implementation
├── main.go               # Package entry point and server implementation
├── go.mod                # Go module file
├── go.sum                # Go dependencies checksum
└── README.md             # Documentation
```

## LiveKit Integration

This package uses LiveKit for real-time communication:

1. **Server-side Integration**: 
   - Room creation and management
   - Token generation for authentication
   - Participant management

2. **Client-side Integration**:
   - Real-time audio/video streaming
   - Screen sharing
   - Chat functionality
   - Participant rendering and UI updates
   - End-to-end encryption support

## Features in Detail

### Room Management

- Create and join rooms
- List active rooms
- Configure room settings (max participants, timeout)

### Media Controls

- Toggle camera on/off
- Toggle microphone on/off
- Share screen
- Switch camera/microphone devices

### Chat Functionality

- Send and receive text messages in a room
- Visual notifications for new messages

### Security

- Token-based authentication
- Optional end-to-end encryption
- Room access control

## Browser Compatibility

The videoconf UI works with modern browsers that support WebRTC:

- Chrome/Chromium (recommended)
- Firefox
- Safari
- Edge

## License

This project is part of the HeroLauncher platform by the Free Flow Universe.
