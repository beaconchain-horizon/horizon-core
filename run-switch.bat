@echo off
cd /d %~dp0
set /p ADMIN_PASSWORD_HASH=<hash.txt
set ADMIN_TOKEN=REPLACE_WITH_ADMIN_TOKEN
set SWITCH_PORT=8080
set SWITCH_DB=%CD%\data\horizon-switch.db
set CHAIN_CONFIG=%CD%\config\chain.json
set CUSTOMERS_CONFIG=%CD%\config\customers.json
set HORIZON_LICENSE_PUBLIC_KEY=048bea3fcc2a793347da31386d15ef67b44b8e8d47143b5c7ed2c7132f241bba969a6973a25dc510f8f091fa9f18b7e03fb1d7ffb10b215e977075ee6c43e5d34a
echo.
echo ========================
echo  HORIZON SWITCH v3.0
echo ========================
echo  Port: %SWITCH_PORT%
echo  Hash: %ADMIN_PASSWORD_HASH%
echo  Token: %ADMIN_TOKEN%
echo.
switch.exe
pause
