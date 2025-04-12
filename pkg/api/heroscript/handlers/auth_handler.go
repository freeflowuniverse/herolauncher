package handlers

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
)

// AuthHandler handles authentication actions
type AuthHandler struct {
	BaseHandler
	secrets []string
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(secrets ...string) *AuthHandler {
	return &AuthHandler{
		BaseHandler: BaseHandler{
			BaseHandler: handlerfactory.BaseHandler{
				ActorName: "auth",
			},
		},
		secrets: secrets,
	}
}

// Auth handles the auth.auth action
func (h *AuthHandler) Auth(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	secret := params.Get("secret")
	if secret == "" {
		return "Error: secret is required"
	}

	for _, validSecret := range h.secrets {
		if secret == validSecret {
			return "Authentication successful"
		}
	}

	return "Authentication failed: invalid secret"
}

// AddSecret adds a new secret to the handler
func (h *AuthHandler) AddSecret(secret string) {
	h.secrets = append(h.secrets, secret)
}

// RemoveSecret removes a secret from the handler
func (h *AuthHandler) RemoveSecret(secret string) bool {
	for i, s := range h.secrets {
		if s == secret {
			// Remove the secret
			h.secrets = append(h.secrets[:i], h.secrets[i+1:]...)
			return true
		}
	}
	return false
}
