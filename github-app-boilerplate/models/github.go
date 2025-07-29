package models

import (
	"time"

	"github.com/google/go-github/v56/github"
)

// WebhookPayload represents the payload received from GitHub webhooks
type WebhookPayload struct {
	Action      string             `json:"action"`
	Repository  *github.Repository `json:"repository,omitempty"`
	PullRequest *github.PullRequest `json:"pull_request,omitempty"`
	Installation *Installation     `json:"installation,omitempty"`
	Sender      *github.User       `json:"sender,omitempty"`
}

// Installation represents a GitHub App installation
type Installation struct {
	ID     int64  `json:"id,omitempty"`
	NodeID string `json:"node_id,omitempty"`
}

// PullRequestEvent represents a pull request event from GitHub
type PullRequestEvent struct {
	Action      string            `json:"action"`
	Number      int               `json:"number"`
	PullRequest *github.PullRequest `json:"pull_request"`
	Repository  *github.Repository `json:"repository"`
	Sender      *github.User      `json:"sender"`
}

// InstallationEvent represents a GitHub App installation event
type InstallationEvent struct {
	Action       string             `json:"action"`
	Installation *github.Installation `json:"installation"`
	Repositories []*github.Repository `json:"repositories,omitempty"`
	Sender       *github.User       `json:"sender"`
}

// CommentEvent represents a comment event from GitHub
type CommentEvent struct {
	Action  string             `json:"action"`
	Issue   *github.Issue      `json:"issue"`
	Comment *github.IssueComment `json:"comment"`
	Repo    *github.Repository `json:"repository"`
	Sender  *github.User       `json:"sender"`
}

// WebhookResponse represents the response sent back to GitHub after processing a webhook
type WebhookResponse struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}
