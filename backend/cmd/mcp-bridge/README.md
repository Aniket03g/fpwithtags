# MCP Bridge Service

A secure local bridge service that interfaces with the GitHub MCP (Model Context Protocol) server to analyze GitHub repositories and generate FeaturePlus-compatible project structures.

## Overview

The MCP Bridge acts as an intermediary between the FeaturePlus backend and the GitHub MCP Docker container. It:
- Accepts HTTP requests with GitHub repository URLs
- Spawns MCP Docker containers for analysis
- Handles stdio communication with Content-Length framing
- Returns structured JSON responses
- Enforces single-concurrency to prevent resource exhaustion

## Architecture

```
FeaturePlus Backend
    ↓ HTTP POST
MCP Bridge (localhost:8089)
    ↓ Docker spawn + stdio
GitHub MCP Container
    ↓ Analysis result
MCP Bridge
    ↓ JSON response
FeaturePlus Backend
```

## Setup

### Prerequisites

- Docker installed and running
- Go 1.21+ (for development)
- GitHub Personal Access Token with `repo` scope

### Environment Variables

Create a `.env` file (see `.env.example`):

```bash
# Required
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token_here
MCP_BRIDGE_API_KEY=your-secure-api-key-here

# Optional
MCP_BRIDGE_PORT=8089                    # Default: 8089
MCP_ANALYSIS_TIMEOUT=90s                # Default: 90s
MCP_CONTAINER_IMAGE=ghcr.io/github/github-mcp-server  # Default
```

## Running the Bridge

### Local Development

```bash
# Set environment variables
export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token_here
export MCP_BRIDGE_API_KEY=featureplus-local

# Run the bridge
go run ./cmd/mcp-bridge

# Or with custom settings
export MCP_BRIDGE_PORT=8090
export MCP_ANALYSIS_TIMEOUT=120s
go run ./cmd/mcp-bridge
```

### Using Docker

```bash
# Build the image
docker build -f cmd/mcp-bridge/Dockerfile -t mcp-bridge .

# Run the container
docker run -d \
  --name mcp-bridge \
  -p 127.0.0.1:8089:8089 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token_here \
  -e MCP_BRIDGE_API_KEY=featureplus-local \
  mcp-bridge
```

## API Endpoints

### POST /mcp/analyze

Analyze a GitHub repository and return FeaturePlus-compatible JSON.

**Request:**
```http
POST http://127.0.0.1:8089/mcp/analyze
Authorization: Bearer featureplus-local
Content-Type: application/json

{
  "repo_url": "https://github.com/username/repository",
  "format": "featureplus"  // Optional, default: "featureplus"
}
```

**Response (Success - 200):**
```json
{
  "id": "repository-name",
  "name": "Repository Name",
  "description": "Repository description",
  "tech_stack": "Node.js",
  "features": [
    {
      "name": "User Authentication",
      "category": "Auth",
      "description": "JWT-based authentication",
      "context": "Development"
    }
  ],
  "tasks": [
    {
      "name": "Setup database",
      "type": "Backend",
      "description": "Initialize PostgreSQL database",
      "priority": "high",
      "context": "Development"
    }
  ],
  "dependencies": ["express", "jsonwebtoken"],
  "setup_steps": ["npm install", "npm run migrate"],
  "environment_variables": ["DATABASE_URL", "JWT_SECRET"]
}
```

**Response (Error - 401):**
```json
{
  "error": "Unauthorized"
}
```

**Response (Error - 429):**
```json
{
  "error": "Analysis already in progress. Please try again later."
}
```

### GET /health

Health check endpoint.

**Request:**
```http
GET http://127.0.0.1:8089/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "mcp-bridge"
}
```

## Testing

### Basic Test

```bash
# Test health endpoint
curl http://127.0.0.1:8089/health

# Test analysis (replace with your API key)
curl -X POST http://127.0.0.1:8089/mcp/analyze \
  -H "Authorization: Bearer featureplus-local" \
  -H "Content-Type: application/json" \
  -d '{
    "repo_url": "https://github.com/expressjs/express",
    "format": "featureplus"
  }'
```

### Full Integration Test

```bash
# 1. Start the bridge
export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token
export MCP_BRIDGE_API_KEY=test-key
go run ./cmd/mcp-bridge

# 2. In another terminal, test the bridge
curl -X POST http://127.0.0.1:8089/mcp/analyze \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/gin-gonic/gin"}' \
  | jq .

# 3. Test with FeaturePlus backend (assuming it's running)
curl -X POST http://localhost:8080/api/imports/from-github \
  -H "Authorization: Bearer <your-featureplus-jwt>" \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/gin-gonic/gin"}'
```

## Security Features

1. **Local-only binding**: Service binds to 127.0.0.1, not accessible from network
2. **API Key authentication**: All requests require valid Bearer token
3. **Single-concurrency**: Only one analysis runs at a time (mutex-protected)
4. **Timeout protection**: Each analysis has a configurable timeout (default 90s)
5. **Container isolation**: MCP containers run with `--network=none`
6. **Input validation**: Basic GitHub URL validation
7. **No token persistence**: Tokens are only in memory/environment

## Troubleshooting

### Bridge won't start

- Check that `MCP_BRIDGE_API_KEY` and `GITHUB_PERSONAL_ACCESS_TOKEN` are set
- Verify port 8089 is not already in use
- Check Docker is running: `docker version`

### Analysis fails

- Check Docker can pull the MCP image: `docker pull ghcr.io/github/github-mcp-server`
- Verify GitHub token has `repo` scope
- Check bridge logs for MCP stderr output
- Try increasing timeout: `export MCP_ANALYSIS_TIMEOUT=180s`

### 429 Too Many Requests

- The bridge enforces single-concurrency
- Wait for current analysis to complete
- Check bridge logs to see if analysis is stuck

### Invalid GitHub URL

- URL must start with `https://github.com/`
- Must have format: `https://github.com/owner/repository`

## Development

### Project Structure

```
cmd/mcp-bridge/
├── main.go          # Bridge service implementation
├── Dockerfile       # Container build file
├── README.md        # This file
└── .env.example     # Environment variable template
```

### Key Functions

- `handleAnalyze()`: HTTP handler for /mcp/analyze
- `runAnalysis()`: Spawns MCP container and manages communication
- `writeFramedMessage()`: Implements Content-Length framing for requests
- `readFramedResponse()`: Parses Content-Length framed responses

### Content-Length Framing

The MCP server uses Content-Length framing for stdio communication:

```
Content-Length: 123\r\n
\r\n
{"json": "payload"}
```

The bridge implements this protocol for both writing requests and reading responses.

## Notes

- The MCP container image (`ghcr.io/github/github-mcp-server`) must be accessible
- Each analysis spawns a new container (stateless)
- Container logs (stderr) are captured for debugging
- The bridge does not cache results - each request triggers fresh analysis

## License

Part of the FeaturePlus project.
