@echo off
set NODE_ID=node-4
set PORT=8084
cd /d "%~dp0"
.\bin\api.exe > logs\node-4.log 2>&1
