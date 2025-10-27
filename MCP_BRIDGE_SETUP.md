# MCP Bridge Setup Guide

## Overview

The MCP Bridge is a secure local service that interfaces between FeaturePlus and the GitHub MCP Docker container. It replaces the previous direct remote MCP API calls that were failing with "Invalid session ID" errors.

## Architecture

```
FeaturePlus Backend (Go + HTMX)
    ↓ HTTP POST to localhost:8089
MCP Bridge Service (Go)
    ↓ Spawns Docker container
GitHub MCP Container (stdio communication)
    ↓ Content-Length framed messages
MCP Bridge Service
    ↓ Returns JSON
FeaturePlus Backend
    ↓ Creates project/features/tasks
```

## Quick Start

### 1. Set Environment Variables

```bash
# Linux/Mac
export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token_here
export MCP_BRIDGE_API_KEY=featureplus-local

# Windows PowerShell
$env:GITHUB_PERSONAL_ACCESS_TOKEN = "ghp_your_token_here"
$env:MCP_BRIDGE_API_KEY = "featureplus-local"
```

### 2. Start the MCP Bridge

```bash
cd backend
go run ./cmd/mcp-bridge
```

You should see:
```
INFO: MCP Bridge starting on 127.0.0.1:8089
INFO: Using MCP container image: ghcr.io/github/github-mcp-server
INFO: Analysis timeout: 90s
```

### 3. Start FeaturePlus Backend

In a new terminal:
```bash
cd backend
go run main.go
```

### 4. Test the Integration

#### Windows PowerShell:
```powershell
.\scripts\test_mcp_bridge.ps1
```

#### Linux/Mac:
```bash
chmod +x ./scripts/test_mcp_bridge.sh
./scripts/test_mcp_bridge.sh
```

## Configuration

### MCP Bridge Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GITHUB_PERSONAL_ACCESS_TOKEN` | Yes | - | GitHub PAT with `repo` scope |
| `MCP_BRIDGE_API_KEY` | Yes | - | API key for bridge authentication |
| `MCP_BRIDGE_PORT` | No | 8089 | Port to bind the bridge |
| `MCP_ANALYSIS_TIMEOUT` | No | 90s | Timeout for each analysis |
| `MCP_CONTAINER_IMAGE` | No | ghcr.io/github/github-mcp-server | Docker image to use |

### Backend Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MCP_BRIDGE_URL` | No | http://127.0.0.1:8089 | MCP Bridge URL |
| `MCP_BRIDGE_API_KEY` | No | featureplus-local | Must match bridge key |

## API Usage

### Direct Bridge Test

```bash
curl -X POST http://127.0.0.1:8089/mcp/analyze \
  -H "Authorization: Bearer featureplus-local" \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/gin-gonic/gin", "format": "featureplus"}'
```

### Via FeaturePlus Backend

```bash
# Get JWT token first (login)
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}' \
  | jq -r '.token')

# Import from GitHub
curl -X POST http://localhost:8080/api/imports/github \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/gin-gonic/gin"}'
```

## Security Features

1. **Local-only binding**: Bridge binds to 127.0.0.1, not accessible from network
2. **API Key authentication**: All requests require Bearer token
3. **Single-concurrency**: Only one analysis at a time (mutex-protected)
4. **Container isolation**: MCP containers run with `--network=none`
5. **No token persistence**: Tokens only in environment variables
6. **Input validation**: GitHub URL validation and sanitization
7. **Timeout protection**: Each analysis has configurable timeout

## Troubleshooting

### Bridge won't start

```
ERROR: MCP_BRIDGE_API_KEY environment variable is required
```
**Solution**: Set the required environment variables

### Docker not found

```
ERROR: Failed to start MCP container: exec: "docker": executable file not found
```
**Solution**: Install Docker and ensure it's running

### Analysis timeout

```
ERROR: MCP analysis failed: context deadline exceeded
```
**Solution**: Increase timeout: `export MCP_ANALYSIS_TIMEOUT=180s`

### Rate limiting (429)

```
{"error": "Analysis already in progress. Please try again later."}
```
**Solution**: Wait for current analysis to complete (single-concurrency enforced)

