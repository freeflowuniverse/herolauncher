# SMTP Library

A simple SMTP server implementation based on [github.com/emersion/go-smtp](https://github.com/emersion/go-smtp) that processes incoming emails and stores them in Redis.

## Features

- SMTP server that receives emails
- Converts email content to Unicode text
- Extracts attachments and encodes them as base64
- Stores emails in Redis as JSON
- Adds emails to a Redis queue for processing

## Structure

The library consists of the following components:

- `smtp.go`: Main SMTP server implementation
- `parser.go`: Email parser that extracts email information
- `utils.go`: Utility functions for processing emails
- `example.go`: Example implementation of the SMTP server

## Usage

### Starting the SMTP Server

```go
package main

import (
    "log"
    "time"

    "github.com/freeflowuniverse/herolauncher/pkg/smtp"
)

func main() {
    // Create a new SMTP server with default configuration
    config := smtp.DefaultConfig()
    config.Port = 2525 // Use a non-privileged port for testing
    
    server, err := smtp.NewServer(config)
    if err != nil {
        log.Fatalf("Failed to create SMTP server: %v", err)
    }
    
    // Start the server
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start SMTP server: %v", err)
    }
}
```

### Processing Emails

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/freeflowuniverse/herolauncher/pkg/smtp"
)

func main() {
    // Create a new SMTP server with default configuration
    config := smtp.DefaultConfig()
    server, err := smtp.NewServer(config)
    if err != nil {
        log.Fatalf("Failed to create SMTP server: %v", err)
    }
    
    // Create a processor for incoming emails
    processor := func(email *smtp.Email) error {
        // Process the email
        fmt.Printf("From: %s\n", email.From)
        fmt.Printf("To: %v\n", email.To)
        fmt.Printf("Subject: %s\n", email.Subject)
        fmt.Printf("Message: %s\n", email.Message)
        fmt.Printf("Attachments: %d\n", len(email.Attachments))
        
        return nil
    }
    
    // Process emails from the queue
    if err := smtp.ProcessEmails(server.GetRedisClient(), processor, 5*time.Second); err != nil {
        log.Fatalf("Error processing emails: %v", err)
    }
}
```

## Email Format

Emails are stored in Redis as JSON with the following structure:

```json
{
  "from": "sender@example.com",
  "to": ["recipient1@example.com", "recipient2@example.com"],
  "subject": "Example Email",
  "message": "This is the email body.",
  "attachments": [
    {
      "filename": "example.txt",
      "content_type": "text/plain",
      "data": "base64-encoded-data"
    }
  ]
}
```

## Redis Storage

Emails are stored in Redis using the following pattern:

- Each email is stored as a hash at `mail:out:<unique-id>`
- The email JSON is stored in the `data` field of the hash
- The email ID is added to the `mail:out` queue for processing
