@echo off
cd /d "%~dp0"

start "Node-1" /MIN cmd /c "set NODE_ID=node-1&& set PORT=8081&& .\bin\api.exe > logs\node-1.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-2" /MIN cmd /c "set NODE_ID=node-2&& set PORT=8082&& .\bin\api.exe > logs\node-2.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-3" /MIN cmd /c "set NODE_ID=node-3&& set PORT=8083&& .\bin\api.exe > logs\node-3.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-4" /MIN cmd /c "set NODE_ID=node-4&& set PORT=8084&& .\bin\api.exe > logs\node-4.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-5" /MIN cmd /c "set NODE_ID=node-5&& set PORT=8085&& .\bin\api.exe > logs\node-5.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-6" /MIN cmd /c "set NODE_ID=node-6&& set PORT=8086&& .\bin\api.exe > logs\node-6.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-7" /MIN cmd /c "set NODE_ID=node-7&& set PORT=8087&& .\bin\api.exe > logs\node-7.log 2>&1"

exit
