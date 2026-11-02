@echo off
cd /d "%~dp0\node7"
set PORT=9007
set NODE_ID=node-7
api.exe > logs\node.log 2>&1
