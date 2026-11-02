@echo off
set NODE_ID=node-2
set PORT=8082
cd /d "%~dp0"
.\bin\api.exe > logs\node-2.log 2>&1
