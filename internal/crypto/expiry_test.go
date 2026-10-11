package crypto

import (
	"testing"
	"time"
)

// verifyExpiry is a small helper that mirrors the check the Switch
// performs: now must be within [issued_at, expires_at].
func verifyExpiry(now time.Time, issuedAt, expiresAt time.Time) bool {
	if now.Before(issuedAt) {
		return false
	}
	if now.After(expiresAt) {
		return false
	}
	return true
}

func TestExpiry_Boundaries(t *testing.T) {
	issued := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	expires := time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"before issue", issued.Add(-time.Hour), false},
		{"at issue", issued, true},
		{"mid life", time.Date(2027, 6, 15, 12, 0, 0, 0, time.UTC), true},
		{"at expiry", expires, true},
		{"after expiry", expires.Add(time.Second), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := verifyExpiry(tc.now, issued, expires); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
