# Redis Mail Feeder

This command generates random test emails and stores them in Redis for use with the IMAP server.

## Overview

The Redis Mail Feeder is a utility that creates random email data and stores it in Redis using the appropriate format for the IMAP server to access. This is useful for testing and development purposes when you need sample email data.

## Usage

```bash
go run main.go [options]
```

### Options

- `-redis-addr`: Redis server address (default: "localhost:6378")
- `-redis-password`: Redis server password (default: "")
- `-redis-db`: Redis database number (default: 0)
- `-num-emails`: Number of emails to generate (default: 100)

## Example

```bash
# Generate 100 emails with default settings
go run main.go

# Generate 50 emails with custom Redis configuration
go run main.go -redis-addr localhost:6379 -num-emails 50
```

## Features

- Generates random emails with realistic content
- Creates emails for multiple test accounts
- Distributes emails across various folders and subfolders
- Occasionally adds attachments to emails
- Ensures unique UIDs for each email
- Stores emails in Redis using the format expected by the IMAP server
