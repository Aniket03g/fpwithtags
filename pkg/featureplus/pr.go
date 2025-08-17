package featureplus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// PRStatus defines the type for PR status values
type PRStatus string

// Define the allowed constant values for PRStatus
const (
	StatusOpen             PRStatus = "Open"
	StatusApproved         PRStatus = "Approved"
	StatusRejected         PRStatus = "Rejected"
	StatusChangesRequested PRStatus = "ChangesRequested"
	StatusMerged           PRStatus = "Merged"
)

// PullRequest represents a pull request in FeaturePlus
type PullRequest struct {
	ID          uint     `json:"id"`
	TaskID      uint     `json:"task_id"`
	FeatureID   uint     `json:"feature_id"`
	URL         string   `json:"pr_url"`
	Title       string   `json:"title"`
	Branch      string   `json:"branch"`
	Description string   `json:"description"`
	Status      PRStatus `json:"status"`
	Tested      bool     `json:"is_tested"`
	Version     string   `json:"version"`
	CreatedAt   int64    `json:"created_at"`
}

// PRInfo represents GitHub PR information
type PRInfo struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Body        string `json:"body"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
}

// UploadRequest represents a request to upload a PR to FeaturePlus
type UploadRequest struct {
	FeatureID   int    `json:"feature_id"`
	TaskID      int    `json:"task_id"`
	PRURL       string `json:"pr_url"`
	Branch      string `json:"branch"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsTested    bool   `json:"is_tested"`
	Version     string `json:"version"`
}

// ReviewRequest represents a request to review a PR in FeaturePlus
type ReviewRequest struct {
	Status     PRStatus `json:"status"`
	Comment    string   `json:"comment,omitempty"`
	ApprovedAt int64    `json:"approved_at,omitempty"`
	Version    string   `json:"version,omitempty"`
}

// GetPRInfo retrieves information about the current PR from GitHub CLI
func GetPRInfo() (*PRInfo, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "title,url,body,headRefName,baseRefName")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting PR info: %w, output: %s", err, string(output))
	}

	var prInfo PRInfo
	if err := json.Unmarshal(output, &prInfo); err != nil {
		return nil, fmt.Errorf("error parsing PR info: %w", err)
	}

	return &prInfo, nil
}

// UploadPR uploads a PR to FeaturePlus
func (c *Client) UploadPR(req *UploadRequest) error {
	url := fmt.Sprintf("%s/api/pr", c.BaseURL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

// ListPRs retrieves all PRs from FeaturePlus, optionally filtered by feature ID
func (c *Client) ListPRs(featureID uint) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/api/pr", c.BaseURL)
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

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return prs, nil
}

// GetPR retrieves a specific PR by ID
func (c *Client) GetPR(id int) (*PullRequest, error) {
	url := fmt.Sprintf("%s/api/pr/%d", c.BaseURL, id)

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

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &pr, nil
}

// ApprovePR approves a PR in FeaturePlus
func (c *Client) ApprovePR(prID int, req *ReviewRequest) error {
	url := fmt.Sprintf("%s/api/pr/%d/review", c.BaseURL, prID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

// MergePR merges a PR using GitHub CLI
func (c *Client) MergePR(prNumber string, method string, deleteBranch bool, comment string) error {
	args := []string{"pr", "merge", prNumber, "--" + method}

	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	if comment != "" {
		args = append(args, "--body", comment)
	}

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error merging PR: %w", err)
	}
	return nil
}
