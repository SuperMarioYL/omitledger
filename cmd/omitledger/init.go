package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the omission ledger in this repo",
	Long: `Creates the local SQLite ledger (.omitledger/ledger.db by default) and its
schema. Any subcommand also self-initializes, so init is the friendly explicit
entry point — run it once in a repo root to confirm the ledger is ready and
see how to install the agent skill snippet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := openStore()
		if err != nil {
			return die(err)
		}
		defer store.Close()
		open, total, err := store.Counts()
		if err != nil {
			return die(err)
		}
		fmt.Printf("omitledger ledger ready: %s\n", store.Path())
		fmt.Printf("  %d omission(s) recorded, %d open\n", total, open)
		fmt.Println()
		fmt.Println("Next: install the agent skill so your agent records omissions automatically:")
		fmt.Println("  mkdir -p ~/.claude/skills/omit-ledger && cp skills/omit-ledger/SKILL.md ~/.claude/skills/omit-ledger/SKILL.md")
		fmt.Println("Then, when your agent deliberately skips something, record it:")
		fmt.Println("  omitledger add --item \"parser unit test\" --reason \"trivial getter, low risk\" --file parser.go --category test")
		return nil
	},
}
