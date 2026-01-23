@echo off
REM AI Commit Hub - Development Runner
REM This script compiles and runs the application

echo ============================================
echo AI Commit Hub - Development Build ^& Run
echo ============================================
echo.

echo [1/2] Building application...
go build -o build/bin/ai-commit-hub.exe .
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

echo [2/2] Running application...
echo.
build\bin\ai-commit-hub.exe

pause
