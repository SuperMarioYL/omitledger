// Command omitledger is the deliberate-omission ledger CLI.
//
// It records each item a coding agent deliberately skips — a file, a test, a
// section — at the moment of the minimization decision, together with the
// agent's own stated reason, and exposes a one-key re-request button. Run an
// agent task; run `omitledger list`; see each skipped item with its reason and
// status. Run `omitledger reopen <id>` to flag the one you want back.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

// run executes the CLI and is the single error print point: the root command
// runs with SilenceErrors, so errors cobra raises itself (unknown command,
// unknown flag) would otherwise exit 1 with no output at all. Every failure —
// cobra-raised or returned from a RunE — prints here in the documented
// `omitledger: <message>` stderr shape, exactly once.
func run(args []string) error {
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "omitledger: %v\n", err)
		return err
	}
	return nil
}
