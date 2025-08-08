@echo off
setlocal ENABLEDELAYEDEXPANSION

echo === GoProxy: Build + Run (One-Click) ===

REM 1) Check Go availability
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

REM 2) Kill any existing processes (ignore errors)
for %%P in (goproxy.exe test_server.exe) do (
    taskkill /IM %%P /F >nul 2>&1
)

REM 3) Ensure deps and build binaries
echo Tidying modules...
go mod tidy
if %errorlevel% neq 0 goto :fail

echo Building proxy...
go build -o goproxy.exe main.go
if %errorlevel% neq 0 goto :fail

echo Building test backend...
go build -tags testserver -o test_server.exe test_server.go
if %errorlevel% neq 0 goto :fail

REM 4) Start backend on :8081
echo Starting test backend on :8081 ...
start "GoProxy Test Backend" cmd /c ""%~dp0test_server.exe""

REM 5) Wait for backend readiness (max 30s)
set /a __tries=0
:wait_backend
curl -s -f http://localhost:8081/ >nul 2>&1
if %errorlevel% neq 0 (
    set /a __tries+=1
    if !__tries! geq 30 (
        echo Backend did not become ready. Aborting.
        goto :fail
    )
    timeout /t 1 >nul
    goto :wait_backend
)
echo Backend is up.

REM 6) Start proxy on :8080 (override with %1 if provided)
set __PORT=%1
if "%__PORT%"=="" set __PORT=8080
echo Starting proxy on :%__PORT% ...
start "GoProxy" cmd /c ""%~dp0goproxy.exe" -port %__PORT% -backend http://localhost:8081"

REM 7) Wait for proxy readiness (max 30s)
set /a __tries=0
:wait_proxy
curl -s -f http://localhost:%__PORT%/health >nul 2>&1
if %errorlevel% neq 0 (
    set /a __tries+=1
    if !__tries! geq 30 (
        echo Proxy did not become ready. Aborting.
        goto :fail
    )
    timeout /t 1 >nul
    goto :wait_proxy
)
echo Proxy is up.

REM 8) Optionally open UI and metrics in browser
start "" http://localhost:%__PORT%/
start "" http://localhost:%__PORT%/metrics

echo.
echo === All set! ===
echo - Test backend:   http://localhost:8081/
echo - Proxy health:   http://localhost:%__PORT%/health
echo - UI dashboard:   http://localhost:%__PORT%/
echo - Metrics:        http://localhost:%__PORT%/metrics
echo.
echo To stop everything, run: stop.bat
echo.
pause
exit /b 0

:fail
echo.
echo Build/Run failed. See messages above.
pause
exit /b 1


