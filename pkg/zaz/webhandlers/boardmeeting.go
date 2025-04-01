package webhandlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// BoardMeetingHandler handles board meeting-related routes
type BoardMeetingHandler struct {
	store *models.Store
}

// NewBoardMeetingHandler creates a new BoardMeetingHandler
func NewBoardMeetingHandler(store *models.Store) *BoardMeetingHandler {
	return &BoardMeetingHandler{
		store: store,
	}
}

// GetBoardMeetings renders the board meetings list page
func (h *BoardMeetingHandler) GetBoardMeetings(c *fiber.Ctx) error {
	// Get board meetings directly from the store and preload company information
	meetings := h.store.BoardMeetingHandler.GetAll()
	
	// Preload company information for each meeting
	for i := range meetings {
		if company, err := h.store.CompanyHandler.GetByID(meetings[i].CompanyID); err == nil {
			meetings[i].Company = company
		}
	}

	// Handle potential rendering errors
	err := RenderWithDefaults(c, "boardmeetings", fiber.Map{
		"title": "Board Meetings",
		"boardMeetings": meetings,
		"search": c.Query("search"),
		"currentYear": time.Now().Year(),
		"formErrors": nil,
		"csrfToken": GetCSRFToken(c),
	})
	
	if err != nil {
		// Log the error
		fmt.Printf("Error rendering boardmeetings template: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	
	return nil
}

// GetCreateBoardMeeting renders the board meeting creation page
func (h *BoardMeetingHandler) GetCreateBoardMeeting(c *fiber.Ctx) error {
	// Get list of companies directly from the store
	companies := h.store.CompanyHandler.GetAll()

	return RenderWithDefaults(c, "boardmeetings_create", fiber.Map{
		"title": "Schedule Board Meeting",
		"companies": companies,
		"currentYear": time.Now().Year(),
		"formErrors": nil,
		"csrfToken": GetCSRFToken(c),
		"form": fiber.Map{},
	})
}

// PostCreateBoardMeeting handles board meeting creation form submission
func (h *BoardMeetingHandler) PostCreateBoardMeeting(c *fiber.Ctx) error {
	// Parse form data
	title := c.FormValue("title")
	companyIDStr := c.FormValue("company_id")
	dateStr := c.FormValue("date")
	location := c.FormValue("location")
	description := c.FormValue("description")
	
	// Simple validation
	if title == "" || companyIDStr == "" || dateStr == "" {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Title, company, and date are required",
		})
	}
	
	// Parse company ID
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Invalid company ID",
		})
	}
	
	// Parse date
	date, err := time.Parse("2006-01-02T15:04", dateStr)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Invalid date format",
		})
	}
	
	// Create board meeting
	meeting := models.BoardMeeting{
		CompanyID:   companyID,
		Title:       title,
		Date:        date,
		Location:    location,
		Description: description,
		Status:      "Scheduled",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Save to database using the store
	_, err = h.store.BoardMeetingHandler.Create(meeting)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		return RenderWithDefaults(c, "boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Failed to create board meeting: " + err.Error(),
		})
	}
	
	// Redirect to board meetings list
	return c.Redirect("/boardmeetings")
}

// GetBoardMeetingDetails renders the board meeting details page
func (h *BoardMeetingHandler) GetBoardMeetingDetails(c *fiber.Ctx) error {
	// Parse board meeting ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid board meeting ID")
	}
	
	// Get board meeting from store
	meeting, err := h.store.BoardMeetingHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Board meeting not found")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(meeting.CompanyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	return RenderWithDefaults(c, "boardmeetings_details", fiber.Map{
		"title": meeting.Title,
		"meeting": meeting,
		"company": company,
	})
}

// GetBoardMeetingMinutes renders the board meeting minutes page
func (h *BoardMeetingHandler) GetBoardMeetingMinutes(c *fiber.Ctx) error {
	// Parse board meeting ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid board meeting ID")
	}
	
	// Get board meeting from store
	meeting, err := h.store.BoardMeetingHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Board meeting not found")
	}
	
	return RenderWithDefaults(c, "boardmeetings_minutes", fiber.Map{
		"title": meeting.Title + " - Minutes",
		"meeting": meeting,
	})
}

// PostBoardMeetingMinutes handles board meeting minutes form submission
func (h *BoardMeetingHandler) PostBoardMeetingMinutes(c *fiber.Ctx) error {
	// Parse board meeting ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid board meeting ID")
	}
	
	// Get board meeting from store
	meeting, err := h.store.BoardMeetingHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Board meeting not found")
	}
	
	// Update minutes
	meeting.Minutes = c.FormValue("minutes")
	meeting.Status = "Completed"
	meeting.UpdatedAt = time.Now()
	
	// Save to database using the store
	err = h.store.BoardMeetingHandler.Update(meeting)
	if err != nil {
		return RenderWithDefaults(c, "boardmeetings_minutes", fiber.Map{
			"title": meeting.Title + " - Minutes",
			"meeting": meeting,
			"error": "Failed to update minutes: " + err.Error(),
		})
	}
	
	// Redirect to board meeting details
	return c.Redirect("/boardmeetings/" + c.Params("id"))
}

// GetBoardMeetingsAPI returns board meetings data as JSON for API consumption
func (h *BoardMeetingHandler) GetBoardMeetingsAPI(c *fiber.Ctx) error {
	// Check if company ID is provided as query parameter
	companyIDStr := c.Query("company_id")
	if companyIDStr != "" {
		companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid company ID",
			})
		}
		
		// Get board meetings for specific company
		meetings := h.store.BoardMeetingHandler.GetByCompanyID(companyID)
		return c.JSON(meetings)
	}
	
	// Get all board meetings
	meetings := h.store.BoardMeetingHandler.GetAll()
	return c.JSON(meetings)
}
