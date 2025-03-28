package models

import (
	"errors"
	"time"
)

// GetAllBoardMeetings returns all board meetings
func GetAllBoardMeetings() []BoardMeeting {
	var meetings []BoardMeeting
	GetDB().Preload("Attendees").Preload("Company").Find(&meetings)
	return meetings
}

// GetBoardMeetingByID returns a board meeting by ID
func GetBoardMeetingByID(id int64) (BoardMeeting, error) {
	var meeting BoardMeeting
	result := GetDB().Preload("Attendees").Preload("Company").First(&meeting, id)
	if result.Error != nil {
		return BoardMeeting{}, errors.New("board meeting not found")
	}
	return meeting, nil
}

// GetBoardMeetingsByCompanyID returns all board meetings for a company
func GetBoardMeetingsByCompanyID(companyID int64) []BoardMeeting {
	var meetings []BoardMeeting
	GetDB().Preload("Attendees").Where("company_id = ?", companyID).Find(&meetings)
	return meetings
}

// AddBoardMeeting adds a new board meeting
func AddBoardMeeting(meeting BoardMeeting) int64 {
	// Set timestamps if not already set
	if meeting.CreatedAt.IsZero() {
		meeting.CreatedAt = time.Now()
	}
	if meeting.UpdatedAt.IsZero() {
		meeting.UpdatedAt = time.Now()
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0
	}
	
	// Create meeting
	if err := tx.Create(&meeting).Error; err != nil {
		tx.Rollback()
		return 0
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0
	}
	
	return meeting.ID
}

// UpdateBoardMeeting updates an existing board meeting
func UpdateBoardMeeting(meeting BoardMeeting) error {
	// Set update timestamp
	meeting.UpdatedAt = time.Now()
	
	// Check if meeting exists
	var existingMeeting BoardMeeting
	if err := GetDB().First(&existingMeeting, meeting.ID).Error; err != nil {
		return errors.New("board meeting not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Update meeting
	if err := tx.Model(&meeting).Updates(map[string]interface{}{
		"company_id":  meeting.CompanyID,
		"title":       meeting.Title,
		"date":        meeting.Date,
		"location":    meeting.Location,
		"description": meeting.Description,
		"status":      meeting.Status,
		"minutes":     meeting.Minutes,
		"updated_at":  meeting.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}

// UpdateBoardMeetingStatus updates the status of a board meeting
func UpdateBoardMeetingStatus(id int64, status string) error {
	// Check if meeting exists
	var meeting BoardMeeting
	if err := GetDB().First(&meeting, id).Error; err != nil {
		return errors.New("board meeting not found")
	}
	
	// Update status
	result := GetDB().Model(&meeting).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})
	
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteBoardMeeting deletes a board meeting
func DeleteBoardMeeting(id int64) error {
	// Check if meeting exists
	var meeting BoardMeeting
	if err := GetDB().First(&meeting, id).Error; err != nil {
		return errors.New("board meeting not found")
	}
	
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Delete attendees
	if err := tx.Where("board_meeting_id = ?", id).Delete(&Attendee{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete meeting
	if err := tx.Delete(&BoardMeeting{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	return tx.Commit().Error
}
