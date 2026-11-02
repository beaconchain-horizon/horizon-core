@echo off
set NODE_ID=node-3
set PORT=8083
cd /d "%~dp0"
.\bin\api.exe > logs\node-3.log 2>&1
