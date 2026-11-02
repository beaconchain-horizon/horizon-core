@echo off
cd /d "%~dp0\node2"
set PORT=9002
set NODE_ID=node-2
api.exe > logs\node.log 2>&1
