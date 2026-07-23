@echo off
REM Prerequisites: Go, Node.js, npm, Wails CLI
echo Building Universal-Ops...
wails build -o universal-ops.exe
if %ERRORLEVEL% EQU 0 (
    echo Build successful: build/bin/universal-ops.exe
) else (
    echo Build failed.
    exit /b %ERRORLEVEL%
)
