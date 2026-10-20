package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Use an owned executable to cover real stdin, artifacts, failure and byte limits.
func TestHarnessCodexGeneration(t *testing.T) {
	for _, test := range []struct {
		name      string
		body      string
		limit     int
		want      string
		wantError bool
	}{
		{"output file", `printf '{"ok":true}' > "$output"`, 64, `{"ok":true}`, false},
		{"stdout fallback", `printf '{"ok":true}'`, 64, `{"ok":true}`, false},
		{"oversized file", `printf 123456789 > "$output"`, 8, "", true},
		{"oversized fallback", `printf 123456789`, 8, "", true},
		{"nonzero exit", `printf failed >&2; exit 7`, 64, "", true},
		{"early exit", `exit 0`, 64, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := fixtureRequest(t, test.body)
			request.OutputLimit = test.limit

			result, err := (Runner{}).RunAgent(t.Context(), request)
			if (err != nil) != test.wantError || string(result.Output) != test.want {
				t.Fatalf("output = %q, error = %v", result.Output, err)
			}

			if result.Provider != Identity || result.Usage.Input != nil || result.Usage.Output != nil {
				t.Fatalf("metadata = %#v", result)
			}
		})
	}
}

// A reused request cannot accept an earlier output file after an empty call.
func TestHarnessCodexStaleOutput(t *testing.T) {
	request := fixtureRequest(t, `exit 0`)
	if err := os.MkdirAll(request.RunDir, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(request.RunDir, request.OutputName), []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if result, err := (Runner{}).RunAgent(t.Context(), request); err == nil || len(result.Output) != 0 {
		t.Fatalf("stale result = %q, %v", result.Output, err)
	}
}

// Preserve read-only, ephemeral generation flags and explicit legacy selections.
func TestHarnessCodexCommand(t *testing.T) {
	request := fixtureRequest(t, `printf '{"ok":true}' > "$output"`)
	request.Model = "fixture-model"
	request.OutputSchema = []byte(`{"type":"object"}`)

	result, err := (Runner{}).RunAgent(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	command := strings.Join(result.Command, " ")
	for _, flag := range []string{"--sandbox read-only", "--ephemeral", "--skip-git-repo-check", "--output-schema", "--model fixture-model"} {
		if !strings.Contains(command, flag) {
			t.Fatalf("missing %q in %q", flag, command)
		}
	}

	for _, path := range []string{result.PromptPath, result.OutputPath, result.StdoutPath, result.StderrPath, result.CommandPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
}

// Frozen CLI-default deliberately omits a model flag even when the process
// environment changes after admission; ordinary legacy calls retain env defaults.
func TestHarnessCodexFrozenDefault(t *testing.T) {
	t.Setenv("CHR_CODEX_MODEL", "replacement-model")

	request := fixtureRequest(t, `printf '{"ok":true}' > "$output"`)
	request.FrozenSelection = true

	result, err := (Runner{}).RunAgent(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(strings.Join(result.Command, " "), "--model") || result.Model != "" {
		t.Fatal("CLI-default rebound through generation environment")
	}
}

// Adapter deadlines and stream overflow must fail without accepting output.
func TestHarnessCodexCancellation(t *testing.T) {
	request := fixtureRequest(t, `sleep 30`)

	request.Timeout = 20 * time.Millisecond
	if _, err := (Runner{}).RunAgent(t.Context(), request); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error = %v", err)
	}

	request = fixtureRequest(t, `printf 123456789; sleep 30`)

	runner := Runner{CommandExecutor: process.Executor{StdoutLimit: 8}}
	if result, err := runner.RunAgent(t.Context(), request); !errors.Is(err, process.ErrLimit) || len(result.Output) != 0 {
		t.Fatalf("overflow = %q, %v", result.Output, err)
	}
}

// Diagnostic artifacts and errors must exclude supplied credentials.
func TestHarnessCodexRedaction(t *testing.T) {
	t.Setenv("EXAMPLE_API_KEY", "fixture-private-key")

	request := fixtureRequest(t, `printf 'fixture-private-key Bearer fixture-bearer' >&2; exit 1`)

	result, err := (Runner{}).RunAgent(t.Context(), request)
	if err == nil {
		t.Fatal("expected process failure")
	}

	data, readErr := os.ReadFile(result.StderrPath)
	if readErr != nil {
		t.Fatal(readErr)
	}

	for _, value := range []string{string(data), string(result.Stderr), err.Error()} {
		if strings.Contains(value, "fixture-private-key") || strings.Contains(value, "fixture-bearer") {
			t.Fatal("diagnostic retained a credential")
		}
	}
}

func fixtureRequest(t *testing.T, body string) harness.AgentRunRequest {
	t.Helper()

	dir := t.TempDir()
	program := filepath.Join(dir, "codex-fixture")

	script := `#!/bin/sh
while [ "$#" -gt 0 ]; do
 if [ "$1" = --output-last-message ]; then shift; output=$1; fi
 shift
done
cat >/dev/null
` + body + "\n"
	if err := os.WriteFile(program, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	return harness.AgentRunRequest{
		ID: "fixture", Purpose: "generation", ProjectDir: dir,
		RunDir: filepath.Join(dir, "run"), Prompt: "Return JSON.",
		OutputName: "output.json", Program: program,
	}
}
