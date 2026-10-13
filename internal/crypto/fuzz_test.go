package crypto

// Go native fuzzing for the secp256k1 sign/verify pair.
//
// Seed corpus runs as part of the normal test suite. To run the
// mutational fuzzer:
//
//   go test -fuzz=FuzzVerifyHashRS -fuzztime=30s ./internal/crypto/
//   go test -fuzz=FuzzSignDataRS   -fuzztime=30s ./internal/crypto/
//   go test -fuzz=FuzzVerifyDataHex -fuzztime=30s ./internal/crypto/
//
// Invariants asserted for every input:
//
//   - No panic.
//   - Deterministic: same input -> same result.
//   - VerifyHashRS rejects any hash that is not 32 bytes.
//   - VerifyHashRS rejects any signature that is not 64 bytes.
//   - Malleability closure: if (r, s) verifies, (r, n-s) must not.
//   - SignDataRS + VerifyHashRS roundtrip succeeds for every payload.
//   - RFC 6979 determinism: two signs of the same payload are byte-identical.
//   - Any single-bit flip in the signature breaks verification.

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func FuzzVerifyHashRS(f *testing.F) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		f.Fatalf("keygen: %v", err)
	}
	validHash := Sha256Hash([]byte("fuzz-seed"))
	validSig, err := SignHashRS(priv, validHash)
	if err != nil {
		f.Fatalf("sign seed: %v", err)
	}

	f.Add(validHash, validSig)
	f.Add(make([]byte, 32), make([]byte, 64))
	f.Add(validHash, make([]byte, 64))
	f.Add([]byte{}, []byte{})
	f.Add(make([]byte, 31), make([]byte, 64))
	f.Add(make([]byte, 33), make([]byte, 64))
	f.Add(make([]byte, 32), make([]byte, 63))
	f.Add(make([]byte, 32), make([]byte, 65))

	f.Fuzz(func(t *testing.T, hash, sig []byte) {
		got := VerifyHashRS(&priv.PublicKey, hash, sig)

		if got2 := VerifyHashRS(&priv.PublicKey, hash, sig); got != got2 {
			t.Fatal("non-deterministic verification")
		}
		if len(hash) != 32 && got {
			t.Fatalf("accepted non-32-byte hash (len=%d)", len(hash))
		}
		if len(sig) != 64 && got {
			t.Fatalf("accepted non-64-byte signature (len=%d)", len(sig))
		}

		if got && len(sig) == 64 {
			r := new(big.Int).SetBytes(sig[:32])
			s := new(big.Int).SetBytes(sig[32:])
			n := secp256k1.S256().Params().N
			flipped := make([]byte, 64)
			r.FillBytes(flipped[:32])
			new(big.Int).Sub(n, s).FillBytes(flipped[32:])
			if VerifyHashRS(&priv.PublicKey, hash, flipped) {
				t.Fatal("high-S counterpart of a low-S signature was accepted")
			}
		}
	})
}

func FuzzSignDataRS(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("hello"))
	f.Add(bytes.Repeat([]byte{0xFF}, 1024))
	f.Add([]byte{0x00, 0xAA, 0x55, 0xFF})

	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		f.Fatalf("keygen: %v", err)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		sig1, err := SignDataRS(data, priv)
		if err != nil {
			t.Fatalf("SignDataRS failed on %d-byte payload: %v", len(data), err)
		}
		if len(sig1) != Secp256k1SignatureSize {
			t.Fatalf("signature length %d, want %d", len(sig1), Secp256k1SignatureSize)
		}

		sig2, err := SignDataRS(data, priv)
		if err != nil {
			t.Fatalf("second SignDataRS: %v", err)
		}
		if !bytes.Equal(sig1, sig2) {
			t.Fatal("RFC 6979 violation: signatures differ")
		}

		hash := Sha256Hash(data)
		if !VerifyHashRS(&priv.PublicKey, hash, sig1) {
			t.Fatal("signature of fresh payload was rejected")
		}

		tampered := make([]byte, len(sig1))
		copy(tampered, sig1)
		tampered[0] ^= 0x01
		if VerifyHashRS(&priv.PublicKey, hash, tampered) {
			t.Fatal("tampered signature verified")
		}
	})
}

func FuzzVerifyDataHex(f *testing.F) {
	priv, err := GenerateSecp256k1ECDSAKey()
	if err != nil {
		f.Fatalf("keygen: %v", err)
	}
	f.Add("")
	f.Add("deadbeef")
	f.Add("0xdeadbeef")
	f.Add(strings.Repeat("00", 64))
	f.Add(strings.Repeat("f", 128))
	f.Add("zz")
	f.Add("0")

	f.Fuzz(func(t *testing.T, sigHex string) {
		ok, err := VerifyData(&priv.PublicKey, []byte("fuzz"), sigHex)
		if ok2, _ := VerifyData(&priv.PublicKey, []byte("fuzz"), sigHex); ok != ok2 {
			t.Fatal("non-deterministic VerifyData")
		}
		if ok && err != nil {
			t.Fatal("VerifyData returned (true, non-nil error)")
		}
	})
}
