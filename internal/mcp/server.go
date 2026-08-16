// Package mcp hosts the native MCP server (m2 milestone). m1 ships the CLI
// ledger only; this package is the reserved home for the stdio MCP transport
// that exposes record_omission / list_omissions / reopen_omission tools so
// agents call the ledger directly without shelling out.
package mcp

import (
	"errors"

	// Pinned for the m2 milestone: the native MCP transport will be built on
	// the modelcontextprotocol/go-sdk mcp package. m1 blank-imports it so the
	// dependency is present and go.mod reflects the planned stack; no m2 code yet.
	_ "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ErrNotImplemented is returned by stubs until the m2 milestone lands.
var ErrNotImplemented = errors.New("omitledger mcp: not implemented until m2")

// Server is the m2 MCP server stub. It holds a modelcontextprotocol/go-sdk
// server and a *ledger.Store in m2; for now it is an empty placeholder so the
// import graph and help tree advertise the roadmap without implementing it.
type Server struct {
	// TODO(m2): store *ledger.Store and *mcp.Server (go-sdk).
}

// New returns an unconfigured m2 server stub.
func New() *Server { return &Server{} }

// Serve runs the MCP server over stdio. Implemented in m2.
func (s *Server) Serve() error { return ErrNotImplemented }
