// Horizon BNPL - secp256k1 (Ethereum-compatible)
// Signs tx with owner's Ethereum private key
// Derives Ethereum address from public key

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/sha3"
	hcrypto "horizon-core/internal/crypto"
)

var (
	issuerKey     *ecdsa.PrivateKey
	issuerPubPEM  string
	issuerAddress string
)

// loadKeyFromEnvOrFile - reads HORIZON_ISSUER_KEY from env, fallback to file
func loadKeyFromEnvOrFile(keyPath string) error {
	hexKey := strings.TrimSpace(os.Getenv("HORIZON_ISSUER_KEY"))
	if hexKey != "" {
		return loadHexKey(hexKey)
	}
	return loadOrCreateFileKey(keyPath)
}

// loadHexKey - loads secp256k1 key from hex string (Ethereum format)
func loadHexKey(hexKey string) error {
	clean := strings.TrimSpace(hexKey)
	clean = strings.TrimPrefix(clean, "0x")
	clean = strings.TrimPrefix(clean, "0X")
	keyBytes, err := hex.DecodeString(clean)
	if err != nil {
		return errors.New("invalid hex: " + err.Error())
	}
	if len(keyBytes) != 32 {
		return errors.New("private key must be 32 bytes (64 hex chars)")
	}
	key, err := hcrypto.Secp256k1PrivateKeyFromBytes(keyBytes)
	if err != nil {
		return fmt.Errorf("invalid secp256k1 private key: %w", err)
	}
	issuerKey = key
	if err := derivePublic(); err != nil {
		issuerKey = nil
		issuerPubPEM = ""
		issuerAddress = ""
		return err
	}
	return nil
}

// loadOrCreateFileKey - dev fallback: file-based key
func loadOrCreateFileKey(keyPath string) error {
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return fmt.Errorf("create issuer key directory: %w", err)
	}

	data, err := os.ReadFile(keyPath)
	if err == nil {
		key, parseErr := hcrypto.PEMToPrivateKey(string(data))
		if parseErr != nil {
			return fmt.Errorf("existing issuer key %q cannot be loaded as a Horizon secp256k1 key; refusing implicit key rotation: %w", keyPath, parseErr)
		}
		issuerKey = key
		if err := derivePublic(); err != nil {
			issuerKey = nil
			issuerPubPEM = ""
			issuerAddress = ""
			return fmt.Errorf("derive issuer public identity: %w", err)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("read issuer key %q: %w", keyPath, err)
	}

	generated, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return fmt.Errorf("generate secp256k1 issuer key: %w", err)
	}
	key := generated.ToECDSA()
	privatePEM, err := hcrypto.PrivateKeyPEM(key)
	if err != nil {
		return fmt.Errorf("encode secp256k1 issuer key: %w", err)
	}
	publicPEM, err := hcrypto.PublicKeyPEM(key)
	if err != nil {
		return fmt.Errorf("encode secp256k1 issuer public key: %w", err)
	}
	address := ethereumAddress(&key.PublicKey)
	if address == "" {
		return errors.New("could not derive secp256k1 issuer address")
	}

	// O_EXCL prevents accidentally overwriting a key created by another process.
	file, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return fmt.Errorf("create issuer key file without overwriting an existing key: %w", err)
	}
	if _, err := file.Write([]byte(privatePEM)); err != nil {
		_ = file.Close()
		_ = os.Remove(keyPath)
		return fmt.Errorf("write issuer key file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(keyPath)
		return fmt.Errorf("close issuer key file: %w", err)
	}

	issuerKey = key
	issuerPubPEM = publicPEM
	issuerAddress = address
	return nil
}

// derivePublic - extracts PEM pubkey + Ethereum address
func derivePublic() error {
	if issuerKey == nil {
		return errors.New("issuer private key is not loaded")
	}
	publicPEM, err := hcrypto.PublicKeyPEM(issuerKey)
	if err != nil {
		return fmt.Errorf("encode secp256k1 issuer public key: %w", err)
	}
	address := ethereumAddress(&issuerKey.PublicKey)
	if address == "" {
		return errors.New("could not derive secp256k1 issuer address")
	}
	issuerPubPEM = publicPEM
	issuerAddress = address
	return nil
}

// ethereumAddress - derives 0x... from secp256k1 public key
func ethereumAddress(pub *ecdsa.PublicKey) string {
	uncompressed := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	if len(uncompressed) != 65 || uncompressed[0] != 0x04 {
		return ""
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(uncompressed[1:])
	hash := h.Sum(nil)
	return "0x" + hex.EncodeToString(hash[12:])
}

// signData - ECDSA secp256k1 sign, returns hex(r||s)
func signData(data []byte) (string, error) {
	if issuerKey == nil {
		return "", errors.New("signer key not loaded")
	}
	sig, err := hcrypto.SignDataRS(data, issuerKey)
	if err != nil {
		return "", fmt.Errorf("secp256k1 sign failed: %w", err)
	}
	return hex.EncodeToString(sig), nil
}

// verifyData - verifies signature with given pubkey PEM
func verifyData(pubPEM string, data []byte, sigHex string) bool {
	pub, err := hcrypto.PEMToPublicKey(pubPEM)
	if err != nil {
		return false
	}
	ok, err := hcrypto.VerifyData(pub, data, sigHex)
	return err == nil && ok
}

// txSignablePayload - canonical bytes signed for a tx
func txSignablePayload(tx *CreditTx) []byte {
	s := tx.TxID + "|" + tx.PlanID + "|" + tx.ProviderID + "|" +
		tx.PersonnelID + "|" + tx.MerchantID + "|" +
		itoa(tx.Amount) + "|" + itoa(int64(tx.Installments)) + "|" +
		tx.Nonce
	return []byte(s)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
