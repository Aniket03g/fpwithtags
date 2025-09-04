package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

// Client represents a FeaturePlus API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	authToken  string
}

// NewClient creates a new FeaturePlus API client
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		authToken:  "",
	}
}

// addAuthHeader adds the Authorization header to the request if an auth token is available
func (c *Client) addAuthHeader(req *http.Request) {
	fmt.Printf("DEBUG AUTH: Adding auth header, token present: %v\n", c.authToken != "")
	if c.authToken != "" {
		fmt.Printf("DEBUG AUTH: Token starts with: %s...\n", c.authToken[:10])
		req.Header.Set("Authorization", "Bearer "+c.authToken)
		fmt.Printf("DEBUG AUTH: Authorization header set: %s\n", req.Header.Get("Authorization"))
	} else {
		fmt.Printf("DEBUG AUTH: No token available, skipping auth header\n")
	}
}

// Project represents a project in FeaturePlus
type Project struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	RepoURL     string `json:"repo_url"`
}

// GetProjects retrieves all projects from the FeaturePlus API
func (c *Client) GetProjects() ([]Project, error) {
	url := fmt.Sprintf("%s/api/projects", c.BaseURL)
	
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
	
	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return projects, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(id uint) (*Project, error) {
	url := fmt.Sprintf("%s/api/projects/%d", c.BaseURL, id)
	
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
	
	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(project *Project) error {
	url := path.Join(c.BaseURL, "/api/projects")
	
	reqBody, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("error marshaling project: %w", err)
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
	} else if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned non-Created status: %d", resp.StatusCode)
	}
	
	return nil
}
