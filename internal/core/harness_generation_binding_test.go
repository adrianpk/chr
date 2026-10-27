package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
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

	first, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: "operation-first", Purpose: "plan_generation"}, resolve)
	if err != nil {
		t.Fatal(err)
	}

	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: beta\n")

	nested, err := FreezeHarnessGeneration(harnesshost.WithHarnessGeneration(t.Context(), first), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: "nested", Purpose: "question"}, nil)
	if err != nil || nested != first {
		t.Fatal("nested call rebound its operation", err)
	}

	retained, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: first.ID, Purpose: first.Purpose}, nil)
	if err != nil || retained != first {
		t.Fatal("restart consulted replacement selection", err)
	}

	checkGenerationBindingFresh(t, project, first)
}

func checkGenerationBindingFresh(t *testing.T, project string, first harnesshost.HarnessGenerationOperation) {
	t.Helper()

	fresh, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: "independent", Purpose: first.Purpose}, generationBindingResolver(filepath.Join(project, "fixture-driver"), "model-beta"))
	if err != nil || fresh.Binding.Selection.Profile != "beta" || fresh.Binding.Selection.SelectionRevision == first.Binding.Selection.SelectionRevision {
		t.Fatal("fresh operation did not resolve current configuration", err)
	}

	if _, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: first.ID, Purpose: "other-purpose"}, nil); err == nil {
		t.Fatal("reused operation identity changed purpose")
	}

	// Even unreadable live settings must not prevent retained identity recovery.
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  unknown: invalid\n")

	if resumed, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: first.ID, Purpose: first.Purpose}, nil); err != nil || resumed != first {
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
	return func(_ context.Context, config WorkspaceConfig) (harnesshost.HarnessGenerationPreparation, error) {
		executable, err := harnesshost.FreezeExecutable([]string{program})
		if err != nil {
			return harnesshost.HarnessGenerationPreparation{}, err
		}

		selection := harnesscore.Selection{
			Role: harnesscore.Generation, Action: "plan_generation", Source: "project", Profile: config.Harnesses.Generation,
			Harness: "codex", Backend: "codex", ModelProvider: "cli_default", Model: model,
			SelectionRevision: config.Revision(), ProfileRevision: contenthash.Sum([]byte("fixture-profile-" + model)),
			Capabilities: harness.Capabilities{Generation: true},
			Readiness:    harnesscore.Readiness{Supported: true, Executable: "available", Credentials: "cli_managed", Runtime: "native"},
		}
		binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{Selection: selection, ExecutableRef: executable.Reference, ProtocolRevision: "fixture-generation/v1"})

		return harnesshost.HarnessGenerationPreparation{Binding: binding, Launch: harnesshost.HarnessGenerationLaunch{Program: executable.Command[0]}}, err
	}
}

// Corrupt records and future schemas fail before a resolver or process can run.
func TestHarnessGenerationBindingCorruption(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*generationBindingWire)
	}{
		{"future schema", func(r *generationBindingWire) { r.Schema = "chr-harness-operation/v2" }},
		{"model mutation", func(r *generationBindingWire) { r.Operation.Binding.Selection.Model = "other" }},
		{"program mutation", func(r *generationBindingWire) { r.Program = "/missing-fixture-driver" }},
		{"operation mutation", func(r *generationBindingWire) { r.Operation.ID = "other" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			replaceGenerationBindingRecord(t, project, operation, test.change)

			if _, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: operation.ID, Purpose: operation.Purpose}, nil); err == nil {
				t.Fatal("retained corruption accepted")
			}
		})
	}
}

