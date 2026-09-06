package switch_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"horizon-core/internal/crypto"
)

// BenchmarkValidatorSigning تست سرعت امضای دیجیتال برای هر ولیدیتور
func BenchmarkValidatorSigning(b *testing.B) {
	// کلید خصوصی تستی
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	data := []byte("Validator Data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.SignData(data, priv)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidatorVerification تست سرعت تأیید امضا (Verify) برای هر ولیدیتور
func BenchmarkValidatorVerification(b *testing.B) {
	// تولید کلید و امضا
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub := &priv.PublicKey
	data := []byte("Validator Data")
	sig, _ := crypto.SignData(data, priv)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !crypto.VerifySignature(pub, data, sig) {
			b.Fatal("Invalid signature")
		}
	}
}
