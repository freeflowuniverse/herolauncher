package models

import (
	"errors"
	"time"
)

// VoteHandler handles all operations related to Vote model
type VoteHandler struct {
	circleID string
	db       *Db
}

// NewVoteHandler creates a new VoteHandler
func NewVoteHandler(circleID string, db *Db) *VoteHandler {
	return &VoteHandler{
		circleID: circleID,
		db:       db,
	}
}

// Vote represents a voting item in the Freezone
type Vote struct {
	ID          int64        `gorm:"primaryKey" json:"id"`
	CompanyID   int64        `gorm:"index;not null" json:"company_id"`
	Title       string       `gorm:"not null" json:"title"`
	Description string       `json:"description"`
	StartDate   time.Time    `gorm:"not null" json:"start_date"`
	EndDate     time.Time    `gorm:"not null" json:"end_date"`
	Status      string       `gorm:"not null" json:"status"` // Open, Closed, Cancelled
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	Options     []VoteOption `gorm:"foreignKey:VoteID" json:"options,omitempty"`
	Ballots     []Ballot     `gorm:"foreignKey:VoteID" json:"ballots,omitempty"`
	Company     Company      `gorm:"foreignKey:CompanyID" json:"-"`
}

// VoteOption represents an option in a vote
type VoteOption struct {
	ID     int64  `gorm:"primaryKey" json:"id"`
	VoteID int64  `gorm:"index;not null" json:"vote_id"`
	Text   string `gorm:"not null" json:"text"`
	Count  int    `gorm:"default:0" json:"count"`
	Vote   Vote   `gorm:"foreignKey:VoteID" json:"-"`
}

// Ballot represents a cast ballot in a vote
type Ballot struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	VoteID       int64      `gorm:"index;not null" json:"vote_id"`
	UserID       int64      `gorm:"index;not null" json:"user_id"`
	VoteOptionID int64      `gorm:"index;not null" json:"vote_option_id"`
	SharesCount  int        `gorm:"not null" json:"shares_count"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	Vote         Vote       `gorm:"foreignKey:VoteID" json:"-"`
	User         User       `gorm:"foreignKey:UserID" json:"-"`
	VoteOption   VoteOption `gorm:"foreignKey:VoteOptionID" json:"-"`
}

// GetAll returns all votes
func (h *VoteHandler) GetAll() []Vote {
	var votes []Vote
	GetDB().Preload("Options").Preload("Ballots").Find(&votes)
	return votes
}

// GetByID returns a vote by ID
func (h *VoteHandler) GetByID(id int64) (Vote, error) {
	var vote Vote
	result := GetDB().Preload("Options").Preload("Ballots").First(&vote, id)
	if result.Error != nil {
		return Vote{}, ErrRecordNotFound
	}
	return vote, nil
}

// GetByCompanyID returns all votes for a company
func (h *VoteHandler) GetByCompanyID(companyID int64) []Vote {
	var votes []Vote
	GetDB().Preload("Options").Preload("Ballots").Where("company_id = ?", companyID).Find(&votes)
	return votes
}

// Create adds a new vote
func (h *VoteHandler) Create(vote Vote) (int64, error) {
	if vote.CreatedAt.IsZero() {
		vote.CreatedAt = time.Now()
	}
	if vote.UpdatedAt.IsZero() {
		vote.UpdatedAt = time.Now()
	}

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// Create vote
	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return vote.ID, nil
}

// Update updates an existing vote
func (h *VoteHandler) Update(vote Vote) error {
	vote.UpdatedAt = time.Now()

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if vote exists
	var existingVote Vote
	if err := tx.First(&existingVote, vote.ID).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
	}

	// Update vote
	if err := tx.Model(&vote).Updates(map[string]interface{}{
		"company_id":  vote.CompanyID,
		"title":       vote.Title,
		"description": vote.Description,
		"start_date":  vote.StartDate,
		"end_date":    vote.EndDate,
		"status":      vote.Status,
		"updated_at":  vote.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update options (delete and re-insert)
	if err := tx.Where("vote_id = ?", vote.ID).Delete(&VoteOption{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Insert new options
	for i := range vote.Options {
		vote.Options[i].VoteID = vote.ID
		if err := tx.Create(&vote.Options[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

// Delete deletes a vote and all associated options and ballots
func (h *VoteHandler) Delete(id int64) error {
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if vote exists
	var vote Vote
	if err := tx.First(&vote, id).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
	}

	// Delete ballots
	if err := tx.Where("vote_id = ?", id).Delete(&Ballot{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete options
	if err := tx.Where("vote_id = ?", id).Delete(&VoteOption{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete vote
	if err := tx.Delete(&Vote{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// CastBallot adds a new ballot to a vote
func (h *VoteHandler) CastBallot(ballot Ballot) error {
	if ballot.CreatedAt.IsZero() {
		ballot.CreatedAt = time.Now()
	}

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if vote exists and is open
	var vote Vote
	if err := tx.First(&vote, ballot.VoteID).Error; err != nil {
		tx.Rollback()
		return errors.New("vote not found")
	}

	if vote.Status != "Open" {
		tx.Rollback()
		return errors.New("vote is not open")
	}

	// Check if option exists
	var option VoteOption
	if err := tx.Where("id = ? AND vote_id = ?", ballot.VoteOptionID, ballot.VoteID).First(&option).Error; err != nil {
		tx.Rollback()
		return errors.New("vote option not found")
	}

	// Check if user has already voted
	var existingBallot Ballot
	result := tx.Where("vote_id = ? AND user_id = ?", ballot.VoteID, ballot.UserID).First(&existingBallot)
	if result.Error == nil {
		tx.Rollback()
		return errors.New("user has already voted")
	}

	// Insert ballot
	if err := tx.Create(&ballot).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update option count
	if err := tx.Model(&VoteOption{}).Where("id = ?", ballot.VoteOptionID).UpdateColumn("count", option.Count+1).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}
