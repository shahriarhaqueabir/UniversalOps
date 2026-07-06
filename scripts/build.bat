@echo off
REM Build script for Hawkward on Windows
REM Requires Go 1.26.4+

echo Building Hawkward...
go build -ldflags="-s -w" -o hawkward.exe .\cmd\hawkward\
if %ERRORLEVEL% EQU 0 (
    echo Build successful: hawkward.exe
) else (
    echo Build failed with error code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
