@echo off
echo Testing Go compilation...
echo.

go build -o test_backend.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo [32m✓ Compilation successful![0m
    echo.
    echo The following routes are now available:
    echo   POST /api/projects/:id/connect
    echo   GET  /api/projects/:id/status
    echo.
    del test_backend.exe
) else (
    echo [31m✗ Compilation failed![0m
    exit /b 1
)
