@echo off
REM Build script for Hawkward GUI on Windows
REM Requires: Go 1.26.4+, Node.js, npm, Wails CLI
echo Building Hawkward GUI...
wails build -o hawkward-gui.exe
if %ERRORLEVEL% EQU 0 (
    echo Build successful: build/bin/hawkward-gui.exe
) else (
    echo Build failed
    exit /b %ERRORLEVEL%
)
