package ledger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreAddListReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()

	o1, err := s.Add(Omission{Item: "parser unit test", Reason: "trivial getter, low risk", File: "parser.go", Category: "test", SessionID: "s1"})
	if err != nil {
		t.Fatalf("add1: %v", err)
	}
	if o1.ID == "" {
		t.Fatal("add1: empty ID")
	}
	if len(o1.ID) != 26 {
		t.Fatalf("add1 ID len = %d, want 26 (ULID)", len(o1.ID))
	}
	if o1.Status != StatusOpen {
		t.Fatalf("add1 status = %q, want open", o1.Status)
	}

	o2, err := s.Add(Omission{Item: "retry-path error log", Reason: "rare branch, deferred", File: "retry.go", Category: "log", SessionID: "s1"})
	if err != nil {
		t.Fatalf("add2: %v", err)
	}

	// list returns oldest first (insertion order via ULID + created_at).
	all, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("list len = %d, want 2", len(all))
	}
	if all[0].ID != o1.ID {
		t.Fatalf("order: first = %s, want %s", all[0].ID, o1.ID)
	}

	// reopen o1.
	r, err := s.Reopen(o1.ID, "needs the test")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if r.Status != StatusReopened {
		t.Fatalf("reopen status = %q, want reopened", r.Status)
	}
	if r.ReopenedAt == nil {
		t.Fatal("reopen: nil ReopenedAt")
	}

	// status filter: open -> only o2.
	open, err := s.List(ListFilter{Status: "open"})
	if err != nil {
		t.Fatalf("list open: %v", err)
	}
	if len(open) != 1 || open[0].ID != o2.ID {
		t.Fatalf("open filter = %+v", open)
	}
	// closed (not open) -> only o1 (reopened).
	closed, err := s.List(ListFilter{Status: "closed"})
	if err != nil {
		t.Fatalf("list closed: %v", err)
	}
	if len(closed) != 1 || closed[0].ID != o1.ID {
		t.Fatalf("closed filter = %+v", closed)
	}

	openN, totalN, err := s.Counts()
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if openN != 1 || totalN != 2 {
		t.Fatalf("counts = open %d total %d, want 1/2", openN, totalN)
	}

	// reopen unknown id errors.
	if _, err := s.Reopen("bogus", ""); err == nil {
		t.Fatal("reopen bogus: expected error")
	}
}

func TestValidate(t *testing.T) {
	if err := (Omission{}).Validate(); err == nil {
		t.Fatal("empty omission should fail validation")
	}
	if err := (Omission{Item: "x"}).Validate(); err == nil {
		t.Fatal("missing reason should fail")
	}
	if err := (Omission{Item: "x", Reason: "y"}).Validate(); err != nil {
		t.Fatalf("valid omission failed: %v", err)
	}
}

func TestResolveRefLineAndID(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	o1, _ := s.Add(Omission{Item: "a", Reason: "r1"})
	o2, _ := s.Add(Omission{Item: "b", Reason: "r2"})

	// line number resolves to the nth omission (oldest first).
	id, err := s.ResolveRef("2")
	if err != nil {
		t.Fatalf("resolve line 2: %v", err)
	}
	if id != o2.ID {
		t.Fatalf("line 2 -> %s, want %s", id, o2.ID)
	}
	// ULID resolves to itself.
	id, err = s.ResolveRef(o1.ID)
	if err != nil {
		t.Fatalf("resolve ulid: %v", err)
	}
	if id != o1.ID {
		t.Fatalf("ulid -> %s, want %s", id, o1.ID)
	}
	// out of range and unknown both error.
	if _, err := s.ResolveRef("99"); err == nil {
		t.Fatal("line 99 should error")
	}
	if _, err := s.ResolveRef("not-a-ulid-or-int"); err == nil {
		t.Fatal("unknown ref should error")
	}
}

