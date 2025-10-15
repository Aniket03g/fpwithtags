#!/bin/bash

# Test script for Import API endpoints
# Make sure the server is running before executing this script

BASE_URL="http://localhost:8080"
TOKEN="your_auth_token_here"

echo "=== Testing Import API Endpoints ==="
echo ""

# Test 1: List available imports
echo "1. Listing available imports..."
curl -X GET "$BASE_URL/api/imports" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
echo -e "\n"

# Test 2: Import a project
echo "2. Importing project 'github_project_demo'..."
curl -X POST "$BASE_URL/api/imports/import" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "github_project_demo",
    "project_name": "Test Imported Project",
    "description": "This is a test import from GitHub MCP"
  }'
echo -e "\n"

# Test 3: Try to import non-existent project (should fail)
echo "3. Testing error handling (non-existent project)..."
curl -X POST "$BASE_URL/api/imports/import" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "non_existent_project",
    "project_name": "Should Fail"
  }'
echo -e "\n"

# Test 4: Test without authentication (should fail)
echo "4. Testing authentication requirement..."
curl -X GET "$BASE_URL/api/imports"
echo -e "\n"

echo "=== Tests Complete ==="
