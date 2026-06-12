package core

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

func TestNewUserEnvironmentUsesXDGPaths(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		HomeDir:   filepath.Join(root, "home"),
		XDGConfig: filepath.Join(root, "xdg-config"),
		XDGData:   filepath.Join(root, "xdg-data"),
		XDGCache:  filepath.Join(root, "xdg-cache"),
		XDGState:  filepath.Join(root, "xdg-state"),
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "xdg-config", "chc"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "xdg-data", "chc"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "xdg-cache", "chc"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "xdg-state", "chc"))
}

func TestNewUserEnvironmentUsesHomeFallbacks(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")

	env, err := NewUserEnvironment(UserEnvironmentOptions{HomeDir: home})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(home, ".config", "chc"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(home, ".local", "share", "chc"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(home, ".cache", "chc"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(home, ".local", "state", "chc"))
}

func TestNewUserEnvironmentDevModeIgnoresXDGPaths(t *testing.T) {
	root := t.TempDir()

	env, err := NewUserEnvironment(UserEnvironmentOptions{
		DevMode:   true,
		DevRoot:   filepath.Join(root, "dev"),
		HomeDir:   filepath.Join(root, "home"),
		XDGConfig: filepath.Join(root, "xdg-config"),
		XDGData:   filepath.Join(root, "xdg-data"),
		XDGCache:  filepath.Join(root, "xdg-cache"),
		XDGState:  filepath.Join(root, "xdg-state"),
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "dev", "config", "chc"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "dev", "data", "chc"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "dev", "cache", "chc"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "dev", "state", "chc"))
}

func TestNewUserEnvironmentRejectsMissingDevRoot(t *testing.T) {
	_, err := NewUserEnvironment(UserEnvironmentOptions{DevMode: true})
	if err == nil {
		t.Fatal("NewUserEnvironment returned nil error")
	}
}

func TestNewUserEnvironmentRejectsMissingPath(t *testing.T) {
	_, err := NewUserEnvironment(UserEnvironmentOptions{})
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
