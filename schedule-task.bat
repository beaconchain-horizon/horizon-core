@echo off
schtasks /Create /TN "HorizonAngel-Boot" /TR "C:\Users\NIL STORE\Beaconchain\horizon-core\start-all-silent.bat" /SC ONLOGON /RL HIGHEST /F
schtasks /Query /TN "HorizonAngel-Boot"
pause