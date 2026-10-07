// Horizon BNPL - secp256k1 (Ethereum-compatible)
// Signs tx with owner's Ethereum private key
// Derives Ethereum address from public key

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"golang.org/x/crypto/sha3"
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
	clean := strings.TrimPrefix(hexKey, "0x")
	bytes, err := hex.DecodeString(clean)
	if err != nil {
		return errors.New("invalid hex: " + err.Error())
	}
	if len(bytes) != 32 {
		return errors.New("private key must be 32 bytes (64 hex chars)")
	}
	priv, _ := btcec.PrivKeyFromBytes(bytes)
	issuerKey = priv.ToECDSA()
	derivePublic()
	return nil
}

// loadOrCreateFileKey - dev fallback: file-based key
func loadOrCreateFileKey(keyPath string) error {
	os.MkdirAll(filepath.Dir(keyPath), 0700)

	if data, err := os.ReadFile(keyPath); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
				issuerKey = key
				derivePublic()
				return nil
			}
		}
	}

	priv, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return err
	}
	issuerKey = priv
	derivePublic()

	kb, _ := x509.MarshalECPrivateKey(priv)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb})
	return os.WriteFile(keyPath, pemBytes, 0400)
}

// derivePublic - extracts PEM pubkey + Ethereum address
func derivePublic() {
	pb, _ := x509.MarshalPKIXPublicKey(&issuerKey.PublicKey)
	issuerPubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pb}))
	issuerAddress = ethereumAddress(&issuerKey.PublicKey)
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
	h := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, issuerKey, h[:])
	if err != nil {
		return "", err
	}
	buf := make([]byte, 64)
	rb := r.Bytes()
	sb := s.Bytes()
	copy(buf[32-len(rb):32], rb)
	copy(buf[64-len(sb):64], sb)
	return hex.EncodeToString(buf), nil
}

// verifyData - verifies signature with given pubkey PEM
func verifyData(pubPEM string, data []byte, sigHex string) bool {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return false
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return false
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != 64 {
		return false
	}
	r := new(bigInt).SetBytes(sig[:32])
	s := new(bigInt).SetBytes(sig[32:])
	h := sha256.Sum256(data)
	return ecdsa.Verify(ecPub, h[:], r, s)
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
