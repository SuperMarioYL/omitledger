// Package report emits the post-session omission report (m3 milestone): a
// markdown summary grouped by category with reopen count, reopen.jsonl
// re-injection into the next agent session, and JSON export for a CI
// merge-gate. m1 ships the CLI `list` view only.
package report

import "errors"

// ErrNotImplemented is returned by stubs until the m3 milestone lands.
var ErrNotImplemented = errors.New("omitledger report: not implemented until m3")

// Writer is the m3 report writer stub.
type Writer struct {
	// TODO(m3): *ledger.Store + output sink.
}

// New returns a stub report writer.
func New() *Writer { return &Writer{} }

// Markdown emits the post-session markdown summary grouped by category.
// Implemented in m3.
func (w *Writer) Markdown() error { return ErrNotImplemented }

// JSON exports the ledger for a CI merge-gate. Implemented in m3.
func (w *Writer) JSON() error { return ErrNotImplemented }
