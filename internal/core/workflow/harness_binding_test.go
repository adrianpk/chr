package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/harness"
)

// Real pending tickets, checkpoint/restart and subsequent attempts retain one
// binding even after the live role points to an unavailable replacement.
func TestHarnessBindingPending(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.ResolveHarness = harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))

	started := workflowTestStart(t, store, request)
	if started.Invocation.Schema != "chr-invocation/v2" || started.Invocation.HarnessBinding == nil {
		t.Fatal("explicit admission was not versioned")
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: unavailable\n  generation: other\n")

	request.ResolveHarness = func(context.Context, workspaceconfig.Config) (HarnessResolution, error) {
		t.Fatal("retained request resolved live configuration")

		return HarnessResolution{}, nil
	}

	replay := workflowTestStart(t, store, request)
	if !replay.Attached || !sameHarnessBinding(replay.Invocation.HarnessBinding, started.Invocation.HarnessBinding) {
		t.Fatal("retry replaced its binding")
	}

	workflowBindingExecute(t, store, started.Invocation.ID)
	checkHarnessBindingPause(t, store, started.Invocation)
	workflowBindingExecute(t, store, started.Invocation.ID)

	retained, err := store.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil || retained.Cursor != 2 || len(retained.Attempts) != 2 || !sameHarnessBinding(retained.HarnessBinding, started.Invocation.HarnessBinding) {
		t.Fatal("pending tickets rebound after recovery", retained.Cursor, err)
	}

	checkHarnessBindingInputs(t, store, retained)
}

func harnessBindingTestResolver(t *testing.T, command []string) HarnessResolver {
	t.Helper()

	executable, err := harnesshost.FreezeExecutable(command)
	if err != nil {
		t.Fatal(err)
	}

	return func(_ context.Context, config workspaceconfig.Config) (HarnessResolution, error) {
		binding, err := harnesscore.SealBinding(harnesscore.Binding{SchemaVersion: 1, Role: harnesscore.Worker, Harness: "codex", Source: "project", SelectionRevision: config.Revision(), Profile: config.Harnesses.Worker, ProfileRevision: contenthash.Sum([]byte("fixture-profile-" + config.Harnesses.Worker)), ModelProvider: "fixture-provider", Model: "fixture-model", ExecutableRef: executable.Reference, RuntimeRef: "native", ProtocolRevision: "fixture-exec/v1", Capabilities: harness.Capabilities{Worker: true}})

		return HarnessResolution{Binding: &binding, Command: slices.Clone(executable.Command)}, err
	}
}

func workflowBindingExecute(t *testing.T, store *Store, id string) {
	t.Helper()

	claim := workflowTestClaim(t, store, id)

	result, err := store.ExecuteWorkflowTask(t.Context(), claim, nil)
	if err != nil || !result.Successful {
		t.Fatal("bound task failed", result.Failure, err)
	}
}

func checkHarnessBindingPause(t *testing.T, store *Store, original workflowstate.Invocation) {
	t.Helper()

	paused, err := store.StopWorkflow(t.Context(), original.ID, "pause")
	if err != nil || paused.Status != workflowstate.InvocationPaused {
		t.Fatal("bound invocation did not pause", err)
	}

	runHarnessBindingHost(t, store, original.ID)

	resumed, err := store.ReadWorkflow(t.Context(), original.ID)
	if err != nil || resumed.Policy != original.Policy || !sameHarnessBinding(resumed.HarnessBinding, original.HarnessBinding) {
		t.Fatal("resume replaced frozen policy or harness", err)
	}
}

func checkHarnessBindingInputs(t *testing.T, store *Store, run workflowstate.Invocation) {
	t.Helper()

	for _, attempt := range run.Attempts {
		data, err := opsfile.Read(filepath.Join(store.layout.Root(), attempt.InputRef))
		if err != nil || contenthash.Sum(data) != attempt.InputHash {
			t.Fatal("worker input digest changed", err)
		}

		var input TaskInput
		if err := json.Unmarshal(data, &input); err != nil {
			t.Fatal(err)
		}

		if input.Schema != "chr-task-input/v2" || !sameHarnessBinding(input.HarnessBinding, run.HarnessBinding) || validateWorkflowTaskHarnessSchema(input) != nil {
			t.Fatal("task input did not retain mandatory binding")
		}
	}
}

// Executable replacement and explicit command overrides are rejected before a
// model can run; read-only inspection still permits diagnosing a missing driver.
func TestHarnessBindingExecutableFence(t *testing.T) {
	for _, mode := range []string{"replaced", "command override"} {
		t.Run(mode, func(t *testing.T) { checkHarnessBindingExecutableFence(t, mode) })
	}
}

