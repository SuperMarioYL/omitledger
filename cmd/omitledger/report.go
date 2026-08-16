package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// reportCmd is the m3 milestone stub: post-session markdown report +
// reopen.jsonl re-injection + JSON export for a CI merge-gate.
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "(m3) Emit a post-session omission report [not yet implemented]",
	Long: `Emit a markdown post-session summary of the ledger (grouped by category,
with reopen count), write reopened items to .omitledger/reopen.jsonl so the
next agent session re-requests them, and ` + "`omitledger export json`" + ` feeds a CI
merge-gate.

This lands in the m3 milestone. For now, review omissions with:

  omitledger list
  omitledger list --status open
  omitledger reopen <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "omitledger: `report` (markdown report + reopen.jsonl + JSON export) is not implemented in m1 — coming in m3.")
		fmt.Fprintln(os.Stderr, "for now, use `omitledger list` to review the ledger.")
		return nil
	},
}
