@echo off
taskkill /F /IM backend.exe /T >nul 2>&1
taskkill /F /IM switch.exe /T >nul 2>&1
taskkill /F /IM angel-node.exe /T >nul 2>&1
taskkill /F /IM bash.exe /T >nul 2>&1
echo System stopped successfully.
timeout /t 3 >nul
