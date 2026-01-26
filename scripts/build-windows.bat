@echo off
REM AI Commit Hub Windows Build Script
REM This script ensures the application icon is embedded in the executable

setlocal enabledelayedexpansion

echo ========================================
echo AI Commit Hub - Windows Build Script
echo ========================================
echo.

REM Check if rsrc tool is installed
where rsrc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [1/5] Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
    if %ERRORLEVEL% NEQ 0 (
        echo ERROR: Failed to install rsrc tool
        pause
        exit /b 1
    )
    echo rsrc tool installed successfully
) else (
    echo [1/5] rsrc tool already installed
)

REM Generate .syso file from icon
echo.
echo [2/5] Generating Windows resource file...
rsrc -ico build/windows/icon.ico -o icon.syso
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to generate .syso file
    pause
    exit /b 1
)
echo icon.syso generated successfully

REM Build the application
echo.
echo [3/5] Building application with wails...
wails build -clean
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: wails build failed
    pause
    exit /b 1
)
echo Application built successfully

REM Verify build output
echo.
echo [4/5] Verifying build output...
if exist build\bin\ai-commit-hub.exe (
    echo Executable found: build\bin\ai-commit-hub.exe

    REM Check if icon is embedded
    powershell -Command "Add-Type -AssemblyName System.Drawing; $icon = [System.Drawing.Icon]::ExtractAssociatedIcon('build\bin\ai-commit-hub.exe'); Write-Host 'Icon size:' $icon.Width 'x' $icon.Height"
    if %ERRORLEVEL% EQU 0 (
        echo Icon verification: PASSED
    ) else (
        echo WARNING: Could not verify icon
    )
) else (
    echo ERROR: Build output not found
    pause
    exit /b 1
)

echo.
echo [5/5] Build completed successfully!
echo.
echo ========================================
echo Output: build\bin\ai-commit-hub.exe
echo.
echo NOTE: If the icon doesn't appear in File Explorer,
echo run: scripts\clear-icon-cache.bat
echo ========================================
echo.

pause
