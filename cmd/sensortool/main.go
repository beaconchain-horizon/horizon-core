package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"
)

type sigASN struct{ R, S *big.Int }

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("  sensortool genkey --out=keys")
		fmt.Println("  sensortool sign --key=keys/private.pem --sensor=temp-001 --value=45.5")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "genkey":
		fs := flag.NewFlagSet("genkey", flag.ExitOnError)
		out := fs.String("out", "keys", "output directory")
		fs.Parse(os.Args[2:])
		os.MkdirAll(*out, 0700)
		priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		privBytes, _ := x509.MarshalECPrivateKey(priv)
		pubBytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
		os.WriteFile(*out+"/private.pem", pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}), 0600)
		os.WriteFile(*out+"/public.pem", pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}), 0644)
		fmt.Println("keys written to " + *out)

	case "sign":
		fs := flag.NewFlagSet("sign", flag.ExitOnError)
		keyPath := fs.String("key", "keys/private.pem", "private key path")
		sensorID := fs.String("sensor", "", "sensor id")
		value := fs.Float64("value", 0, "value")
		fs.Parse(os.Args[2:])
		if *sensorID == "" {
			fmt.Println("--sensor is required")
			os.Exit(1)
		}
		pemBytes, err := os.ReadFile(*keyPath)
		if err != nil {
			fmt.Println("read key:", err)
			os.Exit(1)
		}
		block, _ := pem.Decode(pemBytes)
		priv, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			fmt.Println("parse key:", err)
			os.Exit(1)
		}
		nonce := randHex(16)
		ts := time.Now().Unix()
		payload := map[string]any{
			"sensor_id": *sensorID,
			"value":     *value,
			"nonce":     nonce,
			"timestamp": ts,
		}
		canonical, _ := json.Marshal(payload)
		digest := sha256.Sum256(canonical)
		r, s, _ := ecdsa.Sign(rand.Reader, priv, digest[:])
		der, _ := asn1.Marshal(sigASN{r, s})
		payload["signature"] = hex.EncodeToString(der)
		out, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(out))
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
