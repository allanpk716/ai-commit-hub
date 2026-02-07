@echo off
REM Test auto-update functionality script
REM Usage: Double-click to run or execute in command line

echo ========================================
echo AI Commit Hub - Auto Update Test Mode
echo ========================================
echo.
echo This script will start the app with test mode enabled
echo Test mode simulates a new version detection for testing download functionality
echo.

REM Check if wails is installed
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] wails command not found
    echo Please install wails first: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    pause
    exit /b 1
)

REM Set test mode environment variable
set AI_COMMIT_HUB_TEST_MODE=true

echo [OK] Test mode enabled
echo.
echo Test information:
echo   - Test version: v0.2.0-beta.1
echo   - Download URL: GitHub Releases
echo   - File size: ~14.3 MB
echo.
echo Starting application...
echo.

REM Start development server
wails dev
