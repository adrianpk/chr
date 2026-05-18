package chc

import (
	"path/filepath"
	"testing"
)

func TestNewUserEnvironmentUsesExplicitPaths(t *testing.T) {
	env, err := NewUserEnvironment(UserEnvironmentOptions{
		ConfigDir: "/tmp/config/chc",
		DataDir:   "/tmp/data/chc",
		CacheDir:  "/tmp/cache/chc",
		StateDir:  "/tmp/state/chc",
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), "/tmp/config/chc")
	assertUserEnvironmentPath(t, env.DataDir(), "/tmp/data/chc")
	assertUserEnvironmentPath(t, env.CacheDir(), "/tmp/cache/chc")
	assertUserEnvironmentPath(t, env.StateDir(), "/tmp/state/chc")

	if env.DevMode() {
		t.Fatal("DevMode = true, want false")
	}
}

func TestNewUserEnvironmentUsesDevRoot(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode: true,
		DevRoot: root,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "config", "chc"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "data", "chc"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "cache", "chc"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "state", "chc"))

	if !env.DevMode() {
		t.Fatal("DevMode = false, want true")
	}
}

func TestNewUserEnvironmentRejectsMissingDevRoot(t *testing.T) {
	_, err := NewUserEnvironment(UserEnvironmentOptions{DevMode: true})
	if err == nil {
		t.Fatal("NewUserEnvironment returned nil error")
	}
}

func TestNewUserEnvironmentRejectsMissingPath(t *testing.T) {
	_, err := NewUserEnvironment(UserEnvironmentOptions{
		ConfigDir: "/tmp/config/chc",
		DataDir:   "/tmp/data/chc",
		CacheDir:  "/tmp/cache/chc",
	})
	if err == nil {
		t.Fatal("NewUserEnvironment returned nil error")
	}
}

func assertUserEnvironmentPath(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
