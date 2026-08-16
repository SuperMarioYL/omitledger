package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded omissions, their reasons, and status",
	Long: `Prints the ledger as a table — each deliberately-skipped item with the
agent-stated reason, target file, category, and re-request status. This is the
view that lets a reviewer see, at a glance, what the agent chose NOT to do and
why.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return die(err)
		}
		defer st.Close()
		items, err := st.List(ledger.ListFilter{Status: listStatus})
		if err != nil {
			return die(err)
		}
		if len(items) == 0 {
			if listStatus != "" {
				fmt.Printf("no %s omissions recorded\n", listStatus)
			} else {
				fmt.Println("no omissions recorded yet — run `omitledger add` when your agent skips something")
			}
			return nil
		}
		writeOmissionTable(os.Stdout, items)
		fmt.Printf("\n%d omission(s)\n", len(items))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter: open | reopened | resolved | closed (default: all)")
}

// writeOmissionTable renders the ledger as a fixed-width table. The leading
// # column is the 1-based line number usable directly with `omitledger reopen`.
func writeOmissionTable(w io.Writer, items []ledger.Omission) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tID\tSTATUS\tCATEGORY\tFILE\tITEM\tREASON")
	for i, o := range items {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			i+1,
			o.ID,
			o.Status,
			nonEmpty(o.Category, "-"),
			nonEmpty(o.File, "-"),
			trunc(o.Item, 48),
			trunc(o.Reason, 48),
		)
	}
	tw.Flush()
}

func nonEmpty(s, dflt string) string {
	if strings.TrimSpace(s) == "" {
		return dflt
	}
	return s
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return s[:n-1] + "…"
}
