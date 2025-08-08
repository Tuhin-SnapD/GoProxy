@echo off
echo === GoProxy Test Script ===
echo.

REM Test health endpoint
echo Testing health endpoint...
curl -s -f http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Health endpoint
) else (
    echo ✗ Health endpoint
)

REM Test metrics endpoint
echo Testing metrics endpoint...
curl -s -f http://localhost:8080/metrics >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Metrics endpoint
) else (
    echo ✗ Metrics endpoint
)

REM Test proxy to backend (now under /proxy/)
echo Testing proxy to backend...
curl -s -f http://localhost:8080/proxy/ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Proxy to backend
) else (
    echo ✗ Proxy to backend
)

REM Test cache functionality
echo Testing cache functionality...
echo First request (should miss cache):
curl -s http://localhost:8080/proxy/
echo.
echo Second request (should hit cache):
curl -s http://localhost:8080/proxy/
echo.

REM Test rate limiting
echo Testing rate limiting...
echo Making 10 requests to test rate limiting...
for /l %%i in (1,1,10) do (
    for /f "tokens=*" %%a in ('curl -s -w "%%{http_code}" http://localhost:8080/proxy/ -o nul 2^>^&1') do (
        if "%%a"=="429" (
            echo Rate limit hit after %%i requests
            goto :show_metrics
        )
    )
)

:show_metrics
REM Show metrics
echo.
echo === Current Metrics ===
curl -s http://localhost:8080/metrics | findstr /R "total_requests cache_hits cache_misses blocked_requests"

echo.
echo === Test Complete ===
pause 