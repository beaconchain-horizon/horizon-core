@echo off
cd /d "%~dp0\node6"
set PORT=9006
set NODE_ID=node-6
api.exe > logs\node.log 2>&1
