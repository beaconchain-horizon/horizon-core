@echo off
cd /d "%~dp0\node3"
set PORT=9003
set NODE_ID=node-3
api.exe > logs\node.log 2>&1
