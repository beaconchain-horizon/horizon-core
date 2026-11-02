@echo off
cd /d "%~dp0\node5"
set PORT=9005
set NODE_ID=node-5
api.exe > logs\node.log 2>&1
