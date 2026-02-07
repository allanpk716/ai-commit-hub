@echo off
REM AI Commit Hub Update System Test Script
REM This script launches the app in test mode to verify the auto-update system

echo ========================================
echo AI Commit Hub Update System Test
echo ========================================
echo.

REM Set test mode environment variable
set AI_COMMIT_HUB_TEST_MODE=true

echo [INFO] Test mode enabled
echo [INFO] Using v0.2.0-beta.1 Release (14.3 MB)
echo [INFO] Check https://github.com/allanpk716/ai-commit-hub/releases/tag/v0.2.0-beta.1
echo.

echo [STEP] Starting application...
echo.

REM Start the application
wails dev

echo.
echo [INFO] Application closed
pause
