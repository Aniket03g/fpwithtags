# GitHub App Boilerplate in Go

A production-ready boilerplate for building GitHub Apps with Go, featuring JWT authentication, webhook handling, and GitHub API integration.

## Features

- 🔐 JWT-based authentication for GitHub Apps
- 🔄 Webhook handling with signature verification
- 🚀 Easy GitHub API integration
- 🏗️ Clean, modular project structure
- ⚡ Graceful server shutdown
- 📝 Environment-based configuration

## Prerequisites

- Go 1.21 or higher
- A GitHub App (create one [here](https://github.com/settings/apps))
- Private key for your GitHub App

## Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/github-app-boilerplate.git
   cd github-app-boilerplate
   ```

2. **Set up environment variables**
   - Copy `.env.example` to `.env`
   - Fill in your GitHub App details
   ```bash
   cp .env.example .env
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GITHUB_APP_ID` | Your GitHub App ID | Yes |
| `GITHUB_PRIVATE_KEY_PEM` | Your GitHub App's private key in PEM format | Yes |
| `GITHUB_WEBHOOK_SECRET` | Webhook secret for verifying payloads | Yes |
| `GITHUB_APP_SLUG` | Your GitHub App's slug/name | Yes |
| `PORT` | Port to run the server on (default: 8080) | No |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | No |

## Project Structure

```
.
├── config/           # Configuration management
├── github/           # GitHub API client and services
├── handlers/         # HTTP request handlers
├── models/           # Data models and DTOs
├── routes/           # HTTP route definitions
├── .env.example      # Example environment variables
├── go.mod           # Go module definition
├── go.sum           # Go dependencies checksums
├── main.go          # Application entry point
└── README.md        # This file
```

## Webhook Setup

1. Set your webhook URL to `https://your-domain.com/api/github/webhooks`
2. Set the content type to `application/json`
3. Add a webhook secret (must match `GITHUB_WEBHOOK_SECRET`)
4. Subscribe to the following events:
   - Pull requests
   - Installation

## Testing

To test the webhook locally, you can use [ngrok](https://ngrok.com/):

```bash
ngrok http 8080
```

Then update your GitHub App's webhook URL to use the ngrok URL.

## License

MIT
