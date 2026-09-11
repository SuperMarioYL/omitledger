package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

func sampleLedger() []ledger.Omission {
	reopenedAt := time.Date(2026, 9, 11, 4, 0, 0, 0, time.UTC)
	return []ledger.Omission{
		{ID: "id1", Item: "parser unit test", Reason: "trivial getter, low risk", File: "parser.go", Category: "test", Status: ledger.StatusReopened, ReopenedAt: &reopenedAt, ReopenNote: "needs the test"},
		{ID: "id2", Item: "retry-path error log", Reason: "rare branch, deferred", File: "retry.go", Category: "log", Status: ledger.StatusOpen},
		{ID: "id3", Item: "overview section", Reason: "covered by README", File: "doc.md", Category: "", Status: ledger.StatusResolved},
	}
}

func render(t *testing.T, items []ledger.Omission) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Markdown(&buf, items); err != nil {
		t.Fatalf("markdown: %v", err)
	}
	return buf.String()
}

func TestMarkdownStatusCounts(t *testing.T) {
	out := render(t, sampleLedger())
	want := "3 omission(s) recorded — 1 open, 1 reopened, 1 resolved."
	if !strings.Contains(out, want) {
		t.Fatalf("count line missing:\nwant substring %q\ngot:\n%s", want, out)
	}
}

func TestMarkdownGroupsByCategoryInLedgerOrder(t *testing.T) {
	out := render(t, sampleLedger())
	for _, want := range []string{"## Re-requested (1)", "## test (1)", "## log (1)", "## uncategorized (1)"} {
		if !strings.Contains(out, want) {
			t.Fatalf("section %q missing in:\n%s", want, out)
		}
	}
	if strings.Index(out, "## test") > strings.Index(out, "## log") {
		t.Fatalf("categories out of ledger order:\n%s", out)
	}
}

func TestMarkdownReopenedCarriesNoteAndReason(t *testing.T) {
	out := render(t, sampleLedger())
	for _, want := range []string{"`id1` **parser unit test** (parser.go)", "note: needs the test", "reason: trivial getter, low risk"} {
		if !strings.Contains(out, want) {
			t.Fatalf("reopened detail %q missing in:\n%s", want, out)
		}
	}
}

func TestMarkdownEmptyLedger(t *testing.T) {
	out := render(t, nil)
	for _, want := range []string{"0 omission(s) recorded — 0 open, 0 reopened, 0 resolved.", "## Re-requested (0)", "Nothing re-requested"} {
		if !strings.Contains(out, want) {
			t.Fatalf("empty-ledger line %q missing in:\n%s", want, out)
		}
	}
}

func TestMarkdownFileMissingRendersAny(t *testing.T) {
	out := render(t, []ledger.Omission{{ID: "id9", Item: "x", Reason: "r", Status: ledger.StatusOpen}})
	if !strings.Contains(out, "`id9` **x** (*)") {
		t.Fatalf("repo-wide marker (*) missing in:\n%s", out)
	}
}
