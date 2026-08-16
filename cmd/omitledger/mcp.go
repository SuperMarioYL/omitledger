package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// mcpCmd is the m2 milestone stub: native MCP server mode (record_omission,
// list_omissions, reopen_omission tools over stdio) + interactive TUI viewer.
// m1 ships the CLI + ledger only; this command is reserved so the help tree
// advertises the roadmap without implementing it.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "(m2) Run as an MCP server over stdio [not yet implemented]",
	Long: `Run omitledger as a native Model Context Protocol server, exposing
record_omission / list_omissions / reopen_omission tools so agents call the
ledger directly instead of shelling out, plus an interactive TUI viewer with
cursor-selectable reopen.

This lands in the m2 milestone. For now, record omissions via the CLI:

  omitledger add --item "..." --reason "..." --file ... --category ...
  omitledger list
  omitledger reopen <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "omitledger: `mcp` (MCP server mode + TUI viewer) is not implemented in m1 — coming in m2.")
		fmt.Fprintln(os.Stderr, "for now, use `omitledger add`, `omitledger list`, `omitledger reopen`.")
		return nil
	},
}
