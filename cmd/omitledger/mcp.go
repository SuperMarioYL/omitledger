package main

import (
	"github.com/spf13/cobra"

	mcpserver "github.com/SuperMarioYL/omitledger/internal/mcp"
)

// mcpCmd is the m2 milestone: native MCP server mode. It exposes the ledger
// as record_omission / list_omissions / reopen_omission tools over stdio so
// agents call the ledger directly instead of shelling out; the interactive
// TUI viewer (internal/tui) remains the only stub surface.
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run as an MCP server over stdio (record_omission / list_omissions / reopen_omission)",
	Long: `Run omitledger as a native Model Context Protocol server, exposing the
ledger as tools agents call directly instead of shelling out:

  record_omission  — record a deliberate omission with the agent-stated reason
  list_omissions   — list recorded omissions, oldest first (JSON array)
  reopen_omission  — re-request an omission; appends to .omitledger/reopen.jsonl

The server speaks MCP over stdin/stdout, so point your MCP client at:

  omitledger mcp

and it serves the ledger at the resolved store path (--store / OMITLEDGER_STORE),
exactly like the CLI commands. The interactive TUI viewer is not implemented yet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		defer st.Close()
		return mcpserver.New(st, version).Serve(cmd.Context())
	},
}
