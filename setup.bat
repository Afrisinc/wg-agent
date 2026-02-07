@echo off
REM WireGuard Agent Setup Script for Windows

setlocal enabledelayedexpansion

echo.
echo WireGuard Agent Setup
echo =====================
echo.

REM Check Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed.
    echo Please install Go 1.21+ from https://golang.org/
    pause
    exit /b 1
)

echo OK: Go is installed
for /f "tokens=*" %%i in ('go version') do set GOVERSION=%%i
echo %GOVERSION%
echo.

REM Create .env file
if exist ".env" (
    echo Warning: .env file already exists
    set /p OVERWRITE="Do you want to overwrite it? (y/N): "
    if /i "!OVERWRITE!"=="y" (
        copy /Y .env.example .env >nul
        echo OK: .env file created from example
    )
) else (
    copy /Y .env.example .env >nul
    echo OK: .env file created from example
)

echo.
echo Downloading dependencies...
call go mod download
if %errorlevel% neq 0 (
    echo Error: Failed to download dependencies
    pause
    exit /b 1
)
echo OK: Dependencies downloaded

echo.
echo Building the application...
call go build -o wg-agent.exe
if %errorlevel% neq 0 (
    echo Error: Failed to build
    pause
    exit /b 1
)
echo OK: Build complete: wg-agent.exe

echo.
echo OK: Setup complete!
echo.
echo Next steps:
echo.
echo 1. Start the server:
echo    set /p API_KEY=(API Key from .env file):
echo    wg-agent.exe
echo.
echo 2. Access Swagger UI:
echo    http://localhost:9999/swagger/
echo.
echo 3. Test the API:
echo    curl -H "X-API-Key: YOUR_API_KEY" http://localhost:9999/health
echo.
echo 4. Read the documentation:
echo    type QUICK_START.md
echo.

pause
