package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// A retained operation survives a live switch and missing private configuration.
// Nested calls share the same value, while an independent request resolves anew.
func TestHarnessGenerationBindingFreeze(t *testing.T) {
	project := t.TempDir()
	program := generationBindingProgram(t, project)
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: alpha\n")

	resolve := generationBindingResolver(program, "model-alpha")

	first, err := FreezeHarnessGeneration(t.Context(), project, "operation-first", "plan_generation", resolve)
	if err != nil {
		t.Fatal(err)
	}

	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: beta\n")

	nested, err := FreezeHarnessGeneration(WithHarnessGeneration(t.Context(), first), project, "nested", "question", nil)
	if err != nil || nested != first {
		t.Fatal("nested call rebound its operation", err)
	}

	retained, err := FreezeHarnessGeneration(t.Context(), project, first.ID, first.Purpose, nil)
	if err != nil || retained != first {
		t.Fatal("restart consulted replacement selection", err)
	}

	checkGenerationBindingFresh(t, project, first)
}

func checkGenerationBindingFresh(t *testing.T, project string, first HarnessGenerationOperation) {
	t.Helper()

	fresh, err := FreezeHarnessGeneration(t.Context(), project, "independent", first.Purpose, generationBindingResolver(first.program, "model-beta"))
	if err != nil || fresh.Binding.Selection.Profile != "beta" || fresh.Binding.Selection.SelectionRevision == first.Binding.Selection.SelectionRevision {
		t.Fatal("fresh operation did not resolve current configuration", err)
	}

	if _, err := FreezeHarnessGeneration(t.Context(), project, first.ID, "other-purpose", nil); err == nil {
		t.Fatal("reused operation identity changed purpose")
	}

	// Even unreadable live settings must not prevent retained identity recovery.
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  unknown: invalid\n")

	if resumed, err := FreezeHarnessGeneration(t.Context(), project, first.ID, first.Purpose, nil); err != nil || resumed != first {
		t.Fatal("recovery tried parsing live configuration", err)
	}
}

func generationBindingProgram(t *testing.T, project string) string {
	t.Helper()

	program := filepath.Join(project, "fixture-driver")
	if err := os.WriteFile(program, []byte("#!/bin/sh\nprintf '{\"ok\":true}'\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	return program
}

func generationBindingResolver(program, model string) HarnessGenerationResolver {
	return func(_ context.Context, config WorkspaceConfig) (harnesscore.GenerationBinding, HarnessGenerationLaunch, error) {
		executable, err := FreezeHarnessExecutable([]string{program})
		if err != nil {
			return harnesscore.GenerationBinding{}, HarnessGenerationLaunch{}, err
		}

		selection := harnesscore.Selection{
			Role: harnesscore.Generation, Action: "plan_generation", Source: "project", Profile: config.Harnesses.Generation,
			Harness: "codex", Backend: "codex", ModelProvider: "cli_default", Model: model,
			SelectionRevision: config.Revision(), ProfileRevision: contenthash.Sum([]byte("fixture-profile-" + model)),
			Capabilities: harness.Capabilities{Generation: true},
			Readiness:    harnesscore.Readiness{Supported: true, Executable: "available", Credentials: "cli_managed", Runtime: "native"},
		}
		binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{Selection: selection, ExecutableRef: executable.Reference, ProtocolRevision: "fixture-generation/v1"})

		return binding, HarnessGenerationLaunch{Program: executable.Command[0]}, err
	}
}

// Corrupt records and future schemas fail before a resolver or process can run.
func TestHarnessGenerationBindingCorruption(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*harnessGenerationOperationFile)
	}{
		{"future schema", func(r *harnessGenerationOperationFile) { r.Schema = "chr-harness-operation/v2" }},
		{"model mutation", func(r *harnessGenerationOperationFile) { r.Operation.Binding.Selection.Model = "other" }},
		{"program mutation", func(r *harnessGenerationOperationFile) { r.Program = "/missing-fixture-driver" }},
		{"operation mutation", func(r *harnessGenerationOperationFile) { r.Operation.ID = "other" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			replaceGenerationBindingRecord(t, operation, test.change)

			if _, err := FreezeHarnessGeneration(t.Context(), project, operation.ID, operation.Purpose, nil); err == nil {
				t.Fatal("retained corruption accepted")
			}
		})
	}
}

func replaceGenerationBindingRecord(t *testing.T, operation HarnessGenerationOperation, change func(*harnessGenerationOperationFile)) {
	t.Helper()

	path, err := harnessGenerationOperationPath(operation.project, operation.ID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var record harnessGenerationOperationFile
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}

	change(&record)

	record.Checksum = ""

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	record.Checksum = contenthash.Sum(encoded)
	if err := writeHarnessGenerationFile(path, record); err != nil {
		t.Fatal(err)
	}
}

func generationBindingFixture(t *testing.T) (string, HarnessGenerationOperation) {
	t.Helper()

	project := t.TempDir()
	program := generationBindingProgram(t, project)
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: alpha\n")

	operation, err := FreezeHarnessGeneration(t.Context(), project, "operation-fixture", "plan_generation", generationBindingResolver(program, "model-alpha"))
	if err != nil {
		t.Fatal(err)
	}

	return project, operation
}

