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
	items := []ledger.Omission{
		{
			ID:       "000T13NAHVHCJ3N4PW8EBRXFDV",
			Status:   ledger.StatusOpen,
			Item:     strings.Repeat("跳过", 40),
			Reason:   strings.Repeat("理由", 40),
			Category: ledger.CategoryTest,
			File:     "parser.go",
		},
	}
	var buf bytes.Buffer
	writeOmissionTable(&buf, items)
	if !utf8.ValidString(buf.String()) {
		t.Fatalf("table output is not valid UTF-8:\n%q", buf.String())
	}
}
