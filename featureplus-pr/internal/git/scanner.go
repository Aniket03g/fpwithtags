package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash      string
	Message   string
	Author    string
	Date      time.Time
	Files     []string
	FeatureID string // Extracted feature ID (e.g., FTR-001)
}

// FeatureMapping represents aggregated data for a feature
type FeatureMapping struct {
	FeatureID string
	Commits   []CommitInfo
	Files     map[string]bool // Using map for unique files
}

// featureIDPattern matches FTR-001, FTR-123, etc.
var featureIDPattern = regexp.MustCompile(`FTR-(\d+)`)

// ScanGitHistory scans git log for commits and extracts feature mappings
func ScanGitHistory(commitLimit int) ([]CommitInfo, error) {
	// Run git log command
	cmd := exec.Command("git", "log", 
		"--name-only", 
		"--pretty=format:%H|%an|%ae|%ad|%s",
		"--date=iso",
		fmt.Sprintf("-n %d", commitLimit))
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git log: %w", err)
	}

	return parseGitLog(string(output))
}

// parseGitLog parses the output of git log command
func parseGitLog(output string) ([]CommitInfo, error) {
	var commits []CommitInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	var currentCommit *CommitInfo
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Empty line separates commits
		if line == "" {
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
				currentCommit = nil
			}
			continue
		}
		
		// Check if this is a commit header line (contains |)
		if strings.Contains(line, "|") {
			// Save previous commit if exists
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}
			
			// Parse new commit header
			parts := strings.SplitN(line, "|", 5)
			if len(parts) == 5 {
				hash := parts[0]
				author := parts[1]
				// email := parts[2] // Not used currently
				dateStr := parts[3]
				message := parts[4]
				
				// Parse date
				date, _ := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
				
				// Extract feature ID from commit message
				featureID := extractFeatureID(message)
				
				currentCommit = &CommitInfo{
					Hash:      hash,
					Message:   message,
					Author:    author,
					Date:      date,
					Files:     []string{},
					FeatureID: featureID,
				}
			}
		} else if currentCommit != nil {
			// This is a file path line
			if line != "" {
				currentCommit.Files = append(currentCommit.Files, line)
			}
		}
	}
	
	// Don't forget the last commit
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning git log output: %w", err)
	}
	
	return commits, nil
}

// extractFeatureID extracts feature ID from commit message
func extractFeatureID(message string) string {
	matches := featureIDPattern.FindStringSubmatch(message)
	if len(matches) > 0 {
		return matches[0] // Returns "FTR-001"
	}
	return ""
}

// MapCommitsToFeatures groups commits by feature ID
func MapCommitsToFeatures(commits []CommitInfo) map[string]*FeatureMapping {
	featureMap := make(map[string]*FeatureMapping)
	
	for _, commit := range commits {
		if commit.FeatureID == "" {
			continue // Skip commits without feature ID
		}
		
		// Get or create feature mapping
		mapping, exists := featureMap[commit.FeatureID]
		if !exists {
			mapping = &FeatureMapping{
				FeatureID: commit.FeatureID,
				Commits:   []CommitInfo{},
				Files:     make(map[string]bool),
			}
			featureMap[commit.FeatureID] = mapping
		}
		
		// Add commit to feature
		mapping.Commits = append(mapping.Commits, commit)
		
		// Add files to feature (using map for uniqueness)
		for _, file := range commit.Files {
			mapping.Files[file] = true
		}
	}
	
	return featureMap
}

// GetUniqueFiles returns a sorted list of unique files for a feature
func (fm *FeatureMapping) GetUniqueFiles() []string {
	files := make([]string, 0, len(fm.Files))
	for file := range fm.Files {
		files = append(files, file)
	}
	return files
}

// GetCommitCount returns the number of commits for a feature
func (fm *FeatureMapping) GetCommitCount() int {
	return len(fm.Commits)
}

// GetFileCount returns the number of unique files for a feature
func (fm *FeatureMapping) GetFileCount() int {
	return len(fm.Files)
}
