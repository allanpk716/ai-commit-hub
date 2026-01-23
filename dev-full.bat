@echo off
REM AI Commit Hub - Full Development Environment
REM This script starts both frontend and backend

echo ============================================
echo AI Commit Hub - Full Dev Environment
echo ============================================
echo.

echo [1/3] Starting Frontend Dev Server...
start "AI Commit Hub - Frontend" cmd /k "cd frontend && npm run dev"

echo [2/3] Waiting for frontend to start...
timeout /t 3 /nobreak >nul

echo [3/3] Building and running backend...
go build -o build/bin/ai-commit-hub.exe .
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo ============================================
echo Both services are now running!
echo Frontend: http://localhost:5173
echo Backend:  Running in current window
echo ============================================
echo.

build\bin\ai-commit-hub.exe

pause
