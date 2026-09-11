// Package report emits the post-session omission report (m3 milestone): a
// markdown summary of the ledger grouped by category, with per-status counts
// and the re-requested items called out with their notes.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// Markdown writes a post-session summary of items to w: a status count line,
// the re-requested (reopened) items with their notes, then every omission
// grouped by category in ledger (time) order. Items keep their original
// Category strings; entries with no category group under "uncategorized".
func Markdown(w io.Writer, items []ledger.Omission) error {
	counts := map[string]int{}
	for _, o := range items {
		counts[o.Status]++
	}
	fmt.Fprintf(w, "# OmitLedger session report\n\n")
	fmt.Fprintf(w, "%d omission(s) recorded — %d open, %d reopened, %d resolved.\n\n",
		len(items), counts[ledger.StatusOpen], counts[ledger.StatusReopened], counts[ledger.StatusResolved])

	reopened := 0
	for _, o := range items {
		if o.Status == ledger.StatusReopened || o.ReopenedAt != nil {
			reopened++
		}
	}
	fmt.Fprintf(w, "## Re-requested (%d)\n\n", reopened)
	if reopened == 0 {
		fmt.Fprintln(w, "Nothing re-requested — every omission stands as recorded.")
	} else {
		for _, o := range items {
			if o.Status != ledger.StatusReopened && o.ReopenedAt == nil {
				continue
			}
			note := ""
			if o.ReopenNote != "" {
				note = fmt.Sprintf(" — note: %s", o.ReopenNote)
			}
			fmt.Fprintf(w, "- `%s` **%s** (%s)%s\n", o.ID, o.Item, fileOrAny(o), note)
			fmt.Fprintf(w, "  reason: %s\n", o.Reason)
		}
	}
	fmt.Fprintln(w)

	var order []string
	byCategory := map[string][]ledger.Omission{}
	for _, o := range items {
		cat := o.Category
		if strings.TrimSpace(cat) == "" {
			cat = "uncategorized"
		}
		if _, seen := byCategory[cat]; !seen {
			order = append(order, cat)
		}
		byCategory[cat] = append(byCategory[cat], o)
	}
	for _, cat := range order {
		fmt.Fprintf(w, "## %s (%d)\n\n", cat, len(byCategory[cat]))
		for _, o := range byCategory[cat] {
			fmt.Fprintf(w, "- `%s` **%s** (%s) — %s — %s\n",
				o.ID, o.Item, fileOrAny(o), o.Status, o.Reason)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func fileOrAny(o ledger.Omission) string {
	if strings.TrimSpace(o.File) == "" {
		return "*"
	}
	return o.File
}
