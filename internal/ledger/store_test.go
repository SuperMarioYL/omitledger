package ledger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
