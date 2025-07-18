# featureplus-pr

A CLI tool to link GitHub Pull Requests with FeaturePlus features and tasks.

## Usage

After creating a PR with `gh pr create`, run:

```
featureplus-pr upload --feature-id=123 --task-id=456 --mark-tested
```

- `--feature-id` (required): The FeaturePlus feature ID
- `--task-id` (required): The FeaturePlus task ID
- `--mark-tested` (optional): Mark the PR as tested

## How it works
- Fetches PR info from the current branch using the GitHub CLI
- Combines with your flags
- Sends a POST request to your FeaturePlus backend at `http://localhost:8080/api/pr`

## Build & Install

```
cd featureplus-pr
# Build the binary
go build -o featureplus-pr .
# Move to a directory in your PATH
mv featureplus-pr /usr/local/bin/
```

## Requirements
- Go 1.20+
- [GitHub CLI](https://cli.github.com/) (`gh`)
- FeaturePlus backend running at `http://localhost:8080`

## Example

```
featureplus-pr upload --feature-id=42 --task-id=99 --mark-tested
``` 