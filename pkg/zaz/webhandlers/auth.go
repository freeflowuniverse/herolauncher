package webhandlers

import (
	"github.com/freeflowuniverse/herolauncher/pkg/zaz/models"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related routes
type AuthHandler struct {
	store *models.Store
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(store *models.Store) *AuthHandler {
	return &AuthHandler{
		store: store,
	}
}

// GetLogin renders the login page
func (h *AuthHandler) GetLogin(c *fiber.Ctx) error {
	return RenderWithDefaults(c, "login", fiber.Map{
		"title": "Login",
	})
}

// PostLogin handles login form submission
func (h *AuthHandler) PostLogin(c *fiber.Ctx) error {
	// Parse form data
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Simple validation
	if email == "" || password == "" {
		return c.Render("login", fiber.Map{
			"title": "Login",
			"error": "Email and password are required",
		})
	}

	// TODO: Implement actual authentication using the store
	// For now, just redirect to dashboard
	return c.Redirect("/")
}

// GetRegister renders the registration page
func (h *AuthHandler) GetRegister(c *fiber.Ctx) error {
	return c.Render("register", fiber.Map{
		"title": "Register",
	})
}

// PostRegister handles registration form submission
func (h *AuthHandler) PostRegister(c *fiber.Ctx) error {
	// Parse form data
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	// Simple validation
	if name == "" || email == "" || password == "" {
		return c.Render("register", fiber.Map{
			"title": "Register",
			"error": "All fields are required",
		})
	}

	if password != confirmPassword {
		return c.Render("register", fiber.Map{
			"title": "Register",
			"error": "Passwords do not match",
		})
	}

	// Create user using the store
	user := models.User{
		Name:     name,
		Email:    email,
		Password: password, // In a real app, this would be hashed
	}
	
	_, err := h.store.UserHandler.Create(user)
	if err != nil {
		return c.Render("register", fiber.Map{
			"title": "Register",
			"error": "Failed to create user: " + err.Error(),
		})
	}

	// Redirect to login page
	return c.Redirect("/login")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// TODO: Implement actual logout
	return c.Redirect("/login")
}

// GetForgotPassword renders the forgot password page
func (h *AuthHandler) GetForgotPassword(c *fiber.Ctx) error {
	return c.Render("forgot-password", fiber.Map{
		"title": "Forgot Password",
	})
}

// PostForgotPassword handles forgot password form submission
func (h *AuthHandler) PostForgotPassword(c *fiber.Ctx) error {
	// Parse form data
	email := c.FormValue("email")

	// Simple validation
	if email == "" {
		return c.Render("forgot-password", fiber.Map{
			"title": "Forgot Password",
			"error": "Email is required",
		})
	}

	// TODO: Implement actual password reset
	// For now, just redirect to login page with a success message
	return c.Render("login", fiber.Map{
		"title": "Login",
		"message": "Password reset instructions have been sent to your email",
	})
}
