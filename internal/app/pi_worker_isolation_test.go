package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Actual Pi workers with different credentials reach a shared endpoint barrier
// together. They must retain private state while occupying the same repository
// pool used by Codex and native operations; host Git validation remains decisive.
func TestPiWorkerIsolationContainer(t *testing.T) {
	image := os.Getenv("CHR_TEST_PI_RUNTIME_IMAGE")
	if image == "" {
		t.Skip("set CHR_TEST_PI_RUNTIME_IMAGE to the prepared Pi image")
	}

	t.Setenv("OPENAI_API_KEY", "host-fixture-must-not-reach-the-worker")

	bridge := buildPiWorkerFixture(t)
	operations, hierarchy, _ := serviceWorkflowTaskFixture(t)
	target := preparePiProjectContext(t, operations)
	endpoint := piWorkerEndpoint(t, true)

	configurePiWorkerProfile(t, operations, bridge, image, endpoint.URL, "worker-fixture-first", 2)

	first := startPiIsolationInvocation(t, operations, hierarchy.Tasks[0].Task.Key, target)
	configurePiWorkerProfile(t, operations, bridge, image, endpoint.URL, "worker-fixture-second", 2)

	secondKey := clonePiIsolationTicket(t, operations, hierarchy.Tasks[0], "PRJ-91")
	second := startPiIsolationInvocation(t, operations, secondKey, target)

	workers := []<-chan piIsolationResult{startPiIsolationWorker(t, operations, first.ID), startPiIsolationWorker(t, operations, second.ID)}
	endpoint.wait(t, 2)
	checkPiIsolationCapacity(t, operations, 0)
	checkPiContainerMounts(t, operations, image, []string{first.ID, second.ID})
	checkPiIsolationOverflow(t, operations, hierarchy.Tasks[0], target)
	endpoint.release(t)

	states := map[string]bool{}
	labels := map[string]bool{}

	for _, worker := range workers {
		result := <-worker
		if result.err != nil || !result.task.Successful || len(result.task.Validation) != 1 {
			t.Fatalf("isolated worker failed host validation: %v, %s, worker=%+v", result.err, result.task.Failure, result.task.Worker)
		}

		checkPiIsolationResult(t, operations, result.task, states, labels)
	}

	if len(states) != 2 || !labels["CALL=first"] || !labels["CALL=second"] || endpoint.calls(t) != 4 {
		t.Fatal("concurrent workers shared state, credentials or model results")
	}

	checkPiIsolationCapacity(t, operations, 2)
}

func preparePiProjectContext(t *testing.T, operations service) string {
	t.Helper()

	project := operations.projectContext.ProjectDir
	harnessSelectionWrite(t, filepath.Join(project, "AGENTS.md"), "Explicit project fixture instructions.\n")
	harnessSelectionWrite(t, filepath.Join(project, ".pi", "settings.json"), `{"defaultProvider":"POISON_WORKER_CONTEXT","defaultModel":"POISON_WORKER_CONTEXT"}`)
	harnessSelectionWrite(t, filepath.Join(project, ".pi", "extensions", "poison.mjs"), `throw new Error("POISON_WORKER_CONTEXT");`)

	command := exec.CommandContext(t.Context(), "git", "-C", project, "add", "AGENTS.md", ".pi")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("prepare owned project context: %v: %s", err, output)
	}

	command = exec.CommandContext(t.Context(), "git", "-C", project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Fixture project context")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("commit owned project context: %v: %s", err, output)
	}

	output, err := exec.CommandContext(t.Context(), "git", "-C", project, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(output))
}

func clonePiIsolationTicket(t *testing.T, operations service, original workflowstate.Record, key string) string {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	task := *original.Task
	task.Key = key
	task.SetID, task.SliceID, task.UseCase, task.Order = "", "", "", 0

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(original.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	task.ContractHash = authoringcore.WorkflowContentHash(append([]byte("\x00\x00"), data...))
	if _, err := store.Create(t.Context(), "Independent Pi isolation task", workflowstate.Record{Kind: "ticket", Task: &task, Domain: map[string]json.RawMessage{"execution_contract": data}, Body: original.Body}); err != nil {
		t.Fatal(err)
	}

	return key
}

func startPiIsolationInvocation(t *testing.T, operations service, key, target string) workflowcore.Invocation {
	t.Helper()

	request := workflowcore.StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: key,
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}},
	}}

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return started.Invocation
}

type piIsolationResult struct {
	task workflowstate.TaskResult
	err  error
}

func startPiIsolationWorker(t *testing.T, operations service, id string) <-chan piIsolationResult {
	t.Helper()

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	done := make(chan piIsolationResult, 1)

	t.Cleanup(func() {
		cancel()

		select {
		case <-done:
		case <-time.After(25 * time.Second):
			t.Error("owned Pi isolation worker did not stop")
		}
	})

	go func() {
		defer close(done)

		task, err := operations.ExecuteWorkflowTask(ctx, claim, nil)
		done <- piIsolationResult{task: task, err: errors.Join(err, claim.Close())}
	}()

	return done
}

func checkPiIsolationCapacity(t *testing.T, operations service, available int) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	capacity, err := store.QueryWorkflowIntegrationCapacity(t.Context())
	if err != nil || capacity.Maximum != 2 || capacity.Available != available {
		t.Fatal("Pi did not use the repository capacity boundary", capacity, err)
	}
}

