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
line number (the # column from ` + "`list`" + `), so ` + "`reopen 1`" + ` reopens the
first recorded omission without copying a ULID.

In m3, reopen also re-injects the item into the next agent session via
.omitledger/reopen.jsonl.`,
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
