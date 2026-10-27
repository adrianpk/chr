package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Fail closed on ambiguous private setup without exposing its original contents.
func TestHarnessSelectionProfiles(t *testing.T) {
	for _, test := range []struct {
		name, body string
		invalid    bool
	}{
		{"codex", "profiles:\n  coding:\n    harness: codex\n    command: codex\n", false},
		{"pi", "profiles:\n  writing:\n    harness: pi\n    command: pi\n    model_provider: example\n    model: exact-model\n", false},
		{"duplicate name", "profiles:\n  coding:\n    harness: codex\n    command: codex\n  coding:\n", true},
		{"duplicate field", "profiles:\n  coding:\n    harness: codex\n    command: one\n    command: two\n", true},
		{"missing executable", "profiles:\n  coding:\n    harness: codex\n", true},
		{"unknown harness", "profiles:\n  coding:\n    harness: other\n    command: fixture\n", true},
		{"missing Pi model", "profiles:\n  writing:\n    harness: pi\n    command: pi\n", true},
		{"inline secret", "profiles:\n  coding:\n    harness: codex\n    command: codex\n    api_key: fixture-private-value\n", true},
		{"alias", "profiles:\n  coding: *private\n", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "harnesses.yml"), []byte(test.body), 0o600); err != nil {
				t.Fatal(err)
			}

			profiles, err := LoadProfiles(dir)
			if (err != nil) != test.invalid || (err != nil && strings.Contains(err.Error(), "fixture-private-value")) {
				t.Fatalf("profile error = %v", err)
			}

			if !test.invalid && len(profiles) != 1 {
				t.Fatalf("profile count = %d", len(profiles))
			}
		})
	}
}

// Private references are never projected, and credentials do not enter revisions.
func TestHarnessSelectionProfileIdentity(t *testing.T) {
	profile := Profile{Name: "coding", Harness: "codex", Command: "/private/driver", CredentialRef: "/private/auth", ModelConfigRef: "/private/models", RuntimeRef: "private-runtime", Model: "example-model"}
	profile.Revision = ProfileRevision(profile)

	data, err := json.Marshal(profile)
	if err != nil || strings.Contains(string(data), "/private/") || strings.Contains(string(data), "private-runtime") {
		t.Fatalf("private profile projection = %s, %v", data, err)
	}

	profile.Model = "another-model"
	if ProfileRevision(profile) == profile.Revision {
		t.Fatal("changed semantic model kept its revision")
	}
}

// Bound reads and permissions before parsing a private document.
func TestHarnessSelectionPrivateFile(t *testing.T) {
	dir := t.TempDir()
	if profiles, err := LoadProfiles(dir); err != nil || len(profiles) != 0 {
		t.Fatalf("absent private setup = %v", err)
	}

	path := filepath.Join(dir, "harnesses.yml")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", profileLimit+1)), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadProfiles(dir); err == nil {
		t.Fatal("oversized private setup accepted")
	}

	if err := os.WriteFile(path, []byte("profiles:\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadProfiles(dir); err == nil {
		t.Fatal("non-private setup accepted")
	}
}
