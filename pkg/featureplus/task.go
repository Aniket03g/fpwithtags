package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Task represents a task in FeaturePlus
type Task struct {
	ID          uint   `json:"id"`
	FeatureID   uint   `json:"feature_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
	Priority    string `json:"priority"`
}

// GetTasks retrieves all tasks for a feature from the FeaturePlus API
func (c *Client) GetTasks(featureID uint) ([]Task, error) {
	url := fmt.Sprintf("%s/api/tasks", c.BaseURL)
	if featureID > 0 {
		url = fmt.Sprintf("%s?feature_id=%d", url, featureID)
	}
	
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
	
	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return tasks, nil
}

// GetTask retrieves a specific task by ID
func (c *Client) GetTask(id uint) (*Task, error) {
	url := fmt.Sprintf("%s/api/tasks/%d", c.BaseURL, id)
	
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
	
	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &task, nil
}

// CreateTask creates a new task in FeaturePlus
func (c *Client) CreateTask(task *Task) error {
	url := fmt.Sprintf("%s/api/tasks", c.BaseURL)
	
	reqBody, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error marshaling task: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
	}
	
	return nil
}

// UpdateTask updates an existing task in FeaturePlus
func (c *Client) UpdateTask(task *Task) error {
	url := fmt.Sprintf("%s/api/tasks/%d", c.BaseURL, task.ID)
	
	reqBody, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error marshaling task: %w", err)
	}
	
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
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
