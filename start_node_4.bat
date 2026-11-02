@echo off
cd /d "%~dp0\node4"
set PORT=9004
set NODE_ID=node-4
api.exe > logs\node.log 2>&1
