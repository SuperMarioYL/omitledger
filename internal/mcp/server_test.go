package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// connect starts the MCP server over an in-memory transport and returns a
// connected client session. The server lifecycle is tied to the test through
// the deferred cleanup: the client session closes first, then the server Run
// result is checked.
func connect(t *testing.T, st *ledger.Store) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	serverT, clientT := mcp.NewInMemoryTransports()
	srv := New(st, "test")
	serverDone := make(chan error, 1)
	go func() { serverDone <- srv.server.Run(ctx, serverT) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "omitledger-test-client", Version: "0"}, nil)
	conn, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() {
		conn.Close()
		select {
		case err := <-serverDone:
			if err != nil {
				t.Errorf("server run: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("server run did not return after client disconnect")
		}
	})
	return conn
}

// call invokes a tool and returns its single text content.
func call(t *testing.T, conn *mcp.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	res, err := conn.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	var text string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			text += tc.Text
		}
	}
	return text, res.IsError
}

// m12: `omitledger mcp` must serve the ledger over the MCP stdio transport —
// pinned here with a full offline client round trip on the in-memory transport.
func TestMCPServerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	st, err := ledger.Open(filepath.Join(dir, "ledger.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()
	t.Setenv("OMITLEDGER_SESSION", "mcp-test-session")

	conn := connect(t, st)
	ctx := context.Background()

	// tools/list advertises exactly the three ledger tools.
	listed, err := conn.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	got := map[string]bool{}
	for _, tl := range listed.Tools {
		got[tl.Name] = true
	}
	for _, want := range []string{"record_omission", "list_omissions", "reopen_omission"} {
		if !got[want] {
			t.Fatalf("tools/list = %v, want it to advertise %q", got, want)
		}
	}
	if len(got) != 3 {
		t.Fatalf("tools/list advertised %d tools (%v), want exactly 3", len(got), got)
	}

	// record_omission stores a record with env-derived session provenance.
	text, isErr := call(t, conn, "record_omission", map[string]any{
		"item": "parser unit test", "reason": "trivial getter, low risk", "file": "parser.go", "category": "test",
	})
	if isErr {
		t.Fatalf("record_omission returned an error: %s", text)
	}
	if !strings.Contains(text, "recorded omission ") || !strings.Contains(text, "parser unit test") {
		t.Fatalf("record_omission text = %q", text)
	}

	// reopen_omission accepts a 1-based line ref and re-injects.
	text, isErr = call(t, conn, "reopen_omission", map[string]any{"ref": "1", "note": "needs the test"})
	if isErr {
		t.Fatalf("reopen_omission returned an error: %s", text)
	}
	if !strings.Contains(text, "reopened ") || !strings.Contains(text, "needs the test") {
		t.Fatalf("reopen_omission text = %q", text)
	}
	data, err := st.List(ledger.ListFilter{})
	if err != nil {
		t.Fatalf("list store: %v", err)
	}
	if data[0].Status != ledger.StatusReopened || data[0].SessionID != "mcp-test-session" {
		t.Fatalf("stored record = %+v, want reopened with the env session", data[0])
	}

	// list_omissions returns a JSON array honoring the status filter.
	text, isErr = call(t, conn, "list_omissions", map[string]any{"status": "reopened"})
	if isErr {
		t.Fatalf("list_omissions returned an error: %s", text)
	}
	var records []ledger.Omission
	if err := json.Unmarshal([]byte(text), &records); err != nil {
		t.Fatalf("list_omissions output is not a JSON array: %v\n%s", err, text)
	}
	if len(records) != 1 || records[0].Item != "parser unit test" || records[0].ReopenNote != "needs the test" {
		t.Fatalf("list_omissions = %+v", records)
	}

	// reopen.jsonl gained one re-injection event, matching the CLI path.
	logPath := filepath.Join(filepath.Dir(st.Path()), ledger.ReopenLogName)
	events, err := readJSONLLines(logPath)
	if err != nil {
		t.Fatalf("read reopen log: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("reopen.jsonl has %d event(s), want 1", len(events))
	}
	var event ledger.Omission
	if err := json.Unmarshal([]byte(events[0]), &event); err != nil {
		t.Fatalf("reopen.jsonl line is not JSON: %v", err)
	}
	if event.Status != ledger.StatusReopened || event.ReopenNote != "needs the test" {
		t.Fatalf("reopen event = %+v", event)
	}

	// Tool failures surface as error results, not transport errors: an
	// out-of-range line ref names the range in the error text.
	text, isErr = call(t, conn, "reopen_omission", map[string]any{"ref": "7"})
	if !isErr || !strings.Contains(text, "out of range") {
		t.Fatalf("out-of-range reopen: isErr=%v text=%q, want an error result naming the range", isErr, text)
	}

	// Schema-level failure: reason is required in the tool's inferred input
	// schema, so the client gets a params validation error.
	_, err = conn.CallTool(ctx, &mcp.CallToolParams{
		Name: "record_omission", Arguments: map[string]any{"item": "no reason given"},
	})
	if err == nil || !strings.Contains(err.Error(), "reason") {
		t.Fatalf("record without reason: err=%v, want schema rejection naming the field", err)
	}

	// Store-level failure: a whitespace reason passes the schema but fails
	// ledger validation, surfacing as an error tool result.
	text, isErr = call(t, conn, "record_omission", map[string]any{"item": "x", "reason": "   "})
	if !isErr || !strings.Contains(text, "reason") {
		t.Fatalf("record with blank reason: isErr=%v text=%q, want the validation error", isErr, text)
	}
}

func readJSONLLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}
