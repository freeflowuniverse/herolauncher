# IMAP Server

This command runs a standalone IMAP server that integrates with Redis for mail storage.

## Overview

The IMAP server provides a standard IMAP interface for email clients to connect to, with all mail data stored in Redis. This allows for a lightweight, scalable mail server implementation.

## Usage

```bash
go run main.go [options]
```

### Options

- `-redis-addr`: Redis server address (default: "localhost:6378")
- `-imap-addr`: IMAP server address (default: ":1143")
- `-debug`: Enable debug mode with verbose logging (default: false)

## Example

```bash
# Run with default settings
go run main.go

# Run with custom Redis and IMAP addresses
go run main.go -redis-addr localhost:6379 -imap-addr :1144 -debug
```

## Features

- Implements standard IMAP protocol
- Redis-backed storage for mailboxes and messages
- Graceful shutdown with signal handling
- Debug mode for troubleshooting