func checkHarnessBindingExecutableFence(t *testing.T, mode string) {
	t.Helper()

	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	driver := filepath.Join(t.TempDir(), "driver")
	writeHarnessSelectionTestFile(t, driver, "#!/bin/sh\nexit 0\n")

	if err := os.Chmod(driver, 0o700); err != nil {
		t.Fatal(err)
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.ResolveHarness = harnessBindingTestResolver(t, []string{driver})

	started := workflowTestStart(t, store, request)
	if mode == "replaced" {
		writeHarnessSelectionTestFile(t, driver, "#!/bin/sh\nexit 99\n")

		if _, err := store.ClaimWorkflow(t.Context(), started.Invocation.ID); err == nil {
			t.Fatal("replaced executable acquired a claim")
		}
	} else {
		checkHarnessBindingCommandOverride(t, store, started.Invocation.ID)
	}

	if _, err := store.ReadWorkflow(t.Context(), started.Invocation.ID); err != nil {
		t.Fatal("inspection executed or required the changed driver", err)
	}
}

func checkHarnessBindingCommandOverride(t *testing.T, store *Store, id string) {
	t.Helper()

	claim := workflowTestClaim(t, store, id)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, []string{"sh", "-c", "exit 0"}); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("caller replaced the bound command", err)
	}
}

// Explicit schemas and request hashes protect mandatory semantics. Existing
// request IDs cannot exchange model or executable evidence on replay.
func TestHarnessBindingRejects(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.ResolveHarness = harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))

	started := workflowTestStart(t, store, request)
	for _, change := range []string{"hash", "model", "schema", "future schema", "missing binding", "private command"} {
		t.Run(change, func(t *testing.T) {
			run, err := store.readWorkflowInvocation(started.Invocation.ID)
			if err != nil {
				t.Fatal(err)
			}

			run.HarnessBinding = harnesscore.CloneBinding(run.HarnessBinding)

			switch change {
			case "hash":
				run.HarnessBinding.BindingHash = contenthash.Sum([]byte("changed"))
			case "model":
				run.HarnessBinding.Model = "different-model"
			case "schema":
				run.Schema = "chr-invocation/v1"
			case "future schema":
				run.Schema = "chr-invocation/v3"
			case "missing binding":
				run.HarnessBinding = nil
			case "private command":
				run.HarnessCommand = []string{"/different/driver"}
			}

			if err := validateWorkflowInvocation(run); err == nil {
				t.Fatal("changed mandatory binding accepted")
			}
		})
	}

	checkHarnessBindingReplayModel(t, store, request, *started.Invocation.HarnessBinding)
}

func checkHarnessBindingReplayModel(t *testing.T, store *Store, request StartRequest, binding harnesscore.Binding) {
	t.Helper()

	changed := request
	binding.Model = "different-model"

	sealed, err := harnesscore.SealBinding(binding)
	if err != nil {
		t.Fatal(err)
	}

	changed.HarnessBinding = &sealed
	if _, err := store.StartWorkflow(t.Context(), changed); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("replay exchanged its binding", err)
	}
}

// Descriptive capabilities cannot bypass the original finite whole-call budget
// authority; an uncertified bound worker is rejected before launching a model.
func TestHarnessBindingBudget(t *testing.T) {
	for _, identity := range []string{"codex", "pi"} {
		t.Run(identity, func(t *testing.T) { checkHarnessBindingBudget(t, identity) })
	}
}

func checkHarnessBindingBudget(t *testing.T, identity string) {
	t.Helper()

	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.Policy.AgentTokenBudget = 100
	resolver := harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))
	request.ResolveHarness = func(ctx context.Context, config workspaceconfig.Config) (HarnessResolution, error) {
		resolved, err := resolver(ctx, config)
		if err != nil {
			return HarnessResolution{}, err
		}

		resolved.Binding.Harness = identity
		if identity == "pi" {
			resolved.Binding.RuntimeRef = "sha256:" + contenthash.Sum([]byte("pi-runtime"))
		}

		sealed, err := harnesscore.SealBinding(*resolved.Binding)

		return HarnessResolution{Binding: &sealed, Command: resolved.Command}, err
	}
	started := workflowTestStart(t, store, request)

	claim := workflowTestClaim(t, store, started.Invocation.ID)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, nil); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("binding bypassed finite whole-call authority", err)
	}

	retained, err := store.ReadWorkflow(t.Context(), started.Invocation.ID)
	if err != nil || len(retained.Attempts[0].Processes) != 0 || retained.Policy.AgentTokenBudget != 100 {
		t.Fatal("uncertified driver launched or changed its budget", err)
	}
}
