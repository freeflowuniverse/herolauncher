package models

import (
	"errors"
	"time"
)

// GetAllUsers returns all users
func GetAllUsers() []User {
	var users []User
	GetDB().Find(&users)
	return users
}

// GetUserByID returns a user by ID
func GetUserByID(id int64) (User, error) {
	var user User
	result := GetDB().First(&user, id)
	if result.Error != nil {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

// GetUserByEmail returns a user by email
func GetUserByEmail(email string) (User, error) {
	var user User
	result := GetDB().Where("email = ?", email).First(&user)
	if result.Error != nil {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

// AddUser adds a new user
func AddUser(user User) int64 {
	// Set timestamps if not already set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	result := GetDB().Create(&user)
	if result.Error != nil {
		return 0
	}
	return user.ID
}

// UpdateUser updates an existing user
func UpdateUser(user User) error {
	// Set update timestamp
	user.UpdatedAt = time.Now()

	// Check if user exists
	var existingUser User
	if err := GetDB().First(&existingUser, user.ID).Error; err != nil {
		return errors.New("user not found")
	}

	// Update user
	result := GetDB().Model(&user).Updates(map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"password":   user.Password,
		"company":    user.Company,
		"role":       user.Role,
		"updated_at": user.UpdatedAt,
	})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// DeleteUser deletes a user
func DeleteUser(id int64) error {
	// Check if user exists
	var user User
	if err := GetDB().First(&user, id).Error; err != nil {
		return errors.New("user not found")
	}

	// Delete user
	result := GetDB().Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
