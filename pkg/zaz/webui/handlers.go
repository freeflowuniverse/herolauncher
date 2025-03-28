package webui

import (
	"github.com/gofiber/fiber/v2"
)

// AuthHandler interface defines methods for authentication-related routes
type AuthHandler interface {
	GetLogin(c *fiber.Ctx) error
	PostLogin(c *fiber.Ctx) error
	GetRegister(c *fiber.Ctx) error
	PostRegister(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	GetForgotPassword(c *fiber.Ctx) error
	PostForgotPassword(c *fiber.Ctx) error
}

// CompanyHandler interface defines methods for company-related routes
type CompanyHandler interface {
	GetDashboard(c *fiber.Ctx) error
	GetCompanies(c *fiber.Ctx) error
	GetCreateCompany(c *fiber.Ctx) error
	PostCreateCompany(c *fiber.Ctx) error
	GetCompanyDetails(c *fiber.Ctx) error
	GetEditCompany(c *fiber.Ctx) error
	PostEditCompany(c *fiber.Ctx) error
	GetCompaniesAPI(c *fiber.Ctx) error
	GetCompanyDetailsAPI(c *fiber.Ctx) error
}

// ShareholderHandler interface defines methods for shareholder-related routes
type ShareholderHandler interface {
	GetShareholders(c *fiber.Ctx) error
	GetCreateShareholder(c *fiber.Ctx) error
	PostCreateShareholder(c *fiber.Ctx) error
	GetAddShareholder(c *fiber.Ctx) error
	PostAddShareholder(c *fiber.Ctx) error
	GetShareholdersAPI(c *fiber.Ctx) error
}

// BoardMeetingHandler interface defines methods for board meeting-related routes
type BoardMeetingHandler interface {
	GetBoardMeetings(c *fiber.Ctx) error
	GetCreateBoardMeeting(c *fiber.Ctx) error
	PostCreateBoardMeeting(c *fiber.Ctx) error
	GetBoardMeetingDetails(c *fiber.Ctx) error
	GetBoardMeetingMinutes(c *fiber.Ctx) error
	PostBoardMeetingMinutes(c *fiber.Ctx) error
	GetBoardMeetingsAPI(c *fiber.Ctx) error
}

// VoteHandler interface defines methods for vote-related routes
type VoteHandler interface {
	GetVotes(c *fiber.Ctx) error
	GetCreateVote(c *fiber.Ctx) error
	PostCreateVote(c *fiber.Ctx) error
	GetVoteDetails(c *fiber.Ctx) error
	GetVoteResults(c *fiber.Ctx) error
	GetVotesAPI(c *fiber.Ctx) error
}

// SaleHandler interface defines methods for sale-related routes
type SaleHandler interface {
	GetSales(c *fiber.Ctx) error
	GetProducts(c *fiber.Ctx) error
	GetServices(c *fiber.Ctx) error
	GetCreateSale(c *fiber.Ctx) error
	PostCreateSale(c *fiber.Ctx) error
	GetSaleDetails(c *fiber.Ctx) error
	GetSalesReports(c *fiber.Ctx) error
	GetSalesAPI(c *fiber.Ctx) error
}
