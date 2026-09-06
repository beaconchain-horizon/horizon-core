package main
import (
"bytes";"context";"crypto/aes";"crypto/cipher";"crypto/rand";"crypto/sha256";"encoding/hex";"encoding/json";"fmt";"io";"log";"net/http";"os";"os/signal";"syscall";"time"
"github.com/gorilla/mux"
)
func SHA256Hash(d string) string { h:=sha256.Sum256([]byte(d)); return hex.EncodeToString(h[:]) }
func AESEncrypt(k,p []byte)(string,error){b,_:=aes.NewCipher(k);g,_:=cipher.NewGCM(b);n:=make([]byte,g.NonceSize());io.ReadFull(rand.Reader,n);c:=g.Seal(n,n,p,nil);return hex.EncodeToString(c),nil}
func AESDecrypt(k []byte,ct string)([]byte,error){c,_:=hex.DecodeString(ct);b,_:=aes.NewCipher(k);g,_:=cipher.NewGCM(b);n:=c[:g.NonceSize()];c=c[g.NonceSize():];return g.Open(nil,n,c,nil)}
func main(){
r:=mux.NewRouter()
r.HandleFunc("/health",func(w http.ResponseWriter,r *http.Request){json.NewEncoder(w).Encode(map[string]string{"status":"online"})}).Methods("GET")
r.HandleFunc("/api/v1/license/check",func(w http.ResponseWriter,r *http.Request){json.NewEncoder(w).Encode(map[string]bool{"valid":true})}).Methods("POST")
r.HandleFunc("/api/v1/toolbox/encrypt",func(w http.ResponseWriter,r *http.Request){var req struct{Key,Plain string;json:"key,plaintext"};json.NewDecoder(r.Body).Decode(&req);k,_:=hex.DecodeString(req.Key);e,_:=AESEncrypt(k,[]byte(req.Plain));json.NewEncoder(w).Encode(map[string]string{"ciphertext":e})}).Methods("POST")
r.HandleFunc("/api/v1/toolbox/decrypt",func(w http.ResponseWriter,r *http.Request){var req struct{Key,Cipher string;json:"key,ciphertext"};json.NewDecoder(r.Body).Decode(&req);k,_:=hex.DecodeString(req.Key);p,_:=AESDecrypt(k,req.Cipher);json.NewEncoder(w).Encode(map[string]string{"plaintext":string(p)})}).Methods("POST")
port:=os.Getenv("SWITCH_PORT");if port==""{port="8080"}
log.Printf("Switch running on :%s",port)
log.Fatal(http.ListenAndServe(":"+port,r))
}
