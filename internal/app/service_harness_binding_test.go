package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Actual service admissions freeze the selected executable/model before claim.
// Rebuilding the service and changing private profiles cannot rebind a retry.
func TestHarnessBindingService(t *testing.T) {
	operations, hierarchy, target := serviceWorkflowTaskFixture(t)
	profiles := prepareServiceHarnessBinding(t, operations)
	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "profile-worker",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil || started.Invocation.HarnessBinding == nil {
		t.Fatal("service did not bind explicit selection", err)
	}

	original := *started.Invocation.HarnessBinding
	if original.Model != "fixture-model" || original.Profile != "writing" || original.ModelProvider != "cli_default" {
		t.Fatal("service fabricated or lost model identity", original)
	}

	harnessSelectionWrite(t, profiles, "profiles:\n  writing:\n    harness: codex\n    command: /unavailable\n    model: changed-model\n")

	restarted := newService(t.Context(), operations.env, operations.projectContext, testAssetsFS())

	replay, err := restarted.StartWorkflow(t.Context(), request)
	if err != nil || !replay.Attached || *replay.Invocation.HarnessBinding != original {
		t.Fatal("service retry rebound to changed setup", err)
	}

	checkServiceBoundWorker(t, restarted, started.Invocation)

	request.RequestKey = "fresh-operation"

	request.Selection.Selector.Identity = hierarchy.Tasks[len(hierarchy.Tasks)-1].Task.Key
	if _, err := restarted.StartWorkflow(t.Context(), request); err == nil {
		t.Fatal("new admission ignored current readiness")
	}
}

func prepareServiceHarnessBinding(t *testing.T, operations service) string {
	t.Helper()

	driver := filepath.Join(t.TempDir(), "selected-codex")
	script := "#!/bin/sh\n[ \"$1\" = exec ] || exit 71\n[ \"$5\" = fixture-model ] || exit 72\ncat > observed-input.json\nprintf '%s\\n' \"$@\" > observed-argv.txt\nprintf 'Delivered\\n' > result.txt\ngit add result.txt observed-input.json observed-argv.txt\ngit -c user.name=Test -c user.email=test@example.com commit -qm 'Selected worker result'\n"
	harnessSelectionWrite(t, driver, script)

	if err := os.Chmod(driver, 0o700); err != nil {
		t.Fatal(err)
	}

	profiles := filepath.Join(operations.env.ConfigDir(), "harnesses.yml")
	harnessSelectionWrite(t, profiles, "profiles:\n  writing:\n    harness: codex\n    command: "+driver+"\n    model: fixture-model\n")
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nharnesses:\n  worker: writing\n")

	return profiles
}

func checkServiceBoundWorker(t *testing.T, operations service, invocation workflowstate.Invocation) {
	t.Helper()

	public, err := json.Marshal(invocation)
	if err != nil || bytes.Contains(public, []byte("selected-codex")) {
		t.Fatal("public projection exposed private command", err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	result, err := operations.ExecuteWorkflowTask(t.Context(), claim, nil)
	if err != nil || !result.Successful {
		t.Fatal("selected worker execution failed", result.Failure, err)
	}

	retained, err := operations.ReadWorkflow(t.Context(), invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	checkServiceBoundWorkerInput(t, retained.Lane.Worktree, invocation.HarnessBinding.BindingHash)
}

func checkServiceBoundWorkerInput(t *testing.T, worktree, hash string) {
	t.Helper()

	input, err := os.ReadFile(filepath.Join(worktree, "observed-input.json"))
	if err != nil || !bytes.Contains(input, []byte(`"schema": "chr-task-input/v2"`)) || !bytes.Contains(input, []byte(hash)) {
		t.Fatal("worker did not receive frozen identity", err)
	}

	argv, err := os.ReadFile(filepath.Join(worktree, "observed-argv.txt"))
	if err != nil || !strings.Contains(string(argv), "--model\nfixture-model\n") {
		t.Fatal("worker lost its model selection", string(argv), err)
	}
}

// An explicit Codex provider without adapter-owned dispatch support must fail
// readiness rather than appear frozen while the CLI silently chooses a default.
func TestHarnessBindingProviderReadiness(t *testing.T) {
	operations, project, configDir, command := harnessSelectionFixture(t)
	harnessSelectionWrite(t, filepath.Join(configDir, "harnesses.yml"), "profiles:\n  writing:\n    harness: codex\n    command: "+command+"\n    model_provider: unsupported-provider\n    model: exact-model\n")
	harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  worker: writing\n")

	view, err := operations.ResolveHarnessSelection(t.Context(), harnesscore.Worker, "")
	if err == nil || view.Readiness.Ready() {
		t.Fatal("unimplemented provider dispatch appeared ready")
	}
}

// Legacy worker command semantics remain CLI-default even when generation has
// an explicit environment model. Inspection must not invent a frozen worker model.
func TestHarnessBindingLegacyWorkerModel(t *testing.T) {
	operations, _, _, _ := harnessSelectionFixture(t)
	t.Setenv("CHR_CODEX_MODEL", "generation-only-model")

	view, err := operations.ResolveHarnessSelection(t.Context(), harnesscore.Worker, "")
	if err != nil || view.Model != "cli_default" || view.Source != "legacy" {
		t.Fatal("inspection misrepresented legacy worker model", view, err)
	}
}
