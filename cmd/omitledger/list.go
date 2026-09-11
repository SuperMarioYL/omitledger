package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded omissions, their reasons, and status",
	Long: `Prints the ledger as a table — each deliberately-skipped item with the
agent-stated reason, target file, category, and re-request status. This is the
view that lets a reviewer see, at a glance, what the agent chose NOT to do and
why.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return die(err)
		}
		defer st.Close()
		all, err := st.List(ledger.ListFilter{})
		if err != nil {
			return die(err)
		}
		rows, err := omissionRows(all, listStatus)
		if err != nil {
			return die(err)
		}
		if len(rows) == 0 {
			if listStatus != "" {
				fmt.Printf("no %s omissions recorded\n", listStatus)
			} else {
				fmt.Println("no omissions recorded yet — run `omitledger add` when your agent skips something")
			}
			return nil
		}
		writeOmissionTable(os.Stdout, rows)
		fmt.Printf("\n%d omission(s)\n", len(rows))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter: open | reopened | resolved | closed (default: all)")
}

// row is one table row: the omission plus its ABSOLUTE 1-based position in
// the ledger. The # column must stay directly usable with `reopen <line>`:
// ResolveRef resolves a line number against the full ledger, so a filtered
// view that renumbered from 1 would make `reopen 1` target the wrong record.
type row struct {
	num int
	o   ledger.Omission
}

// omissionRows validates the status filter and selects the rows to display,
// keeping each row's absolute ledger position. Filtering delegates to
// ledger.MatchesStatus — the same single source Store.List uses.
func omissionRows(all []ledger.Omission, status string) ([]row, error) {
	if err := ledger.ValidateStatusFilter(status); err != nil {
		return nil, err
	}
	var rows []row
	for i, o := range all {
		if ledger.MatchesStatus(o, status) {
			rows = append(rows, row{num: i + 1, o: o})
		}
	}
	return rows, nil
}

// writeOmissionTable renders the ledger as a fixed-width table. The leading
// # column is the absolute 1-based ledger position usable directly with
// `omitledger reopen` — under a --status filter too.
func writeOmissionTable(w io.Writer, rows []row) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tID\tSTATUS\tCATEGORY\tFILE\tITEM\tREASON")
	for _, r := range rows {
		o := r.o
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.num,
			o.ID,
			o.Status,
			nonEmpty(o.Category, "-"),
			nonEmpty(o.File, "-"),
			trunc(o.Item, 48),
			trunc(o.Reason, 48),
		)
	}
	tw.Flush()
}

func nonEmpty(s, dflt string) string {
	if strings.TrimSpace(s) == "" {
		return dflt
	}
	return s
}

// trunc shortens s to at most n runes, appending an ellipsis when it cuts.
// It operates on runes, never bytes: byte slicing would split a multi-byte
// UTF-8 rune mid-sequence and emit invalid UTF-8 (the ledger is zh-primary,
// so CJK item/reason text is the common case, not the edge case).
func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}
