package webui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Only a private verifier is stored; replacing it requires an explicit setup action.
func TestLoginCredentials(t *testing.T) {
	credentials := loginTestCredentials(t)

	path := filepath.Join(t.TempDir(), "private", "web-login.json")
	if err := SaveLoginCredentials(path, credentials, false); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadLoginCredentials(path)
	if err != nil || !reflect.DeepEqual(loaded, &credentials) {
		t.Fatal("credential round trip failed", err)
	}

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatal("credential file is not private", err)
	}

	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), loginTestPassword) {
		t.Fatal("password was exposed in storage", err)
	}

	if err := SaveLoginCredentials(path, credentials, false); err == nil {
		t.Fatal("existing credentials were replaced without authority")
	}

	if err := SaveLoginCredentials(path, credentials, true); err != nil {
		t.Fatal(err)
	}
}

// Setup rejects unusable usernames and password lengths before storing anything.
func TestLoginCredentialInputs(t *testing.T) {
	for _, test := range []struct{ name, username, password string }{
		{"short password", "operator", "short"},
		{"oversized password", "operator", strings.Repeat("x", 1025)},
		{"empty account", "", loginTestPassword},
		{"surrounding whitespace", " operator ", loginTestPassword},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := CreateLoginCredentials(test.username, test.password); err == nil {
				t.Fatal("invalid account accepted")
			}
		})
	}
}

// A bad existing configuration fails closed instead of falling back to LAN access.
func TestLoginFileGuards(t *testing.T) {
	for _, test := range []struct {
		name, content string
		mode          os.FileMode
	}{
		{"public", "{}", 0o644},
		{"malformed", "{", 0o600},
		{"unknown version", `{"version":2}`, 0o600},
		{"trailing", "{} {}", 0o600},
		{"oversized", strings.Repeat("x", 4097), 0o600},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "web-login.json")
			if err := os.WriteFile(path, []byte(test.content), test.mode); err != nil {
				t.Fatal(err)
			}

			if _, err := LoadLoginCredentials(path); err == nil {
				t.Fatal("unsafe configuration accepted")
			}
		})
	}

	t.Run("symlink", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "web-login.json")
		if err := os.Symlink("missing-file", path); err != nil {
			t.Fatal(err)
		}

		if _, err := LoadLoginCredentials(path); err == nil {
			t.Fatal("linked configuration accepted")
		}
	})
}
