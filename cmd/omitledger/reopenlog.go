package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// reopenLogCmd prints where the re-injection log lives for the resolved ledger.
// The log sits next to the DB, so a --store / OMITLEDGER_STORE override moves it
// away from the default .omitledger/ path; consumers (the agent skill) must ask
// rather than assume that spelling.
var reopenLogCmd = &cobra.Command{
	Use:   "reopen-log",
	Short: "Print the re-injection log path (reopen.jsonl) for the resolved ledger",
	Long: `Print the absolute path of the re-injection log — reopen.jsonl in the
ledger's directory.

Every reopen appends one JSON re-request event there. The agent skill reads this
path at session start to re-request the listed items; because the log follows
the ledger (not the working directory), resolve it with this command instead of
hard-coding .omitledger/reopen.jsonl, which only holds for the default ledger
location.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := ledger.ReopenLogPath(resolveStorePath())
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}
