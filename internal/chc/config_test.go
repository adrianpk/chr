package chc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkspaceConfigUsesDefaultsWhenMissing(t *testing.T) {
	dir := t.TempDir()

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != DefaultTicketKeyPrefix {
		t.Fatalf("TicketKeyPrefix = %q, want %q", config.TicketKeyPrefix, DefaultTicketKeyPrefix)
	}

	assertDefaultRuntimeConfig(t, config.Runtime)
}

func TestLoadWorkspaceConfigParsesTicketAndRuntimeConfig(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `version: 1
workspace_root: .chc
tickets:
  key_prefix: PRJ
runtime:
  image: custom-agent:latest
  dockerfile: .chc/runtime/custom.Dockerfile
  workdir: /repo
  artifact_dir: .chc/runtime/custom-runs
  validation_command: go test ./...
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "PRJ" {
		t.Fatalf("TicketKeyPrefix = %q, want PRJ", config.TicketKeyPrefix)
	}

	want := RuntimeConfig{
		Image:             "custom-agent:latest",
		DockerfilePath:    ".chc/runtime/custom.Dockerfile",
		Workdir:           "/repo",
		ArtifactDir:       ".chc/runtime/custom-runs",
		ValidationCommand: "go test ./...",
	}
	if config.Runtime != want {
		t.Fatalf("Runtime = %#v, want %#v", config.Runtime, want)
	}
}

func TestLoadWorkspaceConfigKeepsRuntimeDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, `runtime:
  image: custom-agent:latest
`)

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.Runtime.Image != "custom-agent:latest" {
		t.Fatalf("Runtime.Image = %q, want custom-agent:latest", config.Runtime.Image)
	}

	if config.Runtime.DockerfilePath != DefaultRuntimeDockerfilePath {
		t.Fatalf("Runtime.DockerfilePath = %q, want %q", config.Runtime.DockerfilePath, DefaultRuntimeDockerfilePath)
	}
}

func TestLoadWorkspaceConfigSupportsLegacyTopLevelTicketKeyPrefix(t *testing.T) {
	dir := t.TempDir()
	writeWorkspaceConfig(t, dir, "key_prefix: LEGACY\n")

	config, err := LoadWorkspaceConfig(dir)
	if err != nil {
		t.Fatalf("LoadWorkspaceConfig returned error: %v", err)
	}

	if config.TicketKeyPrefix != "LEGACY" {
		t.Fatalf("TicketKeyPrefix = %q, want LEGACY", config.TicketKeyPrefix)
	}
}

func assertDefaultRuntimeConfig(t *testing.T, config RuntimeConfig) {
	t.Helper()

	if config != DefaultRuntimeConfig() {
		t.Fatalf("Runtime = %#v, want %#v", config, DefaultRuntimeConfig())
	}
}

func writeWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	path := filepath.Join(projectDir, DirectoryName, "config.yml")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}
