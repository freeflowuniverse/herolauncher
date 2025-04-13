package proxy

// UserConfig represents the configuration for a user
// It contains information about the user's budget and allowed model groups
type UserConfig struct {
	// Budget represents the virtual money the user has available
	Budget uint32 `json:"budget"`
	
	// ModelGroups is a list of model groups the user has access to
	ModelGroups []string `json:"model_groups"`
	
	// APIKey is the OpenAI API key to use for this user's requests
	OpenAIKey string `json:"openai_key"`
}

// ProxyConfig represents the configuration for the AI proxy server
type ProxyConfig struct {
	// Port is the port to listen on
	Port int `json:"port"`
	
	// OpenAIBaseURL is the base URL for the OpenAI API
	OpenAIBaseURL string `json:"openai_base_url"`
	
	// DefaultOpenAIKey is the default OpenAI API key to use if not specified in UserConfig
	DefaultOpenAIKey string `json:"default_openai_key"`
}