package crypto

import (
	"testing"
)

// hwidMatches is a small helper mirroring the switch's HWID policy:
// empty on the license side means no binding; otherwise exact match.
func hwidMatches(licenseHWID, currentHWID string) bool {
	if licenseHWID == "" {
		return true
	}
	return licenseHWID == currentHWID
}

func TestHWID_Policy(t *testing.T) {
	cases := []struct {
		name    string
		license string
		current string
		want    bool
	}{
		{"both empty", "", "", true},
		{"license empty, machine has HWID", "", "abc123", true},
		{"match", "abc123", "abc123", true},
		{"mismatch", "abc123", "xyz789", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hwidMatches(tc.license, tc.current); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
