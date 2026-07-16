@echo off
REM Prerequisites: Go, Node.js, npm, Wails CLI
echo Building OpsForAll GUI...
wails build -o opsforall-gui.exe
if %ERRORLEVEL% EQU 0 (
    echo Build successful: build/bin/opsforall-gui.exe
) else (
    echo Build failed.
    exit /b %ERRORLEVEL%
)
