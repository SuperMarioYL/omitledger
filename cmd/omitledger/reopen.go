package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reopenNote string

var reopenCmd = &cobra.Command{
	Use:   "reopen <id|line>",
	Short: "Re-request a deliberately omitted item",
	Long: `Marks an omission re-requested — the user's ground-truth signal that this
skipped item should come back. The re-request path is what turns a silent cut
corner into an auditable, actionable record: justified laziness stays open-and-
resolved, suspicious omissions get reopened and addressed.

The argument is either a ULID (the ID column from ` + "`list`" + `) or a 1-based
line number (the # column from ` + "`list`" + ` — also under a --status filter),
so ` + "`reopen 1`" + ` reopens the first recorded omission without copying a ULID.

Reopen also re-injects the item into the next agent session: the record is
appended to .omitledger/reopen.jsonl, which the agent skill reads at the
start of the next session to re-request it.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return die(err)
		}
		defer st.Close()
		id, err := st.ResolveRef(args[0])
		if err != nil {
			return die(err)
		}
		o, err := st.Reopen(id, reopenNote)
		if err != nil {
			return die(err)
		}
		// The ledger record is already reopened; a failed re-injection
		// append must not undo it or lie about the reopen — warn and go on.
		if _, err := st.AppendReinject(o); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "omitledger: warning: reopen recorded, but the re-injection log was not updated: %v\n", err)
		}
		fmt.Printf("reopened %s: %s\n", o.ID, o.Item)
		if o.ReopenNote != "" {
			fmt.Printf("  note: %s\n", o.ReopenNote)
		}
		return nil
	},
}

func init() {
	reopenCmd.Flags().StringVar(&reopenNote, "note", "", "optional note on why this omission is being re-requested")
}
