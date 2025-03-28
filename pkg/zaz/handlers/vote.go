package handlers

import (
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// VoteHandler handles vote-related routes
type VoteHandler struct {}

// NewVoteHandler creates a new VoteHandler
func NewVoteHandler(_ *models.Store) *VoteHandler {
	return &VoteHandler{}
}

// GetVotes renders the votes list page
func (h *VoteHandler) GetVotes(c *fiber.Ctx) error {
	// Get votes using model function directly
	votes := models.GetAllVotes()

	return c.Render("votes", fiber.Map{
		"title": "Votes",
		"votes": votes,
		"search": c.Query("search"),
	})
}

// GetCreateVote renders the vote creation page
func (h *VoteHandler) GetCreateVote(c *fiber.Ctx) error {
	// Get list of companies using model function directly
	companies := models.GetAllCompanies()

	return c.Render("votes_create", fiber.Map{
		"title": "Create Vote",
		"companies": companies,
	})
}

// PostCreateVote handles vote creation form submission
func (h *VoteHandler) PostCreateVote(c *fiber.Ctx) error {
	// Parse form data
	title := c.FormValue("title")
	companyID := c.FormValue("company_id")
	startDate := c.FormValue("start_date")
	endDate := c.FormValue("end_date")
	
	// Simple validation
	if title == "" || companyID == "" || startDate == "" || endDate == "" {
		return c.Render("votes_create", fiber.Map{
			"title": "Create Vote",
			"error": "All fields are required",
		})
	}

	// TODO: Implement actual vote creation
	// For now, just redirect to votes list
	return c.Redirect("/votes")
}

// GetVoteDetails renders the vote details page
func (h *VoteHandler) GetVoteDetails(c *fiber.Ctx) error {
	_ = c.Params("id") // Use the parameter to avoid unused variable warning
	
	// In a real implementation, we would fetch the vote from the database
	vote := models.Vote{
		ID: 1,
		CompanyID: 1,
		Title: "Board Member Election",
		Description: "Election of a new board member to replace retiring member John Smith.",
		StartDate: time.Now().AddDate(0, 0, -10),
		EndDate: time.Now().AddDate(0, 0, -3),
		Status: "Closed",
		Options: []models.VoteOption{
			{
				ID: 1,
				VoteID: 1,
				Text: "Jane Doe",
				Count: 1200,
			},
			{
				ID: 2,
				VoteID: 1,
				Text: "Bob Johnson",
				Count: 800,
			},
			{
				ID: 3,
				VoteID: 1,
				Text: "Sarah Williams",
				Count: 600,
			},
		},
	}

	// Get company info
	company := models.Company{
		ID: 1,
		Name: "TechCorp Inc.",
	}

	return c.Render("vote_details", fiber.Map{
		"title": vote.Title,
		"vote": vote,
		"company": company,
	})
}

// GetVoteResults renders the vote results page
func (h *VoteHandler) GetVoteResults(c *fiber.Ctx) error {
	// Sample completed votes for demonstration
	votes := []models.Vote{
		{
			ID: 1,
			CompanyID: 1,
			Title: "Board Member Election",
			StartDate: time.Now().AddDate(0, 0, -10),
			EndDate: time.Now().AddDate(0, 0, -3),
			Status: "Closed",
			Options: []models.VoteOption{
				{
					ID: 1,
					VoteID: 1,
					Text: "Jane Doe",
					Count: 1200,
				},
				{
					ID: 2,
					VoteID: 1,
					Text: "Bob Johnson",
					Count: 800,
				},
				{
					ID: 3,
					VoteID: 1,
					Text: "Sarah Williams",
					Count: 600,
				},
			},
		},
		{
			ID: 4,
			CompanyID: 2,
			Title: "Budget Approval",
			StartDate: time.Now().AddDate(0, -1, 0),
			EndDate: time.Now().AddDate(0, 0, -15),
			Status: "Closed",
			Options: []models.VoteOption{
				{
					ID: 4,
					VoteID: 4,
					Text: "Approve",
					Count: 2500,
				},
				{
					ID: 5,
					VoteID: 4,
					Text: "Reject",
					Count: 500,
				},
			},
		},
	}

	return c.Render("vote_results", fiber.Map{
		"title": "Vote Results",
		"votes": votes,
	})
}

// GetVotesAPI returns votes data as JSON for API consumption
func (h *VoteHandler) GetVotesAPI(c *fiber.Ctx) error {
	// Sample votes for demonstration
	votes := []models.Vote{
		{
			ID: 1,
			CompanyID: 1,
			Title: "Board Member Election",
			StartDate: time.Now().AddDate(0, 0, -10),
			EndDate: time.Now().AddDate(0, 0, -3),
			Status: "Closed",
		},
		{
			ID: 2,
			CompanyID: 1,
			Title: "New Investment Approval",
			StartDate: time.Now().AddDate(0, 0, -5),
			EndDate: time.Now().AddDate(0, 0, 5),
			Status: "Open",
		},
		{
			ID: 3,
			CompanyID: 2,
			Title: "Company Name Change",
			StartDate: time.Now().AddDate(0, 0, 2),
			EndDate: time.Now().AddDate(0, 0, 12),
			Status: "Open",
		},
	}

	// Filter by company if specified
	companyID := c.Query("company_id")
	if companyID != "" {
		var filtered []models.Vote
		for _, v := range votes {
			if v.CompanyID == 1 { // This would actually use the companyID param
				filtered = append(filtered, v)
			}
		}
		votes = filtered
	}

	// Filter by status if specified
	status := c.Query("status")
	if status != "" {
		var filtered []models.Vote
		for _, v := range votes {
			if v.Status == status {
				filtered = append(filtered, v)
			}
		}
		votes = filtered
	}

	return c.JSON(fiber.Map{
		"success": true,
		"votes": votes,
	})
}
