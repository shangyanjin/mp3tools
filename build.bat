@echo off

REM Set Go environment variables
set GO111MODULE=on

REM Check input parameters
if "%1"=="" (
    set OUTPUT=mp3tools
) else (
    set OUTPUT=%1
)

REM Step 1: Build for Linux
echo Building for Linux...
REM Remove existing output file if exists
if exist %OUTPUT% del /f /q %OUTPUT%
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags="-s -w" -o %OUTPUT% ./cmd/mp3tools
if %errorlevel% neq 0 (
    echo Linux build failed!
    exit /b 1
)
echo Linux build succeeded.

REM Step 2: Build for Windows
echo Building for Windows...
REM Remove existing output file if exists
if exist %OUTPUT%.exe del /f /q %OUTPUT%.exe
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %OUTPUT%.exe ./cmd/mp3tools
if %errorlevel% neq 0 (
    echo Windows build failed!
    exit /b 1
)
echo Windows build succeeded.

echo All builds completed successfully.
exit /b 0

