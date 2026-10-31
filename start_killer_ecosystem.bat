@echo off
cd /d "%~dp0"
set ADMIN_TOKEN=12345678901234567890123456789012
set HORIZON_LICENSE_PUBLIC_KEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214
set SERVER_PORT=8080
set SWITCH_PORT=8091
set HORIZON_DB_PATH=%~dp0horizon-dev.db
set NODE_PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

start "" /min cmd /c "bin\api.exe > api.log 2>&1"
ping 127.0.0.1 -n 3 > nul
start "" /min cmd /c "bin\switch.exe > switch.log 2>&1"
cd cmd\agent-node
start "" /min cmd /c "..\..\bin\agent-node.exe > agent-node.log 2>&1"
cd ..\..

set NODE_PORT=9001 & set NODE_ROLE=merkle & start "" /min cmd /c "bin\angel-node.exe > node_9001.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9002 & set NODE_ROLE=hash & start "" /min cmd /c "bin\angel-node.exe > node_9002.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9003 & set NODE_ROLE=ecdsa & start "" /min cmd /c "bin\angel-node.exe > node_9003.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9004 & set NODE_ROLE=type & start "" /min cmd /c "bin\angel-node.exe > node_9004.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9005 & set NODE_ROLE=volume & start "" /min cmd /c "bin\angel-node.exe > node_9005.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9006 & set NODE_ROLE=time & start "" /min cmd /c "bin\angel-node.exe > node_9006.log 2>&1"
ping 127.0.0.1 -n 2 > nul
set NODE_PORT=9007 & set NODE_ROLE=user & start "" /min cmd /c "bin\angel-node.exe > node_9007.log 2>&1"
