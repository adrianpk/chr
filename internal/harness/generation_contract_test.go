package harness_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/codex"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

type contractAgentRunner interface {
	harness.AgentRunner
	Capabilities() harness.Capabilities
}

// Both native adapters obey the same actual process, bounds and failure contract.
func TestPiGenerationAdapterContract(t *testing.T) {
	for _, adapter := range []string{"codex", "pi"} {
		t.Run(adapter, func(t *testing.T) {
			for _, scenario := range []string{"success", "unknown usage", "output limit", "failed exit", "cancelled", "deadline", "diagnostic", "input limit"} {
				t.Run(scenario, func(t *testing.T) { checkGenerationAdapter(t, adapter, scenario) })
			}

			if adapter == "pi" {
				t.Run("diagnostic artifact", func(t *testing.T) { checkGenerationAdapter(t, adapter, "diagnostic artifact") })
			}
		})
	}
}

func checkGenerationAdapter(t *testing.T, adapter, scenario string) {
	t.Helper()

	runner, request := generationAdapterFixture(t, adapter, scenario)

	capabilities := runner.Capabilities()
	if !capabilities.Generation || capabilities.CertifiedWholeCallUsage {
		t.Fatal("adapter misrepresented its generation or whole-call usage capability")
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	switch scenario {
	case "cancelled":
		cancel()
	case "deadline":
		request.Timeout = 30 * time.Millisecond
	case "output limit":
		request.OutputLimit = 3
	case "input limit":
		request.Prompt = strings.Repeat("x", harness.MaxInput+1)
	}

	result, err := runner.RunAgent(ctx, request)
	checkAdapterOutcome(t, result, err, scenario)

	if scenario == "diagnostic artifact" {
		data, err := os.ReadFile(filepath.Join(request.ProjectDir, "retained"))
		if err != nil || string(data) != "Retain this file." {
			t.Fatal("diagnostic redirection changed an existing file", err)
		}
	}
}

func checkAdapterOutcome(t *testing.T, result harness.AgentRunResult, err error, scenario string) {
	t.Helper()

	success := scenario == "success" || scenario == "unknown usage"
	if success && (err != nil || string(result.Output) != `{"ok":true}`) {
		t.Fatal("native generation failed", err)
	}

	if !success && (err == nil || len(result.Output) != 0) {
		t.Fatal("failed generation exposed output", err)
	}

	checkAdapterFailure(t, result, err, scenario)
}

func checkAdapterFailure(t *testing.T, result harness.AgentRunResult, err error, scenario string) {
	t.Helper()

	switch scenario {
	case "cancelled":
		if !errors.Is(err, context.Canceled) {
			t.Fatal("caller cancellation identity lost", err)
		}
	case "deadline":
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("caller deadline identity lost", err)
		}
	case "unknown usage":
		if result.Usage.Input != nil || result.Usage.Output != nil {
			t.Fatal("missing usage fabricated counters")
		}
	case "diagnostic":
		checkAdapterDiagnostics(t, result, err)
	}
}

func generationAdapterFixture(t *testing.T, adapter, scenario string) (contractAgentRunner, harness.AgentRunRequest) {
	t.Helper()

	dir := t.TempDir()
	program := filepath.Join(dir, "driver")
	request := harness.AgentRunRequest{ID: "contract", Purpose: "test", ProjectDir: dir, RunDir: filepath.Join(dir, "run"), Prompt: "Return the owned fixture result.", OutputName: "output.json", Program: program}

	if adapter == "codex" {
		t.Setenv("CHR_CONTRACT_API_KEY", "FIXTURE_PRIVATE_BOUNDARY_KEY")
		writeAdapterFixture(t, program, codexContractScript(scenario), 0o700)

		return codex.Runner{}, request
	}

	peer, err := filepath.Abs("pi/testdata/rpc-peer.py")
	if err != nil {
		t.Fatal(err)
	}

	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal(err)
	}

	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
	writeAdapterFixture(t, program, "#!/bin/sh\nexec "+quote(python)+" "+quote(peer)+" "+quote(dir)+" \"$@\"\n", 0o700)
	writeAdapterFixture(t, filepath.Join(dir, "result.json"), `{"ok":true}`, 0o600)
	writeAdapterFixture(t, filepath.Join(dir, "scenario"), scenario, 0o600)
	writeAdapterFixture(t, filepath.Join(dir, "auth.json"), `{"fixture":{"type":"api_key","key":"FIXTURE_PRIVATE_BOUNDARY_KEY"}}`, 0o600)

	request.Model, request.ModelProvider, request.FrozenSelection = "model", "fixture", true

	request.Resources.CredentialFile = filepath.Join(dir, "auth.json")
	if scenario == "diagnostic artifact" {
		prepareAdapterDiagnosticFault(t, request)
	}

	return pi.Runner{}, request
}

func prepareAdapterDiagnosticFault(t *testing.T, request harness.AgentRunRequest) {
	t.Helper()
	writeAdapterFixture(t, filepath.Join(request.ProjectDir, "retained"), "Retain this file.", 0o600)

	if err := os.Mkdir(request.RunDir, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(request.ProjectDir, "retained"), filepath.Join(request.RunDir, "stderr.log")); err != nil {
		t.Fatal(err)
	}
}

func codexContractScript(scenario string) string {
	prologue := "#!/bin/sh\ncat >/dev/null\n"

	switch scenario {
	case "failed exit":
		return prologue + "printf '{\"ok\":true}'; exit 7\n"
	case "deadline":
		return prologue + "sleep 30\n"
	case "diagnostic":
		return prologue + "head -c 262141 /dev/zero | tr '\\000' X >&2\nprintf '%s' \"$CHR_CONTRACT_API_KEY\" >&2\nexit 7\n"
	default:
		return prologue + "printf '{\"ok\":true}'\n"
	}
}

func checkAdapterDiagnostics(t *testing.T, result harness.AgentRunResult, runErr error) {
	t.Helper()

	retained, err := os.ReadFile(result.StderrPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, diagnostic := range []string{string(result.Stderr), string(retained), runErr.Error()} {
		if strings.Contains(diagnostic, "FIX") || len(result.Stderr) > harness.MaxDiagnostic {
			t.Fatal("split credential survived the diagnostic excerpt boundary")
		}
	}
}

func writeAdapterFixture(t *testing.T, path, value string, mode os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, []byte(value), mode); err != nil {
		t.Fatal(err)
	}
}
