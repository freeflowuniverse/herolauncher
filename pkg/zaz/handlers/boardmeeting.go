package handlers

import (
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// BoardMeetingHandler handles board meeting-related routes
type BoardMeetingHandler struct {}

// NewBoardMeetingHandler creates a new BoardMeetingHandler
func NewBoardMeetingHandler(_ *models.Store) *BoardMeetingHandler {
	return &BoardMeetingHandler{}
}

// GetBoardMeetings renders the board meetings list page
func (h *BoardMeetingHandler) GetBoardMeetings(c *fiber.Ctx) error {
	// Get meetings using model function directly
	meetings := models.GetAllBoardMeetings()

	return c.Render("boardmeetings", fiber.Map{
		"title": "Board Meetings",
		"meetings": meetings,
		"search": c.Query("search"),
	})
}

// GetCreateBoardMeeting renders the board meeting creation page
func (h *BoardMeetingHandler) GetCreateBoardMeeting(c *fiber.Ctx) error {
	// Get list of companies using model function directly
	companies := models.GetAllCompanies()

	return c.Render("boardmeetings_create", fiber.Map{
		"title": "Schedule Board Meeting",
		"companies": companies,
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
		// Get companies for the form
		companies := models.GetAllCompanies()
		return c.Render("boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Title, company, and date are required",
		})
	}

	// Parse company ID
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid company ID")
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// Get companies for the form
		companies := models.GetAllCompanies()
		return c.Render("boardmeetings_create", fiber.Map{
			"title": "Schedule Board Meeting",
			"companies": companies,
			"error": "Invalid date format",
		})
	}

	// Create a new board meeting
	meeting := models.BoardMeeting{
		ID:          int64(len(models.GetAllBoardMeetings()) + 1),
		CompanyID:   companyID,
		Title:       title,
		Date:        date,
		Location:    location,
		Description: description,
		Status:      "Scheduled",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add the meeting using model function directly
	models.AddBoardMeeting(meeting)

	return c.Redirect("/boardmeetings")
}

// GetBoardMeetingDetails renders the board meeting details page
func (h *BoardMeetingHandler) GetBoardMeetingDetails(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	
	// Fetch the meeting using model function directly
	meeting, err := models.GetBoardMeetingByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Meeting not found")
	}

	// Get company info using model function directly
	company, err := models.GetCompanyByID(meeting.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Company not found")
	}

	return c.Render("boardmeeting_details", fiber.Map{
		"title": meeting.Title,
		"meeting": meeting,
		"company": company,
	})
}

// GetBoardMeetingMinutes renders the board meeting minutes page
func (h *BoardMeetingHandler) GetBoardMeetingMinutes(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	
	// Fetch the meeting using model function directly
	meeting, err := models.GetBoardMeetingByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Meeting not found")
	}

	// Get company info using model function directly
	company, err := models.GetCompanyByID(meeting.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Company not found")
	}

	return c.Render("boardmeeting_minutes", fiber.Map{
		"title": meeting.Title + " - Minutes",
		"meeting": meeting,
		"company": company,
	})
}

// PostBoardMeetingMinutes handles board meeting minutes form submission
func (h *BoardMeetingHandler) PostBoardMeetingMinutes(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	
	// Parse form data
	minutes := c.FormValue("minutes")
	
	// Simple validation
	if minutes == "" {
		return c.Render("boardmeeting_minutes", fiber.Map{
			"title": "Meeting Minutes",
			"error": "Minutes cannot be empty",
		})
	}

	// Fetch the meeting using model function directly
	meeting, err := models.GetBoardMeetingByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Meeting not found")
	}
	
	// Update the minutes
	meeting.Minutes = minutes
	meeting.Status = "Completed"
	
	// Save the updated meeting using model function directly
	err = models.UpdateBoardMeeting(meeting)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update meeting")
	}
	
	return c.Redirect("/boardmeetings/" + idParam)
}

// GetBoardMeetingsAPI returns board meetings data as JSON for API consumption
func (h *BoardMeetingHandler) GetBoardMeetingsAPI(c *fiber.Ctx) error {
	// Sample board meetings for demonstration
	meetings := []models.BoardMeeting{
		{
			ID: 1,
			CompanyID: 1,
			Title: "Q2 Financial Review",
			Date: time.Now().AddDate(0, 1, 0),
			Location: "Virtual (Zoom)",
			Status: "Scheduled",
		},
		{
			ID: 2,
			CompanyID: 1,
			Title: "Annual General Meeting",
			Date: time.Now().AddDate(0, -1, 0),
			Location: "Company HQ",
			Status: "Completed",
		},
		{
			ID: 3,
			CompanyID: 2,
			Title: "Strategic Planning Session",
			Date: time.Now().AddDate(0, 0, 15),
			Location: "Conference Center",
			Status: "Scheduled",
		},
	}

	// Filter by company if specified
	companyID := c.Query("company_id")
	if companyID != "" {
		var filtered []models.BoardMeeting
		for _, m := range meetings {
			if m.CompanyID == 1 { // This would actually use the companyID param
				filtered = append(filtered, m)
			}
		}
		meetings = filtered
	}

	return c.JSON(fiber.Map{
		"success": true,
		"meetings": meetings,
	})
}
