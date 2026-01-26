@echo off
REM Clear Windows Icon Cache

echo Killing Windows Explorer...
taskkill /f /im explorer.exe

echo Clearing icon cache...
del /a /q "%userprofile%\AppData\Local\IconCache.db"
del /a /f /q "%userprofile%\AppData\Local\Microsoft\Windows\Explorer\*.db"

echo Restarting Windows Explorer...
start explorer.exe

echo Icon cache cleared! Please check if the icon appears correctly.
pause
