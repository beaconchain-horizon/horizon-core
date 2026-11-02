@echo off
set NODE_ID=node-5
set PORT=8085
cd /d "%~dp0"
.\bin\api.exe > logs\node-5.log 2>&1
