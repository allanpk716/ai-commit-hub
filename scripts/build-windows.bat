@echo off
REM AI Commit Hub Windows Build Script
REM This script ensures the application icon is embedded in the executable

setlocal enabledelayedexpansion

echo ========================================
echo AI Commit Hub - Windows Build Script
echo ========================================
echo.

REM Check if go-winres is installed
where go-winres >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [1/5] Installing go-winres tool...
    go install github.com/tc-hib/go-winres@latest
    if %ERRORLEVEL% NEQ 0 (
        echo ERROR: Failed to install go-winres tool
        pause
        exit /b 1
    )
    echo go-winres tool installed successfully
) else (
    echo [1/5] go-winres tool already installed
)

REM Check if Python is available for icon generation
where python >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [2/5] Python not found, skipping icon generation
    echo WARNING: Icon generation may be incomplete
) else (
    echo [2/5] Generating multi-size icon PNG files...
    python scripts\prepare_icons.py
    echo Icon files generated
)

REM Generate resource files
echo.
echo [3/5] Generating Windows resource files...
go-winres make
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to generate resource files
    pause
    exit /b 1
)
echo Resource files generated

REM Build the application
echo.
echo [4/5] Building application with wails...
wails build -clean
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: wails build failed
    pause
    exit /b 1
)
echo Application built successfully

REM Verify build output
echo.
echo [5/5] Verifying build output...
if exist build\bin\ai-commit-hub.exe (
    echo Executable found: build\bin\ai-commit-hub.exe
    echo.
    echo Resource files included:
    dir /b *.syso 2>nul
    echo.
    echo NOTE: The icon should now display correctly at ALL sizes
    echo (16x16, 32x32, 48x48, 64x64, 128x128, 256x256)
) else (
    echo ERROR: Build output not found
    pause
    exit /b 1
)

echo.
echo ========================================
echo Build completed successfully!
echo.
echo Output: build\bin\ai-commit-hub.exe
echo.
echo Icons included:
echo   - 16x16  (small icons, list view)
echo   - 32x32  (standard icons)
echo   - 48x48  (large icons)
echo   - 64x64  (extra large icons)
echo   - 128x128 (extra extra large)
echo   - 256x256 (high DPI displays)
echo.
echo If icons still don't display correctly:
echo   1. Run: scripts\clear-icon-cache.bat
echo   2. Or restart Windows Explorer
echo   3. Or change folder view and back
echo ========================================
echo.

pause
