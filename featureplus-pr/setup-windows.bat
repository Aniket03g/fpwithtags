@echo off
REM FeaturePlus CLI Setup Script for Windows (Batch version)

echo ========================================
echo   FeaturePlus CLI Setup for Windows
echo ========================================
echo.

REM Get the current directory
set CLI_DIR=%~dp0
set CLI_DIR=%CLI_DIR:~0,-1%

echo CLI Directory: %CLI_DIR%
echo.

REM Step 1: Build the executable
echo [1/3] Building featureplus-pr.exe...
cd /d "%CLI_DIR%"
go build -o featureplus-pr.exe

if %ERRORLEVEL% EQU 0 (
    echo [32m✓ Build successful![0m
) else (
    echo [31m✗ Build failed![0m
    pause
    exit /b 1
)

echo.
echo [2/3] To add to PATH, please run this PowerShell command:
echo.
echo [33m[Environment]::SetEnvironmentVariable("Path", $env:Path + ";%CLI_DIR%", "User")[0m
echo.
echo Or use the PowerShell setup script: setup-windows.ps1
echo.

REM Step 3: Test
echo [3/3] Testing...
"%CLI_DIR%\featureplus-pr.exe" --version

echo.
echo ========================================
echo   Build Complete!
echo ========================================
echo.
echo To complete setup, either:
echo   1. Run: setup-windows.ps1 (PowerShell - Recommended)
echo   2. Manually add to PATH (see SETUP.md)
echo.
pause
