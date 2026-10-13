package mtls

import "testing"

func TestLoad_MissingPaths(t *testing.T) {
	if _, err := Load(Config{}); err == nil {
		t.Fatal("expected error on missing paths")
	}
}

func TestEnabled_DefaultOff(t *testing.T) {
	if Enabled() {
		t.Fatal("mtls must be off by default")
	}
}
