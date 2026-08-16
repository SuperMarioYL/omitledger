package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

var (
	addItem     string
	addReason   string
	addFile     string
	addCategory string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Record a deliberate omission with the agent-stated reason",
	Long: `Records one item the agent deliberately skipped, captured AT the decision
moment together with the agent's own stated reason. The reason is the asset:
it distinguishes "trivial getter, low risk" from a cut corner.

This is the command the agent skill snippet shells out to whenever the agent
chooses to minimize — e.g. omitting a test, skipping a file, excluding a
section. Capturing the omission here, at decision time, is what makes the
ledger a queryable audit surface instead of a lost chat message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return die(err)
		}
		defer st.Close()
		o := ledger.Omission{
			Item:      addItem,
			Reason:    addReason,
			File:      addFile,
			Category:  addCategory,
			SessionID: sessionID(),
		}
		stored, err := st.Add(o)
		if err != nil {
			return die(err)
		}
		fmt.Printf("recorded omission %s\n", stored.ID)
		fmt.Printf("  item:     %s\n", stored.Item)
		if stored.File != "" {
			fmt.Printf("  file:     %s\n", stored.File)
		}
		if stored.Category != "" {
			fmt.Printf("  category: %s\n", stored.Category)
		}
		fmt.Printf("  reason:   %s\n", stored.Reason)
		fmt.Printf("  session:  %s\n", stored.SessionID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addItem, "item", "", "what was omitted (e.g. \"parser unit test\")")
	addCmd.Flags().StringVar(&addReason, "reason", "", "the agent-stated reason AT decision time")
	addCmd.Flags().StringVar(&addFile, "file", "", "target path, or \"*\" for repo-wide")
	addCmd.Flags().StringVar(&addCategory, "category", "", "test | file | section | refactor | doc | log")
	_ = addCmd.MarkFlagRequired("item")
	_ = addCmd.MarkFlagRequired("reason")
}
