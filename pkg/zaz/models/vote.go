package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// GetAllVotes returns all votes from the database
func GetAllVotes() []Vote {
	var votes []Vote
	GetDB().Preload("Options").Preload("Ballots").Find(&votes)
	return votes
}

// GetVoteByID returns a vote by ID
func GetVoteByID(id int64) (Vote, error) {
	var vote Vote
	result := GetDB().Preload("Options").Preload("Ballots").First(&vote, id)
	if result.Error != nil {
		return Vote{}, errors.New("vote not found")
	}
	return vote, nil
}

// GetVotesByCompanyID returns all votes for a company
func GetVotesByCompanyID(companyID int64) []Vote {
	var votes []Vote
	GetDB().Preload("Options").Preload("Ballots").Where("company_id = ?", companyID).Find(&votes)
	return votes
}

// AddVote adds a new vote to the database
func AddVote(vote Vote) int64 {
	// Set timestamps if not already set
	if vote.CreatedAt.IsZero() {
		vote.CreatedAt = time.Now()
	}
	if vote.UpdatedAt.IsZero() {
		vote.UpdatedAt = time.Now()
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0
	}
	
	// Create vote and associated options
	if err := tx.Create(&vote).Error; err != nil {
		tx.Rollback()
		return 0
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0
	}
	
	return vote.ID
}

// UpdateVote updates an existing vote
func UpdateVote(vote Vote) error {
	// Set update timestamp
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
		return errors.New("vote not found")
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

// DeleteVote deletes a vote and all associated options and ballots
func DeleteVote(id int64) error {
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Check if vote exists
	var vote Vote
	if err := tx.First(&vote, id).Error; err != nil {
		tx.Rollback()
		return errors.New("vote not found")
	}
	
	// Delete ballots (will be automatically deleted due to foreign key constraints with GORM)
	if err := tx.Where("vote_id = ?", id).Delete(&Ballot{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete options (will be automatically deleted due to foreign key constraints with GORM)
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
func CastBallot(ballot Ballot) error {
	// Set timestamp if not already set
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
	} else if result.Error != gorm.ErrRecordNotFound {
		tx.Rollback()
		return result.Error
	}
	
	// Insert ballot
	if err := tx.Create(&ballot).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Update option count
	if err := tx.Model(&option).Update("count", option.Count+ballot.SharesCount).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}

// getVoteOptionsByVoteID returns all options for a vote
func getVoteOptionsByVoteID(voteID int64) []VoteOption {
	var options []VoteOption
	GetDB().Where("vote_id = ?", voteID).Find(&options)
	return options
}

// getBallotsByVoteID returns all ballots for a vote
func getBallotsByVoteID(voteID int64) []Ballot {
	var ballots []Ballot
	GetDB().Where("vote_id = ?", voteID).Find(&ballots)
	return ballots
}
