package ledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (no cgo → single static binary)
)

// DefaultStorePath is the ledger location relative to the working directory:
// a repo-local .omitledger/ledger.db, matching the happy path (`omitledger init`
// in repo root creates .omitledger/ledger.db). Override with --store or
// the OMITLEDGER_STORE env var (e.g. ~/.omitledger/ledger.db for a global
// ledger).
const DefaultStorePath = ".omitledger/ledger.db"

// Store is the local SQLite-backed omission ledger.
type Store struct {
	db   *sql.DB
	path string
}

// Open opens (creating the parent directory and schema if needed) the ledger
// at path. Any command is self-initializing, so `add` works even before an
// explicit `init`; `init` exists as the friendly explicit setup entry point.
func Open(path string) (*Store, error) {
	if path == "" {
		path = DefaultStorePath
	}
	path = expandHomePath(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve store path: %w", err)
	}
	dir := filepath.Dir(abs)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create ledger dir %s: %w", dir, err)
	}
	dsn := "file:" + abs + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open ledger %s: %w", abs, err)
	}
	// A CLI opens, mutates, closes per invocation; one connection keeps SQLite
	// happy and avoids "database is locked" on the writer.
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: abs}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Path returns the absolute path of the ledger DB file.
func (s *Store) Path() string { return s.path }

// expandHomePath expands a leading "~" or "~/" to the user's home directory.
// The documented global-ledger path (~/.omitledger/ledger.db via
// OMITLEDGER_STORE) relies on this: without expansion the literal "~" becomes
// a directory in the working directory and the ledger silently lands in the
// wrong place. Other ~-prefixed forms (~user/…) are returned unchanged.
func expandHomePath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS omissions (
    id          TEXT PRIMARY KEY,
    session_id  TEXT NOT NULL DEFAULT '',
    file        TEXT NOT NULL DEFAULT '',
    item        TEXT NOT NULL,
    reason      TEXT NOT NULL,
    category    TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open',
    reopened_at TEXT,
    reopen_note TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_omissions_status  ON omissions(status);
CREATE INDEX IF NOT EXISTS idx_omissions_created ON omissions(created_at);
CREATE INDEX IF NOT EXISTS idx_omissions_session ON omissions(session_id);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate ledger schema: %w", err)
	}
	return nil
}

