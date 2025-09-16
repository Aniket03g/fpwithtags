package models

// PRResponse extends the PullRequest model with additional computed fields
// for API responses and UI display
type PRResponse struct {
	PullRequest
	IsBlocked           bool                   `json:"is_blocked"`
	BlockingDependencies []map[string]interface{} `json:"blocking_dependencies,omitempty"`
	MergeableState      string                 `json:"mergeable_state,omitempty"`
}

// NewPRResponse creates a PRResponse from a PullRequest
func NewPRResponse(pr *PullRequest) *PRResponse {
	return &PRResponse{
		PullRequest: *pr,
		IsBlocked:   false, // Default to false, will be set by service
	}
}
