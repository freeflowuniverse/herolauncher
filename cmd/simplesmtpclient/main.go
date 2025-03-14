package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	// Parse command line flags
	host := flag.String("host", "localhost", "SMTP server host")
	port := flag.Int("port", 2526, "SMTP server port")
	from := flag.String("from", "sender@example.com", "Sender email address")
	to := flag.String("to", "recipient@example.com", "Recipient email address (comma-separated for multiple)")
	subject := flag.String("subject", "Test Email", "Email subject")
	message := flag.String("message", "This is a test email sent from the SMTP client.", "Email message body")
	flag.Parse()

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

	// Send email content with proper MIME structure
	emailContent := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"Content-Transfer-Encoding: 7bit\r\n"+
		"\r\n"+
		"%s\r\n"+
		".\r\n",
		*from, strings.Join(recipients, ", "), *subject, 
		time.Now().Format(time.RFC1123Z), *message)
	
	log.Printf("Client: [Sending email content]")
	_, err = conn.Write([]byte(emailContent))
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
