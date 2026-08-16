#!/usr/bin/env bash
# Example: the shape of an agent session that deliberately skips two items and
# records each in the OmitLedger at the moment of the decision, then reviews
# and re-requests the suspicious one. Run from a repo root.
set -euo pipefail

omitledger init
omitledger add --item "parser unit test" --reason "trivial getter, low risk" --file parser.go --category test
omitledger add --item "retry-path error log" --reason "rare branch, deferred to follow-up" --file retry.go --category log
omitledger list
omitledger reopen 1 --note "needs the test"
omitledger list
