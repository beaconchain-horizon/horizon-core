package main

import "testing"

func TestBuildCanonical(t *testing.T) {
	l := &license{
		LicenseID: "HORIZON-GOLD-001",
		ProductID: "horizon-core",
		UserID:    "bank-markazi",
		Volume:    50000000,
		Status:    "active",
		ExpiresAt: 1823257127,
	}
	got := buildCanonical(l)
	want := "HORIZON-GOLD-001|horizon-core|bank-markazi|50000000||active|2027-10-11T12:18:47Z"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
