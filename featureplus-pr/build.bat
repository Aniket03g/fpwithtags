@echo off
echo Building FeaturePlus PR CLI...
echo.

REM Build for Windows
go build -o featureplus-pr.exe .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Executable created: featureplus-pr.exe
    echo.
    echo You can now run the tool from anywhere with:
    echo   featureplus-pr.exe config
    echo   featureplus-pr.exe upload --feature-id 1 --task-id 1
    echo.
) else (
    echo.
    echo Build failed! Please check for errors above.
    echo.
)

pause 