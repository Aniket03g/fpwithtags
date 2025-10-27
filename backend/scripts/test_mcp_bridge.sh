#!/bin/bash

# Test script for MCP Bridge integration
# This script demonstrates the full flow from GitHub URL to project creation

set -e

# Configuration
MCP_BRIDGE_URL="http://127.0.0.1:8089"
BACKEND_URL="http://localhost:8080"
MCP_BRIDGE_API_KEY="${MCP_BRIDGE_API_KEY:-featureplus-local}"
GITHUB_TOKEN="${GITHUB_PERSONAL_ACCESS_TOKEN:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== MCP Bridge Integration Test ===${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}ERROR: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set${NC}"
    echo "Please set: export GITHUB_PERSONAL_ACCESS_TOKEN=ghp_your_token"
    exit 1
fi

# Test 1: Check MCP Bridge health
echo -e "\n${YELLOW}Test 1: Checking MCP Bridge health...${NC}"
if curl -s -f "${MCP_BRIDGE_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ MCP Bridge is healthy${NC}"
else
    echo -e "${RED}✗ MCP Bridge is not running${NC}"
    echo "Please start the bridge: go run ./cmd/mcp-bridge"
    exit 1
fi

# Test 2: Test MCP Bridge analysis directly
echo -e "\n${YELLOW}Test 2: Testing MCP Bridge analysis...${NC}"
REPO_URL="https://github.com/gin-gonic/gin"
echo "Analyzing repository: $REPO_URL"

RESPONSE=$(curl -s -X POST "${MCP_BRIDGE_URL}/mcp/analyze" \
    -H "Authorization: Bearer ${MCP_BRIDGE_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "{\"repo_url\": \"${REPO_URL}\", \"format\": \"featureplus\"}")

if echo "$RESPONSE" | grep -q '"id"'; then
    echo -e "${GREEN}✓ MCP Bridge analysis successful${NC}"
    echo "Response preview:"
    echo "$RESPONSE" | jq -r '. | {id, name, features: (.features | length), tasks: (.tasks | length)}' 2>/dev/null || echo "$RESPONSE" | head -c 200
else
    echo -e "${RED}✗ MCP Bridge analysis failed${NC}"
    echo "Response: $RESPONSE"
    exit 1
fi

# Test 3: Test backend import endpoint (requires authentication)
echo -e "\n${YELLOW}Test 3: Testing backend import (requires authentication)...${NC}"

# First, try to login (adjust credentials as needed)
echo "Attempting to login to FeaturePlus backend..."
LOGIN_RESPONSE=$(curl -s -X POST "${BACKEND_URL}/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email": "test@example.com", "password": "password"}' 2>/dev/null || echo "{}")

JWT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null || echo "")

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
    echo -e "${YELLOW}⚠ Could not authenticate with backend${NC}"
    echo "Skipping backend test. To test the full flow:"
    echo "1. Ensure the backend is running"
    echo "2. Create a test user or adjust login credentials in this script"
    echo ""
    echo "Manual test command:"
    echo "curl -X POST ${BACKEND_URL}/api/imports/github \\"
    echo "  -H 'Authorization: Bearer <your-jwt-token>' \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{\"repo_url\": \"${REPO_URL}\"}'"
else
    echo -e "${GREEN}✓ Authenticated with backend${NC}"
    
    echo "Importing repository via backend..."
    IMPORT_RESPONSE=$(curl -s -X POST "${BACKEND_URL}/api/imports/github" \
        -H "Authorization: Bearer ${JWT_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"repo_url\": \"${REPO_URL}\"}")
    
    if echo "$IMPORT_RESPONSE" | grep -q '"status":"success"'; then
        echo -e "${GREEN}✓ Backend import successful${NC}"
        echo "Response:"
        echo "$IMPORT_RESPONSE" | jq '.' 2>/dev/null || echo "$IMPORT_RESPONSE"
    else
        echo -e "${RED}✗ Backend import failed${NC}"
        echo "Response: $IMPORT_RESPONSE"
    fi
fi

echo -e "\n${GREEN}=== Test Summary ===${NC}"
echo "• MCP Bridge is operational"
echo "• Direct MCP analysis works"
if [ ! -z "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
    echo "• Backend integration works"
else
    echo "• Backend integration not tested (authentication required)"
fi

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Check the /backend/data/imports/ directory for saved templates"
echo "2. View imported projects in the FeaturePlus UI"
echo "3. Monitor logs for any issues"
