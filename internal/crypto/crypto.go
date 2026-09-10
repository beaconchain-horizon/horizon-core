package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
)

type ecdsaSignature struct {
	R *big.Int
	S *big.Int
}

func Sha256Hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func SHA256HashString(text string) string {
	h := sha256.Sum256([]byte(text))
	return hex.EncodeToString(h[:])
}

func SHA512HashString(text string) string {
	h := sha512.Sum512([]byte(text))
	return hex.EncodeToString(h[:])
}

func SignData(data []byte, privKey *ecdsa.PrivateKey) (string, error) {
	if privKey == nil {
		return "", errors.New("private key is nil")
	}
	digest := Sha256Hash(data)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, digest)
	if err != nil {
		return "", err
	}
	der, err := asn1.Marshal(ecdsaSignature{r, s})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(der), nil
}

func VerifySignature(pubKey *ecdsa.PublicKey, data []byte, signatureHex string) bool {
	if pubKey == nil {
		return false
	}
	der, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}
	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(der, &sig); err != nil || sig.R == nil || sig.S == nil {
		return false
	}
	digest := Sha256Hash(data)
	return ecdsa.Verify(pubKey, digest, sig.R, sig.S)
}

func GenerateKeyPair() (string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})), nil
}

func PEMToPrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	return x509.ParseECPrivateKey(block.Bytes)
}

func PEMToPublicKey(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not ECDSA")
	}
	return key, nil
}

func PublicKeyPEM(priv *ecdsa.PrivateKey) (string, error) {
	if priv == nil {
		return "", errors.New("private key is nil")
	}
	b, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b})), nil
}

func EncryptAES(key []byte, plaintext string) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("key length must be 16, 24, or 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

func DecryptAES(key []byte, encoded string) (string, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("key length must be 16, 24, or 32 bytes")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	return stringData(gcm.Open(nil, data[:ns], data[ns:], nil))
}

func stringData(data []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(data), nil
}
