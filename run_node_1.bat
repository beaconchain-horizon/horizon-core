@echo off
set NODE_ID=node-1
set PORT=8081
cd /d "%~dp0"
.\bin\api.exe > logs\node-1.log 2>&1
