#!/usr/bin/env bash
set -euo pipefail
demo_dir="$(mktemp -d)"
trap 'rm -rf "$demo_dir"' EXIT
export OMITLEDGER_STORE="$demo_dir/ledger.db"
export OMITLEDGER_SESSION="presentation-demo"
./bin/omitledger add --item "retry-path test" --reason "deferred during initial implementation" --file retry.go --category test
./bin/omitledger list
./bin/omitledger reopen 1 --note "cover the retry branch"
./bin/omitledger list --status reopened
