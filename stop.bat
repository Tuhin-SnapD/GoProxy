@echo off
echo Stopping GoProxy and test backend (if running)...
taskkill /IM goproxy.exe /F >nul 2>&1
taskkill /IM test_server.exe /F >nul 2>&1
echo Done.
pause


