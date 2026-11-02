@echo off
set NODE_ID=node-7
set PORT=8087
cd /d "%~dp0"
.\bin\api.exe > logs\node-7.log 2>&1
