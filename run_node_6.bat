@echo off
set NODE_ID=node-6
set PORT=8086
cd /d "%~dp0"
.\bin\api.exe > logs\node-6.log 2>&1
