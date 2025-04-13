package proxy

import (
	"errors"
	"os"
	"sync"
)

// Factory manages the proxy server and user configurations
type Factory struct {
	// Config is the proxy server configuration
	Config ProxyConfig

	// userConfigs is a map of API keys to user configurations
	userConfigs map[string]UserConfig

	// Lock for concurrent access to userConfigs
	mu sync.RWMutex
}

// NewFactory creates a new proxy factory with the given configuration
func NewFactory(config ProxyConfig) *Factory {
	// Check for OPENAIKEY environment variable and use it if available
	if envKey := os.Getenv("OPENAIKEY"); envKey != "" {
		config.DefaultOpenAIKey = envKey
	}

	return &Factory{
		Config:      config,
		userConfigs: make(map[string]UserConfig),
	}
}

// AddUserConfig adds or updates a user configuration with the associated API key
func (f *Factory) AddUserConfig(apiKey string, config UserConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.userConfigs[apiKey] = config
}

// GetUserConfig retrieves a user configuration by API key
func (f *Factory) GetUserConfig(apiKey string) (UserConfig, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	config, exists := f.userConfigs[apiKey]
	if !exists {
		return UserConfig{}, errors.New("invalid API key")
	}

	return config, nil
}

// RemoveUserConfig removes a user configuration by API key
func (f *Factory) RemoveUserConfig(apiKey string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.userConfigs, apiKey)
}

// GetOpenAIKey returns the OpenAI API key to use for a given proxy API key
// Always returns the default OpenAI key from environment variable
func (f *Factory) GetOpenAIKey(proxyAPIKey string) string {
	// Always use the default OpenAI key from environment variable
	// This ensures that all requests to OpenAI use our key, not the user's key
	return f.Config.DefaultOpenAIKey
}

// DecreaseBudget decreases a user's budget by the specified amount
// Returns error if the user doesn't have enough budget
func (f *Factory) DecreaseBudget(apiKey string, amount uint32) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	config, exists := f.userConfigs[apiKey]
	if !exists {
		return errors.New("invalid API key")
	}

	if config.Budget < amount {
		return errors.New("insufficient budget")
	}

	config.Budget -= amount
	f.userConfigs[apiKey] = config
	return nil
}

// IncreaseBudget increases a user's budget by the specified amount
func (f *Factory) IncreaseBudget(apiKey string, amount uint32) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	config, exists := f.userConfigs[apiKey]
	if !exists {
		return errors.New("invalid API key")
	}

	config.Budget += amount
	f.userConfigs[apiKey] = config
	return nil
}

// CanAccessModel checks if a user can access a specific model
func (f *Factory) CanAccessModel(apiKey string, model string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	config, exists := f.userConfigs[apiKey]
	if !exists {
		return false
	}

	// If no model groups are specified, allow access to all models
	if len(config.ModelGroups) == 0 {
		return true
	}

	// Check if the model is in any of the allowed model groups
	// This is a placeholder - the actual implementation would depend on
	// how model groups are defined and mapped to specific models
	for _, group := range config.ModelGroups {
		if group == "all" {
			return true
		}
		// Add logic to check if model is in group
		// For now we'll just check if the model contains the group name
		if group == model {
			return true
		}
	}

	return false
}