// Add records a new omission and returns the stored record (with ID and
// CreatedAt filled). The item and reason must be non-empty (see Validate).
func (s *Store) Add(o Omission) (Omission, error) {
	if err := o.Validate(); err != nil {
		return o, err
	}
	if o.ID == "" {
		id, err := NewULID(time.Now())
		if err != nil {
			return o, err
		}
		o.ID = id
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	if o.Status == "" {
		o.Status = StatusOpen
	}
	const q = `INSERT INTO omissions
(id, session_id, file, item, reason, category, created_at, status, reopened_at, reopen_note)
VALUES (?,?,?,?,?,?,?,?,?,?)`
	var reopenedAt any
	if o.ReopenedAt != nil {
		reopenedAt = o.ReopenedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err := s.db.Exec(q,
		o.ID, o.SessionID, o.File, o.Item, o.Reason, o.Category,
		o.CreatedAt.UTC().Format(time.RFC3339Nano), o.Status, reopenedAt, o.ReopenNote,
	)
	if err != nil {
		return o, fmt.Errorf("insert omission: %w", err)
	}
	return o, nil
}

// ListFilter narrows List. Empty Status means "all". SessionID empty means
// all sessions. Limit<=0 means no cap (ordered by creation time ascending).
type ListFilter struct {
	Status    string
	SessionID string
	Limit     int
}

// List returns omissions matching filter, oldest first (insertion order,
// leveraging ULID time-ordering). Status filtering delegates to MatchesStatus
// so the store and the CLI list view share one definition; the limit applies
// after filtering (it caps the result set, not the scan).
func (s *Store) List(f ListFilter) ([]Omission, error) {
	if err := ValidateStatusFilter(f.Status); err != nil {
		return nil, err
	}
	q := `SELECT id, session_id, file, item, reason, category, created_at, status, reopened_at, reopen_note
FROM omissions`
	var args []any
	if f.SessionID != "" {
		q += " WHERE session_id = ?"
		args = append(args, f.SessionID)
	}
	q += " ORDER BY created_at ASC, id ASC"
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list omissions: %w", err)
	}
	defer rows.Close()
	var out []Omission
	for rows.Next() {
		o, err := scanOmission(rows)
		if err != nil {
			return nil, err
		}
		if !MatchesStatus(o, f.Status) {
			continue
		}
		if f.Limit > 0 && len(out) == f.Limit {
			break
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Get returns a single omission by ID.
func (s *Store) Get(id string) (Omission, error) {
	const q = `SELECT id, session_id, file, item, reason, category, created_at, status, reopened_at, reopen_note
FROM omissions WHERE id = ?`
	row := s.db.QueryRow(q, id)
	o, err := scanOmissionRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Omission{}, fmt.Errorf("omission %s not found", id)
		}
		return Omission{}, err
	}
	return o, nil
}

// GetByIndex returns the 1-based nth omission in creation order (oldest first),
// matching the row number shown by `list`. n must be >= 1. This lets a user
// reopen by line number (`omitledger reopen 1`) without copying a ULID.
func (s *Store) GetByIndex(n int) (Omission, error) {
	if n < 1 {
		return Omission{}, fmt.Errorf("line %d out of range (must be >= 1)", n)
	}
	items, err := s.List(ListFilter{})
	if err != nil {
		return Omission{}, err
	}
	if n > len(items) {
		return Omission{}, fmt.Errorf("line %d out of range (only %d omission(s) recorded)", n, len(items))
	}
	return items[n-1], nil
}

// ResolveRef maps a reopen argument to a concrete ID: a pure positive integer
// is treated as a 1-based line number (per `list`); anything else is treated as
// a ULID and looked up directly.
func (s *Store) ResolveRef(ref string) (string, error) {
	if n, err := strconv.Atoi(ref); err == nil && n >= 0 {
		if n == 0 {
			return "", fmt.Errorf("line 0 out of range (line numbers start at 1)")
		}
		o, err := s.GetByIndex(n)
		if err != nil {
			return "", err
		}
		return o.ID, nil
	}
	if _, err := s.Get(ref); err != nil {
		return "", err
	}
	return ref, nil
}

// Reopen marks an open omission re-requested: status→reopened, ReopenedAt
// set to now, optional note attached. Re-opening an already-reopened record
// refreshes ReopenedAt and the note (idempotent intent: "still want this").
func (s *Store) Reopen(id, note string) (Omission, error) {
	o, err := s.Get(id)
	if err != nil {
		return o, err
	}
	now := time.Now().UTC()
	const q = `UPDATE omissions SET status = ?, reopened_at = ?, reopen_note = ? WHERE id = ?`
	_, err = s.db.Exec(q, StatusReopened, now.Format(time.RFC3339Nano), note, id)
	if err != nil {
		return o, fmt.Errorf("reopen %s: %w", id, err)
	}
	o.Status = StatusReopened
	o.ReopenedAt = &now
	o.ReopenNote = note
	return o, nil
}

// Resolve marks an omission resolved (the re-requested item was addressed).
// Provided for completeness; m1 does not wire a CLI flag to it.
func (s *Store) Resolve(id string) (Omission, error) {
	o, err := s.Get(id)
	if err != nil {
		return o, err
	}
	const q = `UPDATE omissions SET status = ? WHERE id = ?`
	_, err = s.db.Exec(q, StatusResolved, id)
	if err != nil {
		return o, fmt.Errorf("resolve %s: %w", id, err)
	}
	o.Status = StatusResolved
	return o, nil
}

// ReopenLogName is the re-injection log file, kept next to the ledger DB.
// Every reopen appends one JSON line per re-request event; the agent skill
// reads it at the next session start and re-requests the listed items.
const ReopenLogName = "reopen.jsonl"

// AppendReinject appends o as one JSON line to <ledger-dir>/reopen.jsonl and
// returns the log path. The log is append-only history — one event per
// re-request with its note, not latest state; a consumer wanting current
// re-requests dedupes by id.
func (s *Store) AppendReinject(o Omission) (string, error) {
	line, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("encode reopen record: %w", err)
	}
	path := filepath.Join(filepath.Dir(s.path), ReopenLogName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", fmt.Errorf("open reopen log %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return "", fmt.Errorf("append reopen record: %w", err)
	}
	return path, nil
}

// Counts returns open and total counts (for the init/status summary).
func (s *Store) Counts() (open, total int, err error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM omissions`)
	if e := row.Scan(&total); e != nil {
		return 0, 0, fmt.Errorf("count total: %w", e)
	}
	row = s.db.QueryRow(`SELECT COUNT(*) FROM omissions WHERE status = ?`, StatusOpen)
	if e := row.Scan(&open); e != nil {
		return 0, 0, fmt.Errorf("count open: %w", e)
	}
	return open, total, nil
}

func scanOmission(rows *sql.Rows) (Omission, error) {
	var (
		o                Omission
		createdAt        string
		status           string
		reopenedAt       sql.NullString
	)
	err := rows.Scan(&o.ID, &o.SessionID, &o.File, &o.Item, &o.Reason, &o.Category,
		&createdAt, &status, &reopenedAt, &o.ReopenNote)
	if err != nil {
		return o, err
	}
	o.Status = status
	o.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		o.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	}
	if reopenedAt.Valid {
		t, e := time.Parse(time.RFC3339Nano, reopenedAt.String)
		if e != nil {
			t, e = time.Parse(time.RFC3339, reopenedAt.String)
		}
		if e == nil {
			o.ReopenedAt = &t
		}
	}
	return o, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOmissionRow(row rowScanner) (Omission, error) {
	var (
		o          Omission
		createdAt  string
		status     string
		reopenedAt sql.NullString
	)
	err := row.Scan(&o.ID, &o.SessionID, &o.File, &o.Item, &o.Reason, &o.Category,
		&createdAt, &status, &reopenedAt, &o.ReopenNote)
	if err != nil {
		return o, err
	}
	o.Status = status
	o.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		o.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	}
	if reopenedAt.Valid {
		t, e := time.Parse(time.RFC3339Nano, reopenedAt.String)
		if e != nil {
			t, e = time.Parse(time.RFC3339, reopenedAt.String)
		}
		if e == nil {
			o.ReopenedAt = &t
		}
	}
	return o, err
}
