package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
	"github.com/SuperMarioYL/omitledger/internal/report"
)

var (
	reportSession string
	reportOut     string
)

// reportCmd is the m3 milestone: post-session markdown report. Reopened items
// also land in reopen.jsonl beside the ledger at reopen time (see reopen.go), and
// the export command feeds a CI merge-gate.
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Emit a post-session omission report (markdown)",
	Long: `Emit a markdown post-session summary of the ledger: per-status counts,
the re-requested (reopened) items with their notes, and every omission
grouped by category. Reopened items are also re-injected into the next
agent session via reopen.jsonl beside the ledger (written on reopen; print the
path with ` + "`omitledger reopen-log`" + `); for a
machine-readable dump see ` + "`omitledger export json`" + `.

Prints to stdout by default; --out writes the report to a file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		defer st.Close()
		items, err := st.List(ledger.ListFilter{SessionID: reportSession})
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := report.Markdown(&buf, items); err != nil {
			return err
		}
		// --out is a file contract: write it even when the selection is empty
		// (a 0-record report), so callers never read a stale file from an
		// earlier run after an exit-0 "success".
		if reportOut != "" {
			if err := os.WriteFile(reportOut, buf.Bytes(), 0o644); err != nil {
				return fmt.Errorf("write report %s: %w", reportOut, err)
			}
			fmt.Printf("report written to %s\n", reportOut)
			return nil
		}
		if len(items) == 0 {
			if reportSession != "" {
				fmt.Printf("no omissions recorded for session %s\n", reportSession)
			} else {
				fmt.Println("no omissions recorded yet — run `omitledger add` when your agent skips something")
			}
			return nil
		}
		_, err = os.Stdout.Write(buf.Bytes())
		return err
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportSession, "session", "", "restrict the report to one agent session id")
	reportCmd.Flags().StringVar(&reportOut, "out", "", "write the report to this file (default: stdout)")
}
