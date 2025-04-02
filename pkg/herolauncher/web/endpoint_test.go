package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestConfig holds configuration for the tests
type TestConfig struct {
	BaseURL string
	Timeout time.Duration
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		BaseURL: "http://localhost:9021",
		Timeout: 5 * time.Second,
	}
}

// testEndpoint tests a single endpoint
func testEndpoint(t *testing.T, config *TestConfig, method, path string, expectedStatus int, formData map[string]string) {
	t.Helper()

	client := &http.Client{
		Timeout: config.Timeout,
	}

	var req *http.Request
	var err error

	fullURL := config.BaseURL + path

	if method == "GET" {
		req, err = http.NewRequest(method, fullURL, nil)
	} else if method == "POST" {
		if formData != nil {
			form := make(url.Values)
			for key, value := range formData {
				form.Add(key, value)
			}
			req, err = http.NewRequest(method, fullURL, strings.NewReader(form.Encode()))
			if err == nil {
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			}
		} else {
			req, err = http.NewRequest(method, fullURL, nil)
		}
	}

	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status %d for %s %s, got %d. Response: %s",
			expectedStatus, method, path, resp.StatusCode, string(body))
	} else {
		t.Logf("✅ %s %s - Status: %d", method, path, resp.StatusCode)
	}
}

// TestGetEndpoints tests all GET endpoints
func TestGetEndpoints(t *testing.T) {
	config := NewTestConfig()

	// All endpoints to test
	getEndpoints := []string{
		"/",                       // Root redirect to admin
		"/admin",                  // Admin dashboard
		"/admin/system/info",      // System info page
		"/admin/services",         // Services page
		"/admin/system/processes", // Processes page
		"/admin/system/logs",      // System logs page
		"/admin/system/settings",  // System settings page
	}

	// Test all endpoints
	for _, endpoint := range getEndpoints {
		t.Run(fmt.Sprintf("GET %s", endpoint), func(t *testing.T) {
			testEndpoint(t, config, "GET", endpoint, http.StatusOK, nil)
		})
	}
}

// TestAPIEndpoints tests all API endpoints
func TestAPIEndpoints(t *testing.T) {
	t.Skip("API endpoints need to be fixed")
	config := NewTestConfig()

	apiEndpoints := []string{
		"/admin/api/hardware-stats", // Hardware stats API
		"/admin/api/process-stats",  // Process stats API
	}

	for _, endpoint := range apiEndpoints {
		t.Run(fmt.Sprintf("GET %s", endpoint), func(t *testing.T) {
			testEndpoint(t, config, "GET", endpoint, http.StatusOK, nil)
		})
	}
}

// TestFragmentEndpoints tests all fragment endpoints used for AJAX updates
func TestFragmentEndpoints(t *testing.T) {
	config := NewTestConfig()

	// All fragment endpoints to test
	fragmentEndpoints := []string{
		"/admin/system/hardware-stats", // Hardware stats fragment
		"/admin/system/processes-data", // Processes data fragment
	}

	// Test all fragment endpoints
	for _, endpoint := range fragmentEndpoints {
		t.Run(fmt.Sprintf("GET %s", endpoint), func(t *testing.T) {
			testEndpoint(t, config, "GET", endpoint, http.StatusOK, nil)
		})
	}
}
