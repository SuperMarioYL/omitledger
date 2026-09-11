package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what
// it printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	return buf.String()
}

// seedStore points the CLI at a fresh temp ledger and returns the store path.
func seedStore(t *testing.T, records ...ledger.Omission) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.db")
	st, err := ledger.Open(path)
	if err != nil {
		t.Fatalf("open seed store: %v", err)
	}
	defer st.Close()
	for _, o := range records {
		if _, err := st.Add(o); err != nil {
			t.Fatalf("seed add %q: %v", o.Item, err)
		}
	}
	storeFlag = path
	t.Cleanup(func() { storeFlag = "" })
	return path
}

func TestExportJSONDecodesIntoOmissions(t *testing.T) {
	seedStore(t,
		ledger.Omission{Item: "parser unit test", Reason: "trivial", File: "parser.go", Category: "test", SessionID: "s1"},
		ledger.Omission{Item: "beta doc", Reason: "r2", SessionID: "s2"},
	)
	exportStatus, exportSession = "", ""
	out := captureStdout(t, func() {
		if err := exportCmd.RunE(exportCmd, []string{"json"}); err != nil {
			t.Errorf("export run: %v", err)
		}
	})
	var got []ledger.Omission
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("export output is not a JSON array: %v\n%s", err, out)
	}
	if len(got) != 2 {
		t.Fatalf("exported %d records, want 2", len(got))
	}
	if got[0].Item != "parser unit test" || got[0].File != "parser.go" || got[0].SessionID != "s1" {
		t.Fatalf("record 0 = %+v", got[0])
	}
}

func TestExportJSONEmptyLedgerRendersArray(t *testing.T) {
	seedStore(t)
	exportStatus, exportSession = "", ""
	out := captureStdout(t, func() {
		if err := exportCmd.RunE(exportCmd, []string{}); err != nil {
			t.Errorf("export run: %v", err)
		}
	})
	if out == "null\n" || out == "null" {
		t.Fatalf("empty ledger exported null, want []:\n%q", out)
	}
	var got []ledger.Omission
	if err := json.Unmarshal([]byte(out), &got); err != nil || len(got) != 0 {
		t.Fatalf("empty export = %q (err %v)", out, err)
	}
}

func TestExportJSONStatusAndSessionFilters(t *testing.T) {
	st := seedStore(t,
		ledger.Omission{Item: "a", Reason: "r", SessionID: "s1"},
		ledger.Omission{Item: "b", Reason: "r", SessionID: "s2"},
	)
	// reopen line 1 (record a) so --status open leaves only b.
	s, err := ledger.Open(st)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	id, err := s.ResolveRef("1")
	if err != nil {
		t.Fatalf("resolve line 1: %v", err)
	}
	if _, err := s.Reopen(id, ""); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	s.Close()

	exportStatus = "open"
	out := captureStdout(t, func() {
		if err := exportCmd.RunE(exportCmd, []string{"json"}); err != nil {
			t.Errorf("export run: %v", err)
		}
	})
	var got []ledger.Omission
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("status-filtered export not JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Item != "b" {
		t.Fatalf("status-filtered export = %+v", got)
	}

	exportStatus = ""
	exportSession = "s1"
	out = captureStdout(t, func() {
		if err := exportCmd.RunE(exportCmd, []string{"json"}); err != nil {
			t.Errorf("export run: %v", err)
		}
	})
	got = nil
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("session-filtered export not JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Item != "a" || got[0].Status != ledger.StatusReopened {
		t.Fatalf("session-filtered export = %+v", got)
	}
	exportSession = ""
}

func TestReportCommandSessionFilter(t *testing.T) {
	seedStore(t,
		ledger.Omission{Item: "in session", Reason: "r", Category: "test", SessionID: "s1"},
		ledger.Omission{Item: "other session", Reason: "r", Category: "doc", SessionID: "s2"},
	)
	reportSession = "s1"
	reportOut = ""
	defer func() { reportSession, reportOut = "", "" }()
	out := captureStdout(t, func() {
		if err := reportCmd.RunE(reportCmd, []string{}); err != nil {
			t.Errorf("report run: %v", err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("in session")) {
		t.Fatalf("session record missing from report:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("other session")) {
		t.Fatalf("other session leaked into --session s1 report:\n%s", out)
	}
}

func TestReopenCommandAppendsReinjectionLog(t *testing.T) {
	path := seedStore(t, ledger.Omission{Item: "parser unit test", Reason: "trivial", Category: "test"})
	reopenNote = "needs the test"
	defer func() { reopenNote = "" }()
	if err := reopenCmd.RunE(reopenCmd, []string{"1"}); err != nil {
		t.Fatalf("reopen run: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(path), ledger.ReopenLogName))
	if err != nil {
		t.Fatalf("reopen.jsonl not written next to the ledger: %v", err)
	}
	var rec ledger.Omission
	if err := json.Unmarshal(data[:len(data)-1], &rec); err != nil { // drop trailing \n
		t.Fatalf("reopen.jsonl line is not JSON: %v\n%s", err, data)
	}
	if rec.Item != "parser unit test" || rec.ReopenNote != "needs the test" || rec.Status != ledger.StatusReopened {
		t.Fatalf("reinject record = %+v", rec)
	}
}
