package core

import (
	"path/filepath"
	"testing"
)

func TestNewUserEnvironmentUsesExplicitPaths(t *testing.T) {
	env, err := NewUserEnvironment(UserEnvironmentOptions{
		ConfigDir: "/tmp/config/chr",
		DataDir:   "/tmp/data/chr",
		CacheDir:  "/tmp/cache/chr",
		StateDir:  "/tmp/state/chr",
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), "/tmp/config/chr")
	assertUserEnvironmentPath(t, env.DataDir(), "/tmp/data/chr")
	assertUserEnvironmentPath(t, env.CacheDir(), "/tmp/cache/chr")
	assertUserEnvironmentPath(t, env.StateDir(), "/tmp/state/chr")

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

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "config", "chr"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "data", "chr"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "cache", "chr"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "state", "chr"))

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

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "xdg-config", "chr"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "xdg-data", "chr"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "xdg-cache", "chr"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "xdg-state", "chr"))
}

func TestNewUserEnvironmentUsesHomeFallbacks(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")

	env, err := NewUserEnvironment(UserEnvironmentOptions{HomeDir: home})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(home, ".config", "chr"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(home, ".local", "share", "chr"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(home, ".cache", "chr"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(home, ".local", "state", "chr"))
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

	assertUserEnvironmentPath(t, env.ConfigDir(), filepath.Join(root, "dev", "config", "chr"))
	assertUserEnvironmentPath(t, env.DataDir(), filepath.Join(root, "dev", "data", "chr"))
	assertUserEnvironmentPath(t, env.CacheDir(), filepath.Join(root, "dev", "cache", "chr"))
	assertUserEnvironmentPath(t, env.StateDir(), filepath.Join(root, "dev", "state", "chr"))
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
