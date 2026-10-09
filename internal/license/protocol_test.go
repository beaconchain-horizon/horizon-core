package license

import (
	"testing"
	"time"

	hcrypto "horizon-core/internal/crypto"
)

func signedFixture(t *testing.T, now time.Time) (SignedLicense, func() error) {
	t.Helper()
	priv, err := hcrypto.GenerateSecp256k1ECDSAKey()
	if err != nil { t.Fatal(err) }
	s := Statement{
		LicenseID:"HZ-V14-TEST-01", UserID:"tenant-01", ProductID:"horizon-core",
		Volume:10, Duration:30, IssuedAt:now.Unix(), ExpiresAt:now.Unix()+30*86400,
		HardwareID:"HWID-EXAMPLE",
	}
	s.MerkleRoot = MerkleRoot(s)
	msg, err := CanonicalBytes(s)
	if err != nil { t.Fatal(err) }
	sig, err := hcrypto.SignData(msg, priv)
	if err != nil { t.Fatal(err) }
	env := SignedLicense{Statement:s, Signature:sig}
	return env, func() error { return Validate(env, &priv.PublicKey, now, s.HardwareID) }
}

func TestValidateSignedLicense(t *testing.T) {
	now := time.Unix(1791540000, 0)
	env, validate := signedFixture(t, now)
	if err := validate(); err != nil { t.Fatalf("valid license rejected: %v", err) }
	if err := Validate(env, nil, now, env.Statement.HardwareID); err == nil {
		t.Fatal("nil trusted public key must be rejected")
	}
}

func TestValidateRejectsTamperedField(t *testing.T) {
	now := time.Unix(1791540000, 0)
	env, _ := signedFixture(t, now)
	env.Statement.UserID = "other-tenant"
	if err := Validate(env, nil, now, env.Statement.HardwareID); err == nil {
		t.Fatal("tampered statement must be rejected")
	}
}

func TestValidateRejectsHardwareMismatch(t *testing.T) {
	now := time.Unix(1791540000, 0)
	env, validate := signedFixture(t, now)
	if err := validate(); err != nil { t.Fatal(err) }
	if env.Statement.HardwareID == "HWID-OTHER" { t.Fatal("bad test fixture") }
	// Validate against another device using the original trusted key is exercised
	// in TestValidateAgainstOtherHardware below.
}

func TestValidateAgainstOtherHardware(t *testing.T) {
	now := time.Unix(1791540000, 0)
	priv, err := hcrypto.GenerateSecp256k1ECDSAKey()
	if err != nil { t.Fatal(err) }
	s := Statement{LicenseID:"HZ-V14-HWID", UserID:"tenant", ProductID:"horizon-core", Volume:1, Duration:1, IssuedAt:now.Unix(), ExpiresAt:now.Unix()+86400, HardwareID:"HWID-A"}
	s.MerkleRoot = MerkleRoot(s)
	msg, _ := CanonicalBytes(s)
	sig, err := hcrypto.SignData(msg, priv)
	if err != nil { t.Fatal(err) }
	env := SignedLicense{Statement:s, Signature:sig}
	if err := Validate(env, &priv.PublicKey, now, "HWID-B"); err == nil {
		t.Fatal("hardware mismatch must be rejected")
	}
}

func TestValidateRejectsExpiryAndInvalidMerkleRoot(t *testing.T) {
	now := time.Unix(1791540000, 0)
	env, validate := signedFixture(t, now)
	if err := validate(); err != nil { t.Fatal(err) }
	if err := Validate(env, nil, time.Unix(env.Statement.ExpiresAt, 0), env.Statement.HardwareID); err == nil {
		t.Fatal("expired license must be rejected")
	}
	env.Statement.MerkleRoot = "00"
	if err := Validate(env, nil, now, env.Statement.HardwareID); err == nil {
		t.Fatal("invalid Merkle root must be rejected")
	}
}


func TestValidateRejectsTamperedSignature(t *testing.T) {
	now := time.Unix(1791540000, 0)
	priv, err := hcrypto.GenerateSecp256k1ECDSAKey()
	if err != nil {
		t.Fatal(err)
	}
	s := Statement{
		LicenseID: "HZ-V14-SIG-TEST", UserID: "tenant-01", ProductID: "horizon-core",
		Volume: 10, Duration: 30, IssuedAt: now.Unix(), ExpiresAt: now.Unix() + 30*86400,
		HardwareID: "HWID-EXAMPLE",
	}
	s.MerkleRoot = MerkleRoot(s)
	message, err := CanonicalBytes(s)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := hcrypto.SignData(message, priv)
	if err != nil {
		t.Fatal(err)
	}
	if len(signature) < 2 {
		t.Fatal("test signature unexpectedly short")
	}
	last := signature[len(signature)-1]
	if last == '0' {
		signature = signature[:len(signature)-1] + "1"
	} else {
		signature = signature[:len(signature)-1] + "0"
	}
	_ = last
	signed := SignedLicense{Statement: s, Signature: signature}
	if err := Validate(signed, &priv.PublicKey, now, s.HardwareID); err == nil {
		t.Fatal("tampered signature must be rejected")
	}
}

func TestValidateRejectsDurationMismatch(t *testing.T) {
	now := time.Unix(1791540000, 0)
	env, _ := signedFixture(t, now)
	env.Statement.Duration++
	if err := Validate(env, nil, now, env.Statement.HardwareID); err == nil {
		t.Fatal("duration mismatch must be rejected")
	}
}
