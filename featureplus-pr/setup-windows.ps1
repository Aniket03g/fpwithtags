# FeaturePlus CLI Setup Script for Windows
# This script builds the CLI and adds it to your PATH

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  FeaturePlus CLI Setup for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get the current directory (where the CLI source is)
$CLI_DIR = $PSScriptRoot
Write-Host "CLI Directory: $CLI_DIR" -ForegroundColor Yellow

# Step 1: Build the executable
Write-Host ""
Write-Host "[1/3] Building featureplus-pr.exe..." -ForegroundColor Green
try {
    Set-Location $CLI_DIR
    go build -o featureplus-pr.exe
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build successful!" -ForegroundColor Green
    } else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Error building: $_" -ForegroundColor Red
    exit 1
}

# Step 2: Check if already in PATH
Write-Host ""
Write-Host "[2/3] Checking PATH..." -ForegroundColor Green
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -like "*$CLI_DIR*") {
    Write-Host "✅ Already in PATH!" -ForegroundColor Green
} else {
    Write-Host "⚠️  Not in PATH. Adding now..." -ForegroundColor Yellow
    
    # Add to User PATH
    $newPath = $currentPath + ";$CLI_DIR"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    
    # Also update current session
    $env:Path += ";$CLI_DIR"
    
    Write-Host "✅ Added to PATH!" -ForegroundColor Green
}

# Step 3: Verify installation
Write-Host ""
Write-Host "[3/3] Verifying installation..." -ForegroundColor Green
Write-Host ""

# Refresh PATH for current session
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

# Test the command
try {
    $version = & "$CLI_DIR\featureplus-pr.exe" --version 2>&1 | Select-String "FeaturePlus CLI"
    if ($version) {
        Write-Host "✅ Installation successful!" -ForegroundColor Green
        Write-Host ""
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host "  Setup Complete!" -ForegroundColor Cyan
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "You can now use 'featureplus-pr' from any directory!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Try these commands:" -ForegroundColor Yellow
        Write-Host "  featureplus-pr --help" -ForegroundColor White
        Write-Host "  featureplus-pr init" -ForegroundColor White
        Write-Host "  featureplus-pr connect <project-id>" -ForegroundColor White
        Write-Host "  featureplus-pr status" -ForegroundColor White
        Write-Host ""
        Write-Host "⚠️  IMPORTANT: Close and reopen your terminal for PATH changes to take effect!" -ForegroundColor Yellow
        Write-Host ""
    }
} catch {
    Write-Host "⚠️  Warning: Could not verify installation" -ForegroundColor Yellow
    Write-Host "   Please restart your terminal and try: featureplus-pr --version" -ForegroundColor Yellow
}
