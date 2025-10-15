# Test script for Import API endpoints (PowerShell)
# Make sure the server is running before executing this script

$BaseUrl = "http://localhost:8080"
$Token = "your_auth_token_here"

Write-Host "=== Testing Import API Endpoints ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: List available imports
Write-Host "1. Listing available imports..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/imports" `
        -Method Get `
        -Headers @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: Import a project
Write-Host "2. Importing project 'github_project_demo'..." -ForegroundColor Yellow
try {
    $body = @{
        project_id = "github_project_demo"
        project_name = "Test Imported Project"
        description = "This is a test import from GitHub MCP"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$BaseUrl/api/imports/import" `
        -Method Post `
        -Headers @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        } `
        -Body $body
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: Try to import non-existent project (should fail)
Write-Host "3. Testing error handling (non-existent project)..." -ForegroundColor Yellow
try {
    $body = @{
        project_id = "non_existent_project"
        project_name = "Should Fail"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$BaseUrl/api/imports/import" `
        -Method Post `
        -Headers @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        } `
        -Body $body
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Expected error: $_" -ForegroundColor Green
}
Write-Host ""

# Test 4: Test without authentication (should fail)
Write-Host "4. Testing authentication requirement..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/imports" -Method Get
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "Expected error (no auth): $_" -ForegroundColor Green
}
Write-Host ""

Write-Host "=== Tests Complete ===" -ForegroundColor Cyan
