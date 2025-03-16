package smtpserver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	mailmodel "github.com/freeflowuniverse/herolauncher/pkg/mail"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Email and Attachment types are defined in pkg/mail/model.go

// parseEmail parses the raw email data and extracts the necessary information
func parseEmail(from string, to []string, data []byte) (*mailmodel.Email, error) {
	// Add debugging to see what data we're receiving
	fmt.Printf("DEBUG: Email data received (%d bytes): %s\n", len(data), string(data))

	// Ensure data ends with proper CRLF sequence for mail.ReadMessage
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\r\n")) {
		if bytes.HasSuffix(data, []byte("\n")) {
			// Replace trailing \n with \r\n
			data = bytes.TrimSuffix(data, []byte("\n"))
			data = append(data, []byte("\r\n")...)
		} else {
			// Add \r\n if no trailing newline
			data = append(data, []byte("\r\n")...)
		}
	}

	// Parse the email message
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		// If standard parsing fails, try a more lenient approach
		return parseEmailManually(from, to, data)
	}

	// Extract subject
	subject := ""
	if subjectHeader := msg.Header.Get("Subject"); subjectHeader != "" {
		// Use the subject as is - we'll handle Unicode conversion
		subject = subjectHeader
	}

	// Ensure subject is in Unicode
	unicodeDecoder := unicode.UTF8.NewDecoder()
	subjectReader := transform.NewReader(strings.NewReader(subject), unicodeDecoder)
	subjectBytes, err := io.ReadAll(subjectReader)
	if err == nil {
		subject = string(subjectBytes)
	}

	// Print debug info
	fmt.Printf("DEBUG: Parsed subject: %s\n", subject)

	// Create email
	email := &mailmodel.Email{
		Attachments: []mailmodel.Attachment{},
	}

	// Set envelope fields
	email.SetFrom(from)
	email.SetTo(to)
	email.SetSubject(subject)

	// Process the body based on content type
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Default to plain text if we can't parse the content type
		mediaType = "text/plain"
		params = map[string]string{}
	}

	// Handle multipart messages
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary, ok := params["boundary"]
		if !ok {
			return nil, fmt.Errorf("no boundary found in multipart message")
		}

		mr := multipart.NewReader(msg.Body, boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("error reading multipart: %w", err)
			}

			partContentType := part.Header.Get("Content-Type")
			if partContentType == "" {
				partContentType = "text/plain"
			}

			partMediaType, _, err := mime.ParseMediaType(partContentType)
			if err != nil {
				// Skip parts with invalid content type
				continue
			}

			// Handle text parts
			if strings.HasPrefix(partMediaType, "text/") {
				content, err := io.ReadAll(part)
				if err != nil {
					continue
				}

				// Convert to Unicode
				unicodeDecoder := unicode.UTF8.NewDecoder()
				unicodeReader := transform.NewReader(bytes.NewReader(content), unicodeDecoder)
				unicodeContent, err := io.ReadAll(unicodeReader)
				if err == nil {
					content = unicodeContent
				}

				// If this is the first text part or it's text/plain, use it as the message
				if email.Message == "" || partMediaType == "text/plain" {
					email.Message = string(content)
				}
			} else {
				// Handle attachments
				filename := part.FileName()
				if filename == "" {
					// Skip attachments without a filename
					continue
				}

				content, err := io.ReadAll(part)
				if err != nil {
					continue
				}

				// Base64 encode the attachment data
				encodedData := base64.StdEncoding.EncodeToString(content)

				attachment := mailmodel.Attachment{
					Filename:    filename,
					ContentType: partMediaType,
					Data:        encodedData,
				}

				email.Attachments = append(email.Attachments, attachment)
			}
		}
	} else {
		// Handle single part messages
		content, err := io.ReadAll(msg.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading message body: %w", err)
		}

		// Convert to Unicode
		unicodeDecoder := unicode.UTF8.NewDecoder()
		unicodeReader := transform.NewReader(bytes.NewReader(content), unicodeDecoder)
		unicodeContent, err := io.ReadAll(unicodeReader)
		if err == nil {
			content = unicodeContent
		}

		email.Message = string(content)
	}

	return email, nil
}

// parseEmailManually parses the email data manually when the standard library fails
func parseEmailManually(from string, to []string, data []byte) (*mailmodel.Email, error) {
	fmt.Printf("DEBUG: Attempting manual email parsing\n")

	// Create a basic email structure
	email := &mailmodel.Email{
		Message:     "",
		Attachments: []mailmodel.Attachment{},
	}

	// Set envelope fields
	email.SetFrom(from)
	email.SetTo(to)
	email.SetSubject("")

	// Convert data to string for easier processing
	content := string(data)

	// Split the email into headers and body
	parts := strings.SplitN(content, "\r\n\r\n", 2)
	if len(parts) < 2 {
		// Try with just \n\n if \r\n\r\n not found
		parts = strings.SplitN(content, "\n\n", 2)
		if len(parts) < 2 {
			// If still can't split, use the whole content as the message
			email.Message = content
			return email, nil
		}
	}

	// Process headers
	headers := parts[0]
	body := parts[1]

	// Parse headers line by line
	headerLines := strings.Split(headers, "\n")
	for _, line := range headerLines {
		// Remove any trailing \r
		line = strings.TrimSuffix(line, "\r")

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if this is a continuation of the previous header
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// This is a continuation, but we'll ignore it for simplicity
			continue
		}

		// Split header into name and value
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) != 2 {
			continue
		}

		headerName := strings.TrimSpace(headerParts[0])
		headerValue := strings.TrimSpace(headerParts[1])

		// Process specific headers
		switch strings.ToLower(headerName) {
		case "subject":
			email.SetSubject(headerValue)
		}
	}

	// Set the message body
	email.Message = body

	// Debug output
	fmt.Printf("DEBUG: Manually parsed email - Subject: %s, Body length: %d\n",
		email.Subject, len(email.Message))

	return email, nil
}
