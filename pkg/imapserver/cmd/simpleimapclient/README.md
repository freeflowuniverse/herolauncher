# Simple IMAP Client

This command provides a simple IMAP client for interacting with the IMAP server.

## Overview

The Simple IMAP Client connects to an IMAP server and performs various operations to demonstrate and test the IMAP functionality. It's designed to be a more comprehensive and user-friendly client compared to the basic imaptest tool.

## Usage

```bash
go run main.go [options]
```

### Options

- `-imap-addr`: IMAP server address (default: "localhost:1143")
- `-username`: IMAP username (default: "jan")
- `-password`: IMAP password (default: "testpass")

## Example

```bash
# Run with default settings
go run main.go

# Run with custom IMAP server address and credentials
go run main.go -imap-addr localhost:2143 -username testuser -password testpass
```

## Features

- Lists all available mailboxes
- Searches for mailboxes containing messages
- Fetches and displays message details including:
  - Subject
  - From/To addresses
  - Flags
  - Date
  - Body preview
- Tests message deletion functionality
- Tests flag setting operations (marking messages as seen)
- Provides formatted output of email addresses and message content
