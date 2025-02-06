package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PocketBaseClient interacts with the PocketBase HTTP API
type PocketBaseClient struct {
	BaseURL    string
	SuperToken string
}

// Check pocketbase connection is working
func (p *PocketBaseClient) CheckConnection() error {
	url := fmt.Sprintf("%s/api/health", p.BaseURL)

	// Create a new HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make GET request to health endpoint
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to PocketBase: %w", err)
	}
	defer resp.Body.Close()

	// Check if status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PocketBase health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// NewPocketBaseClient initializes a new PocketBase client and authenticates the superuser
func NewPocketBaseClient(baseURL, superuserEmail, superuserPassword string) (*PocketBaseClient, error) {
	client := &PocketBaseClient{BaseURL: baseURL}

	// Authenticate superuser and store the token
	token, err := client.authenticateSuperuser(superuserEmail, superuserPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate superuser: %w", err)
	}

	client.SuperToken = token
	return client, nil
}

// authenticateSuperuser logs in the superuser and retrieves the authentication token
func (p *PocketBaseClient) authenticateSuperuser(email, password string) (string, error) {
	url := fmt.Sprintf("%s/api/collections/_superusers/auth-with-password", p.BaseURL)

	// Data payload for authentication
	data := map[string]interface{}{
		"identity": email,
		"password": password,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate superuser: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate superuser: status %d", resp.StatusCode)
	}

	// Parse response
	var respData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respData.Token, nil
}
