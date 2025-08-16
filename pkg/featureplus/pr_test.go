package featureplus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func TestListPRs(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pr" {
			t.Errorf("Expected to request '/api/pr', got: %s", r.URL.Path)
		}
		
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}
		
		// Check query parameters
		query := r.URL.Query()
		featureID := query.Get("feature_id")
		if featureID != "1" {
			t.Errorf("Expected feature_id=1, got: %s", featureID)
		}
		
		// Return mock response
		prs := []PullRequest{
			{
				ID:        1,
				TaskID:    2,
				FeatureID: 1,
				URL:       "https://github.com/org/repo/pull/1",
				Title:     "Test PR",
				Branch:    "feature/test",
				Status:    "Open",
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prs)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Call the function
	prs, err := client.ListPRs(1)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify response
	if len(prs) != 1 {
		t.Fatalf("Expected 1 PR, got: %d", len(prs))
	}
	
	pr := prs[0]
	if pr.ID != 1 {
		t.Errorf("Expected PR ID 1, got: %d", pr.ID)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Expected PR title 'Test PR', got: %s", pr.Title)
	}
	if pr.Branch != "feature/test" {
		t.Errorf("Expected PR branch 'feature/test', got: %s", pr.Branch)
	}
}

func TestUploadPR(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pr" {
			t.Errorf("Expected to request '/api/pr', got: %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		
		// Decode request body
		var req UploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Error decoding request body: %v", err)
		}
		
		// Verify request fields
		if req.FeatureID != 1 {
			t.Errorf("Expected feature_id 1, got: %d", req.FeatureID)
		}
		if req.TaskID != 2 {
			t.Errorf("Expected task_id 2, got: %d", req.TaskID)
		}
		if req.Title != "Test PR" {
			t.Errorf("Expected title 'Test PR', got: %s", req.Title)
		}
		
		// Return success
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Create upload request
	req := &UploadRequest{
		FeatureID:   1,
		TaskID:      2,
		PRURL:       "https://github.com/org/repo/pull/1",
		Branch:      "feature/test",
		Title:       "Test PR",
		Description: "Test description",
		IsTested:    true,
		Version:     "1.0.0",
	}
	
	// Call the function
	err := client.UploadPR(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestGetPR(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pr/123" {
			t.Errorf("Expected to request '/api/pr/123', got: %s", r.URL.Path)
		}
		
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}
		
		// Return mock response
		pr := PullRequest{
			ID:          123,
			TaskID:      456,
			FeatureID:   789,
			URL:         "https://github.com/org/repo/pull/123",
			Title:       "Test PR",
			Branch:      "feature/test",
			Description: "Test description",
			Status:      "Open",
			Tested:      true,
			Version:     "1.0.0",
			CreatedAt:   1629123456,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pr)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Call the function
	pr, err := client.GetPR(123)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify response
	if pr.ID != 123 {
		t.Errorf("Expected PR ID 123, got: %d", pr.ID)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Expected PR title 'Test PR', got: %s", pr.Title)
	}
	if pr.Branch != "feature/test" {
		t.Errorf("Expected PR branch 'feature/test', got: %s", pr.Branch)
	}
	if pr.FeatureID != 789 {
		t.Errorf("Expected PR feature ID 789, got: %d", pr.FeatureID)
	}
}

func TestApprovePR(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pr/123/review" {
			t.Errorf("Expected to request '/api/pr/123/review', got: %s", r.URL.Path)
		}
		
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		
		// Decode request body
		var req ReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Error decoding request body: %v", err)
		}
		
		// Verify request fields
		if req.Status != "approved" {
			t.Errorf("Expected status 'approved', got: %s", req.Status)
		}
		if req.Comment != "LGTM" {
			t.Errorf("Expected comment 'LGTM', got: %s", req.Comment)
		}
		
		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create client with test server URL
	client := NewClient(server.URL, nil)
	
	// Create review request
	req := &ReviewRequest{
		Status:     "approved",
		Comment:    "LGTM",
		ApprovedAt: 1629123456,
		Version:    "abc123",
	}
	
	// Call the function
	err := client.ApprovePR(123, req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// Variable used to mock exec.Command in tests
var execCommand = exec.Command

func TestMergePR(t *testing.T) {
	// Save the original exec.Command function
	originalExecCommand := execCommand
	// Restore the original function after the test
	defer func() { execCommand = originalExecCommand }()
	
	// Mock the exec.Command to return a successful result
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Verify the command and arguments
		if command != "gh" {
			t.Errorf("Expected command 'gh', got: %s", command)
		}
		
		// Check that we're calling gh pr merge with the right arguments
		if len(args) < 3 || args[0] != "pr" || args[1] != "merge" {
			t.Errorf("Expected 'pr merge' subcommand, got: %v", args)
		}
		
		// Create a fake successful command
		cmd := exec.Command("echo", "PR merged successfully")
		return cmd
	}
	
	// Create client
	client := NewClient("http://example.com", nil)
	
	// Test with different merge options
	testCases := []struct {
		name       string
		prNumber   string
		method     string
		delBranch  bool
		comment    string
		expectArgs []string
	}{
		{
			name:      "Basic merge",
			prNumber:  "123",
			method:    "merge",
			delBranch: false,
			comment:   "",
			expectArgs: []string{"pr", "merge", "123", "--merge"},
		},
		{
			name:      "Squash merge with delete branch",
			prNumber:  "456",
			method:    "squash",
			delBranch: true,
			comment:   "",
			expectArgs: []string{"pr", "merge", "456", "--squash", "--delete-branch"},
		},
		{
			name:      "Rebase merge with comment",
			prNumber:  "789",
			method:    "rebase",
			delBranch: false,
			comment:   "Merging feature",
			expectArgs: []string{"pr", "merge", "789", "--rebase", "--body", "Merging feature"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Override the executor for this specific test case
			execCommand = func(command string, args ...string) *exec.Cmd {
				// Verify the command is gh
				if command != "gh" {
					t.Errorf("Expected command 'gh', got: %s", command)
				}
				
				// Check that the PR number is correct
				if args[2] != tc.prNumber {
					t.Errorf("Expected PR number %s, got: %s", tc.prNumber, args[2])
				}
				
				// Check for merge method flag
				methodFlagFound := false
				for _, arg := range args {
					if arg == "--"+tc.method {
						methodFlagFound = true
						break
					}
				}
				if !methodFlagFound {
					t.Errorf("Expected merge method flag '--%s', not found in args: %v", tc.method, args)
				}
				
				// Check for delete branch flag if needed
				if tc.delBranch {
					delBranchFound := false
					for _, arg := range args {
						if arg == "--delete-branch" {
							delBranchFound = true
							break
						}
					}
					if !delBranchFound {
						t.Errorf("Expected --delete-branch flag, not found in args: %v", args)
					}
				}
				
				// Check for comment if provided
				if tc.comment != "" {
					commentFound := false
					for i, arg := range args {
						if arg == "--body" && i+1 < len(args) && args[i+1] == tc.comment {
							commentFound = true
							break
						}
					}
					if !commentFound {
						t.Errorf("Expected --body flag with comment '%s', not found in args: %v", tc.comment, args)
					}
				}
				
				// Create a fake successful command
				cmd := exec.Command("echo", "PR merged successfully")
				return cmd
			}
			
			// Call the function
			err := client.MergePR(tc.prNumber, tc.method, tc.delBranch, tc.comment)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
		})
	}
	
	// Test error case
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Create a command that will fail
		cmd := exec.Command("false")
		return cmd
	}
	
	// Call the function with the failing command
	err := client.MergePR("999", "merge", false, "")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
