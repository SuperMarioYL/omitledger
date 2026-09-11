package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

const longDesc = `OmitLedger — the ledger that records what your coding agent deliberately skipped.

When an agent deliberately minimizes output — skips a file, omits a test,
writes less, excludes a section — it leaves no record of what it chose NOT to
do. OmitLedger captures each skipped item with the agent-stated reason and a
re-request button at the moment the minimization decision is made, so justified
laziness becomes a queryable, re-requestable asset distinct from a silently cut
corner.

The agent skill snippet (skills/omit-ledger/SKILL.md) makes an agent call
` + "`omitledger add`" + ` the instant it deliberately skips something.`

var rootCmd = &cobra.Command{
	Use:           "omitledger",
	Short:         "Ledger what your coding agent deliberately skipped",
	Long:          longDesc,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// storeFlag overrides the ledger DB path; resolved via --store > OMITLEDGER_STORE > default.
var storeFlag string

func init() {
	rootCmd.PersistentFlags().StringVar(&storeFlag, "store", "", "ledger DB path (default .omitledger/ledger.db; env OMITLEDGER_STORE)")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(reopenCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(exportCmd)
}

// resolveStorePath picks the ledger location: explicit flag → env → repo-local default.
func resolveStorePath() string {
	switch {
	case storeFlag != "":
		return storeFlag
	case os.Getenv("OMITLEDGER_STORE") != "":
		return os.Getenv("OMITLEDGER_STORE")
	default:
		return ledger.DefaultStorePath
	}
}

// openStore opens (and migrates) the ledger at the resolved path.
func openStore() (*ledger.Store, error) {
	return ledger.Open(resolveStorePath())
}

// sessionID returns the agent run id from the environment, or "local".
func sessionID() string {
	for _, k := range []string{"OMITLEDGER_SESSION", "CLAUDE_SESSION_ID", "CURSOR_SESSION_ID"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return "local"
}

// die prints an error to stderr in a consistent shape and returns it so RunE
// can `return die(err)` (cobra prints nothing thanks to SilenceErrors).
func die(err error) error {
	fmt.Fprintf(os.Stderr, "omitledger: %v\n", err)
	return err
}
