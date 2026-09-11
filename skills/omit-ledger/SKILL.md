---
name: omit-ledger
description: When you deliberately skip or omit something while coding (a test, a file, a section, an error log, a refactor, a doc), record it in the OmitLedger at the moment of the decision with your stated reason. Call omitledger add with what was skipped and why, so the omission is auditable and re-requestable instead of lost to chat scrollback.
---

# omit-ledger

When you deliberately omit something, record it immediately at the decision moment with your stated reason:

```bash
omitledger add --item "<what you skipped>" --reason "<why, at decision time>" --file <path-or-*> --category <test|file|section|refactor|doc|log>
```

- `omitledger add --item "parser unit test" --reason "trivial getter, low risk" --file parser.go --category test`
- `omitledger add --item "retry-path error log" --reason "rare branch, deferred to follow-up" --file retry.go --category log`

The user reviews with `omitledger list` and re-requests an item with `omitledger reopen <id-or-line>`.

At the START of every session, check whether the user re-requested anything:

```bash
[ -f .omitledger/reopen.jsonl ] && cat .omitledger/reopen.jsonl
```

Each line is one re-request event (JSON: id, item, reason, file, reopened_at, reopen_note). Treat every line as an instruction to address that omitted item in this session (dedupe by id — the log is append-only history, not latest state). Acknowledge each re-requested item before starting new work.

Capture the omission the instant you decide to skip it — the reason is the asset that distinguishes justified laziness from a cut corner.
