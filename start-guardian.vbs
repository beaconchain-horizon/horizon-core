Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c cd /d C:\Users\NIL
STORE\Beaconchain\horizon-core && ""C:\Program Files\Git\usr\bin\bash.exe"" -c ""./horizon-guardian.sh""", 0, False
