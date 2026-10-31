@echo off
setlocal
cd /d "%~dp0"

set ADMIN_TOKEN=12345678901234567890123456789012
set HORIZON_LICENSE_PUBLIC_KEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set SERVER_PORT=8080
set SWITCH_PORT=8091
set HORIZON_DB_PATH=%~dp0horizon-dev.db

echo Starting Backend...
start "Horizon-API" /min cmd /c "bin\api.exe > api.log 2>&1"
timeout /t 3 /nobreak >nul

echo Starting Switch...
start "Horizon-Switch" /min cmd /c "bin\switch.exe > switch.log 2>&1"

echo Starting Agent Node (Node-08)...
cd cmd\agent-node
start "Agent-Node" /min cmd /c "..\..\bin\agent-node.exe > agent-node.log 2>&1"
cd ..\..

echo Starting 7 Angel Nodes...
set NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

set NODE_PORT=9001 & set NODE_ROLE=merkle & start "Angel-9001" /min cmd /c "bin\angel-node.exe > node_9001.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9002 & set NODE_ROLE=hash & start "Angel-9002" /min cmd /c "bin\angel-node.exe > node_9002.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9003 & set NODE_ROLE=ecdsa & start "Angel-9003" /min cmd /c "bin\angel-node.exe > node_9003.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9004 & set NODE_ROLE=type & start "Angel-9004" /min cmd /c "bin\angel-node.exe > node_9004.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9005 & set NODE_ROLE=volume & start "Angel-9005" /min cmd /c "bin\angel-node.exe > node_9005.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9006 & set NODE_ROLE=time & start "Angel-9006" /min cmd /c "bin\angel-node.exe > node_9006.log 2>&1"
timeout /t 1 /nobreak >nul
set NODE_PORT=9007 & set NODE_ROLE=user & start "Angel-9007" /min cmd /c "bin\angel-node.exe > node_9007.log 2>&1"

echo All services dispatched.
