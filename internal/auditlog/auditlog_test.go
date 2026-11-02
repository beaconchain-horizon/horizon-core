package auditlog

import "testing"

func TestAudit_ChainIntact(t *testing.T) {
	l := New()
	l.Append("license_generate", "admin", "id=1")
	l.Append("license_generate", "admin", "id=2")
	l.Append("suspend_all", "admin", "")
	if got := l.Verify(); got != 0 {
		t.Fatalf("chain broken at %d", got)
	}
}

func TestAudit_TamperDetected(t *testing.T) {
	l := New()
	l.Append("a", "x", "1")
	l.Append("b", "y", "2")
	l.Append("c", "z", "3")
	l.mu.Lock()
	l.entries[1].Payload = "TAMPERED"
	l.mu.Unlock()
	if got := l.Verify(); got != 2 {
		t.Fatalf("expected break at 2, got %d", got)
	}
}
