package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

var (
	exportStatus  string
	exportSession string
)

// exportCmd is the m3 CI-gate surface: a machine-readable dump of the ledger.
// `omitledger export json > omissions.json` gives a merge-gate everything it
// needs (ids, items, reasons, statuses) without parsing the human views.
var exportCmd = &cobra.Command{
	Use:   "export [format]",
	Short: "Export the ledger in a machine-readable format",
	Long: `Export the ledger as JSON (a pretty-printed array of omission records
with id, item, reason, file, category, session, created_at, status, and
reopen fields). Meant for CI merge-gates and scripting:

  omitledger export json > omissions.json
  omitledger export json --status open | jq 'length'

Only "json" is supported. Filters mirror ` + "`list`" + `: --status
(open|reopened|resolved|closed) and --session.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := "json"
		if len(args) == 1 {
			format = args[0]
		}
		if format != "json" {
			return die(fmt.Errorf("unknown export format %q (only json is supported)", format))
		}
		st, err := openStore()
		if err != nil {
			return die(err)
		}
		defer st.Close()
		items, err := st.List(ledger.ListFilter{Status: exportStatus, SessionID: exportSession})
		if err != nil {
			return die(err)
		}
		if items == nil {
			items = []ledger.Omission{} // render [] instead of null on an empty ledger
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(items)
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportStatus, "status", "", "filter: open | reopened | resolved | closed (default: all)")
	exportCmd.Flags().StringVar(&exportSession, "session", "", "restrict the export to one agent session id")
}
