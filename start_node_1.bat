@echo off
cd /d "%~dp0\node1"
set PORT=9001
set NODE_ID=node-1
api.exe > logs\node.log 2>&1
