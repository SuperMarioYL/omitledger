# Changelog

## [0.3.0] — 2026-10-16
### Fixed — three verified defects from the v0.2.0 source
- The ledger no longer mis-orders records added within the same second: `Store.List` ordered by the RFC3339Nano TEXT column, whose variable-length fraction sorts byte-wise ("…0.91Z" sorts before "…0.9Z"), so insertion order — and the `#` line column and `reopen <line>` resolution derived from it — could silently target the wrong record. "Oldest first" now has one owner: rows are stable-sorted by parsed time with the ULID as tie-break. No schema change; existing ledgers keep working and now order correctly.
- The CLI no longer fails silently: the root command runs with `SilenceErrors`, so an unknown command (`omitledger lst`) or unknown flag exited 1 with zero output. Errors now print exactly once from a single choke point in `main`, in the documented `omitledger: <message>` shape, for cobra-raised and RunE errors alike. Exit codes are unchanged.
- A `?`, `#` or `%` in the `--store` / `OMITLEDGER_STORE` path no longer silently relocates the ledger: the raw path was spliced into a SQLite URI DSN, so SQLite parsed the tail as query/fragment and created the DB in a truncated file (repro: `…/weird?name/ledger.db` landed in `…/weird` with exit 0). The DSN is now built with `net/url` percent-encoding, which SQLite decodes back — every filesystem-legal path round-trips to the requested location.

### Added — m2: native MCP server
- `omitledger mcp` is real: a native MCP server over stdio exposing `record_omission` / `list_omissions` / `reopen_omission`, built on the already-pinned `modelcontextprotocol/go-sdk`, so agents call the ledger directly instead of shelling out. `reopen_omission` keeps CLI parity (ULID or `#` line ref, note, reopen.jsonl re-injection event). Session provenance resolves through `ledger.SessionID()` — one owner shared with the CLI.

### Changed
- The interactive TUI viewer is now the only stub surface; the MCP server it was reserved alongside has shipped.
- CLI error output has a single print point (`main`); the `die()` helper is gone. No exit-code changes.
## [0.2.0] — 2026-09-11
### Fixed — three verified defects from the v0.1.0 source
- `list` no longer corrupts CJK item/reason text: `trunc()` sliced bytes and split multi-byte UTF-8 runes mid-sequence, so any Chinese item or reason over 48 bytes (~17 CJK chars) rendered with U+FFFD replacement characters on the main list view. Truncation is now rune-based and the table output is always valid UTF-8.
- Filtered lists no longer renumber rows from 1: `list --status open` showed its rows as `#1, #2…` while `reopen <line>` resolves line numbers against the full ledger, so `reopen 1` after a filtered list silently re-opened the wrong (possibly already reopened) record with exit 0. Rows now carry their absolute ledger position in every view, so each displayed `#` is directly usable with `reopen`.
- An unknown `--status` value (typo, wrong case) now errors with exit 1 instead of printing `no <value> omissions recorded` — previously indistinguishable from an empty ledger. Status filtering has one definition (`ledger.MatchesStatus`), shared by the store and the CLI.
- The documented `~/.omitledger/ledger.db` global-ledger path (via `OMITLEDGER_STORE` or `--store`) now works: nothing expanded the tilde, so the ledger silently landed in a directory literally named `~` inside the working directory. `Store.Open` expands a leading `~`/`~/` to the home directory.

### Added — m3: report, re-injection, export
- `omitledger report` — post-session markdown summary: per-status counts, the re-requested items with their notes, and all omissions grouped by category; `--session <id>` scopes it, `--out <path>` writes a file (stdout by default).
- `reopen` re-injects into the next agent session: every re-request appends one JSON event to `.omitledger/reopen.jsonl` next to the ledger (append-only history); the agent skill snippet now reads it at session start and re-requests the listed items.
- `omitledger export json` — machine-readable ledger dump for CI merge-gates with the same `--status` / `--session` filters as `list`; `Omission` records gained snake_case JSON tags.
- `omitledger --version` — the CLI now reports its version (previously an error).

### Changed
- `mcp` (MCP server mode + TUI viewer, m2) remains the only stub command.
- Version surfaces are pinned in lockstep by `TestVersionLockstep`: VERSION file, CLI `--version`, `web/site.json` `meta.content_version`, and this changelog.

## [0.1.0] — 2026-08-16
### Added — m1: record and list
- `omitledger init` — create the local SQLite ledger (`.omitledger/ledger.db`).
- `omitledger add --item --reason --file --category` — record a deliberate omission at decision time with the agent-stated reason.
- `omitledger list [--status open|reopened|resolved|closed]` — table view of the ledger with a `#` line-number column.
- `omitledger reopen <id|line>` — re-request an omission by ULID or 1-based line number.
- ULID-based, time-ordered omission IDs (pure-Go, stdlib `crypto/rand` — single static binary, no cgo).
- Agent skill snippet (`skills/omit-ledger/SKILL.md`) that makes an agent call `omitledger add` when it deliberately skips something.
- Animated hero + architecture SVGs (dark/light), bilingual README (zh-primary + English), vhs demo tape + demo CI.

### Reserved (stubs present, not implemented in m1)
- `omitledger mcp` — native MCP server mode (`record_omission` / `list_omissions` / `reopen_omission` over stdio) + interactive TUI viewer (m2).
- ~~`omitledger report` — post-session markdown report + `.omitledger/reopen.jsonl` re-injection + JSON export for a CI merge-gate (m3).~~ — shipped in 0.2.0.

[0.3.0]: https://github.com/SuperMarioYL/omitledger/compare/v0.2.0...v0.3.0
