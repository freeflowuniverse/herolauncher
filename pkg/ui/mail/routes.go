package mail

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the mail server
func (m *MailServer) SetupRoutes() {
	// Home route - shows the mailboxes list
	m.App.All("/", func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusOK)
		}
		
		// Get mailboxes (mock data for now)
		mailboxes := getMockMailboxes()
		
		// Render the layout template with the mailboxes
		return c.Render("layout", fiber.Map{
			"mailboxes": mailboxes,
			"title":     m.Config.Title,
			"view":      "mailboxes",
		})
	})

	// Mailbox route - displays a list of emails in the selected mailbox
	m.App.Get("/mailbox/:id", func(c *fiber.Ctx) error {
		mailboxID := c.Params("id")
		
		// Get mailboxes (mock data for now)
		mailboxes := getMockMailboxes()
		
		// Get mail previews for the selected mailbox (mock data for now)
		mailPreviews := getMockMailPreviews(mailboxID)
		
		// Find the active mailbox
		var activeMailbox Mailbox
		for _, mb := range mailboxes {
			if mb.ID == mailboxID {
				activeMailbox = mb
				break
			}
		}
		
		// Render the layout template with the mailbox and mail previews
		return c.Render("layout", fiber.Map{
			"mailboxes":     mailboxes,
			"mailPreviews":  mailPreviews,
			"activeMailbox": activeMailbox,
			"activeMail":    nil,
			"title":         m.Config.Title,
			"view":          "mailbox",
		})
	})

	// Mail view route - displays a single email
	m.App.Get("/mail/:id", func(c *fiber.Ctx) error {
		mailID := c.Params("id")
		
		// Get mailboxes (mock data for now)
		mailboxes := getMockMailboxes()
		
		// Get the mail (mock data for now)
		mail := getMockMail(mailID)
		
		// Get mail previews for the selected mailbox (mock data for now)
		mailPreviews := getMockMailPreviews(mail.MailboxID)
		
		// Find the active mailbox
		var activeMailbox Mailbox
		for _, mb := range mailboxes {
			if mb.ID == mail.MailboxID {
				activeMailbox = mb
				break
			}
		}
		
		// Render the layout template with the mail content
		return c.Render("layout", fiber.Map{
			"mailboxes":     mailboxes,
			"mailPreviews":  mailPreviews,
			"activeMailbox": activeMailbox,
			"activeMail":    mail,
			"title":         m.Config.Title,
			"view":          "mail",
		})
	})

	// API endpoints for AJAX requests
	
	// Get mailboxes
	m.App.Get("/api/mailboxes", func(c *fiber.Ctx) error {
		mailboxes := getMockMailboxes()
		return c.JSON(mailboxes)
	})
	
	// Get mail previews for a mailbox
	m.App.Get("/api/mailbox/:id/mails", func(c *fiber.Ctx) error {
		mailboxID := c.Params("id")
		mailPreviews := getMockMailPreviews(mailboxID)
		return c.JSON(mailPreviews)
	})
	
	// Get a single mail
	m.App.Get("/api/mail/:id", func(c *fiber.Ctx) error {
		mailID := c.Params("id")
		mail := getMockMail(mailID)
		return c.JSON(mail)
	})
}

// Mock data functions

// getMockMailboxes returns a list of mock mailboxes
func getMockMailboxes() []Mailbox {
	return []Mailbox{
		{ID: "inbox", Name: "Inbox", Count: 5, Type: "inbox"},
		{ID: "sent", Name: "Sent", Count: 2, Type: "sent"},
		{ID: "drafts", Name: "Drafts", Count: 1, Type: "drafts"},
		{ID: "trash", Name: "Trash", Count: 3, Type: "trash"},
		{ID: "spam", Name: "Spam", Count: 10, Type: "spam"},
	}
}

