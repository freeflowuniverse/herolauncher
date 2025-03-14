package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	// Parse command line flags
	host := flag.String("host", "localhost", "SMTP server host")
	port := flag.Int("port", 2525, "SMTP server port")
	from := flag.String("from", "sender@example.com", "Sender email address")
	to := flag.String("to", "recipient@example.com", "Recipient email address (comma-separated for multiple)")
	subject := flag.String("subject", "Test Email", "Email subject")
	message := flag.String("message", "This is a test email sent from the SMTP client.", "Email message body")
	attachment := flag.String("attachment", "", "Path to file to attach (optional)")
	flag.Parse()

	// Parse recipient list
	recipients := strings.Split(*to, ",")
	for i, recipient := range recipients {
		recipients[i] = strings.TrimSpace(recipient)
	}

	// Connect to the SMTP server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Connecting to SMTP server at %s", addr)
	
	client, err := smtp.Dial(addr)
	if err != nil {
		log.Fatalf("Failed to connect to SMTP server: %v", err)
	}
	defer client.Close()

	// Set the sender and recipient
	if err := client.Mail(*from); err != nil {
		log.Fatalf("Failed to set sender: %v", err)
	}

	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			log.Fatalf("Failed to set recipient %s: %v", recipient, err)
		}
	}

	// Create a properly formatted RFC 822 message
	var messageBuffer strings.Builder
	
	// Generate a boundary for multipart messages
	boundary := fmt.Sprintf("_boundary_%d", time.Now().UnixNano())
	
	// Add headers
	messageBuffer.WriteString(fmt.Sprintf("From: %s\r\n", *from))
	messageBuffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ", ")))
	messageBuffer.WriteString(fmt.Sprintf("Subject: %s\r\n", *subject))
	messageBuffer.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	messageBuffer.WriteString("MIME-Version: 1.0\r\n")
	
	// Set content type based on whether there's an attachment
	if *attachment == "" {
		messageBuffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		messageBuffer.WriteString("\r\n")
		// Add random number to the message
		randomNum := rand.Intn(10000)
		messageBuffer.WriteString(fmt.Sprintf("%s\n\nRandom Number: %d", *message, randomNum))
	} else {
		messageBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		messageBuffer.WriteString("\r\n")
		
		// Add text part
		messageBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		messageBuffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		messageBuffer.WriteString("\r\n")
		// Add random number to the message
		randomNum := rand.Intn(10000)
		messageBuffer.WriteString(fmt.Sprintf("%s\n\nRandom Number: %d", *message, randomNum))
		messageBuffer.WriteString("\r\n\r\n")
		
		// Add attachment
		file, err := os.Open(*attachment)
		if err != nil {
			log.Fatalf("Failed to open attachment: %v", err)
		}
		defer file.Close()
		
		// Check if file exists and is readable
		_, err = file.Stat()
		if err != nil {
			log.Fatalf("Failed to stat attachment: %v", err)
		}
		
		// Determine content type based on file extension
		contentType := "application/octet-stream"
		ext := filepath.Ext(*attachment)
		switch strings.ToLower(ext) {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".pdf":
			contentType = "application/pdf"
		case ".txt":
			contentType = "text/plain"
		case ".html":
			contentType = "text/html"
		}
		
		// Add attachment headers
		filename := filepath.Base(*attachment)
		messageBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		messageBuffer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
		messageBuffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filename))
		messageBuffer.WriteString("Content-Transfer-Encoding: base64\r\n")
		messageBuffer.WriteString("\r\n")
		
		// Read file and encode as base64
		fileData, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("Failed to read attachment: %v", err)
		}
		
		// Encode to base64 and add to message
		encoded := base64.StdEncoding.EncodeToString(fileData)
		
		// Add base64 data in lines of 76 characters
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			messageBuffer.WriteString(encoded[i:end] + "\r\n")
		}
		
		// Close the multipart message
		messageBuffer.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	}
	
	// Send the DATA command
	wc, err := client.Data()
	if err != nil {
		log.Fatalf("Failed to send DATA command: %v", err)
	}
	defer wc.Close()
	
	// Write the complete message
	if _, err = fmt.Fprint(wc, messageBuffer.String()); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	// Send the QUIT command
	if err := client.Quit(); err != nil {
		log.Fatalf("Failed to quit: %v", err)
	}

	log.Printf("Email sent successfully from %s to %s", *from, strings.Join(recipients, ", "))
	if *attachment != "" {
		log.Printf("Attachment: %s", *attachment)
	}
}

// No longer needed as the functionality is now inline