func TestListUnknownStatusErrors(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if _, err := s.Add(Omission{Item: "a", Reason: "r1"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	// A typo or wrong casing must be an error, never a silently empty list.
	for _, bad := range []string{"bogus", "Open", "CLOSED"} {
		if _, err := s.List(ListFilter{Status: bad}); err == nil {
			t.Fatalf("List status %q: expected error, got rows", bad)
		}
	}
	// The documented values still work, including the closed (not-open) view.
	for _, ok := range []string{"", "open", "reopened", "resolved", "closed"} {
		if _, err := s.List(ListFilter{Status: ok}); err != nil {
			t.Fatalf("List status %q: %v", ok, err)
		}
	}
}

func TestExpandHomePath(t *testing.T) {
	t.Setenv("HOME", "/fake/home")
	cases := []struct{ in, want string }{
		{"~", "/fake/home"},
		{"~/.omitledger/ledger.db", "/fake/home/.omitledger/ledger.db"},
		{"relative/ledger.db", "relative/ledger.db"},
		{"/abs/ledger.db", "/abs/ledger.db"},
		{"~user/ledger.db", "~user/ledger.db"}, // only ~/ is documented
	}
	for _, c := range cases {
		if got := expandHomePath(c.in); got != c.want {
			t.Fatalf("expandHomePath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestOpenExpandsTilde(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s, err := Open("~/.omitledger/ledger.db")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	if got := s.Path(); !strings.HasPrefix(got, os.Getenv("HOME")) || strings.Contains(got, "~") {
		t.Fatalf("Path() = %q, want under $HOME with no literal ~", got)
	}
	// A tilde store must not leave a literal "~" directory in the CWD.
	if _, err := os.Stat("~"); err == nil {
		t.Fatal("a literal ~ directory exists in the working directory")
	}
}

func TestAppendReinject(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	o1, err := s.Add(Omission{Item: "parser unit test", Reason: "r1", Category: "test"})
	if err != nil {
		t.Fatalf("add1: %v", err)
	}
	r1, err := s.Reopen(o1.ID, "needs the test")
	if err != nil {
		t.Fatalf("reopen1: %v", err)
	}
	path, err := s.AppendReinject(r1)
	if err != nil {
		t.Fatalf("append1: %v", err)
	}
	if want := filepath.Join(dir, ReopenLogName); path != want {
		t.Fatalf("reinject path = %q, want %q", path, want)
	}

	// A second re-request appends a second line (append-only event log).
	o2, err := s.Add(Omission{Item: "beta doc", Reason: "r2"})
	if err != nil {
		t.Fatalf("add2: %v", err)
	}
	r2, err := s.Reopen(o2.ID, "")
	if err != nil {
		t.Fatalf("reopen2: %v", err)
	}
	if _, err := s.AppendReinject(r2); err != nil {
		t.Fatalf("append2: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read reopen log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("reopen log has %d line(s), want 2:\n%s", len(lines), data)
	}
	var first Omission
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("line 1 is not JSON: %v", err)
	}
	if first.ID != o1.ID || first.Item != "parser unit test" || first.ReopenNote != "needs the test" {
		t.Fatalf("line 1 = %+v", first)
	}
	if first.ReopenedAt == nil {
		t.Fatal("line 1 has no reopened_at")
	}
	if !strings.Contains(lines[0], `"reopened_at"`) {
		t.Fatalf("line 1 lacks the reopened_at JSON key: %s", lines[0])
	}
}

func TestListLimitAppliesAfterStatusFilter(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	for _, item := range []string{"a", "b", "c"} {
		if _, err := s.Add(Omission{Item: item, Reason: "r"}); err != nil {
			t.Fatalf("add %s: %v", item, err)
		}
	}
	limited, err := s.List(ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(limited) != 2 || limited[0].Item != "a" || limited[1].Item != "b" {
		t.Fatalf("limit = %+v", limited)
	}
}

// m9: "oldest first" must follow TIME, not the RFC3339Nano TEXT column — the
// variable-length fraction makes byte order wrong for sub-second timestamps.
func TestListPreservesInsertionOrderAcrossFractionDigits(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	// ".9Z" (1 fraction digit) is stored before ".91Z" (2 digits); TEXT order
	// would put the later record first because 'Z' > '1'.
	first, err := s.Add(Omission{Item: "added FIRST", Reason: "r", CreatedAt: base.Add(900 * time.Millisecond)})
	if err != nil {
		t.Fatalf("add first: %v", err)
	}
	second, err := s.Add(Omission{Item: "added SECOND", Reason: "r", CreatedAt: base.Add(910 * time.Millisecond)})
	if err != nil {
		t.Fatalf("add second: %v", err)
	}
	got, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 || got[0].ID != first.ID || got[1].ID != second.ID {
		t.Fatalf("List broke insertion order: got [%s %s], want FIRST then SECOND", got[0].Item, got[1].Item)
	}
}

func TestListOrderByTimeNotText(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	// Chronological order 100ms < 120ms < 123ms; the stored TEXT strings
	// ("…0.1Z", "…0.12Z", "…0.123Z") sort in the exact reverse byte order.
	for i, ms := range []time.Duration{100, 120, 123} {
		if _, err := s.Add(Omission{Item: fmt.Sprintf("rec-%d", i), Reason: "r", CreatedAt: base.Add(ms * time.Millisecond)}); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}
	got, err := s.List(ListFilter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	want := []string{"rec-0", "rec-1", "rec-2"}
	for i, w := range want {
		if got[i].Item != w {
			t.Fatalf("List[%d] = %s, want %s (order must follow time, not created_at TEXT)", i, got[i].Item, w)
		}
	}
}

// The # column and `reopen <line>` resolve through List order, so the index
// view must agree with the sorted view on the same shuffled ledger.
func TestGetByIndexMatchesListOrder(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	items := []struct {
		item string
		at   time.Duration
	}{{"earliest", 100}, {"middle", 120}, {"latest", 123}}
	for _, it := range items {
		if _, err := s.Add(Omission{Item: it.item, Reason: "r", CreatedAt: base.Add(it.at * time.Millisecond)}); err != nil {
			t.Fatalf("add %s: %v", it.item, err)
		}
	}
	for n, want := range []string{"earliest", "middle", "latest"} {
		o, err := s.GetByIndex(n + 1)
		if err != nil {
			t.Fatalf("GetByIndex(%d): %v", n+1, err)
		}
		if o.Item != want {
			t.Fatalf("GetByIndex(%d) = %s, want %s", n+1, o.Item, want)
		}
	}
	id, err := s.ResolveRef("3")
	if err != nil {
		t.Fatalf("ResolveRef(3): %v", err)
	}
	if o, err := s.Get(id); err != nil || o.Item != "latest" {
		t.Fatalf("ResolveRef(3) -> %s (err %v), want latest", o.Item, err)
	}
}

// m11: the store path is spliced into a file: URI — '?', '#' and '%' are
// filesystem-legal and must round-trip instead of being parsed as query or
// fragment separators (which silently relocated the ledger).
func TestOpenSpecialCharPathRoundTrips(t *testing.T) {
	for _, name := range []string{"weird?name", "weird#name", "weird%20name"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			want := filepath.Join(dir, name, "ledger.db")
			s, err := Open(want)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer s.Close()
			if s.Path() != want {
				t.Fatalf("Store.Path() = %s, want %s", s.Path(), want)
			}
			added, err := s.Add(Omission{Item: "round trip", Reason: "r"})
			if err != nil {
				t.Fatalf("add: %v", err)
			}
			if _, err := s.Get(added.ID); err != nil {
				t.Fatalf("get after add: %v", err)
			}
			if _, err := os.Stat(want); err != nil {
				t.Fatalf("ledger DB not at the requested path: %v", err)
			}
			// Re-open: the data must be found through the same path.
			s2, err := Open(want)
			if err != nil {
				t.Fatalf("reopen: %v", err)
			}
			defer s2.Close()
			got, err := s2.List(ListFilter{})
			if err != nil {
				t.Fatalf("list after reopen: %v", err)
			}
			if len(got) != 1 || got[0].Item != "round trip" {
				t.Fatalf("reopened ledger = %+v, want the stored record", got)
			}
		})
	}
}
