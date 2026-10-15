package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/SuperMarioYL/omitledger/internal/ledger"
)

// repoRoot is the repo root relative to this package's directory (tests run
// from cmd/omitledger), both locally and on a CI checkout.
const repoRoot = "../.."

// TestVersionLockstep pins every version surface together: the CLI const,
// the VERSION file, web/site.json meta.content_version, both README hero tag
// lines, and the CHANGELOG sections. Bump them in lockstep or this test fails.
func TestVersionLockstep(t *testing.T) {
	versionFile, err := os.ReadFile(repoRoot + "/VERSION")
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	if got := strings.TrimSpace(string(versionFile)); got != version {
		t.Fatalf("VERSION file = %q, CLI const version = %q — bump in lockstep", got, version)
	}

	siteRaw, err := os.ReadFile(repoRoot + "/web/site.json")
	if err != nil {
		t.Fatalf("read web/site.json: %v", err)
	}
	var site struct {
		Meta struct {
			ContentVersion string `json:"content_version"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(siteRaw, &site); err != nil {
		t.Fatalf("parse web/site.json: %v", err)
	}
	if site.Meta.ContentVersion != version {
		t.Fatalf("web/site.json meta.content_version = %q, want %q", site.Meta.ContentVersion, version)
	}

	// Both README hero tag lines carry the version too.
	for _, readme := range []string{"README.md", "README.en.md"} {
		raw, err := os.ReadFile(repoRoot + "/" + readme)
		if err != nil {
			t.Fatalf("read %s: %v", readme, err)
		}
		if want := "`v" + version + "` · `Go 1.24+`"; !strings.Contains(string(raw), want) {
			t.Fatalf("%s hero line is missing %q — bump every version surface in lockstep", readme, want)
		}
	}

	changelog, err := os.ReadFile(repoRoot + "/CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	for _, want := range []string{"## [0.1.0]", "## [0.2.0]", "## [" + version + "]"} {
		if !strings.Contains(string(changelog), want) {
			t.Fatalf("CHANGELOG.md is missing the %q section", want)
		}
	}
}

// TestVersionFlag exercises the real surface: `omitledger --version` must
// print the version and exit 0 (cobra wires the flag from rootCmd.Version
// at Execute time).
func TestVersionFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--version"})
	defer rootCmd.SetArgs(nil)
	out := captureStdout(t, func() {
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("execute --version: %v", err)
		}
	})
	if want := "omitledger version " + version; !strings.Contains(out, want) {
		t.Fatalf("--version output = %q, want %q", out, want)
	}
}

// captureStderr runs fn with os.Stderr redirected to a pipe and returns what
// it printed (the mirror of captureStdout; run() is the CLI's single error
// print point, so its stderr output is the contract under test).
func captureStderr(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w
	runErr := fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stderr: %v", err)
	}
	return buf.String(), runErr
}

// m10: the root command runs with SilenceErrors, so before the fix an unknown
// command exited 1 with ZERO output. Every cobra-raised error must surface in
// the documented `omitledger: ` stderr shape.
func TestUnknownCommandPrintsError(t *testing.T) {
	stderr, err := captureStderr(t, func() error { return run([]string{"lst"}) })
	if err == nil {
		t.Fatal("unknown command must return an error")
	}
	if !strings.Contains(stderr, "omitledger:") || !strings.Contains(stderr, "lst") {
		t.Fatalf("unknown command stderr = %q, want the omitledger-prefixed error naming the command", stderr)
	}
}

func TestUnknownFlagPrintsError(t *testing.T) {
	stderr, err := captureStderr(t, func() error { return run([]string{"list", "--stat"}) })
	if err == nil {
		t.Fatal("unknown flag must return an error")
	}
	if !strings.Contains(stderr, "omitledger:") || !strings.Contains(stderr, "--stat") {
		t.Fatalf("unknown flag stderr = %q, want the omitledger-prefixed error naming the flag", stderr)
	}
}

// RunE errors travel through the same single print point — exactly once, not
// once per layer.
func TestRunEErrorPrintsOnce(t *testing.T) {
	seedStore(t, ledger.Omission{Item: "only record", Reason: "r"})
	stderr, err := captureStderr(t, func() error { return run([]string{"reopen", "12345"}) })
	if err == nil {
		t.Fatal("out-of-range reopen must return an error")
	}
	if n := strings.Count(stderr, "omitledger:"); n != 1 {
		t.Fatalf("stderr printed the omitledger prefix %d times, want exactly once:\n%s", n, stderr)
	}
	if !strings.Contains(stderr, "out of range") {
		t.Fatalf("stderr = %q, want the out-of-range cause", stderr)
	}
}
