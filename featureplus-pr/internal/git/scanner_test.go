package git

import (
	"testing"
)

func TestExtractFeatureID(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"FTR-001: Add user login", "FTR-001"},
		{"[FTR-123] Fix authentication bug", "FTR-123"},
		{"Implement dashboard (FTR-042)", "FTR-042"},
		{"No feature ID here", ""},
		{"ftr-001 lowercase should not match", ""},
		{"FTR-1 single digit", "FTR-1"},
		{"Multiple FTR-001 and FTR-002 IDs", "FTR-001"}, // Returns first match
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := extractFeatureID(tt.message)
			if result != tt.expected {
				t.Errorf("extractFeatureID(%q) = %q, want %q", tt.message, result, tt.expected)
			}
		})
	}
}

func TestParseGitLog(t *testing.T) {
	// Sample git log output
	gitLogOutput := `abc123|John Doe|john@example.com|2025-11-10 12:00:00 +0530|FTR-001: Add user login
backend/auth/login.go
backend/routes.go

def456|Jane Smith|jane@example.com|2025-11-10 13:00:00 +0530|FTR-002: Dashboard UI
ui/components/dashboard.tsx
ui/styles/dashboard.css

ghi789|John Doe|john@example.com|2025-11-10 14:00:00 +0530|Fix typo in README
README.md
`

	commits, err := parseGitLog(gitLogOutput)
	if err != nil {
		t.Fatalf("parseGitLog() error = %v", err)
	}

	if len(commits) != 3 {
		t.Errorf("parseGitLog() returned %d commits, want 3", len(commits))
	}

	// Test first commit
	if commits[0].Hash != "abc123" {
		t.Errorf("commits[0].Hash = %q, want %q", commits[0].Hash, "abc123")
	}
	if commits[0].FeatureID != "FTR-001" {
		t.Errorf("commits[0].FeatureID = %q, want %q", commits[0].FeatureID, "FTR-001")
	}
	if len(commits[0].Files) != 2 {
		t.Errorf("commits[0].Files has %d files, want 2", len(commits[0].Files))
	}

	// Test second commit
	if commits[1].FeatureID != "FTR-002" {
		t.Errorf("commits[1].FeatureID = %q, want %q", commits[1].FeatureID, "FTR-002")
	}

	// Test third commit (no feature ID)
	if commits[2].FeatureID != "" {
		t.Errorf("commits[2].FeatureID = %q, want empty string", commits[2].FeatureID)
	}
}

func TestMapCommitsToFeatures(t *testing.T) {
	commits := []CommitInfo{
		{
			Hash:      "abc123",
			Message:   "FTR-001: Add login",
			FeatureID: "FTR-001",
			Files:     []string{"auth/login.go", "routes.go"},
		},
		{
			Hash:      "def456",
			Message:   "FTR-001: Fix login bug",
			FeatureID: "FTR-001",
			Files:     []string{"auth/login.go", "auth/session.go"},
		},
		{
			Hash:      "ghi789",
			Message:   "FTR-002: Add dashboard",
			FeatureID: "FTR-002",
			Files:     []string{"ui/dashboard.tsx"},
		},
	}

	featureMap := MapCommitsToFeatures(commits)

	if len(featureMap) != 2 {
		t.Errorf("MapCommitsToFeatures() returned %d features, want 2", len(featureMap))
	}

	// Test FTR-001
	ftr001, exists := featureMap["FTR-001"]
	if !exists {
		t.Fatal("FTR-001 not found in feature map")
	}
	if ftr001.GetCommitCount() != 2 {
		t.Errorf("FTR-001 has %d commits, want 2", ftr001.GetCommitCount())
	}
	if ftr001.GetFileCount() != 3 {
		t.Errorf("FTR-001 has %d unique files, want 3", ftr001.GetFileCount())
	}

	// Test FTR-002
	ftr002, exists := featureMap["FTR-002"]
	if !exists {
		t.Fatal("FTR-002 not found in feature map")
	}
	if ftr002.GetCommitCount() != 1 {
		t.Errorf("FTR-002 has %d commits, want 1", ftr002.GetCommitCount())
	}
	if ftr002.GetFileCount() != 1 {
		t.Errorf("FTR-002 has %d unique files, want 1", ftr002.GetFileCount())
	}
}
