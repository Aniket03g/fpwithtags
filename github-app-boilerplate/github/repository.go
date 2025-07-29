package github

import (
	"context"
	"log"

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
	if opts == nil {
		opts = &github.PullRequestListOptions{
			State: "open",
			ListOptions: github.ListOptions{
				PerPage: 30, // Default page size
			},
		}
	}

	prs, _, err := r.client.PullRequests.List(context.Background(), owner, repo, opts)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d pull requests in %s/%s\n", len(prs), owner, repo)
	return prs, nil
}

// GetPullRequestFiles fetches the list of files changed in a pull request
func (r *RepositoryClient) GetPullRequestFiles(owner, repo string, number int) ([]*github.CommitFile, error) {
	files, _, err := r.client.PullRequests.ListFiles(
		context.Background(),
		owner,
		repo,
		number,
		&github.ListOptions{PerPage: 100},
	)

	if err != nil {
		return nil, err
	}

	return files, nil
}

// CreateComment creates a new comment on a pull request
func (r *RepositoryClient) CreateComment(owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, error) {
	newComment, _, err := r.client.Issues.CreateComment(
		context.Background(),
		owner,
		repo,
		number,
		comment,
	)

	if err != nil {
		return nil, err
	}

	return newComment, nil
}

// GetRepository fetches repository details
func (r *RepositoryClient) GetRepository(owner, repo string) (*github.Repository, error) {
	repository, _, err := r.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return nil, err
	}

	return repository, nil
}
