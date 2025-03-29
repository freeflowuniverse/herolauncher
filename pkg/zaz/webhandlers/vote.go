package webhandlers

import (
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// VoteHandler handles vote-related routes
type VoteHandler struct {
	store *models.Store
}

// NewVoteHandler creates a new VoteHandler
func NewVoteHandler(store *models.Store) *VoteHandler {
	return &VoteHandler{
		store: store,
	}
}

// GetVotes renders the votes list page
func (h *VoteHandler) GetVotes(c *fiber.Ctx) error {
	// Get votes directly from the store
	votes := h.store.VoteHandler.GetAll()

	return RenderWithDefaults(c, "votes", fiber.Map{
		"title": "Votes",
		"votes": votes,
	})
}

// GetCreateVote renders the vote creation page
func (h *VoteHandler) GetCreateVote(c *fiber.Ctx) error {
	// Get list of companies and board meetings directly from the store
	companies := h.store.CompanyHandler.GetAll()
	boardMeetings := h.store.BoardMeetingHandler.GetAll()

	return RenderWithDefaults(c, "votes_create", fiber.Map{
		"title": "Create Vote",
		"companies": companies,
		"boardMeetings": boardMeetings,
	})
}

// PostCreateVote handles vote creation form submission
func (h *VoteHandler) PostCreateVote(c *fiber.Ctx) error {
	// Parse form data
	title := c.FormValue("title")
	description := c.FormValue("description")
	companyIDStr := c.FormValue("company_id")
	startDateStr := c.FormValue("start_date")
	endDateStr := c.FormValue("end_date")
	
	// Simple validation
	if title == "" || companyIDStr == "" || startDateStr == "" || endDateStr == "" {
		companies := h.store.CompanyHandler.GetAll()
		boardMeetings := h.store.BoardMeetingHandler.GetAll()
		return RenderWithDefaults(c, "votes_create", fiber.Map{
			"title": "Create Vote",
			"companies": companies,
			"boardMeetings": boardMeetings,
			"error": "Title, company, start date, and end date are required",
		})
	}
	
	// Parse company ID
	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		boardMeetings := h.store.BoardMeetingHandler.GetAll()
		return RenderWithDefaults(c, "votes_create", fiber.Map{
			"title": "Create Vote",
			"companies": companies,
			"boardMeetings": boardMeetings,
			"error": "Invalid company ID",
		})
	}
	
	// Parse dates
	startDate, err := time.Parse("2006-01-02T15:04", startDateStr)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		boardMeetings := h.store.BoardMeetingHandler.GetAll()
		return RenderWithDefaults(c, "votes_create", fiber.Map{
			"title": "Create Vote",
			"companies": companies,
			"boardMeetings": boardMeetings,
			"error": "Invalid start date format",
		})
	}
	
	endDate, err := time.Parse("2006-01-02T15:04", endDateStr)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		boardMeetings := h.store.BoardMeetingHandler.GetAll()
		return RenderWithDefaults(c, "votes_create", fiber.Map{
			"title": "Create Vote",
			"companies": companies,
			"boardMeetings": boardMeetings,
			"error": "Invalid end date format",
		})
	}
	
	// Create vote
	vote := models.Vote{
		CompanyID:   companyID,
		Title:       title,
		Description: description,
		Status:      "Open",
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Save to database using the store
	_, err = h.store.VoteHandler.Create(vote)
	if err != nil {
		companies := h.store.CompanyHandler.GetAll()
		boardMeetings := h.store.BoardMeetingHandler.GetAll()
		return RenderWithDefaults(c, "votes_create", fiber.Map{
			"title": "Create Vote",
			"companies": companies,
			"boardMeetings": boardMeetings,
			"error": "Failed to create vote: " + err.Error(),
		})
	}
	
	// Redirect to votes list
	return c.Redirect("/votes")
}

