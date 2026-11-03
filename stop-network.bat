@echo off
echo Stopping Horizon Network...
taskkill /F /IM horizon-api.exe 2>nul
taskkill /F /IM horizon-switch.exe 2>nul
taskkill /F /IM horizon-angel.exe 2>nul
taskkill /F /IM horizon-agent.exe 2>nul
echo All services stopped.
pause
