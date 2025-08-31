package github

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v56/github"
)

type RepositoryService interface {
	GetPullRequests(owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error)
}

type RepositoryClient struct {
	client *github.Client
}

func NewRepositoryClient(githubClient *github.Client) *RepositoryClient {
	return &RepositoryClient{
		client: githubClient,
	}
}

// GetPullRequests fetches pull requests for the given repository
func (r *RepositoryClient) GetPullRequests(owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Fetching pull requests for %s/%s\n", owner, repo)
	}

	if opts == nil {
		opts = &github.PullRequestListOptions{
			State: "open",
			ListOptions: github.ListOptions{
				PerPage: 30, // Default page size
			},
		}
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Calling GitHub API: PullRequests.List for %s/%s with state=%s\n", 
			owner, repo, opts.State)
	}

	prs, _, err := r.client.PullRequests.List(context.Background(), owner, repo, opts)
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Error fetching pull requests: %v\n", err)
		}
		return nil, err
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully fetched %d pull requests from %s/%s\n", len(prs), owner, repo)
	}

	log.Printf("Found %d pull requests in %s/%s\n", len(prs), owner, repo)
	return prs, nil
}

// GetPullRequestFiles fetches the list of files changed in a pull request
func (r *RepositoryClient) GetPullRequestFiles(owner, repo string, number int) ([]*github.CommitFile, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Fetching files for PR #%d in %s/%s\n", number, owner, repo)
	}

	files, _, err := r.client.PullRequests.ListFiles(
		context.Background(),
		owner,
		repo,
		number,
		&github.ListOptions{PerPage: 100},
	)

	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Error fetching PR files: %v\n", err)
		}
		return nil, err
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully fetched %d files for PR #%d\n", len(files), number)
	}

	return files, nil
}

// CreateComment creates a new comment on a pull request
func (r *RepositoryClient) CreateComment(owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Creating comment on PR #%d in %s/%s\n", number, owner, repo)
	}

	newComment, _, err := r.client.Issues.CreateComment(
		context.Background(),
		owner,
		repo,
		number,
		comment,
	)

	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Error creating comment: %v\n", err)
		}
		return nil, err
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully created comment with ID: %d\n", *newComment.ID)
	}

	return newComment, nil
}

// GetRepository fetches repository details
func (r *RepositoryClient) GetRepository(owner, repo string) (*github.Repository, error) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Fetching repository details for %s/%s\n", owner, repo)
	}

	repository, _, err := r.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		if os.Getenv("DEBUG") == "1" {
			fmt.Printf("[DEBUG][PR_RELEASE] Error fetching repository details: %v\n", err)
		}
		return nil, err
	}

	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("[DEBUG][PR_RELEASE] Successfully fetched repository details for %s/%s\n", owner, repo)
	}

	return repository, nil
}
