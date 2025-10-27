# Test script for MCP Bridge integration (PowerShell version)
# This script demonstrates the full flow from GitHub URL to project creation

$ErrorActionPreference = "Stop"

# Configuration
$MCP_BRIDGE_URL = "http://127.0.0.1:8089"
$BACKEND_URL = "http://localhost:8080"
$MCP_BRIDGE_API_KEY = if ($env:MCP_BRIDGE_API_KEY) { $env:MCP_BRIDGE_API_KEY } else { "featureplus-local" }
$GITHUB_TOKEN = $env:GITHUB_PERSONAL_ACCESS_TOKEN

Write-Host "=== MCP Bridge Integration Test ===" -ForegroundColor Green
Write-Host ""

# Check prerequisites
Write-Host "Checking prerequisites..." -ForegroundColor Yellow

if (-not $GITHUB_TOKEN) {
    Write-Host "ERROR: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set" -ForegroundColor Red
    Write-Host "Please set: `$env:GITHUB_PERSONAL_ACCESS_TOKEN = 'ghp_your_token'"
    exit 1
}

# Test 1: Check MCP Bridge health
Write-Host "`nTest 1: Checking MCP Bridge health..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$MCP_BRIDGE_URL/health" -Method Get
    Write-Host "✓ MCP Bridge is healthy" -ForegroundColor Green
} catch {
    Write-Host "✗ MCP Bridge is not running" -ForegroundColor Red
    Write-Host "Please start the bridge: go run ./cmd/mcp-bridge"
    exit 1
}

# Test 2: Test MCP Bridge analysis directly
Write-Host "`nTest 2: Testing MCP Bridge analysis..." -ForegroundColor Yellow
$REPO_URL = "https://github.com/gin-gonic/gin"
Write-Host "Analyzing repository: $REPO_URL"

$headers = @{
    "Authorization" = "Bearer $MCP_BRIDGE_API_KEY"
    "Content-Type" = "application/json"
}

$body = @{
    repo_url = $REPO_URL
    format = "featureplus"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$MCP_BRIDGE_URL/mcp/analyze" -Method Post -Headers $headers -Body $body
    Write-Host "✓ MCP Bridge analysis successful" -ForegroundColor Green
    Write-Host "Response preview:"
    Write-Host "ID: $($response.id)"
    Write-Host "Name: $($response.name)"
    Write-Host "Features: $($response.features.Count)"
    Write-Host "Tasks: $($response.tasks.Count)"
} catch {
    Write-Host "✗ MCP Bridge analysis failed" -ForegroundColor Red
    Write-Host "Error: $_"
    exit 1
}

# Test 3: Test backend import endpoint (requires authentication)
Write-Host "`nTest 3: Testing backend import (requires authentication)..." -ForegroundColor Yellow

# First, try to login (adjust credentials as needed)
Write-Host "Attempting to login to FeaturePlus backend..."

$loginBody = @{
    email = "test@example.com"
    password = "password"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BACKEND_URL/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $JWT_TOKEN = $loginResponse.token
    
    if ($JWT_TOKEN) {
        Write-Host "✓ Authenticated with backend" -ForegroundColor Green
        
        Write-Host "Importing repository via backend..."
        $importHeaders = @{
            "Authorization" = "Bearer $JWT_TOKEN"
            "Content-Type" = "application/json"
        }
        
        $importBody = @{
            repo_url = $REPO_URL
        } | ConvertTo-Json
        
        try {
            $importResponse = Invoke-RestMethod -Uri "$BACKEND_URL/api/imports/github" -Method Post -Headers $importHeaders -Body $importBody
            Write-Host "✓ Backend import successful" -ForegroundColor Green
            Write-Host "Response:"
            $importResponse | ConvertTo-Json -Depth 3
        } catch {
            Write-Host "✗ Backend import failed" -ForegroundColor Red
            Write-Host "Error: $_"
        }
    }
} catch {
    Write-Host "⚠ Could not authenticate with backend" -ForegroundColor Yellow
    Write-Host "Skipping backend test. To test the full flow:"
    Write-Host "1. Ensure the backend is running"
    Write-Host "2. Create a test user or adjust login credentials in this script"
    Write-Host ""
    Write-Host "Manual test command:"
    Write-Host @"
Invoke-RestMethod -Uri "$BACKEND_URL/api/imports/github" ``
    -Method Post ``
    -Headers @{
        "Authorization" = "Bearer <your-jwt-token>"
        "Content-Type" = "application/json"
    } ``
    -Body '{"repo_url": "$REPO_URL"}'
"@
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "• MCP Bridge is operational"
Write-Host "• Direct MCP analysis works"
if ($JWT_TOKEN) {
    Write-Host "• Backend integration works"
} else {
    Write-Host "• Backend integration not tested (authentication required)"
}

Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Check the /backend/data/imports/ directory for saved templates"
Write-Host "2. View imported projects in the FeaturePlus UI"
Write-Host "3. Monitor logs for any issues"
