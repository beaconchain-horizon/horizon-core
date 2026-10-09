package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"horizon-core/internal/crypto"
	"os"
	"path/filepath"
	"time"
)

// MUST match industrial.SignableReading field order exactly
type SignableReading struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Nonce     string  `json:"nonce"`
	Timestamp int64   `json:"timestamp"`
}

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

		if err := os.MkdirAll(*out, 0700); err != nil {
			fmt.Println("mkdir:", err)
			os.Exit(1)
		}

		privatePEM, err := crypto.GenerateKeyPair()
		if err != nil {
			fmt.Println("generate key:", err)
			os.Exit(1)
		}

		priv, err := crypto.PEMToPrivateKey(privatePEM)
		if err != nil {
			fmt.Println("parse generated key:", err)
			os.Exit(1)
		}

		publicPEM, err := crypto.PublicKeyPEM(priv)
		if err != nil {
			fmt.Println("public key:", err)
			os.Exit(1)
		}

		if err := os.WriteFile(
			filepath.Join(*out, "private.pem"),
			[]byte(privatePEM),
			0600,
		); err != nil {
			fmt.Println("write private key:", err)
			os.Exit(1)
		}

		if err := os.WriteFile(
			filepath.Join(*out, "public.pem"),
			[]byte(publicPEM),
			0644,
		); err != nil {
			fmt.Println("write public key:", err)
			os.Exit(1)
		}

		fmt.Println("secp256k1 keys written to " + *out)

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

		priv, err := crypto.PEMToPrivateKey(string(pemBytes))
		if err != nil {
			fmt.Println("parse key:", err)
			os.Exit(1)
		}

		sr := SignableReading{
			SensorID:  *sensorID,
			Value:     *value,
			Nonce:     randHex(16),
			Timestamp: time.Now().Unix(),
		}

		canonical, _ := json.Marshal(sr)

		signature, err := crypto.SignData(canonical, priv)
		if err != nil {
			fmt.Println("sign:", err)
			os.Exit(1)
		}

		out := map[string]any{
			"sensor_id": sr.SensorID,
			"value":     sr.Value,
			"nonce":     sr.Nonce,
			"timestamp": sr.Timestamp,
			"signature": signature,
		}

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Println("marshal:", err)
			os.Exit(1)
		}

		fmt.Println(string(b))
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
