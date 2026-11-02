@echo off
cd /d "%~dp0"
echo Starting Horizon Core Nodes in background...

start "Node-1" /MIN cmd /c "set NODE_ID=node-1&& set PORT=8081&& .\bin\api.exe > logs\node-1.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-2" /MIN cmd /c "set NODE_ID=node-2&& set PORT=8082&& .\bin\api.exe > logs\node-2.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-3" /MIN cmd /c "set NODE_ID=node-3&& set PORT=8083&& .\bin\api.exe > logs\node-3.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-4" /MIN cmd /c "set NODE_ID=node-4&& set PORT=8084&& .\bin\api.exe > logs\node-4.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-5" /MIN cmd /c "set NODE_ID=node-5&& set PORT=8085&& .\bin\api.exe > logs\node-5.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-6" /MIN cmd /c "set NODE_ID=node-6&& set PORT=8086&& .\bin\api.exe > logs\node-6.log 2>&1"
timeout /t 1 /nobreak > NUL

start "Node-7" /MIN cmd /c "set NODE_ID=node-7&& set PORT=8087&& .\bin\api.exe > logs\node-7.log 2>&1"
timeout /t 1 /nobreak > NUL

echo Starting Horizon Switch (if exists)...
if exist .\horizon-bank-package\switch.exe (
    start "Horizon-Switch" /MIN cmd /c ".\horizon-bank-package\switch.exe > logs\switch.log 2>&1"
)

exit
