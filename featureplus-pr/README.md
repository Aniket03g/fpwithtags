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

## Configuration

The CLI automatically looks for a `config.json` file in multiple locations (in order of preference):

1. **Current working directory** - `./config.json`
2. **Executable directory** - `[executable-dir]/config.json`
3. **User home directory** - `~/.featureplus-pr/config.json` or `~/config.json`

If no config file is found, it defaults to `http://localhost:8080`.

### Creating a config file

Create a `config.json` file in any of the locations above:

```json
{
  "api_url": "https://featureplus.onrender.com"
}
```

### Checking your configuration

Run the config command to see current settings and available config file locations:

```bash
featureplus-pr config
```

This will show:
- Current API URL being used
- All possible config file locations
- Which config files exist (✅) or don't exist (❌)

## Commands

- `featureplus-pr` - Show help and current configuration
- `featureplus-pr config` - Show detailed configuration information
- `featureplus-pr upload` - Upload current branch's PR info to FeaturePlus
- `featureplus-pr list` - List features and tasks from FeaturePlus
- `featureplus-pr checkout` - Checkout a feature or task

## How it works
- Fetches PR info from the current branch using the GitHub CLI
- Combines with your flags
- Sends a POST request to your FeaturePlus backend API

## Build & Install

### Quick Build (Windows)
```bash
cd featureplus-pr
build.bat
```

### Quick Build (Linux/Mac)
```bash
cd featureplus-pr
chmod +x build.sh
./build.sh
```

### Manual Build
```bash
cd featureplus-pr
# Build the binary
go build -o featureplus-pr .
# Move to a directory in your PATH (optional)
mv featureplus-pr /usr/local/bin/
```

## Requirements
- Go 1.20+
- [GitHub CLI](https://cli.github.com/) (`gh`)
- FeaturePlus backend running (default: `http://localhost:8080`)

## Example

```bash
# Check your configuration
featureplus-pr config

# Upload a PR
featureplus-pr upload --feature-id=42 --task-id=99 --mark-tested
``` 