package webui

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
		BaseURL: "http://localhost:9030",
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

	getEndpoints := []string{
		"/",                    // Dashboard
		"/companies",           // Companies list
		"/companies/create",    // Create company form
		"/shareholders",        // Shareholders list
		"/shareholders/create", // Create shareholder form
		"/boardmeetings",       // Board meetings list
		"/boardmeetings/create", // Create board meeting form
		"/votes",               // Votes list
		"/votes/create",        // Create vote form
	}

	for _, endpoint := range getEndpoints {
		t.Run(fmt.Sprintf("GET %s", endpoint), func(t *testing.T) {
			testEndpoint(t, config, "GET", endpoint, http.StatusOK, nil)
		})
	}
}

// TestPostEndpoints tests all POST endpoints
func TestPostEndpoints(t *testing.T) {
	config := NewTestConfig()

	// Test POST endpoints with sample data
	t.Run("POST /companies/create", func(t *testing.T) {
		companyData := map[string]string{
			"name":          "Test Company",
			"email":         "test@example.com",
			"phone":         "+1234567890",
			"website":       "https://example.com",
			"address":       "123 Test Street",
			"business_type": "corporation",
			"industry":      "technology",
			"description":   "A test company",
			"csrf_token":    "sample-token", // Add CSRF token
		}
		testEndpoint(t, config, "POST", "/companies/create", http.StatusFound, companyData)
	})

	t.Run("POST /shareholders/create", func(t *testing.T) {
		shareholderData := map[string]string{
			"name":           "Test Shareholder",
			"email":          "shareholder@example.com",
			"phone":          "+1234567890",
			"address":        "123 Shareholder Street",
			"identification": "ID12345",
			"nationality":    "Test Nation",
			"tax_id":         "TAX12345",
			"notes":          "Test notes",
			"csrf_token":     "sample-token", // Add CSRF token
		}
		testEndpoint(t, config, "POST", "/shareholders/create", http.StatusFound, shareholderData)
	})

	t.Run("POST /boardmeetings/create", func(t *testing.T) {
		boardMeetingData := map[string]string{
			"title":       "Test Board Meeting",
			"date":        time.Now().Format("2006-01-02"),
			"time":        "14:00",
			"location":    "Test Location",
			"description": "Test board meeting description",
			"csrf_token":  "sample-token", // Add CSRF token
		}
		testEndpoint(t, config, "POST", "/boardmeetings/create", http.StatusFound, boardMeetingData)
	})

	t.Run("POST /votes/create", func(t *testing.T) {
		voteData := map[string]string{
			"title":       "Test Vote",
			"description": "Test vote description",
			"start_date":  time.Now().Format("2006-01-02"),
			"end_date":    time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
			"csrf_token":  "sample-token", // Add CSRF token
		}
		testEndpoint(t, config, "POST", "/votes/create", http.StatusFound, voteData)
	})
}
