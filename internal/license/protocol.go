// Package license defines the v1.4 signed-license wire protocol shared by
// Horizon Authority, the API, Switch, and offline verifiers.
package license

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	hcrypto "horizon-core/internal/crypto"
	"horizon-core/internal/merkle"
)

// Statement contains immutable fields covered by the Authority signature.
// Mutable status and usage counters are deliberately stored outside this
// statement and therefore cannot silently change its cryptographic meaning.
type Statement struct {
	LicenseID  string `json:"license_id"`
	UserID     string `json:"user_id"`
	ProductID  string `json:"product_id"`
	Volume     int    `json:"volume"`
	Duration   int    `json:"duration"`
	MerkleRoot string `json:"merkle_root"`
	IssuedAt   int64  `json:"issued_at"`
	ExpiresAt  int64  `json:"expires_at"`
	HardwareID string `json:"hardware_id"`
}

// SignedLicense is the one wire envelope accepted by Switch and the offline
// verifier. The trusted issuer public key is provisioned out of band; it is
// intentionally not accepted from this envelope.
type SignedLicense struct {
	Statement Statement `json:"statement"`
	Signature string    `json:"signature"`
}

// CanonicalBytes returns the unique JSON encoding used by all signers and
// verifiers. Keep Statement field order stable; changes require a protocol
// version change, not an implicit compatibility fallback.
func CanonicalBytes(s Statement) ([]byte, error) {
	return json.Marshal(s)
}

// MerkleRoot computes an RFC 6962 root over the immutable license attributes.
// The root is then included in the signed canonical Statement.
func MerkleRoot(s Statement) string {
	leaves := [][]byte{
		[]byte("license_id=" + s.LicenseID),
		[]byte("user_id=" + s.UserID),
		[]byte("product_id=" + s.ProductID),
		[]byte(fmt.Sprintf("volume=%d", s.Volume)),
		[]byte(fmt.Sprintf("duration=%d", s.Duration)),
		[]byte(fmt.Sprintf("issued_at=%d", s.IssuedAt)),
		[]byte(fmt.Sprintf("expires_at=%d", s.ExpiresAt)),
		[]byte("hardware_id=" + s.HardwareID),
	}
	return merkle.BinaryRootNoPrefix(leaves)
}

// Validate verifies protocol fields, expiry, hardware binding, recomputed
// Merkle root, and the secp256k1 signature against a separately trusted key.
// expectedHardwareID may be empty only for an independent cryptographic audit;
// Switch must pass the local hardware ID and therefore enforce binding.
func Validate(signed SignedLicense, publicKey *ecdsa.PublicKey, now time.Time, expectedHardwareID string) error {
	s := signed.Statement
	if strings.TrimSpace(s.LicenseID) == "" || strings.TrimSpace(s.UserID) == "" || strings.TrimSpace(s.ProductID) == "" {
		return errors.New("license_id, user_id, and product_id are required")
	}
	if s.Volume <= 0 || s.Duration <= 0 {
		return errors.New("volume and duration must be positive")
	}
	if strings.TrimSpace(s.HardwareID) == "" {
		return errors.New("hardware_id is required")
	}
	if s.IssuedAt <= 0 || s.ExpiresAt <= s.IssuedAt {
		return errors.New("invalid issued_at/expires_at interval")
	}
	if s.ExpiresAt-s.IssuedAt != int64(s.Duration)*86400 {
		return errors.New("duration does not match issued_at/expires_at")
	}
	if now.Unix() < s.IssuedAt {
		return errors.New("license is not active yet")
	}
	if now.Unix() >= s.ExpiresAt {
		return errors.New("license expired")
	}
	if expectedHardwareID != "" && !strings.EqualFold(strings.TrimSpace(expectedHardwareID), strings.TrimSpace(s.HardwareID)) {
		return errors.New("hardware_id mismatch")
	}
	root, err := hex.DecodeString(s.MerkleRoot)
	if err != nil || len(root) != 32 {
		return errors.New("merkle_root must be a 32-byte hex digest")
	}
	if !strings.EqualFold(s.MerkleRoot, MerkleRoot(s)) {
		return errors.New("merkle_root does not match immutable license fields")
	}
	if strings.TrimSpace(signed.Signature) == "" {
		return errors.New("signature is required")
	}
	message, err := CanonicalBytes(s)
	if err != nil {
		return fmt.Errorf("canonicalize license: %w", err)
	}
	ok, err := hcrypto.VerifyData(publicKey, message, signed.Signature)
	if err != nil {
		return fmt.Errorf("verify secp256k1 signature: %w", err)
	}
	if !ok {
		return errors.New("license signature is invalid")
	}
	return nil
}
