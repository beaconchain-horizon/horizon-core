package db

import (
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	"horizon-core/internal/crypto"
)

var DB *sql.DB
var PrivateKey *ecdsa.PrivateKey

func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite", "./horizon.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS licenses (
			id TEXT PRIMARY KEY,
			volume INTEGER,
			root TEXT,
			seed TEXT,
			created_at DATETIME,
			signature TEXT
		);
		CREATE TABLE IF NOT EXISTS keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			private_key TEXT
		);
	`)
	if err != nil {
		return err
	}

	var privPEM string
	err = DB.QueryRow("SELECT private_key FROM keys LIMIT 1").Scan(&privPEM)
	if err != nil {
		privPEM, err = generateAndSaveKey()
		if err != nil {
			return err
		}
	}
	PrivateKey, err = loadPrivateKey(privPEM)
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func generateAndSaveKey() (string, error) {
	privPEM, err := crypto.GenerateKeyPair()
	if err != nil {
		return "", err
	}
	err = SavePrivateKey(privPEM)
	if err != nil {
		return "", err
	}
	return privPEM, nil
}

func loadPrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func SavePrivateKey(pemData string) error {
	_, err := DB.Exec("INSERT INTO keys (private_key) VALUES (?)", pemData)
	return err
}
