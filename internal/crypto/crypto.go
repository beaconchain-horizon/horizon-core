package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

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

	sig, err := SignDataRS(data, privKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}

func GenerateKeyPair() (string, error) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		return "", err
	}

	return PrivateKeyPEM(priv)
}

func PEMToPrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
	return ParsePrivateKeyPEM(pemStr)
}

func PEMToPublicKey(pemStr string) (*ecdsa.PublicKey, error) {
	return ParsePublicKeyPEM(pemStr)
}

func PublicKeyPEM(priv *ecdsa.PrivateKey) (string, error) {
	return secp256k1PublicKeyPEM(priv)
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
