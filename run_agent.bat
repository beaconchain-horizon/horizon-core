@echo off
cd /d "%~dp0cmd\agent-node"
..\..\bin\agent-node.exe > agent-node.log 2>&1
