// Command omitledger is the deliberate-omission ledger CLI.
//
// It records each item a coding agent deliberately skips — a file, a test, a
// section — at the moment of the minimization decision, together with the
// agent's own stated reason, and exposes a one-key re-request button. Run an
// agent task; run `omitledger list`; see each skipped item with its reason and
// status. Run `omitledger reopen <id>` to flag the one you want back.
package main

import "os"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
