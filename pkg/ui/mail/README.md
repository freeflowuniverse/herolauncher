# Mail UI Package

A simple mail client UI implementation using Go, Fiber, Pug templates, PicoCSS, and Unpoly.

## Overview

This package provides a web-based mail client interface with the following features:

- Mailbox sidebar for navigation between different mail folders
- Mail list view showing previews of emails in the selected mailbox
- Mail detail view for reading individual emails
- Responsive design that works on desktop and mobile devices
- Dynamic content loading using Unpoly for a smooth user experience

## Structure

The package follows a clean structure:

```
pkg/ui/mail/
├── cmd/               # Command-line tools
│   └── main.go        # Main executable for running the mail server
├── web/               # Web assets
│   ├── static/        # Static files (CSS, JS)
│   │   ├── css/       # CSS stylesheets
│   │   └── js/        # JavaScript files
│   └── templates/     # Pug templates
│       ├── layout.pug         # Main layout template
│       ├── mailbox-sidebar.pug # Sidebar template
│       ├── mail-list.pug      # Mail list template
│       └── mail-view.pug      # Mail detail view template
├── factory.go         # Server initialization
├── main.go            # Package entry point
├── routes.go          # HTTP route definitions
└── types.go           # Data structures
```

## Technologies

- **Go**: Backend language
- **Fiber**: Web framework
- **Pug**: Template engine
- **PicoCSS**: Minimal CSS framework
- **Unpoly**: JavaScript library for smooth page transitions

## Usage

To run the mail server:

```bash
go run cmd/main.go --port 8080
```

Then open your browser to http://localhost:8080 to view the mail client.

## Mock Data

Currently, the package uses mock data for demonstration purposes. In a real implementation, you would replace the mock data functions with actual API calls to a mail server.

## Customization

You can customize the appearance by modifying the CSS files in the `web/static/css` directory.
