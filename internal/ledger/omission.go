// Package ledger is the deliberate-omission ledger: a queryable, re-requestable
// record of every item a coding agent deliberately skipped, captured at the
// moment of the minimization decision together with the agent's own reason.
//
// The Omission record is the defensible abstraction this tool owns: it is
// distinct from any minimization engine that produces it (those engines make
// an agent lazy; this ledger accounts for what the laziness skipped).
package ledger

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Status of an Omission record.
const (
	StatusOpen     = "open"
	StatusReopened = "reopened"
	StatusResolved = "resolved"
)

// StatusClosed is a virtual filter value for `list --status closed`: every
// record that is no longer open (reopened or resolved).
const StatusClosed = "closed"

// ValidateStatusFilter returns an error for a --status value the ledger does
// not know. Unknown values must not silently render an empty ledger — a typo
// like "Open" would be indistinguishable from "nothing recorded".
func ValidateStatusFilter(status string) error {
	switch status {
	case "", StatusOpen, StatusReopened, StatusResolved, StatusClosed:
		return nil
	}
	return fmt.Errorf("unknown status filter %q (want open, reopened, resolved, closed, or empty)", status)
}

// MatchesStatus reports whether o is shown by `list --status <status>`: empty
// matches all; "closed" matches anything not open; anything else is an exact
// status match. Single source of truth for status filtering — Store.List and
// the CLI list view both delegate here.
func MatchesStatus(o Omission, status string) bool {
	switch status {
	case "":
		return true
	case StatusClosed:
		return o.Status != StatusOpen
	default:
		return o.Status == status
	}
}

// Category buckets the kind of thing that was skipped. Free-form strings are
// tolerated on read; these are the canonical values the agent skill emits.
const (
	CategoryTest    = "test"
	CategoryFile    = "file"
	CategorySection = "section"
	CategoryRefactor = "refactor"
	CategoryDoc     = "doc"
	CategoryLog     = "log"
)

// Omission is a single deliberately-skipped item, captured at decision time.
// It is the new primitive: per-omission provenance (the agent-stated reason)
// plus a re-request path (reopen), persisted across sessions in local SQLite.
type Omission struct {
	// ID is a ULID (monotonic, time-ordered, 26 Crockford-base32 chars). It is
	// assigned by the store at insert time; callers leave it empty.
	ID string `json:"id"`
	// SessionID is the agent run id (sourced from CLAUDE/Cursor env). "local"
	// when no session context is available.
	SessionID string `json:"session_id"`
	// File is the target path, or "*" for repo-wide omissions. May be empty.
	File string `json:"file"`
	// Item is what was omitted, e.g. "parser unit test".
	Item string `json:"item"`
	// Reason is the agent-stated reason AT decision time. The reason is the
	// asset: it distinguishes "trivial getter, low risk" from a cut corner.
	Reason string `json:"reason"`
	// Category buckets the omission (test | file | section | refactor | doc | log).
	Category string `json:"category"`
	// CreatedAt is when the omission was recorded (decision time).
	CreatedAt time.Time `json:"created_at"`
	// Status is open | reopened | resolved.
	Status string `json:"status"`
	// ReopenedAt is when the user re-requested the item, if any.
	ReopenedAt *time.Time `json:"reopened_at,omitempty"`
	// ReopenNote is an optional user note attached on reopen.
	ReopenNote string `json:"reopen_note,omitempty"`
}

// Validate returns an error if the record is missing the fields that make an
// omission meaningful (the item and the agent-stated reason). Everything else
// is optional provenance.
func (o Omission) Validate() error {
	if strings.TrimSpace(o.Item) == "" {
		return errors.New("omission: --item is required (what was skipped)")
	}
	if strings.TrimSpace(o.Reason) == "" {
		return errors.New("omission: --reason is required (the agent-stated reason)")
	}
	return nil
}

// crockford is the ULID alphabet (excludes I, L, O, U to avoid ambiguity).
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// NewULID returns a 26-char Crockford-base32 ULID: 48 bits of millisecond
// timestamp (time-ordered, monotonic within a millisecond across processes)
// followed by 80 bits of crypto random. The result sorts lexicographically
// by creation time, which makes the ledger read in insertion order.
func NewULID(now time.Time) (string, error) {
	var b [16]byte
	ms := uint64(now.UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	if _, err := rand.Read(b[6:]); err != nil {
		return "", err
	}
	return encodeBase32(b), nil
}

// encodeBase32 encodes a 16-byte (128-bit) value as 26 Crockford-base32 chars,
// matching the ULID spec (the final char carries 2 padding bits).
func encodeBase32(b [16]byte) string {
	// Work in 17 bytes (1 leading zero byte) so the bit stream aligns to
	// 5-bit groups; ULID consumes 26 groups = 130 bits, dropping the low 6
	// padding bits of the 136-bit window.
	var buf [17]byte
	buf[0] = 0
	copy(buf[1:], b[:])

	var out [26]byte
	idx := 0
	bits := 0
	nbits := 0
	for i := 0; i < len(buf) && idx < len(out); i++ {
		bits = (bits << 8) | int(buf[i])
		nbits += 8
		for nbits >= 5 && idx < len(out) {
			nbits -= 5
			out[idx] = crockford[(bits>>nbits)&0x1f]
			idx++
		}
	}
	if nbits > 0 && idx < len(out) {
		out[idx] = crockford[(bits<<(5-nbits))&0x1f]
	}
	return string(out[:])
}