// getMockMailPreviews returns a list of mock mail previews for a mailbox
func getMockMailPreviews(mailboxID string) []MailPreview {
	switch mailboxID {
	case "inbox":
		return []MailPreview{
			{ID: "1", Subject: "Welcome to Mail", From: "admin@example.com", Date: "2025-03-28", Preview: "Welcome to the new mail system! We hope you enjoy...", Read: false, Starred: true, HasAttach: false, MailboxID: "inbox"},
			{ID: "2", Subject: "Meeting Tomorrow", From: "manager@example.com", Date: "2025-03-27", Preview: "Don't forget our team meeting tomorrow at 10 AM...", Read: true, Starred: false, HasAttach: true, MailboxID: "inbox"},
			{ID: "3", Subject: "Project Update", From: "team@example.com", Date: "2025-03-26", Preview: "Here's the latest update on our project status...", Read: true, Starred: false, HasAttach: false, MailboxID: "inbox"},
			{ID: "4", Subject: "Vacation Request", From: "hr@example.com", Date: "2025-03-25", Preview: "Your vacation request has been approved...", Read: false, Starred: false, HasAttach: false, MailboxID: "inbox"},
			{ID: "5", Subject: "New Feature Release", From: "product@example.com", Date: "2025-03-24", Preview: "We're excited to announce our new feature release...", Read: false, Starred: false, HasAttach: true, MailboxID: "inbox"},
		}
	case "sent":
		return []MailPreview{
			{ID: "6", Subject: "Re: Project Timeline", From: "you@example.com", Date: "2025-03-27", Preview: "I've reviewed the timeline and have some suggestions...", Read: true, Starred: false, HasAttach: false, MailboxID: "sent"},
			{ID: "7", Subject: "Vacation Request", From: "you@example.com", Date: "2025-03-25", Preview: "I'd like to request vacation days for next month...", Read: true, Starred: false, HasAttach: false, MailboxID: "sent"},
		}
	case "drafts":
		return []MailPreview{
			{ID: "8", Subject: "Draft: Quarterly Report", From: "you@example.com", Date: "2025-03-26", Preview: "Here's the quarterly report for Q1 2025...", Read: true, Starred: false, HasAttach: false, MailboxID: "drafts"},
		}
	case "trash":
		return []MailPreview{
			{ID: "9", Subject: "Old Newsletter", From: "news@example.com", Date: "2025-03-20", Preview: "Check out our latest newsletter with updates...", Read: true, Starred: false, HasAttach: false, MailboxID: "trash"},
			{ID: "10", Subject: "Expired Offer", From: "sales@example.com", Date: "2025-03-19", Preview: "Limited time offer expires today!...", Read: true, Starred: false, HasAttach: false, MailboxID: "trash"},
			{ID: "11", Subject: "Outdated Information", From: "info@example.com", Date: "2025-03-18", Preview: "Please review this outdated information...", Read: true, Starred: false, HasAttach: false, MailboxID: "trash"},
		}
	case "spam":
		return []MailPreview{
			{ID: "12", Subject: "You Won a Prize!", From: "prize@example.com", Date: "2025-03-28", Preview: "Congratulations! You've been selected as a winner...", Read: false, Starred: false, HasAttach: false, MailboxID: "spam"},
			// Add more spam emails as needed
		}
	default:
		return []MailPreview{}
	}
}

// getMockMail returns a mock mail for the given ID
func getMockMail(mailID string) Mail {
	switch mailID {
	case "1":
		return Mail{
			ID:        "1",
			Subject:   "Welcome to Mail",
			From:      "admin@example.com",
			To:        []string{"user@example.com"},
			Date:      "2025-03-28",
			Body:      "Welcome to the new mail system! We hope you enjoy using it. This is a simple mail client built with Go, Fiber, and Pug templates.",
			BodyHTML:  "<p>Welcome to the new mail system! We hope you enjoy using it.</p><p>This is a simple mail client built with Go, Fiber, and Pug templates.</p>",
			Read:      true,
			Starred:   true,
			MailboxID: "inbox",
		}
	case "2":
		return Mail{
			ID:          "2",
			Subject:     "Meeting Tomorrow",
			From:        "manager@example.com",
			To:          []string{"user@example.com"},
			Cc:          []string{"team@example.com"},
			Date:        "2025-03-27",
			Body:        "Don't forget our team meeting tomorrow at 10 AM. Please prepare a short update on your current projects. The meeting will be held in the main conference room.",
			BodyHTML:    "<p>Don't forget our team meeting tomorrow at 10 AM.</p><p>Please prepare a short update on your current projects.</p><p>The meeting will be held in the main conference room.</p>",
			Attachments: []string{"meeting_agenda.pdf"},
			Read:        true,
			Starred:     false,
			MailboxID:   "inbox",
		}
	// Add more cases for other mail IDs
	default:
		return Mail{
			ID:        mailID,
			Subject:   "Unknown Mail",
			From:      "unknown@example.com",
			To:        []string{"user@example.com"},
			Date:      "2025-03-28",
			Body:      "This mail does not exist or has been deleted.",
			BodyHTML:  "<p>This mail does not exist or has been deleted.</p>",
			Read:      true,
			Starred:   false,
			MailboxID: "inbox",
		}
	}
}
