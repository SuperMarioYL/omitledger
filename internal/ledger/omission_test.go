package ledger

import (
	"testing"
	"time"
)

func TestNewULID(t *testing.T) {
	now := time.Now()
	a, err := NewULID(now)
	if err != nil {
		t.Fatalf("ulid: %v", err)
	}
	if len(a) != 26 {
		t.Fatalf("ulid len = %d, want 26", len(a))
	}
	// every char must be in the Crockford alphabet.
	for _, r := range a {
		if !isValidCrockford(byte(r)) {
			t.Fatalf("ulid %q contains non-Crockford char %q", a, r)
		}
	}
	// time-ordered: an earlier timestamp sorts before a later one.
	b, _ := NewULID(now.Add(time.Millisecond))
	if a > b {
		t.Fatalf("ulid not time-ordered: %s > %s", a, b)
	}
}

func TestEncodeBase32RoundtripAlphabet(t *testing.T) {
	var zero [16]byte
	id := encodeBase32(zero)
	if len(id) != 26 {
		t.Fatalf("zero-len = %d, want 26", len(id))
	}
	for _, r := range id {
		if !isValidCrockford(byte(r)) {
			t.Fatalf("zero id %q has bad char %q", id, r)
		}
	}
}

func isValidCrockford(b byte) bool {
	for i := 0; i < len(crockford); i++ {
		if crockford[i] == b {
			return true
		}
	}
	return false
}

// Session provenance has one owner: the CLI add path and the MCP
// record_omission tool both resolve through SessionID.
func TestSessionIDFromEnv(t *testing.T) {
	for _, key := range []string{"OMITLEDGER_SESSION", "CLAUDE_SESSION_ID", "CURSOR_SESSION_ID"} {
		t.Setenv(key, "")
	}
	if got := SessionID(); got != "local" {
		t.Fatalf("SessionID() with no env = %q, want local", got)
	}
	t.Setenv("OMITLEDGER_SESSION", "s-explicit")
	if got := SessionID(); got != "s-explicit" {
		t.Fatalf("SessionID() = %q, want s-explicit", got)
	}
	t.Setenv("OMITLEDGER_SESSION", "")
	t.Setenv("CLAUDE_SESSION_ID", "s-claude")
	if got := SessionID(); got != "s-claude" {
		t.Fatalf("SessionID() = %q, want s-claude (first non-empty wins)", got)
	}
}
