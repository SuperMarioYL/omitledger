package main

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

func TestTruncIsRuneSafe(t *testing.T) {
	// 60 CJK runes = 180 bytes: the byte-based truncation splits a rune.
	cjk := strings.Repeat("测", 60)
	got := trunc(cjk, 48)
	if !utf8.ValidString(got) {
		t.Fatalf("trunc produced invalid UTF-8: %q", got)
	}
	if want := strings.Repeat("测", 47) + "…"; got != want {
		t.Fatalf("trunc CJK = %q, want %q", got, want)
	}
}

func TestTruncShortStringsPassThrough(t *testing.T) {
	for _, s := range []string{"", "short", strings.Repeat("测", 10), "  padded  "} {
		if got := trunc(s, 48); got != strings.TrimSpace(s) {
			t.Fatalf("trunc(%q) = %q, want passthrough %q", s, got, strings.TrimSpace(s))
		}
	}
}

func TestWriteOmissionTableValidUTF8(t *testing.T) {
	all := []ledger.Omission{
		{
			ID:       "000T13NAHVHCJ3N4PW8EBRXFDV",
			Status:   ledger.StatusOpen,
			Item:     strings.Repeat("跳过", 40),
			Reason:   strings.Repeat("理由", 40),
			Category: ledger.CategoryTest,
			File:     "parser.go",
		},
	}
	rows, err := omissionRows(all, "")
	if err != nil {
		t.Fatalf("omissionRows: %v", err)
	}
	var buf bytes.Buffer
	writeOmissionTable(&buf, rows)
	if !utf8.ValidString(buf.String()) {
		t.Fatalf("table output is not valid UTF-8:\n%q", buf.String())
	}
}

func TestListFilteredRowsKeepAbsoluteNumbers(t *testing.T) {
	// Three omissions, the first one already reopened. The open-filtered
	// view must number its rows 2 and 3 (absolute ledger positions), NOT
	// renumber from 1 — `reopen 1` resolves against the full ledger, so a
	// renumbered view would send it to the wrong record.
	all := []ledger.Omission{
		{ID: "id1", Status: ledger.StatusReopened, Item: "alpha", Reason: "r1"},
		{ID: "id2", Status: ledger.StatusOpen, Item: "beta", Reason: "r2"},
		{ID: "id3", Status: ledger.StatusOpen, Item: "gamma", Reason: "r3"},
	}
	rows, err := omissionRows(all, ledger.StatusOpen)
	if err != nil {
		t.Fatalf("omissionRows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("open filter returned %d rows, want 2", len(rows))
	}
	for i, want := range []int{2, 3} {
		if rows[i].num != want {
			t.Fatalf("row %d numbered %d, want absolute position %d", i, rows[i].num, want)
		}
		if rows[i].o.ID != all[want-1].ID {
			t.Fatalf("row %d is %s, want %s", i, rows[i].o.ID, all[want-1].ID)
		}
	}
	// The unfiltered view keeps 1-based positions of the full ledger.
	allRows, err := omissionRows(all, "")
	if err != nil {
		t.Fatalf("omissionRows all: %v", err)
	}
	for i, r := range allRows {
		if r.num != i+1 {
			t.Fatalf("unfiltered row %d numbered %d", i, r.num)
		}
	}
}

func TestOmissionRowsUnknownStatusErrors(t *testing.T) {
	all := []ledger.Omission{{ID: "id1", Status: ledger.StatusOpen, Item: "a", Reason: "r"}}
	// A typo or wrong casing must be an error, never an empty ledger.
	for _, bad := range []string{"bogus", "Open", "CLOSED", "reopened "} {
		if _, err := omissionRows(all, bad); err == nil {
			t.Fatalf("status %q: expected error, got rows", bad)
		}
	}
}
