package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/process"
)

// Failed calls retain their actual transport outcome without normalizing any
// partial output into a final artifact or claiming schema validation occurred.
func TestHarnessGenerationBindingFailures(t *testing.T) {
	for _, test := range []struct {
		name, outcome string
		err           error
	}{
		{"process failure", "failed", errors.New("fixture process failed")},
		{"cancelled", "cancelled", context.Canceled},
		{"deadline", "timed_out", context.DeadlineExceeded},
		{"stream overflow", "protocol_error", process.ErrLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			request := generationBindingRequest(project)

			runner := &generationBindingRunner{result: harness.AgentRunResult{Output: []byte("partial")}, err: test.err}
			if _, err := RunAgent(WithHarnessGeneration(t.Context(), operation), runner, request); !errors.Is(err, test.err) {
				t.Fatal("lost transport error", err)
			}

			record := generationBindingReadMetadata(t, request)
			if record.ProcessOutcome != test.outcome || record.OutputRef != "" || record.OutputHash != "" || record.OutputValidation != "unknown" {
				t.Fatal("failed transport published a final output", record)
			}
		})
	}
}

// Metadata includes measured zero only when actually supplied. Missing counts
// and requested model identity cannot turn into observed whole-call accounting.
func TestHarnessGenerationBindingUsage(t *testing.T) {
	zero, measured, negative := int64(0), int64(17), int64(-1)
	for _, test := range []struct {
		name, status string
		usage        harness.Usage
	}{
		{"unknown", "unknown", harness.Usage{}},
		{"partial input", "partial", harness.Usage{Input: &measured}},
		{"partial output", "partial", harness.Usage{Output: &zero}},
		{"measured zero", "known", harness.Usage{Input: &measured, Output: &zero}},
		{"invalid measurement", "unknown", harness.Usage{Input: &negative, Output: &measured}},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			request := generationBindingRequest(project)

			runner := &generationBindingRunner{result: harness.AgentRunResult{Output: []byte(`{}`), Usage: test.usage}}
			if _, err := RunAgent(WithHarnessGeneration(t.Context(), operation), runner, request); err != nil {
				t.Fatal(err)
			}

			record := generationBindingReadMetadata(t, request)
			if record.UsageStatus != test.status || (record.Usage == nil) != (test.status == "unknown") {
				t.Fatal("invented usage certainty", record)
			}
		})
	}
}

// The non-secret projection and run metadata exclude executable paths and input
// content; digests retain correlation without copying private source material.
func TestHarnessGenerationBindingPrivacy(t *testing.T) {
	project, operation := generationBindingFixture(t)
	request := generationBindingRequest(project)
	request.Prompt = "fixture-private-prompt"
	runner := &generationBindingRunner{result: harness.AgentRunResult{Output: []byte(`{}`)}}

	ctx := WithHarnessGeneration(t.Context(), operation)
	if _, err := RunAgent(ctx, runner, request); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(harnessRunMetadataPath(request))
	if err != nil {
		t.Fatal(err)
	}

	checkGenerationMetadataPrivacy(t, data, request)

	if err := ValidateHarnessGenerationOutput(ctx, request, nil); err != nil {
		t.Fatal(err)
	}

	if record := generationBindingReadMetadata(t, request); record.OutputValidation != "valid" || record.ObservedModel != "" {
		t.Fatal("caller validity or unknown observed model was lost", record)
	}
}

func checkGenerationMetadataPrivacy(t *testing.T, data []byte, request harness.AgentRunRequest) {
	t.Helper()

	if strings.Contains(string(data), request.ProjectDir) || strings.Contains(string(data), request.Prompt) {
		t.Fatal("run metadata copied private inputs or executable paths")
	}

	info, err := os.Stat(harnessRunMetadataPath(request))
	if err != nil || info.Mode().Perm()&0o077 != 0 {
		t.Fatal("metadata was not private", err)
	}
}

// A retained artifact is mandatory: corrupt or public files cannot silently
// reset an operation. FIFO and symlink substitutions must fail without blocking.
func TestHarnessGenerationBindingFiles(t *testing.T) {
	for _, test := range []struct {
		name    string
		replace func(*testing.T, string)
	}{
		{"public mode", func(t *testing.T, path string) {
			if err := os.Chmod(path, 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink", generationBindingSymlink},
		{"oversized record", func(t *testing.T, path string) {
			if err := os.WriteFile(path, []byte(strings.Repeat("x", harnessGenerationRecordLimit+1)), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"named pipe", generationBindingFIFO},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)

			path, err := harnessGenerationOperationPath(project, operation.ID)
			if err != nil {
				t.Fatal(err)
			}

			test.replace(t, path)

			if _, err := FreezeHarnessGeneration(t.Context(), project, operation.ID, operation.Purpose, nil); err == nil {
				t.Fatal("invalid artifact was treated as a fresh operation")
			}
		})
	}
}

func generationBindingSymlink(t *testing.T, path string) {
	t.Helper()

	if err := os.Rename(path, path+".original"); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(path+".original", path); err != nil {
		t.Fatal(err)
	}
}

func generationBindingFIFO(t *testing.T, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
}

// Reusing one call artifact in a different operation cannot overwrite its
// original selection evidence, even if the current profile name is unchanged.
func TestHarnessGenerationBindingCallIdentity(t *testing.T) {
	project, operation := generationBindingFixture(t)
	request := generationBindingRequest(project)

	runner := &generationBindingRunner{result: harness.AgentRunResult{Output: []byte(`{}`)}}
	if _, err := RunAgent(WithHarnessGeneration(t.Context(), operation), runner, request); err != nil {
		t.Fatal(err)
	}

	second, err := FreezeHarnessGeneration(t.Context(), project, "second-operation", operation.Purpose, generationBindingResolver(operation.program, "model-alpha"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := RunAgent(WithHarnessGeneration(t.Context(), second), runner, request); err == nil || runner.calls != 1 {
		t.Fatal("reused call artifact was overwritten by another operation", err)
	}

	if _, err := os.Stat(filepath.Join(request.RunDir, "output.json")); err != nil {
		t.Fatal("conflicting request damaged original output", err)
	}
}
