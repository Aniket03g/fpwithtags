# FeaturePlus CLI

A comprehensive command-line interface tool for managing the integration between GitHub Pull Requests and the FeaturePlus project management system. This CLI streamlines the workflow for developers by providing commands to link PRs with features and tasks, manage releases, approve PRs, and more.

## Overview

FeaturePlus CLI helps development teams maintain a clear connection between code changes (PRs) and project management items (features and tasks). It enables:

- Linking GitHub PRs to FeaturePlus features and tasks
- Managing PR approvals and testing status
- Creating and finalizing releases
- Checking out PR branches locally
- Listing and filtering PRs by feature

## Installation

### Prerequisites

- Go 1.20+
- [GitHub CLI](https://cli.github.com/) (`gh`) for GitHub integration
- FeaturePlus backend running (default: `http://localhost:8080`)

### Build & Install

```bash
# Clone the repository
git clone https://github.com/your-org/featureplus-pr.git
cd featureplus-pr

# Build the binary
go build -o featureplus-pr .

# Move to a directory in your PATH (Linux/macOS)
mv featureplus-pr /usr/local/bin/

# For Windows, add the binary location to your PATH or move to a directory in your PATH
```

## Configuration

The CLI reads configuration from a `config.json` file in the current working directory:

```json
{
  "api_url": "https://your-featureplus-instance.com"
}
```

If no config file is found, it defaults to `http://localhost:8080`.

## Authentication

Before using most commands, you need to authenticate with the FeaturePlus backend:

```bash
featureplus-pr login
```

This will prompt for your username/email and password, then store an authentication token locally.

## Available Commands

### Login

Authenticate with the FeaturePlus backend:

```bash
featureplus-pr login
```

### List Pull Requests

List all pull requests or filter by feature ID:

```bash
# List all PRs
featureplus-pr list

# Filter by feature ID
featureplus-pr list --feature-id=42
```

### Upload PR Information

Link the current branch's PR with a FeaturePlus feature and task:

```bash
featureplus-pr upload --feature-id=123 --task-id=456 --mark-tested
```

Options:
- `--feature-id` (required): The FeaturePlus feature ID
- `--task-id` (required): The FeaturePlus task ID
- `--mark-tested` (optional): Mark the PR as tested
- `--version` (optional): Specify a version for the PR

### Approve a PR

Approve a pull request in FeaturePlus (and optionally on GitHub):

```bash
featureplus-pr approve --id=42 --comment="Looks good!"
```

Options:
- `--id` (required): The PR ID to approve
- `--comment` (optional): Add a comment with your approval

### Checkout a PR Branch

Fetch and checkout a PR's branch locally:

```bash
featureplus-pr checkout --id=21
```

Options:
- `--id` (required): The PR ID to checkout

### Release Management

Create a new release:

```bash
featureplus-pr release create --name="v1.2.0" --prs=21,22,23
```

List releases:

```bash
featureplus-pr release list
```

Finalize a release:

```bash
featureplus-pr release finalize --id=5
```

## Workflow Examples

### Complete PR Workflow

```bash
# 1. Create a PR on GitHub
gh pr create --title "Add new feature" --body "Implements feature XYZ"

# 2. Link the PR to FeaturePlus
featureplus-pr upload --feature-id=123 --task-id=456

# 3. After testing, mark as tested
featureplus-pr upload --feature-id=123 --task-id=456 --mark-tested

# 4. List PRs for the feature
featureplus-pr list --feature-id=123

# 5. Approve the PR (assuming ID is 42)
featureplus-pr approve --id=42 --comment="Approved after testing"
```

### Release Management Workflow

```bash
# 1. List all approved PRs
featureplus-pr list

# 2. Create a release with selected PRs
featureplus-pr release create --name="v1.2.0" --prs=21,22,23

# 3. Finalize the release (assuming release ID is 5)
featureplus-pr release finalize --id=5
```

### Collaborative Development Workflow

```bash
# 1. List PRs to review
featureplus-pr list

# 2. Checkout a colleague's PR to review locally
featureplus-pr checkout --id=21

# 3. After review, approve the PR
featureplus-pr approve --id=21 --comment="Code looks good, tests pass"
```

## Debugging

Set the `DEBUG` environment variable to enable verbose output:

```bash
DEBUG=1 featureplus-pr list
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)