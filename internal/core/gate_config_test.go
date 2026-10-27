package core

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadGateConfigUsesDefaultsWhenMissing(t *testing.T) {
	config, err := LoadGateConfig(filepath.Join(t.TempDir(), "gates.yml"))
	if err != nil {
		t.Fatalf("LoadGateConfig returned error: %v", err)
	}

	if !reflect.DeepEqual(config, DefaultGateConfig()) {
		t.Fatalf("config = %#v, want defaults %#v", config, DefaultGateConfig())
	}
}

func TestLoadGateConfigParsesDevMainConfig(t *testing.T) {
	projectDir := t.TempDir()
	writeGateConfig(t, projectDir, `gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: false
      warn_on_manual_changes: true
    manual_changes:
      list: false
      warn_on_high_risk_paths: true
    high_risk_paths:
      - internal/runtime/**
      - docs/reference/**
`)

	config, err := LoadGateConfig(UserPath(projectDir, "config", "gates.yml"))
	if err != nil {
		t.Fatalf("LoadGateConfig returned error: %v", err)
	}

	if config.DevMain.ValidationCommand != "make verify" {
		t.Fatalf("ValidationCommand = %q, want make verify", config.DevMain.ValidationCommand)
	}

	if config.DevMain.Traceability.FailOnMissingAgentEvidence {
		t.Fatal("FailOnMissingAgentEvidence = true, want false")
	}

	if !config.DevMain.Traceability.WarnOnManualChanges {
		t.Fatal("WarnOnManualChanges = false, want true")
	}

	if config.DevMain.ManualChanges.List {
		t.Fatal("ManualChanges.List = true, want false")
	}

	if !config.DevMain.ManualChanges.WarnOnHighRiskPaths {
		t.Fatal("WarnOnHighRiskPaths = false, want true")
	}

	if len(config.DevMain.HighRiskPaths) != 2 {
		t.Fatalf("HighRiskPaths = %#v, want two entries", config.DevMain.HighRiskPaths)
	}
}

func writeGateConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	path := UserPath(projectDir, "config", "gates.yml")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create gate config dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write gate config: %v", err)
	}
}
