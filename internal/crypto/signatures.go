package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// ============================================================================
// METHOD 1: Standard Transaction Signature (Online)
// Signs raw data using a private key, returns 64-byte R||S hex.
// ============================================================================
func SignTransaction(priv *secp256k1.PrivateKey, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	sig := secp256k1ecdsa.Sign(priv, hash[:])
	
	// Extract R and S and serialize to 64 bytes (32 bytes R || 32 bytes S)
	// This matches the offline dashboard's HEX-32 output
	sigBytes := serializeRawRS(sig)
	return hex.EncodeToString(sigBytes), nil
}

// ============================================================================
// METHOD 2: Offline License Verification (Air-Gapped)
// Verifies the exact JSON payload signed by the offline dashboard.
// Handles the "0x04" uncompressed public key format.
// ============================================================================
func VerifyOfflineLicense(publicKeyHex string, jsonPayload []byte, signatureHex string) (bool, error) {
	pub, err := Secp256k1PublicKeyFromHex(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key: %w", err)
	}
	return VerifyData(pub, jsonPayload, signatureHex)
}

// ============================================================================
// METHOD 3: Authority Block Signature (Consensus Level)
// Signs the Merkle Root of a block. Used by the 7 nodes for consensus.
// Supports both Binary and Ternary Merkle Roots.
// ============================================================================
func SignBlockAuthority(priv *secp256k1.PrivateKey, merkleRoot string, blockHash string) (string, error) {
	payload := []byte(fmt.Sprintf("%s|%s", merkleRoot, blockHash))
	hash := sha256.Sum256(payload)
	sig := secp256k1ecdsa.Sign(priv, hash[:])
	
	sigBytes := serializeRawRS(sig)
	return hex.EncodeToString(sigBytes), nil
}

// ============================================================================
// UTILITY: Serialize a secp256k1 signature to raw 64-byte R||S format
// ============================================================================
func serializeRawRS(sig *secp256k1ecdsa.Signature) []byte {
	r := sig.R()
	s := sig.S()
	
	var rBytes [32]byte
	var sBytes [32]byte
	r.PutBytesUnchecked(rBytes[:])
	s.PutBytesUnchecked(sBytes[:])
	
	// Concatenate R || S (64 bytes total)
	out := make([]byte, 64)
	copy(out[:32], rBytes[:])
	copy(out[32:], sBytes[:])
	return out
}

// ============================================================================
// UTILITY: Generate a new keypair (for testing or node setup)
// ============================================================================
func GenerateKeyPair() (string, string, error) {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	privHex := hex.EncodeToString(priv.Serialize())
	
	// Uncompressed public key with 0x04 prefix
	pubBytes := priv.PubKey().SerializeUncompressed()
	pubHex := "0x" + hex.EncodeToString(pubBytes)
	
	return privHex, pubHex, nil
}

// Helper to parse a hex private key
func ParsePrivateKey(hexKey string) (*secp256k1.PrivateKey, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	priv := secp256k1.PrivKeyFromBytes(b)
	return priv, nil
}
