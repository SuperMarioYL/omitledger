**English** | [简体中文](README.md)

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/hero-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/hero-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/hero-dark.svg">
  <img src="assets/presentation/hero-light.svg" width="1000" alt="Record skipped items, reasons, files and sessions in a local ledger so reviewers can decide what needs another pass.">
</picture>

**Record skipped items, reasons, files and sessions in a local ledger so reviewers can decide what needs another pass.**

`v0.1.0` · `Go 1.24+` · [MIT](LICENSE)

[Website](https://omitledger.lei6393.com) · [Demo record](docs/demo-results.json)

## Why use it

During review, completed changes are visible while deliberately omitted work can disappear. OmitLedger lets a caller record the item and stated reason at the time of omission, then use list and reopen to make follow-up review traceable.

## Architecture

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/architecture-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/architecture-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/architecture-dark.svg">
  <img src="assets/presentation/architecture-light.svg" width="1000" alt="The Cobra CLI maps add/list/reopen to ledger.Store. SQLite stores an Omission’s category, target file, session, status and review note. reopen updates ledger state; it does not execute omitted work. The current mcp and report command entry points remain placeholders.">
</picture>

The Cobra CLI maps add/list/reopen to ledger.Store. SQLite stores an Omission’s category, target file, session, status and review note. reopen updates ledger state; it does not execute omitted work. The current mcp and report command entry points remain placeholders.

Source entry points: [cmd/omitledger/root.go](cmd/omitledger/root.go) · [cmd/omitledger/add.go](cmd/omitledger/add.go) · [cmd/omitledger/list.go](cmd/omitledger/list.go) · [cmd/omitledger/reopen.go](cmd/omitledger/reopen.go) · [cmd/omitledger/mcp.go](cmd/omitledger/mcp.go) · [cmd/omitledger/report.go](cmd/omitledger/report.go) · [internal/ledger/store.go](internal/ledger/store.go)

## Install

Requires Go 1.24+ and uses a pure-Go SQLite driver. The script operates only on a temporary database and does not change the referenced retry.go.

```bash
git clone https://github.com/SuperMarioYL/omitledger.git
cd omitledger
go build -o bin/omitledger ./cmd/omitledger
```

## Quickstart

A complete shell script creates a temporary ledger, records one synthetic omission, lists it, reopens it and lists it again. IDs are generated on each run.

```bash
bash examples/presentation-demo.sh
```

Complete inputs and execution steps are included in the commands above and the [demo record](docs/demo-results.json).

## Usage

```bash
./bin/omitledger init
./bin/omitledger add --item "API example" --reason "deferred for review" --file README.md --category doc
./bin/omitledger list --status open
./bin/omitledger reopen 1 --note "include a complete request"
```
`--item` and `--reason` are required. Conventional categories are test/file/section/refactor/doc/log; repository-wide items can use `--file "*"`. reopen accepts a stable ID. Do not assume a filtered list’s row number matches its position in the complete ledger.

## Recorded demo

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/process-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/process-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/process-dark.svg">
  <img src="assets/presentation/process-light.svg" width="1000" alt="A complete shell script creates a temporary ledger, records one synthetic omission, lists it, reopens it and lists it again. IDs are generated on each run.">
</picture>

### Record and reopen

The same item moves from open to reopened with a note requesting retry-branch coverage.

```text
$ bash examples/presentation-demo.sh
recorded omission 000T0ZPHK7EB5WN6SCHAH9546H
  item:     retry-path test
  file:     retry.go
  category: test
  reason:   deferred during initial implementation
  session:  presentation-demo
#  ID                          STATUS  CATEGORY  FILE      ITEM             REASON
1  000T0ZPHK7EB5WN6SCHAH9546H  open    test      retry.go  retry-path test  deferred during initial implementation

1 omission(s)
reopened 000T0ZPHK7EB5WN6SCHAH9546H: retry-path test
  note: cover the retry branch
#  ID                          STATUS    CATEGORY  FILE      ITEM             REASON
1  000T0ZPHK7EB5WN6SCHAH9546H  reopened  test      retry.go  retry-path test  deferred during initial implementation

1 omission(s)
```

## Capabilities and integration

<picture>
  <source media="(max-width: 640px) and (prefers-color-scheme: dark)" srcset="assets/presentation/integrations-mobile-dark.svg">
  <source media="(max-width: 640px)" srcset="assets/presentation/integrations-mobile-light.svg">
  <source media="(prefers-color-scheme: dark)" srcset="assets/presentation/integrations-dark.svg">
  <img src="assets/presentation/integrations-light.svg" width="1000" alt="Workflows can explicitly call the CLI or use the supplied skill snippet as guidance. Recording is cooperative: it does not infer omissions from code or decide whether a reason is justified.">
</picture>

Workflows can explicitly call the CLI or use the supplied skill snippet as guidance. Recording is cooperative: it does not infer omissions from code or decide whether a reason is justified.



## Configuration

Ledger location precedence is `--store` > `OMITLEDGER_STORE` > `.omitledger/ledger.db`. Session precedence is `OMITLEDGER_SESSION`, CLAUDE_SESSION_ID, CURSOR_SESSION_ID, then local. SQLite uses WAL and a 5000ms busy timeout.

## Roadmap and scope

Current usable entry points are init/add/list/reopen. MCP, TUI, session reports, automatic reinjection and team aggregation still require wiring or implementation; internal modules alone do not establish CLI support.

- Records are caller statements: missing entries do not establish absence of omissions, and reopened does not mean completed.
- The current mcp/report placeholders can exit successfully without providing the described full functionality.

![Terminal recording](assets/demo.gif) · [Recording script](docs/demo.tape)

## License

[MIT](LICENSE)
