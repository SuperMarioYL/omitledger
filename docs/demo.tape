# OmitLedger m1 demo tape — the 10-minute happy path.
# Drives the real `omitledger` binary: init -> two deliberate omissions ->
# list -> reopen the first by line number -> list again (status flips to
# reopened). Rendered to assets/demo.gif by `vhs docs/demo.tape` (locally or
# via the `demo` GitHub Actions workflow).
Output assets/demo.gif
Require omitledger

Set FontSize 13
Set Width 1080
Set Height 560
Set Padding 20
Set Theme "Dracula"
Set TypingSpeed 6ms

# Fresh scratch dir so the demo starts from a clean ledger every run.
Type "rm -rf /tmp/ol-demo && mkdir -p /tmp/ol-demo && cd /tmp/ol-demo"
Sleep 200ms
Enter
Sleep 500ms

Type "omitledger init"
Sleep 200ms
Enter
Sleep 900ms

Type "omitledger add --item 'parser unit test' --reason 'trivial getter, low risk' --file parser.go --category test"
Sleep 200ms
Enter
Sleep 900ms

Type "omitledger add --item 'retry-path error log' --reason 'rare branch, deferred to follow-up' --file retry.go --category log"
Sleep 200ms
Enter
Sleep 900ms

Type "omitledger list"
Sleep 200ms
Enter
Sleep 1200ms

Type "omitledger reopen 1 --note 'needs the test'"
Sleep 200ms
Enter
Sleep 900ms

Type "omitledger list"
Sleep 200ms
Enter
Sleep 2500ms
