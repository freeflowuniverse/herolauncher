package models

import (
	"time"
)

// UserHandler handles all operations related to User model
type UserHandler struct {
	circleID string
	db       *Db
}

// User represents a user in the Freezone Manager system
type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Don't serialize password
	Company   string    `json:"company"`
	Role      string    `json:"role"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(circleID string, db *Db) *UserHandler {
	return &UserHandler{
		circleID: circleID,
		db:       db,
	}
}

// GetAll returns all users
func (h *UserHandler) GetAll() []User {
	var users []User
	GetDB().Find(&users)
	return users
}

// GetByID returns a user by ID
func (h *UserHandler) GetByID(id int64) (User, error) {
	var user User
	result := GetDB().First(&user, id)
	if result.Error != nil {
		return User{}, ErrRecordNotFound
	}
	return user, nil
}

// GetByEmail returns a user by email
func (h *UserHandler) GetByEmail(email string) (User, error) {
	var user User
	result := GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		return User{}, ErrRecordNotFound
	}
	return user, nil
}

// Create adds a new user
func (h *UserHandler) Create(user User) (int64, error) {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	result := GetDB().Create(&user)
	if result.Error != nil {
		return 0, result.Error
	}
	return user.ID, nil
}

// Update updates an existing user
func (h *UserHandler) Update(user User) error {
	user.UpdatedAt = time.Now()

	var existingUser User
	if err := GetDB().First(&existingUser, user.ID).Error; err != nil {
		return ErrRecordNotFound
	}

	result := GetDB().Model(&user).Updates(map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"company":    user.Company,
		"role":       user.Role,
		"updated_at": user.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdatePassword updates a user's password
func (h *UserHandler) UpdatePassword(id int64, password string) error {
	var user User
	if err := GetDB().First(&user, id).Error; err != nil {
		return ErrRecordNotFound
	}

	user.Password = password
	user.UpdatedAt = time.Now()

	result := GetDB().Model(&user).Updates(map[string]interface{}{
		"password":   user.Password,
		"updated_at": user.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete deletes a user
func (h *UserHandler) Delete(id int64) error {
	var user User
	if err := GetDB().First(&user, id).Error; err != nil {
		return ErrRecordNotFound
	}

	result := GetDB().Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
