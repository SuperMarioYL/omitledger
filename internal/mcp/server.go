// Package mcp hosts the native MCP server (m2 milestone): `omitledger mcp`
// serves record_omission / list_omissions / reopen_omission over stdio so
// agents call the ledger directly instead of shelling out. The interactive
// TUI viewer remains a stub in internal/tui.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// Server exposes the omission ledger as MCP tools. It wraps a *ledger.Store
// opened by the caller (the `omitledger mcp` command), which also owns the
// store's lifecycle.
type Server struct {
	store  *ledger.Store
	server *mcp.Server
}

// New builds the MCP server over st, advertising impl as its implementation
// info (name "omitledger", the CLI version).
func New(st *ledger.Store, version string) *Server {
	s := &Server{store: st}
	s.server = mcp.NewServer(&mcp.Implementation{Name: "omitledger", Version: version}, nil)
	mcp.AddTool(s.server, &mcp.Tool{
		Name: "record_omission",
		Description: "Record one item the coding agent deliberately skipped, captured at the " +
			"decision moment with the agent-stated reason. item and reason are required.",
	}, s.record)
	mcp.AddTool(s.server, &mcp.Tool{
		Name: "list_omissions",
		Description: "List recorded omissions, oldest first. Optional filters: status " +
			"(open | reopened | resolved | closed), session_id, limit. Returns a JSON array.",
	}, s.list)
	mcp.AddTool(s.server, &mcp.Tool{
		Name: "reopen_omission",
		Description: "Re-request a recorded omission: status becomes reopened and the item is " +
			"appended to .omitledger/reopen.jsonl for re-injection into the next agent session. " +
			"ref is an omission ULID or a 1-based line number from list_omissions.",
	}, s.reopen)
	return s
}

// Serve runs the MCP server over stdio until the client disconnects.
func (s *Server) Serve(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// recordInput is the record_omission tool input. Fields without omitempty are
// required in the inferred input schema.
type recordInput struct {
	Item     string `json:"item"`
	Reason   string `json:"reason"`
	File     string `json:"file,omitempty"`
	Category string `json:"category,omitempty"`
}

func (s *Server) record(_ context.Context, _ *mcp.CallToolRequest, in recordInput) (*mcp.CallToolResult, any, error) {
	stored, err := s.store.Add(ledger.Omission{
		Item:      in.Item,
		Reason:    in.Reason,
		File:      in.File,
		Category:  in.Category,
		SessionID: ledger.SessionID(),
	})
	if err != nil {
		return nil, nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "recorded omission %s\n", stored.ID)
	fmt.Fprintf(&b, "  item:     %s\n", stored.Item)
	if stored.File != "" {
		fmt.Fprintf(&b, "  file:     %s\n", stored.File)
	}
	if stored.Category != "" {
		fmt.Fprintf(&b, "  category: %s\n", stored.Category)
	}
	fmt.Fprintf(&b, "  reason:   %s\n", stored.Reason)
	fmt.Fprintf(&b, "  session:  %s", stored.SessionID)
	return textResult(b.String()), nil, nil
}

// listInput is the list_omissions tool input; empty fields mean "no filter".
type listInput struct {
	Status    string `json:"status,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func (s *Server) list(_ context.Context, _ *mcp.CallToolRequest, in listInput) (*mcp.CallToolResult, any, error) {
	items, err := s.store.List(ledger.ListFilter{Status: in.Status, SessionID: in.SessionID, Limit: in.Limit})
	if err != nil {
		return nil, nil, err
	}
	if items == nil {
		items = []ledger.Omission{} // render [] instead of null on an empty ledger
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode omissions: %w", err)
	}
	return textResult(string(data)), nil, nil
}

// reopenInput is the reopen_omission tool input.
type reopenInput struct {
	Ref  string `json:"ref"`
	Note string `json:"note,omitempty"`
}

func (s *Server) reopen(_ context.Context, _ *mcp.CallToolRequest, in reopenInput) (*mcp.CallToolResult, any, error) {
	id, err := s.store.ResolveRef(in.Ref)
	if err != nil {
		return nil, nil, err
	}
	o, err := s.store.Reopen(id, in.Note)
	if err != nil {
		return nil, nil, err
	}
	// Parity with the CLI reopen: the ledger record is already reopened; a
	// failed re-injection append must not undo it — warn and go on.
	warning := ""
	if _, err := s.store.AppendReinject(o); err != nil {
		warning = fmt.Sprintf("\nwarning: reopen recorded, but the re-injection log was not updated: %v", err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "reopened %s: %s\n", o.ID, o.Item)
	if o.ReopenNote != "" {
		fmt.Fprintf(&b, "  note: %s\n", o.ReopenNote)
	}
	b.WriteString(strings.TrimPrefix(warning, "\n"))
	return textResult(strings.TrimRight(b.String(), "\n")), nil, nil
}

// textResult wraps a plain-text tool response.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}