// A changed executable, explicit override or cross-project call fails before
// the injected runner. Reading the retained operation remains diagnostic only.
func TestHarnessGenerationBindingDispatchFence(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*testing.T, HarnessGenerationOperation, *harness.AgentRunRequest)
	}{
		{"program override", func(_ *testing.T, _ HarnessGenerationOperation, r *harness.AgentRunRequest) {
			r.Program = "/other-fixture"
		}},
		{"model override", func(_ *testing.T, _ HarnessGenerationOperation, r *harness.AgentRunRequest) { r.Model = "other" }},
		{"artifact traversal", func(_ *testing.T, _ HarnessGenerationOperation, r *harness.AgentRunRequest) { r.ID = "../escape" }},
		{"cross project", func(t *testing.T, _ HarnessGenerationOperation, r *harness.AgentRunRequest) {
			r.ProjectDir = t.TempDir()
		}},
		{"changed executable", func(t *testing.T, o HarnessGenerationOperation, _ *harness.AgentRunRequest) {
			if err := os.WriteFile(o.program, []byte("#!/bin/sh\nexit 99\n"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			request := generationBindingRequest(project)
			test.change(t, operation, &request)

			runner := &generationBindingRunner{}
			if _, err := RunAgent(WithHarnessGeneration(t.Context(), operation), runner, request); err == nil || runner.calls != 0 {
				t.Fatal("dispatch fence did not precede model execution", err)
			}

			if _, err := FreezeHarnessGeneration(t.Context(), project, operation.ID, operation.Purpose, nil); err != nil {
				t.Fatal("diagnostic recovery required a dispatch-ready executable", err)
			}
		})
	}
}

func generationBindingRequest(project string) harness.AgentRunRequest {
	return harness.AgentRunRequest{ID: "fixture-call", Purpose: "plan_outline_generation", ProjectDir: project, RunDir: workspacecore.UserPath(project, "tmp", "generation-fixture"), Prompt: "Private fixture prompt.", OutputSchema: []byte(`{"type":"object"}`), OutputName: "output.json"}
}

type generationBindingRunner struct {
	calls   int
	request harness.AgentRunRequest
	result  harness.AgentRunResult
	err     error
	before  func(harness.AgentRunRequest)
}

func (r *generationBindingRunner) RunAgent(_ context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	r.calls++

	r.request = request
	if r.before != nil {
		r.before(request)
	}

	return r.result, r.err
}

// Validation and transport outcomes are independent; no usage or model
// observation may be invented from a successful process or requested model.
func TestHarnessGenerationBindingMetadata(t *testing.T) {
	project, operation := generationBindingFixture(t)
	ctx := WithHarnessGeneration(t.Context(), operation)
	request := generationBindingRequest(project)

	runner := &generationBindingRunner{
		result: harness.AgentRunResult{Output: []byte(`{"ok":true}`), Model: "requested-not-observed"},
		before: func(request harness.AgentRunRequest) { checkGenerationMetadataStarted(t, request) },
	}
	if _, err := RunAgent(ctx, runner, request); err != nil {
		t.Fatal(err)
	}

	if runner.request.Program != operation.program || runner.request.Model != "model-alpha" || !runner.request.FrozenSelection {
		t.Fatal("adapter request did not carry frozen selection")
	}

	record := generationBindingReadMetadata(t, request)
	checkGenerationMetadataCompleted(t, record, operation, runner.result.Output)

	invalid := errors.New("schema rejects fixture")
	if err := ValidateHarnessGenerationOutput(ctx, request, invalid); !errors.Is(err, invalid) {
		t.Fatal("lost caller validation failure", err)
	}

	record = generationBindingReadMetadata(t, request)
	if record.ProcessOutcome != "completed" || record.OutputValidation != "invalid" {
		t.Fatal("schema failure replaced process outcome", record)
	}
}

func checkGenerationMetadataStarted(t *testing.T, request harness.AgentRunRequest) {
	t.Helper()

	record := generationBindingReadMetadata(t, request)
	if record.ProcessOutcome != "started" || record.FinishedAt != nil || record.OutputHash != "" {
		t.Fatal("host did not retain started metadata before model launch", record)
	}
}

func checkGenerationMetadataCompleted(t *testing.T, record HarnessRunMetadata, operation HarnessGenerationOperation, output []byte) {
	t.Helper()

	if record.ProcessOutcome != "completed" || record.OutputValidation != "unknown" || record.FinishedAt == nil {
		t.Fatal("transport invented validity", record)
	}

	if record.Usage != nil || record.UsageStatus != "unknown" || record.ObservedModel != "" {
		t.Fatal("transport invented usage or observed model", record)
	}

	checkGenerationMetadataOutput(t, record, operation, output)
}

func checkGenerationMetadataOutput(t *testing.T, record HarnessRunMetadata, operation HarnessGenerationOperation, output []byte) {
	t.Helper()

	if record.OutputRef != "output.json" || record.OutputHash != contenthash.Sum(output) || record.SelectionRevision != operation.Binding.Selection.SelectionRevision {
		t.Fatal("normalized artifact identity was not retained", record)
	}
}

func generationBindingReadMetadata(t *testing.T, request harness.AgentRunRequest) HarnessRunMetadata {
	t.Helper()

	data, err := readHarnessGenerationFile(harnessRunMetadataPath(request))
	if err != nil {
		t.Fatal(err)
	}

	var record HarnessRunMetadata
	if err := decodeHarnessGenerationFile(data, &record); err != nil {
		t.Fatal(err)
	}

	return record
}
