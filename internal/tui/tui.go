// Package tui hosts the interactive ledger viewer (m2 milestone): a
// charmbracelet/bubbles-based table with cursor-selectable reopen. m1 ships
// the CLI `list` table view only; this package is the reserved home for the
// interactive viewer.
package tui

import (
	"errors"

	// Pinned for the m2 milestone: the interactive viewer will be built on
	// charmbracelet/bubbles. m1 blank-imports the table subpackage so the
	// dependency is present and go.mod reflects the planned stack; no m2 code yet.
	_ "github.com/charmbracelet/bubbles/table"
)

// ErrNotImplemented is returned by stubs until the m2 milestone lands.
var ErrNotImplemented = errors.New("omitledger tui: not implemented until m2")

// Viewer is the m2 interactive viewer stub.
type Viewer struct {
	// TODO(m2): bubbles table.Model + a *ledger.Store.
}

// New returns a stub viewer.
func New() *Viewer { return &Viewer{} }

// Run launches the interactive viewer with cursor-selectable reopen.
// Implemented in m2.
func (v *Viewer) Run() error { return ErrNotImplemented }
