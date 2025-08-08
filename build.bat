@echo off
echo Building GoProxy...

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM Install dependencies
echo Installing dependencies...
go mod tidy

REM Build the main application
echo Building main application...
go build -o goproxy.exe main.go

REM Build the test server
echo Building test server...
go build -tags testserver -o test_server.exe test_server.go

echo Build complete!
echo.
echo To run the proxy server:
echo   goproxy.exe
echo.
echo To run the test backend server:
echo   test_server.exe
echo.
pause 