package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Parse command line flags
	host := flag.String("host", "localhost", "SMTP server host")
	port := flag.Int("port", 2526, "SMTP server port")
	from := flag.String("from", "sender@example.com", "Sender email address")
	to := flag.String("to", "recipient@example.com", "Recipient email address (comma-separated for multiple)")
	emailFile := flag.String("file", "", "Path to email file (.eml)")
	flag.Parse()

	// Validate email file
	if *emailFile == "" {
		log.Fatal("Email file path is required")
	}

	// Read email file
	emailData, err := os.ReadFile(*emailFile)
	if err != nil {
		log.Fatalf("Failed to read email file: %v", err)
	}

	// Parse recipient list
	recipients := strings.Split(*to, ",")
	for i, recipient := range recipients {
		recipients[i] = strings.TrimSpace(recipient)
	}

	// Connect to the SMTP server
	addr := net.JoinHostPort(*host, fmt.Sprintf("%d", *port))
	log.Printf("Connecting to SMTP server at %s", addr)
	
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Read the greeting
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read greeting: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	// Send HELO
	cmd := fmt.Sprintf("HELO localhost\r\n")
	log.Printf("Client: %s", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		log.Fatalf("Failed to send HELO: %v", err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read HELO response: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	// Send MAIL FROM
	cmd = fmt.Sprintf("MAIL FROM:<%s>\r\n", *from)
	log.Printf("Client: %s", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		log.Fatalf("Failed to send MAIL FROM: %v", err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read MAIL FROM response: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	// Send RCPT TO for each recipient
	for _, recipient := range recipients {
		cmd = fmt.Sprintf("RCPT TO:<%s>\r\n", recipient)
		log.Printf("Client: %s", cmd)
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			log.Fatalf("Failed to send RCPT TO: %v", err)
		}
		n, err = conn.Read(buf)
		if err != nil {
			log.Fatalf("Failed to read RCPT TO response: %v", err)
		}
		log.Printf("Server: %s", string(buf[:n]))
	}

	// Send DATA
	cmd = "DATA\r\n"
	log.Printf("Client: %s", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		log.Fatalf("Failed to send DATA: %v", err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read DATA response: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	// Send email content
	log.Printf("Client: [Sending email content from file]")
	
	// Ensure email ends with proper CRLF sequence and dot
	emailStr := string(emailData)
	if !strings.HasSuffix(emailStr, "\r\n") {
		if strings.HasSuffix(emailStr, "\n") {
			emailStr = strings.TrimSuffix(emailStr, "\n") + "\r\n"
		} else {
			emailStr += "\r\n"
		}
	}
	
	// Add the final dot on a line by itself
	emailStr += ".\r\n"
	
	_, err = conn.Write([]byte(emailStr))
	if err != nil {
		log.Fatalf("Failed to send email content: %v", err)
	}
	
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read email content response: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	// Send QUIT
	cmd = "QUIT\r\n"
	log.Printf("Client: %s", cmd)
	_, err = conn.Write([]byte(cmd))
	if err != nil {
		log.Fatalf("Failed to send QUIT: %v", err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read QUIT response: %v", err)
	}
	log.Printf("Server: %s", string(buf[:n]))

	log.Printf("Email sent successfully from %s to %s", *from, strings.Join(recipients, ", "))
}
