package core

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// An admission holds the same config owner as role changes. Its complete old
// revision remains coherent; the next independent operation sees both new roles.
func TestHarnessGenerationBindingSelectionOrder(t *testing.T) {
	project := t.TempDir()
	program := generationBindingProgram(t, project)
	original := "harnesses:\n  worker: alpha\n  generation: alpha\n"
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), original)

	entered, release := make(chan struct{}), make(chan struct{})
	resolve := generationBindingResolver(program, "model-alpha")
	finished := make(chan generationBindingAdmissionResult, 1)

	go func() {
		operation, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: "ordered-operation", Purpose: "plan_generation"}, func(ctx context.Context, c WorkspaceConfig) (harnesshost.GenerationPreparation, error) {
			close(entered)
			<-release

			return resolve(ctx, c)
		})
		finished <- generationBindingAdmissionResult{operation, err}
	}()

	select {
	case <-entered:
	case result := <-finished:
		t.Fatal("admission failed before reaching its resolver", result.err)
	}

	changed := make(chan error, 1)

	go func() {
		request := harnessSelectionTestRequest("ordered-switch", contenthash.Sum([]byte(original)), "beta")

		_, err := ChangeHarnessSelections(t.Context(), project, harnessSelectionTestOptions(request))
		changed <- err
	}()

	close(release)

	admission := <-finished
	if admission.err != nil || admission.operation.Binding.Selection.SelectionRevision != contenthash.Sum([]byte(original)) {
		t.Fatal("admission observed a mixed configuration image", admission.err)
	}

	if err := <-changed; err != nil {
		t.Fatal("ordered switch failed", err)
	}

	fresh, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: "after-switch", Purpose: "plan_generation"}, generationBindingResolver(program, "model-beta"))
	if err != nil || fresh.Binding.Selection.Profile != "beta" {
		t.Fatal("independent admission retained old role", err)
	}
}

type generationBindingAdmissionResult struct {
	operation harnesshost.GenerationOperation
	err       error
}

// Direct API identity is retained without inventing a CLI executable or routing
// a legacy API operation into a replacement harness process.
func TestHarnessGenerationBindingDirectAPI(t *testing.T) {
	project := t.TempDir()

	operation, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: "api-operation", Purpose: "quiz_next_question"}, func(_ context.Context, c WorkspaceConfig) (harnesshost.GenerationPreparation, error) {
		binding, err := harnesscore.SealGenerationBinding(harnesscore.GenerationBinding{
			Selection: harnesscore.Selection{
				Role: harnesscore.Generation, Action: "quiz_next_question", Backend: "openai", Source: "legacy", ModelProvider: "openai", Model: "fixture-api-model",
				SelectionRevision: c.Revision(), ProfileRevision: contenthash.Sum([]byte("fixture-api-profile")),
				Readiness: harnesscore.Readiness{Supported: true, Executable: "not_required", Credentials: "api_managed", Runtime: "not_required"},
			}, ProtocolRevision: "responses/v1",
		})

		return harnesshost.GenerationPreparation{Binding: binding}, err
	})
	if err != nil {
		t.Fatal(err)
	}

	runner := &generationBindingRunner{}
	if _, err := harnesshost.RunAgent(harnesshost.WithGeneration(t.Context(), operation), runner, generationBindingRequest(project)); err == nil || runner.calls != 0 {
		t.Fatal("direct API operation dispatched through a harness", err)
	}

	retained, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: operation.ID, Purpose: operation.Purpose}, nil)
	if err != nil || retained != operation {
		t.Fatal("direct API recovery changed identity", err)
	}
}

// Reaching the retained-identity limit fails before model dispatch and preserves
// prior identities; no eviction can silently rebind an interrupted operation.
func TestHarnessGenerationBindingCapacity(t *testing.T) {
	project, operation := generationBindingFixture(t)

	path, err := generationBindingOperationPath(project, operation.ID)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < workflowstate.RecordLimit; i++ {
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "fixture-"+strconv.Itoa(i)+".json"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	_, err = FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: "over-capacity", Purpose: operation.Purpose}, generationBindingResolver(filepath.Join(project, "fixture-driver"), "model-alpha"))
	if err == nil {
		t.Fatal("identity capacity silently evicted an existing operation")
	}

	if retained, err := FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: project, ID: operation.ID, Purpose: operation.Purpose}, nil); err != nil || retained != operation {
		t.Fatal("capacity exhaustion broke existing recovery", err)
	}
}

// The original owner may refresh bounded input on a legitimate provider retry.
// Metadata records current input while the operation's selection stays frozen.
func TestHarnessGenerationBindingInputIdentity(t *testing.T) {
	project, operation := generationBindingFixture(t)
	ctx := harnesshost.WithGeneration(t.Context(), operation)
	request := generationBindingRequest(project)

	runner := &generationBindingRunner{result: harness.AgentRunResult{Output: []byte(`{}`)}}
	if _, err := harnesshost.RunAgent(ctx, runner, request); err != nil {
		t.Fatal(err)
	}

	request.Prompt = "replacement input"
	if _, err := harnesshost.RunAgent(ctx, runner, request); err != nil || runner.calls != 2 {
		t.Fatal("retry introduced another argument-replay authority", err)
	}

	record := generationBindingReadMetadata(t, request)
	if record.InputRef != "sha256:"+contenthash.Sum([]byte(request.Prompt)) || record.Selection != operation.Binding {
		t.Fatal("retry lost current input or original selection", record)
	}
}
