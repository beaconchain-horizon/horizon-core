@echo off
cd /d "%~dp0"

start "Node-1" /MIN cmd /c "cd node1 && set PORT=9001 && set NODE_ID=node-1 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-2" /MIN cmd /c "cd node2 && set PORT=9002 && set NODE_ID=node-2 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-3" /MIN cmd /c "cd node3 && set PORT=9003 && set NODE_ID=node-3 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-4" /MIN cmd /c "cd node4 && set PORT=9004 && set NODE_ID=node-4 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-5" /MIN cmd /c "cd node5 && set PORT=9005 && set NODE_ID=node-5 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-6" /MIN cmd /c "cd node6 && set PORT=9006 && set NODE_ID=node-6 && api.exe > logs\node.log 2>&1"
ping 127.0.0.1 -n 2 > nul

start "Node-7" /MIN cmd /c "cd node7 && set PORT=9007 && set NODE_ID=node-7 && api.exe > logs\node.log 2>&1"

if exist horizon-bank-package\switch.exe (
    start "Horizon-Switch" /MIN cmd /c "cd horizon-bank-package && set PORT=9000 && switch.exe > data\switch.log 2>&1"
)
exit
