package mail

import (
	"github.com/gofiber/fiber/v2"
)

// Mailbox represents a mail folder/mailbox
type Mailbox struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Type  string `json:"type"` // inbox, sent, drafts, trash, etc.
}

// Mail represents an email message
type Mail struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	From        string   `json:"from"`
	To          []string `json:"to"`
	Cc          []string `json:"cc,omitempty"`
	Bcc         []string `json:"bcc,omitempty"`
	Date        string   `json:"date"`
	Body        string   `json:"body"`
	BodyHTML    string   `json:"bodyHtml,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
	Read        bool     `json:"read"`
	Starred     bool     `json:"starred"`
	MailboxID   string   `json:"mailboxId"`
}

// MailPreview represents a preview of an email in a list
type MailPreview struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	From      string `json:"from"`
	Date      string `json:"date"`
	Preview   string `json:"preview"`
	Read      bool   `json:"read"`
	Starred   bool   `json:"starred"`
	HasAttach bool   `json:"hasAttach"`
	MailboxID string `json:"mailboxId"`
}

// Configuration represents the mail client configuration
type Configuration struct {
	Title   string `json:"Title,omitempty"`
	BaseURL string `json:"BaseURL,omitempty"`
}

// MailServer represents the mail server with its configuration
type MailServer struct {
	App     *fiber.App
	Config  Configuration
}
