@echo off
cd /d "%~dp0"

:: --- Start 7 Nodes on Ports 9001-9007 ---
for %%N in (1 2 3 4 5 6 7) do (
    set /A PORT=9000+%%N
    start "Node-%%N" /MIN cmd /c "cd node%%N && set PORT=%%PORT%% && set NODE_ID=node-%%N && api.exe > logs\node.log 2>&1"
    ping 127.0.0.1 -n 2 > nul
)

:: --- Start Bank Switch on Port 9000 ---
if exist horizon-bank-package\switch.exe (
    start "Horizon-Switch" /MIN cmd /c "cd horizon-bank-package && set PORT=9000 && switch.exe > data\switch.log 2>&1"
) else (
    echo Switch not found.
)

exit
