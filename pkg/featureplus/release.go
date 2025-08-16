package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Release represents a release in FeaturePlus
type Release struct {
	ID        uint   `json:"id"`
	Tag       string `json:"tag"`
	Notes     string `json:"notes"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	PRIDs     []uint `json:"pr_ids"`
}

// CreateReleaseRequest represents a request to create a release
type CreateReleaseRequest struct {
	Tag   string `json:"tag"`
	Notes string `json:"notes"`
	PRIDs []uint `json:"pr_ids"`
}

// CreateRelease creates a new release in FeaturePlus
func (c *Client) CreateRelease(req *CreateReleaseRequest) (*Release, error) {
	url := fmt.Sprintf("%s/api/releases", c.BaseURL)
	
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
	}
	
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &release, nil
}

// ListReleases retrieves all releases from FeaturePlus
func (c *Client) ListReleases() ([]Release, error) {
	url := fmt.Sprintf("%s/api/releases", c.BaseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return releases, nil
}

// GetRelease retrieves a specific release by ID
func (c *Client) GetRelease(id uint) (*Release, error) {
	url := fmt.Sprintf("%s/api/releases/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &release, nil
}

// FinalizeRelease finalizes a release in FeaturePlus
func (c *Client) FinalizeRelease(id uint) error {
	url := fmt.Sprintf("%s/api/releases/%d/finalize", c.BaseURL, id)
	
	// Empty JSON body
	emptyBody := []byte("{}")
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(emptyBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	return nil
}
