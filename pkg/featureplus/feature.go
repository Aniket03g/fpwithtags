package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Feature represents a feature in FeaturePlus
type Feature struct {
	ID          uint   `json:"id"`
	ProjectID   uint   `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// GetFeatures retrieves all features from the FeaturePlus API
func (c *Client) GetFeatures(projectID uint) ([]Feature, error) {
	url := fmt.Sprintf("%s/api/features", c.BaseURL)
	if projectID > 0 {
		url = fmt.Sprintf("%s?project_id=%d", url, projectID)
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add authentication header
	c.addAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required: please login first using 'featureplus-pr login'")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	var features []Feature
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return features, nil
}

// GetFeature retrieves a specific feature by ID
func (c *Client) GetFeature(id uint) (*Feature, error) {
	url := fmt.Sprintf("%s/api/features/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add authentication header
	c.addAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication required: please login first using 'featureplus-pr login'")
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	var feature Feature
	if err := json.NewDecoder(resp.Body).Decode(&feature); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &feature, nil
}

// CreateFeature creates a new feature in FeaturePlus
func (c *Client) CreateFeature(feature *Feature) error {
	url := fmt.Sprintf("%s/api/features", c.BaseURL)
	
	reqBody, err := json.Marshal(feature)
	if err != nil {
		return fmt.Errorf("error marshaling feature: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication header
	c.addAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication required: please login first using 'featureplus-pr login'")
	} else if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
	}
	
	return nil
}

// UpdateFeature updates an existing feature in FeaturePlus
func (c *Client) UpdateFeature(feature *Feature) error {
	url := fmt.Sprintf("%s/api/features/%d", c.BaseURL, feature.ID)
	
	reqBody, err := json.Marshal(feature)
	if err != nil {
		return fmt.Errorf("error marshaling feature: %w", err)
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Add authentication header
	c.addAuthHeader(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication required: please login first using 'featureplus-pr login'")
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}
	
	return nil
}
