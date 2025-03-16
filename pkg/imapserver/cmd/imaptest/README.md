# IMAP Test Client

This command provides a test client for the IMAP server, demonstrating various IMAP operations.

## Overview

The IMAP test client connects to an IMAP server and performs a series of operations to test functionality, including:
- Listing mailboxes
- Selecting mailboxes
- Fetching messages
- Moving messages between folders
- Marking messages for deletion
- Expunging deleted messages

## Usage

```bash
go run main.go [options]
```

### Options

- `-imap-addr`: IMAP server address (default: "localhost:2143")
- `-username`: IMAP username (default: "jan")
- `-password`: IMAP password (default: "password")

## Example

```bash
# Run with default settings
go run main.go

# Run with custom IMAP server address and credentials
go run main.go -imap-addr localhost:1143 -username testuser -password testpass
```

## Test Operations

The client performs the following test operations:
1. Connect to the IMAP server
2. List available mailboxes
3. Select the inbox
4. Fetch and display message envelopes
5. Test moving a message to the "archive" folder
6. Test marking a message for deletion and expunging it
