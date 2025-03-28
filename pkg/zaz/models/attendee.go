package models

import (
	"errors"
	"time"
)

// GetAttendeesByBoardMeetingID returns all attendees for a board meeting
func GetAttendeesByBoardMeetingID(meetingID int64) []Attendee {
	var attendees []Attendee
	GetDB().Where("board_meeting_id = ?", meetingID).Find(&attendees)
	return attendees
}

// GetAttendeeByID returns an attendee by ID
func GetAttendeeByID(id int64) (Attendee, error) {
	var attendee Attendee
	result := GetDB().First(&attendee, id)
	if result.Error != nil {
		return Attendee{}, errors.New("attendee not found")
	}
	return attendee, nil
}

// AddAttendee adds a new attendee to a board meeting
func AddAttendee(attendee Attendee) int64 {
	// Set timestamp if not already set
	if attendee.CreatedAt.IsZero() {
		attendee.CreatedAt = time.Now()
	}
	
	result := GetDB().Create(&attendee)
	if result.Error != nil {
		return 0
	}
	return attendee.ID
}

// UpdateAttendeeStatus updates the status of an attendee
func UpdateAttendeeStatus(id int64, status string) error {
	// Check if attendee exists
	var attendee Attendee
	if err := GetDB().First(&attendee, id).Error; err != nil {
		return errors.New("attendee not found")
	}
	
	// Update status
	result := GetDB().Model(&attendee).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}

// DeleteAttendee deletes an attendee
func DeleteAttendee(id int64) error {
	// Check if attendee exists
	var attendee Attendee
	if err := GetDB().First(&attendee, id).Error; err != nil {
		return errors.New("attendee not found")
	}
	
	// Delete attendee
	result := GetDB().Delete(&Attendee{}, id)
	if result.Error != nil {
		return result.Error
	}
	
	return nil
}
