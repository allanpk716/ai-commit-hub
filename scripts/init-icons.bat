@echo off
REM Initialize build icons from source
REM Run this after cloning the repository

echo Copying icon files to build directory...

REM Ensure build directory exists
if not exist "build" mkdir "build"
if not exist "build\windows" mkdir "build\windows"

REM Copy source icon to build directory
copy /Y "assets\icons\appicon.png" "build\appicon.png" >nul 2>&1

echo.
echo [OK] Build icons initialized!
echo You can now run: wails dev
echo.
