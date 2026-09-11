package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// repoRoot is the repo root relative to this package's directory (tests run
// from cmd/omitledger), both locally and on a CI checkout.
const repoRoot = "../.."

// TestVersionLockstep pins every version surface together: the CLI const,
// the VERSION file, web/site.json meta.content_version, and the CHANGELOG
// sections. Bump them in lockstep or this test fails.
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

	changelog, err := os.ReadFile(repoRoot + "/CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	for _, want := range []string{"## [0.1.0]", "## [" + version + "]"} {
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
