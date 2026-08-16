<div align="right"><sub><b>English</b>&nbsp;&nbsp;⇄&nbsp;&nbsp;<a href="./README.md">简体中文</a></sub></div>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./assets/hero-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./assets/hero-light.svg">
  <img src="./assets/hero-light.svg" width="880" alt="OmitLedger — record what your agent deliberately skipped, and why">
</picture>

<p align="center"><sub>OmitLedger records each item your coding agent deliberately skipped, with the reason — so justified laziness is distinguishable from a cut corner.</sub></p>

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/SuperMarioYL/omitledger?color=0071E3&label=license" alt="license"></a>
  <a href="https://github.com/SuperMarioYL/omitledger/releases"><img src="https://img.shields.io/github/v/release/SuperMarioYL/omitledger?color=5E5CE6&label=release" alt="latest release"></a>
  <img src="https://img.shields.io/github/actions/workflow/status/SuperMarioYL/omitledger/ci.yml?branch=main&label=ci" alt="CI">
  <img src="https://img.shields.io/badge/Go-1.24-0071E3?logo=go&logoColor=white" alt="Go 1.24">
</p>

**When your agent learns to be lazy, every item it skips should leave a record.**

<h2><img src="https://api.iconify.design/tabler:topology-star-3.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Architecture</h2>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./assets/atlas-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset="./assets/atlas-light.svg">
  <img src="./assets/atlas-light.svg" width="880" alt="Architecture: Coding Agent → omitledger CLI → SQLite ledger, with reopen re-injecting into the next session">
</picture>

One binary, one local DB file. At the moment your coding agent decides to skip something, it calls `omitledger add` to record what it omitted, its stated reason, and the target file; `omitledger list` lays the whole ledger open; `omitledger reopen` flags the one suspicious omission as re-requested. v0.1 has no network and no server process — everything stays in your repo's `.omitledger/ledger.db`.

## Table of contents

