@echo off
setlocal
cd /d "%~dp0"

set PUBKEY=04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214

set NODE_PORT=9001
set NODE_ROLE=merkle
start "Node-9001" /min cmd /c "bin\angel-node.exe > node_9001.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9002
set NODE_ROLE=hash
start "Node-9002" /min cmd /c "bin\angel-node.exe > node_9002.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9003
set NODE_ROLE=ecdsa
start "Node-9003" /min cmd /c "bin\angel-node.exe > node_9003.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9004
set NODE_ROLE=type
start "Node-9004" /min cmd /c "bin\angel-node.exe > node_9004.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9005
set NODE_ROLE=volume
start "Node-9005" /min cmd /c "bin\angel-node.exe > node_9005.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9006
set NODE_ROLE=time
start "Node-9006" /min cmd /c "bin\angel-node.exe > node_9006.log 2>&1"
ping 127.0.0.1 -n 2 > nul

set NODE_PORT=9007
set NODE_ROLE=user
start "Node-9007" /min cmd /c "bin\angel-node.exe > node_9007.log 2>&1"
