# Changelog

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
- `omitledger report` — post-session markdown report + `.omitledger/reopen.jsonl` re-injection + JSON export for a CI merge-gate (m3).
