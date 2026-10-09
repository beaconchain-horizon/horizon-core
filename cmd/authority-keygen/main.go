// authority-keygen creates a fresh Horizon v1.4 secp256k1 Authority identity.
// Refuses to overwrite existing outputs. Keep the private key offline or
// provision it to Backend using a secret mount; never put it in a browser.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	hcrypto "horizon-core/internal/crypto"
)

func writeNew(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}

func main() {
	privateOut := flag.String("private-out", "secrets/issuer-private.pem", "Authority private key path")
	publicOut := flag.String("public-out", "secrets/issuer-public.hex", "uncompressed secp256k1 public key hex path")
	flag.Parse()
	if _, err := os.Stat(*privateOut); err == nil {
		fmt.Fprintln(os.Stderr, "refusing to overwrite existing private key:", *privateOut)
		os.Exit(2)
	}
	if _, err := os.Stat(*publicOut); err == nil {
		fmt.Fprintln(os.Stderr, "refusing to overwrite existing public key:", *publicOut)
		os.Exit(2)
	}
	priv, err := hcrypto.GenerateSecp256k1ECDSAKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate secp256k1 key:", err)
		os.Exit(1)
	}
	pemData, err := hcrypto.PrivateKeyPEM(priv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encode private key:", err)
		os.Exit(1)
	}
	pub := make([]byte, 65)
	pub[0] = 0x04
	priv.X.FillBytes(pub[1:33])
	priv.Y.FillBytes(pub[33:65])
	if err := writeNew(*privateOut, []byte(pemData), 0600); err != nil {
		fmt.Fprintln(os.Stderr, "write private key:", err)
		os.Exit(1)
	}
	if err := writeNew(*publicOut, []byte(hex.EncodeToString(pub)+"\n"), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "private key was created but public key write failed:", err)
		os.Exit(1)
	}
	fmt.Println("Horizon Authority keypair created.")
	fmt.Println("Private key:", *privateOut, "(secret; not for Switch or browser)")
	fmt.Println("Public key:", *publicOut, "(provision to Switch)")
}