### MCP container image not found

```
MCP STDERR: Unable to find image 'ghcr.io/github/github-mcp-server:latest'
```
**Solution**: Pull the image manually: `docker pull ghcr.io/github/github-mcp-server`

## Implementation Details

### Content-Length Framing

The MCP server uses Content-Length framing for stdio communication:

```
Request:
Content-Length: 123\r\n
\r\n
{"repo_url": "...", "format": "..."}

Response:
Content-Length: 456\r\n
\r\n
{"id": "...", "features": [...], "tasks": [...]}
```

### Files Structure

```
backend/
├── cmd/mcp-bridge/
│   ├── main.go              # Bridge service implementation
│   ├── Dockerfile            # Container build file
│   ├── README.md            # Bridge documentation
│   └── .env.example         # Environment template
├── internal/mcpbridge/
│   └── client.go            # Bridge client for backend
├── handlers/
│   └── import_handler.go    # Updated to use bridge
├── services/
│   └── github_mcp_service.go # Deprecated, commented out
└── scripts/
    ├── test_mcp_bridge.sh   # Linux/Mac test script
    └── test_mcp_bridge.ps1  # Windows test script
```

### Key Changes from Previous Implementation

1. **Removed direct MCP calls**: All code calling `https://api.githubcopilot.com/mcp/` has been commented out
2. **Added local bridge**: New service at `cmd/mcp-bridge/` handles MCP communication
3. **Updated import handler**: Now uses `mcpbridge.CallLocalMCPBridge()` instead of direct calls
4. **Better error handling**: Specific error codes for different failure scenarios
5. **Security hardening**: Local-only binding, API keys, container isolation

## Development Workflow

1. **Start services**:
   ```bash
   # Terminal 1: MCP Bridge
   go run ./cmd/mcp-bridge
   
   # Terminal 2: Backend
   go run main.go
   
   # Terminal 3: Frontend (if separate)
   npm run dev
   ```

2. **Test import**:
   - Navigate to Projects page
   - Click "GitHub MCP" button
   - Enter repository URL
   - Wait for analysis (30-60 seconds)
   - View imported project

3. **Monitor logs**:
   - Bridge logs: Shows MCP container stderr
   - Backend logs: Shows import process
   - Check `/backend/data/imports/` for saved templates

## Production Deployment

### Using Docker Compose

```yaml
version: '3.8'
services:
  mcp-bridge:
    build:
      context: ./backend
      dockerfile: cmd/mcp-bridge/Dockerfile
    environment:
      - GITHUB_PERSONAL_ACCESS_TOKEN=${GITHUB_PERSONAL_ACCESS_TOKEN}
      - MCP_BRIDGE_API_KEY=${MCP_BRIDGE_API_KEY}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    ports:
      - "127.0.0.1:8089:8089"
    
  backend:
    build: ./backend
    environment:
      - MCP_BRIDGE_URL=http://mcp-bridge:8089
      - MCP_BRIDGE_API_KEY=${MCP_BRIDGE_API_KEY}
    depends_on:
      - mcp-bridge
    ports:
      - "8080:8080"
```

### Using systemd (Linux)

Create `/etc/systemd/system/mcp-bridge.service`:
```ini
[Unit]
Description=MCP Bridge Service
After=docker.service
Requires=docker.service

[Service]
Type=simple
User=featureplus
WorkingDirectory=/opt/featureplus/backend
Environment="GITHUB_PERSONAL_ACCESS_TOKEN=ghp_..."
Environment="MCP_BRIDGE_API_KEY=..."
ExecStart=/opt/featureplus/backend/mcp-bridge
Restart=always

[Install]
WantedBy=multi-user.target
```

## Summary

The MCP Bridge provides a robust, secure solution for integrating GitHub MCP analysis into FeaturePlus. It replaces the failing direct API calls with a local service that properly handles the MCP container's stdio communication protocol.

Key benefits:
- ✅ Works with actual GitHub MCP container
- ✅ Secure (local-only, API key protected)
- ✅ Reliable (timeout protection, single-concurrency)
- ✅ Observable (comprehensive logging)
- ✅ Maintainable (clean separation of concerns)
