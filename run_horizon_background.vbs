Set WshShell = CreateObject("WScript.Shell")
Set FSO = CreateObject("Scripting.FileSystemObject")
CurrentDir = FSO.GetAbsolutePathName(".")

' Set Environment Variables
WshShell.Environment("Process")("ADMIN_TOKEN") = "12345678901234567890123456789012"
WshShell.Environment("Process")("HORIZON_LICENSE_PUBLIC_KEY") = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
WshShell.Environment("Process")("SERVER_PORT") = "8080"
WshShell.Environment("Process")("SWITCH_PORT") = "8091"
WshShell.Environment("Process")("HORIZON_DB_PATH") = CurrentDir & "\horizon-dev.db"

' 1. Start API
WshShell.Run "cmd /c bin\api.exe > api.log 2>&1", 0, False
WScript.Sleep 3000

' 2. Start Switch
WshShell.Run "cmd /c bin\switch.exe > switch.log 2>&1", 0, False

' 3. Start Agent Node (Node-08)
WshShell.Run "cmd /c cd /d """ & CurrentDir & "\cmd\agent-node"" && ..\..\bin\agent-node.exe > agent-node.log 2>&1", 0, False

' 4. Start 7 Angel Nodes
Dim roles, ports, i
roles = Array("merkle", "hash", "ecdsa", "type", "volume", "time", "user")
ports = Array("9001", "9002", "9003", "9004", "9005", "9006", "9007")

For i = 0 To 6
    WshShell.Environment("Process")("NODE_ROLE") = roles(i)
    WshShell.Environment("Process")("NODE_PORT") = ports(i)
    WshShell.Environment("Process")("NODE_PUBKEY") = "04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
    
    WshShell.Run "cmd /c bin\angel-node.exe > node_" & ports(i) & ".log 2>&1", 0, False
    WScript.Sleep 500
Next