- [Why this exists](#why-this-exists)
- [Install](#install)
- [Quickstart](#quickstart)
- [Usage](#usage)
- [Demo](#demo)
- [Configuration](#configuration)
- [Paid](#paid)
- [Roadmap](#roadmap)
- [License](#license)

## Why this exists

In 2026, making your agent "deliberately do less" is mainstream — [ponytail](https://github.com/DietrichGebert/ponytail) (100k+ stars) makes an agent think like the laziest senior dev in the room, and [taste-skill](https://github.com/Leonxlnx/taste-skill) (76k+ stars) actively suppresses generic slop. Both produce large sets of "things that were never produced," yet neither records a single one of those omissions. So when a "lazy" agent hands you less than you expected, you can't tell whether it prudently minimized or silently cut a corner — you can only diff the output against the mental model of completeness in your own head.

OmitLedger captures each deliberately-skipped item at the moment of the decision, together with the agent's own stated reason and a one-key re-request button. This omission asset is a new primitive: distinct from any minimization engine that makes an agent lazy, it turns omissions into a queryable, re-requestable, auditable record — so justified laziness and a cut corner are distinguishable at a glance.

<h2><img src="https://api.iconify.design/tabler:rocket.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Install</h2>

Requires Go 1.24+. A single binary with a pure-Go SQLite driver (no cgo → one static binary):

```bash
go install github.com/SuperMarioYL/omitledger@latest
```

<h2><img src="https://api.iconify.design/tabler:rocket.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Quickstart</h2>

Cold clone to first visible result, three steps:

```bash
omitledger init                                          # creates .omitledger/ledger.db in the repo root
omitledger add --item "parser unit test" \
  --reason "trivial getter, low risk" --file parser.go --category test
omitledger list                                      # see the omission you just recorded
```

<details><summary>Sample output</summary>

```
$ omitledger init
omitledger ledger ready: /your-repo/.omitledger/ledger.db
  0 omission(s) recorded, 0 open
Next: install the agent skill ...
  mkdir -p ~/.claude/skills/omit-ledger && cp skills/omit-ledger/SKILL.md ~/.claude/skills/omit-ledger/SKILL.md

$ omitledger add --item "parser unit test" --reason "trivial getter, low risk" --file parser.go --category test
recorded omission 000T02CDSCFS5XYTBBFE9DSQNT
  item:     parser unit test
  file:     parser.go
  category: test
  reason:   trivial getter, low risk
  session:  local

$ omitledger list
#  ID                          STATUS  CATEGORY  FILE       ITEM                REASON
1  000T02CDSCFS5XYTBBFE9DSQNT  open    test      parser.go  parser unit test    trivial getter, low risk
```
</details>

<h2><img src="https://api.iconify.design/tabler:terminal-2.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Usage</h2>

Let the agent record omissions itself — drop `skills/omit-ledger/SKILL.md` into the agent's skill directory and it will call `omitledger add` whenever it deliberately skips something:

```bash
mkdir -p ~/.claude/skills/omit-ledger && cp skills/omit-ledger/SKILL.md ~/.claude/skills/omit-ledger/SKILL.md
```

The five most common workflows:

```bash
# 1) Record an omission (the agent calls this at the decision moment)
omitledger add --item "retry-path error log" --reason "rare branch, deferred to follow-up" --file retry.go --category log

# 2) See the whole ledger
omitledger list

# 3) Only what's still open (awaiting your review)
omitledger list --status open

# 4) Re-request one — by # line number, or by ULID
omitledger reopen 1 --note "needs the test"

# 5) m2/m3 placeholders (not implemented in m1)
omitledger mcp      # MCP server mode + TUI viewer (m2)
omitledger report   # post-session report + re-injection + JSON export (m3)
```

`--category` is one of `test | file | section | refactor | doc | log`. `--file` is the target path; use `*` for a repo-wide omission.

<h2><img src="https://api.iconify.design/tabler:photo.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Demo</h2>

`init` → the agent deliberately skips two items → `list` → `reopen 1` → `list` (status flips to reopened):

![demo](assets/demo.gif)

> The recording script is [`docs/demo.tape`](./docs/demo.tape) ([vhs](https://github.com/charmbracelet/vhs)); `.github/workflows/demo.yml` re-renders it to `assets/demo.gif` on demand.

<h2><img src="https://api.iconify.design/tabler:adjustments.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Configuration</h2>

OmitLedger has no config file — behavior is controlled by flags and environment variables:

| Key | Type | Default | Meaning |
|---|---|---|---|
| `--store` | path | `.omitledger/ledger.db` | ledger DB location; the `OMITLEDGER_STORE` env var does the same — use `~/.omitledger/ledger.db` for a global ledger |
| `OMITLEDGER_SESSION` | string | `local` | agent run id stamped on each omission; falls back to `CLAUDE_SESSION_ID` / `CURSOR_SESSION_ID` when unset |
| `.omitledger/ledger.db` | file | (repo-local) | the pure-Go SQLite ledger; WAL + `busy_timeout=5000` |

<h2><img src="https://api.iconify.design/tabler:credit-card.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Paid</h2>

The CLI is free OSS forever. The **team tier** for 5–30-dev engineering teams (regulated fintech / health / security teams especially, where a skipped test can break prod) is the paid layer in v0.2:

- **aggregates each member's local ledger** into a team view;
- **surfaces unreviewed omissions per PR**;
- emits a **merge-block webhook** (CI gate) when unreviewed omissions exceed a threshold — the auditable omission asset is what a team pays for; the CLI stays free.

Priced at **$8 / dev-seat / month**, team floor **$49 / mo** for 10 seats then $8 / seat beyond. It lands between a CI add-on and CodeRabbit-class review (~$15/seat), justified by the narrower omission-audit surface. v0.1 ships **no paid features** — this is a roadmap hook, not a paywall. Registering Stripe Trusted Publisher and the Fly.io hosted aggregation service is a one-time v0.2 step.

<h2><img src="https://api.iconify.design/tabler:map-2.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> Roadmap</h2>

- [x] **m1 — record and list**: `init` / `add` / `list` / `reopen` CLI + pure-Go SQLite ledger + agent skill snippet (this release)
- [ ] **m2 — MCP and TUI**: `omitledger mcp` serves native MCP tools (`record_omission` / `list_omissions` / `reopen_omission`) over stdio, and `omitledger tui` is an interactive viewer with cursor-selectable reopen
- [ ] **m3 — report and re-inject**: `omitledger report` emits a post-session markdown summary grouped by category; `reopen` writes `.omitledger/reopen.jsonl` to re-inject into the next session; `omitledger export json` feeds a CI merge-gate
- [ ] **team tier (commercial, v0.2)**: aggregate member ledgers, surface unreviewed omissions per PR, merge-block webhook

<h2><img src="https://api.iconify.design/tabler:license.svg?color=%230071E3&width=24" height="22" align="absmiddle" alt=""> License</h2>

MIT — see [LICENSE](./LICENSE). Issues and PRs welcome.

<p align="center"><sub><a href="./LICENSE">MIT</a> © 2026 SuperMarioYL</sub></p>
