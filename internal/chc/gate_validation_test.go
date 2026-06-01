package chc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGateValidationRunnerPassesDefaultGoTest(t *testing.T) {
	projectDir := newGateValidationGoProject(t, "package gatevalidation\n")

	result, err := NewGateValidationRunner().Run(t.Context(), GateValidationRequest{
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.Passed {
		t.Fatalf("Passed = false, want true: %#v", result)
	}

	if result.Command != DefaultGateValidationCommand {
		t.Fatalf("Command = %q, want %q", result.Command, DefaultGateValidationCommand)
	}
}

func TestGateValidationRunnerReportsFailingGoTest(t *testing.T) {
	projectDir := newGateValidationGoProject(t, `package gatevalidation

import "testing"

func TestFailure(t *testing.T) {
	t.Fatal("expected failure")
}
`)

	result, err := NewGateValidationRunner().Run(t.Context(), GateValidationRequest{
		ProjectDir: projectDir,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Passed {
		t.Fatalf("Passed = true, want false: %#v", result)
	}

	if result.ExitCode == 0 {
		t.Fatalf("ExitCode = 0, want non-zero: %#v", result)
	}

	if !strings.Contains(result.Output, "expected failure") {
		t.Fatalf("Output = %q, want failing test output", result.Output)
	}
}

func TestGateValidationRunnerUsesConfiguredCommand(t *testing.T) {
	projectDir := newGateValidationGoProject(t, "package gatevalidation\n")

	result, err := NewGateValidationRunner().Run(t.Context(), GateValidationRequest{
		ProjectDir: projectDir,
		Command:    "go test ./...",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.Passed {
		t.Fatalf("Passed = false, want true: %#v", result)
	}
}

func TestGateValidationRunnerRejectsMissingProjectDir(t *testing.T) {
	_, err := NewGateValidationRunner().Run(t.Context(), GateValidationRequest{})
	if err == nil {
		t.Fatal("Run returned nil error")
	}
}

func newGateValidationGoProject(t *testing.T, testContent string) string {
	t.Helper()

	projectDir := t.TempDir()
	writeGateValidationFile(t, projectDir, "go.mod", "module gatevalidation.test\n\ngo 1.25\n")
	writeGateValidationFile(t, projectDir, "gatevalidation_test.go", testContent)

	return projectDir
}

func writeGateValidationFile(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create test file dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