func replaceGenerationBindingRecord(t *testing.T, project string, operation harnesshost.HarnessGenerationOperation, change func(*generationBindingWire)) {
	t.Helper()

	path, err := generationBindingOperationPath(project, operation.ID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var record generationBindingWire
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}

	image, err := json.Marshal(record)
	if err != nil || !bytes.Equal(image, data) {
		t.Fatal("corruption fixture changed the native wire image before mutation", err)
	}

	change(&record)

	record.Checksum = ""

	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	record.Checksum = contenthash.Sum(encoded)

	data, err = json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func generationBindingFixture(t *testing.T) (string, harnesshost.HarnessGenerationOperation) {
	t.Helper()

	project := t.TempDir()
	program := generationBindingProgram(t, project)
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: alpha\n")

	operation, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: "operation-fixture", Purpose: "plan_generation"}, generationBindingResolver(program, "model-alpha"))
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
		change func(*testing.T, string, *harness.AgentRunRequest)
	}{
		{"program override", func(_ *testing.T, _ string, r *harness.AgentRunRequest) {
			r.Program = "/other-fixture"
		}},
		{"model override", func(_ *testing.T, _ string, r *harness.AgentRunRequest) {
			r.Model = "other"
		}},
		{"artifact traversal", func(_ *testing.T, _ string, r *harness.AgentRunRequest) {
			r.ID = "../escape"
		}},
		{"cross project", func(t *testing.T, _ string, r *harness.AgentRunRequest) {
			r.ProjectDir = t.TempDir()
		}},
		{"changed executable", func(t *testing.T, program string, _ *harness.AgentRunRequest) {
			if err := os.WriteFile(program, []byte("#!/bin/sh\nexit 99\n"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			project, operation := generationBindingFixture(t)
			request := generationBindingRequest(project)
			test.change(t, filepath.Join(project, "fixture-driver"), &request)

			runner := &generationBindingRunner{}
			if _, err := harnesshost.RunAgent(harnesshost.WithHarnessGeneration(t.Context(), operation), runner, request); err == nil || runner.calls != 0 {
				t.Fatal("dispatch fence did not precede model execution", err)
			}

			if _, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: operation.ID, Purpose: operation.Purpose}, nil); err != nil {
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
	ctx := harnesshost.WithHarnessGeneration(t.Context(), operation)
	request := generationBindingRequest(project)

	runner := &generationBindingRunner{
		result: harness.AgentRunResult{Output: []byte(`{"ok":true}`), Model: "requested-not-observed"},
		before: func(request harness.AgentRunRequest) { checkGenerationMetadataStarted(t, request) },
	}
	if _, err := harnesshost.RunAgent(ctx, runner, request); err != nil {
		t.Fatal(err)
	}

	if runner.request.Program != filepath.Join(project, "fixture-driver") || runner.request.Model != "model-alpha" || !runner.request.FrozenSelection {
		t.Fatal("adapter request did not carry frozen selection")
	}

	record := generationBindingReadMetadata(t, request)
	checkGenerationMetadataCompleted(t, record, operation, runner.result.Output)

	invalid := errors.New("schema rejects fixture")
	if err := harnesshost.ValidateHarnessGenerationOutput(ctx, request, invalid); !errors.Is(err, invalid) {
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

func checkGenerationMetadataCompleted(t *testing.T, record harnesshost.HarnessRunMetadata, operation harnesshost.HarnessGenerationOperation, output []byte) {
	t.Helper()

	if record.ProcessOutcome != "completed" || record.OutputValidation != "unknown" || record.FinishedAt == nil {
		t.Fatal("transport invented validity", record)
	}

	if record.Usage != nil || record.UsageStatus != "unknown" || record.ObservedModel != "" {
		t.Fatal("transport invented usage or observed model", record)
	}

	checkGenerationMetadataOutput(t, record, operation, output)
}

func checkGenerationMetadataOutput(t *testing.T, record harnesshost.HarnessRunMetadata, operation harnesshost.HarnessGenerationOperation, output []byte) {
	t.Helper()

	if record.OutputRef != "output.json" || record.OutputHash != contenthash.Sum(output) || record.SelectionRevision != operation.Binding.Selection.SelectionRevision {
		t.Fatal("normalized artifact identity was not retained", record)
	}
}

func generationBindingReadMetadata(t *testing.T, request harness.AgentRunRequest) harnesshost.HarnessRunMetadata {
	t.Helper()

	data, err := os.ReadFile(generationBindingMetadataPath(request))
	if err != nil {
		t.Fatal(err)
	}

	var record harnesshost.HarnessRunMetadata
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}

	return record
}
