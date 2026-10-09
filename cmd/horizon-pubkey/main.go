package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"horizon-core/internal/crypto"
)

func main() {
	keyPEM, err := os.ReadFile(".horizon-dev-key.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "read key:", err)
		os.Exit(1)
	}
	priv, err := crypto.PEMToPrivateKey(string(keyPEM))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse key:", err)
		os.Exit(1)
	}
	var xb, yb [32]byte
	priv.PublicKey.X.FillBytes(xb[:])
	priv.PublicKey.Y.FillBytes(yb[:])
	fmt.Println(hex.EncodeToString(xb[:]) + hex.EncodeToString(yb[:]))
}