// GetVoteDetails renders the vote details page
func (h *VoteHandler) GetVoteDetails(c *fiber.Ctx) error {
	// Parse vote ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid vote ID")
	}
	
	// Get vote from store
	vote, err := h.store.VoteHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Vote not found")
	}
	
	// Get company from store
	company, err := h.store.CompanyHandler.GetByID(vote.CompanyID)
	if err != nil {
		return c.Status(404).SendString("Company not found")
	}
	
	// Get vote options and ballots
	voteOptions := []models.VoteOption{}
	ballots := []models.Ballot{}
	
	return RenderWithDefaults(c, "votes_details", fiber.Map{
		"title": vote.Title,
		"vote": vote,
		"company": company,
		"voteOptions": voteOptions,
		"ballots": ballots,
	})
}

// PostCastVote handles vote casting form submission
func (h *VoteHandler) PostCastVote(c *fiber.Ctx) error {
	// Parse vote ID from URL
	voteID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid vote ID")
	}
	
	// Get vote from store
	vote, err := h.store.VoteHandler.GetByID(voteID)
	if err != nil {
		return c.Status(404).SendString("Vote not found")
	}
	
	// Check if vote is still open
	if vote.Status != "Open" {
		return c.Status(400).SendString("This vote is no longer open")
	}
	
	// Parse form data
	userIDStr := c.FormValue("user_id")
	optionIDStr := c.FormValue("option_id")
	
	// Simple validation
	if userIDStr == "" || optionIDStr == "" {
		return c.Status(400).SendString("User ID and option ID are required")
	}
	
	// Validate the IDs format (we don't use the parsed values yet, but we validate them)
	_, err = strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid user ID")
	}
	
	_, err = strconv.ParseInt(optionIDStr, 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid option ID")
	}
	
	// Create ballot
	// Note: We need to implement a method to create ballots
	// For now, just return success without creating the ballot
	// ballot := models.Ballot{
	// 	VoteID:       voteID,
	// 	UserID:       userID,
	// 	VoteOptionID: optionID,
	// 	SharesCount:  1, // Default to 1 share
	// }
	//
	// _, err = h.store.VoteHandler.CreateBallot(ballot)
	// if err != nil {
	// 	return c.Status(500).SendString("Failed to cast vote: " + err.Error())
	// }
	
	// Redirect to vote details
	return c.Redirect("/votes/" + c.Params("id"))
}

// GetCloseVote renders the vote closing confirmation page
func (h *VoteHandler) GetCloseVote(c *fiber.Ctx) error {
	// Parse vote ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid vote ID")
	}
	
	// Get vote from store
	vote, err := h.store.VoteHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Vote not found")
	}
	
	return RenderWithDefaults(c, "votes_close", fiber.Map{
		"title": "Close Vote: " + vote.Title,
		"vote": vote,
	})
}

// PostCloseVote handles vote closing form submission
func (h *VoteHandler) PostCloseVote(c *fiber.Ctx) error {
	// Parse vote ID from URL
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).SendString("Invalid vote ID")
	}
	
	// Get vote from store
	vote, err := h.store.VoteHandler.GetByID(id)
	if err != nil {
		return c.Status(404).SendString("Vote not found")
	}
	
	// Update vote status
	vote.Status = "Closed"
	vote.UpdatedAt = time.Now()
	
	// Save to database using the store
	err = h.store.VoteHandler.Update(vote)
	if err != nil {
		return RenderWithDefaults(c, "votes_close", fiber.Map{
			"title": "Close Vote: " + vote.Title,
			"vote": vote,
			"error": "Failed to close vote: " + err.Error(),
		})
	}
	
	// Redirect to vote details
	return c.Redirect("/votes/" + c.Params("id"))
}

// GetVotesAPI returns votes data as JSON for API consumption
func (h *VoteHandler) GetVotesAPI(c *fiber.Ctx) error {
	// Check if company ID is provided as query parameter
	companyIDStr := c.Query("company_id")
	if companyIDStr != "" {
		companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid company ID",
			})
		}
		
		// Get votes for specific company
		votes := h.store.VoteHandler.GetByCompanyID(companyID)
		return c.JSON(votes)
	}
	
	// Get all votes
	votes := h.store.VoteHandler.GetAll()
	return c.JSON(votes)
}