func checkPiIsolationOverflow(t *testing.T, operations service, original workflowstate.Record, target string) {
	t.Helper()

	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), "profiles:\n  worker:\n    harness: codex\n    command: sh\n")

	key := clonePiIsolationTicket(t, operations, original, "PRJ-92")

	run := startPiIsolationInvocation(t, operations, key, target)
	if run.HarnessBinding == nil || run.HarnessBinding.Harness != "codex" {
		t.Fatal("overflow fixture did not retain a Codex selection")
	}

	claim, err := operations.ClaimWorkflow(t.Context(), run.ID)
	if claim != nil {
		t.Error(claim.Close())
	}

	if !errors.Is(err, workflowcore.ErrWorkflowCapacity) {
		t.Fatal("Codex bypassed occupied Pi slots", err)
	}

	lock, err := workflowcore.AcquireAgentOperationLock(t.Context(), workflowcore.AgentOperationLockOptions{ProjectDir: operations.projectContext.ProjectDir, Operation: workflowcore.AgentOperationDevelopment, Owner: "native-fixture"})
	if lock != nil {
		t.Error(lock.Release())
	}

	if !errors.Is(err, workflowcore.ErrWorkflowCapacity) {
		t.Fatal("native operation bypassed occupied Pi slots", err)
	}

	retained, err := operations.ReadWorkflow(t.Context(), run.ID)
	if err != nil || len(retained.Attempts) != 0 {
		t.Fatal("overflow consumed a worker attempt", err)
	}
}

func checkPiIsolationResult(t *testing.T, operations service, result workflowstate.TaskResult, states, labels map[string]bool) {
	t.Helper()

	text := result.Validation[0].Stdout
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "STATE=") {
			states[line] = true
		}
	}

	output, err := exec.CommandContext(t.Context(), "git", "-C", operations.projectContext.ProjectDir, "show", result.Commit+":call.txt").Output()
	if err != nil {
		t.Fatal(err)
	}

	labels[strings.TrimSpace(string(output))] = true
}

func piIsolationInput(t *testing.T, operations service, id string) workflowcore.TaskInput {
	t.Helper()

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || len(run.Attempts) != 1 {
		t.Fatal("missing owned worker attempt", err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(layout.Root(), run.Attempts[0].InputRef))
	if err != nil {
		t.Fatal(err)
	}

	var input workflowcore.TaskInput
	if json.Unmarshal(data, &input) != nil {
		t.Fatal("invalid owned worker input")
	}

	return input
}

type piContainerInspection struct {
	Mounts []struct {
		Source, Destination string
		RW                  bool
	}
	Config struct {
		WorkingDir string
		Entrypoint []string
	}
	HostConfig struct {
		NetworkMode    string
		Privileged     bool
		ReadonlyRootfs bool
		CapDrop        []string
		SecurityOpt    []string
		Tmpfs          map[string]string
	}
}

func checkPiContainerMounts(t *testing.T, operations service, image string, runs []string) {
	t.Helper()

	wanted := map[string]bool{}
	for _, id := range runs {
		wanted[piIsolationInput(t, operations, id).Lane.Worktree] = true
	}

	ids := strings.Fields(dockerPiFixture(t, "ps", "--filter", "ancestor="+image, "--filter", "name=chr-pi-worker-", "--format", "{{.ID}}"))
	if len(ids) == 0 {
		t.Fatal("owned Pi containers missing at endpoint barrier")
	}

	var containers []piContainerInspection
	if json.Unmarshal([]byte(dockerPiFixture(t, append([]string{"inspect", "--type", "container"}, ids...)...)), &containers) != nil {
		t.Fatal("invalid container inspection")
	}

	for _, container := range containers {
		if !wanted[container.Config.WorkingDir] {
			continue
		}

		assertPiContainerBoundary(t, container)
		delete(wanted, container.Config.WorkingDir)
	}

	if len(wanted) != 0 {
		t.Fatal("both owned Pi containers must be inspectable while active")
	}
}

func assertPiContainerBoundary(t *testing.T, container piContainerInspection) {
	t.Helper()

	host := container.HostConfig
	if host.NetworkMode == "host" || host.Privileged || !host.ReadonlyRootfs || len(host.Tmpfs) != 2 {
		t.Fatal("worker runtime changed the intended container isolation")
	}

	if !slices.Equal(host.CapDrop, []string{"ALL"}) || !slices.Equal(host.SecurityOpt, []string{"no-new-privileges"}) || !slices.Equal(container.Config.Entrypoint, []string{"/usr/local/bin/chr-worker-bridge"}) {
		t.Fatal("worker runtime changed its owned entrypoint or process restrictions")
	}

	assertPiResourceMounts(t, container)
}

func assertPiResourceMounts(t *testing.T, container piContainerInspection) {
	t.Helper()

	if len(container.Mounts) != 7 {
		t.Fatal("worker changed the explicit mount boundary")
	}

	for _, mount := range container.Mounts {
		switch {
		case mount.Destination == "/chr-auth/auth.json", mount.Destination == "/chr-models.json", mount.Destination == "/usr/local/bin/chr-worker-bridge", strings.HasSuffix(mount.Destination, ".lock"):
			if mount.RW {
				t.Fatal("immutable selected resource was writable")
			}
		default:
			if mount.Source != mount.Destination || !mount.RW {
				t.Fatal("worker checkout or Git metadata mount changed")
			}
		}
	}
}
