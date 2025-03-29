package models

import (
	"time"
)

// BoardMeetingHandler handles all operations related to BoardMeeting model
type BoardMeetingHandler struct {
	circleID string
	db       *Db
}

// NewBoardMeetingHandler creates a new BoardMeetingHandler
func NewBoardMeetingHandler(circleID string, db *Db) *BoardMeetingHandler {
	return &BoardMeetingHandler{
		circleID: circleID,
		db:       db,
	}
}

// BoardMeeting represents a board meeting of a company
type BoardMeeting struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	CompanyID   int64      `gorm:"index;not null" json:"company_id"`
	Title       string     `gorm:"not null" json:"title"`
	Date        time.Time  `gorm:"not null" json:"date"`
	Location    string     `json:"location"`
	Description string     `json:"description"`
	Status      string     `gorm:"not null" json:"status"` // Scheduled, Completed, Cancelled
	Minutes     string     `json:"minutes,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	Attendees   []Attendee `gorm:"foreignKey:BoardMeetingID" json:"attendees,omitempty"`
	Company     Company    `gorm:"foreignKey:CompanyID" json:"-"`
}

// Attendee represents an attendee of a board meeting
type Attendee struct {
	ID             int64        `gorm:"primaryKey" json:"id"`
	BoardMeetingID int64        `gorm:"index;not null" json:"board_meeting_id"`
	UserID         int64        `gorm:"index;not null" json:"user_id"`
	Name           string       `gorm:"not null" json:"name"`
	Role           string       `json:"role"`
	Status         string       `gorm:"not null" json:"status"` // Confirmed, Pending, Declined
	CreatedAt      time.Time    `gorm:"autoCreateTime" json:"created_at"`
	BoardMeeting   BoardMeeting `gorm:"foreignKey:BoardMeetingID" json:"-"`
	User           User         `gorm:"foreignKey:UserID" json:"-"`
}

// GetAll returns all board meetings
func (h *BoardMeetingHandler) GetAll() []BoardMeeting {
	var meetings []BoardMeeting
	GetDB().Preload("Attendees").Find(&meetings)
	return meetings
}

// GetByID returns a board meeting by ID
func (h *BoardMeetingHandler) GetByID(id int64) (BoardMeeting, error) {
	var meeting BoardMeeting
	result := GetDB().Preload("Attendees").First(&meeting, id)
	if result.Error != nil {
		return BoardMeeting{}, ErrRecordNotFound
	}
	return meeting, nil
}

// GetByCompanyID returns all board meetings for a company
func (h *BoardMeetingHandler) GetByCompanyID(companyID int64) []BoardMeeting {
	var meetings []BoardMeeting
	GetDB().Preload("Attendees").Where("company_id = ?", companyID).Find(&meetings)
	return meetings
}

// Create adds a new board meeting
func (h *BoardMeetingHandler) Create(meeting BoardMeeting) (int64, error) {
	if meeting.CreatedAt.IsZero() {
		meeting.CreatedAt = time.Now()
	}
	if meeting.UpdatedAt.IsZero() {
		meeting.UpdatedAt = time.Now()
	}

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	// Create meeting
	if err := tx.Create(&meeting).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Create attendees
	for i := range meeting.Attendees {
		meeting.Attendees[i].BoardMeetingID = meeting.ID
		if err := tx.Create(&meeting.Attendees[i]).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return meeting.ID, nil
}

// Update updates an existing board meeting
func (h *BoardMeetingHandler) Update(meeting BoardMeeting) error {
	meeting.UpdatedAt = time.Now()

	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if meeting exists
	var existingMeeting BoardMeeting
	if err := tx.First(&existingMeeting, meeting.ID).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
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

	// Update attendees (delete and re-insert)
	if err := tx.Where("board_meeting_id = ?", meeting.ID).Delete(&Attendee{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Insert new attendees
	for i := range meeting.Attendees {
		meeting.Attendees[i].BoardMeetingID = meeting.ID
		if err := tx.Create(&meeting.Attendees[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

// Delete deletes a board meeting and all associated attendees
func (h *BoardMeetingHandler) Delete(id int64) error {
	// Begin transaction
	tx := GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Check if meeting exists
	var meeting BoardMeeting
	if err := tx.First(&meeting, id).Error; err != nil {
		tx.Rollback()
		return ErrRecordNotFound
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
