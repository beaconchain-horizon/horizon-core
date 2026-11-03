package main
import("bufio";"encoding/hex";"fmt";"os";"strings";s "github.com/decred/dcrd/dcrec/secp256k1/v4";e "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa")
func main(){
sc:=bufio.NewScanner(os.Stdin);sc.Scan()
h:=strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(sc.Text(),"0x"),"0X"))
b,err:=hex.DecodeString(h)
if err!=nil||len(b)!=32{fmt.Fprintln(os.Stderr,"bad key");os.Exit(1)}
p:=s.PrivKeyFromBytes(b);pb:=p.PubKey()
exp:="04cf6aaaba8e7a6fddde63a2e2588553dc95503d2de6b2a5c78f4a30a0b91b0aca4d28d06fc1ccbaf730abe1dcda7c29726dab18008353b521c9253c04b1641214"
if hex.EncodeToString(pb.SerializeUncompressed())!=exp{fmt.Fprintln(os.Stderr,"MISMATCH");os.Exit(2)}
gh,_:=hex.DecodeString("8c0763dd5b1345cba0782f03163e8a69de49bede54c91bb878126742c7990603")
sg:=e.Sign(p,gh)
fmt.Fprintln(os.Stderr,"OK")
fmt.Println(hex.EncodeToString(sg.Serialize()))
}
