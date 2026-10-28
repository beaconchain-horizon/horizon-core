Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c ""C:\Program Files\Git\bin\bash.exe"" -c ""./horizon-guardian.sh""", 0, False
