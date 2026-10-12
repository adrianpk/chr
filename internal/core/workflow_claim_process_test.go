package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type workflowClaimProcessResult struct {
	Claimed bool
	Owned   bool
}

// TestWorkflowClaimProcess is the owned subprocess fixture for exclusion and lost cleanup.
func TestWorkflowClaimProcess(t *testing.T) {
	dir := os.Getenv("CHR_CLAIM_DIR")
	if dir == "" {
		return
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	configureWorkflowClaimCrash(t, store)

	claim, err := store.ClaimWorkflow(t.Context(), os.Getenv("CHR_CLAIM_ID"))
	result := workflowClaimProcessResult{Claimed: err == nil, Owned: errors.Is(err, ErrWorkflowOwned) || errors.Is(err, ErrWorkflowCapacity)}

	writeWorkflowClaimProcessResult(t, result)

	if err != nil && !result.Owned {
		t.Fatal(err)
	}

	if claim != nil {
		waitWorkflowClaimFile(t, os.Getenv("CHR_CLAIM_RELEASE"))

		if err := store.CheckpointWorkflow(t.Context(), claim); err != nil {
			t.Fatal(err)
		}
	}
}

func configureWorkflowClaimCrash(t *testing.T, store *WorkflowStore) {
	t.Helper()

	if stop := os.Getenv("CHR_CLAIM_CRASH"); stop != "" {
		index, err := strconv.Atoi(stop)
		if err != nil {
			t.Fatal(err)
		}

		store.afterReplace = func(replaced int) error {
			if replaced == index {
				os.Exit(17)
			}

			return nil
		}
	}
}

func writeWorkflowClaimProcessResult(t *testing.T, result workflowClaimProcessResult) {
	t.Helper()

	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}

	if err := os.WriteFile(os.Getenv("CHR_CLAIM_RESULT"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func workflowClaimProcess(t *testing.T, store *WorkflowStore, id, result, release, crash string) *exec.Cmd {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowClaimProcess$")

	command.Env = append(os.Environ(), "CHR_CLAIM_DIR="+store.layout.ProjectDir(), "CHR_CLAIM_ID="+id, "CHR_CLAIM_RESULT="+result, "CHR_CLAIM_RELEASE="+release, "CHR_CLAIM_CRASH="+crash)

	t.Cleanup(func() {
		if command.Process != nil && command.ProcessState == nil {
			if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				t.Error(err)
			}
			// The owned fixture is stopped on test failure, then reaped.
			var exited *exec.ExitError
			if err := command.Wait(); err != nil && !errors.As(err, &exited) {
				t.Error(err)
			}
		}
	})

	return command
}

func waitWorkflowClaimFile(t *testing.T, path string) {
	t.Helper()

	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()

	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		if _, err := os.Stat(path); err == nil {
			return
		}

		select {
		case <-deadline.C:
			t.Fatal("owned claim fixture did not reach its checkpoint")
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case <-tick.C:
		}
	}
}

// TestWorkflowClaimProcessRace admits one subprocess and retains one durable attempt.
func TestWorkflowClaimProcessRace(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	release := filepath.Join(fixture, "release")
	commands := []*exec.Cmd{}

	paths := []string{filepath.Join(fixture, "first.json"), filepath.Join(fixture, "second.json")}
	for _, path := range paths {
		command := workflowClaimProcess(t, store, run.Invocation.ID, path, release, "")
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}

		commands = append(commands, command)
	}

	claimed := 0

	for _, path := range paths {
		waitWorkflowClaimFile(t, path)

		result := readWorkflowClaimProcessResult(t, path)

		if result.Claimed {
			claimed++
		} else if !result.Owned {
			t.Fatal("losing subprocess did not observe ownership")
		}
	}

	finishWorkflowClaimProcesses(t, commands, release)

	retained, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
	if claimed != 1 || err != nil || len(retained.Attempts) != 1 || retained.Status != "paused" {
		t.Fatalf("race = %d owners, %+v, %v", claimed, retained, err)
	}
}

func readWorkflowClaimProcessResult(t *testing.T, path string) workflowClaimProcessResult {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	result := workflowClaimProcessResult{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	return result
}

func finishWorkflowClaimProcesses(t *testing.T, commands []*exec.Cmd, release string) {
	t.Helper()

	if err := os.WriteFile(release, []byte("checkpoint\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWorkflowClaimJournal recovers every partial claim without admitting another worker.
func TestWorkflowClaimJournal(t *testing.T) {
	for _, index := range []int{0, 2} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			store, hierarchy, request := workflowClaimFixture(t)
			run := workflowTestStart(t, store, request)
			command := workflowClaimProcess(t, store, run.Invocation.ID, "", "", strconv.Itoa(index))
			output, err := command.CombinedOutput()

			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 17 {
				t.Fatalf("owned crash: %v: %s", err, output)
			}

			recovered, err := store.ReadWorkflow(t.Context(), run.Invocation.ID)
			if err != nil || recovered.Status != "running" || len(recovered.Attempts) != 1 {
				t.Fatalf("recovery: %+v, %v", recovered, err)
			}

			checkRecoveredWorkflowClaim(t, store, hierarchy.Tasks[0].Path, recovered)

			if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowOwned) {
				t.Fatal("unreconciled crash admitted another worker")
			}
		})
	}
}

func checkRecoveredWorkflowClaim(t *testing.T, store *WorkflowStore, path string, recovered WorkflowInvocation) {
	t.Helper()

	record, err := store.Read(t.Context(), path)
	if err != nil || record.Task.AttemptID != recovered.Attempts[0].ID || record.Revision != recovered.Members[0].ExpectedRevision || record.SourceHash != recovered.Members[0].ExpectedSourceHash {
		t.Fatal("partial claim was not replayed as one image")
	}
}

// TestWorkflowClaimLinkedRoots shares attachments and capacity with a newly added worktree.
func TestWorkflowClaimLinkedRoots(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	run := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, run.Invocation.ID)
	dir := filepath.Join(t.TempDir(), "linked")
	workflowSelectionGit(t, store, "worktree", "add", "-qb", "linked", dir, "HEAD")

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	linked := NewWorkflowStore(layout, "project-alias")

	attached := workflowTestStart(t, linked, request)
	if !attached.Attached || attached.Invocation.ID != run.Invocation.ID {
		t.Fatal("linked root allocated duplicate ownership")
	}

	if _, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: dir, Operation: AgentOperationDevelopment, Owner: "legacy"}); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("new legacy root bypassed serial worker: %v", err)
	}

	if err := linked.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}

	legacy, err := AcquireAgentOperationLock(t.Context(), AgentOperationLockOptions{ProjectDir: dir, Operation: AgentOperationDevelopment, Owner: "legacy"})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ClaimWorkflow(t.Context(), run.Invocation.ID); !errors.Is(err, ErrWorkflowCapacity) {
		t.Fatalf("native claim bypassed legacy worker: %v", err)
	}

	if err := legacy.Release(); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowClaimForeignClone refuses copied active ownership without the registered common directory.
func TestWorkflowClaimForeignClone(t *testing.T) {
	store, _, request := workflowClaimFixture(t)
	workflowTestStart(t, store, request)
	workflowSelectionGit(t, store, "add", "ops")
	workflowSelectionCommit(t, store)

	dir := filepath.Join(t.TempDir(), "clone")

	command := exec.CommandContext(t.Context(), "git", "clone", "-q", store.layout.ProjectDir(), dir)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("clone fixture: %v: %s", err, output)
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	before := nativeQueueImages(t, layout.Root())
	if _, err := NewWorkflowStore(layout, "project-1").StartWorkflow(t.Context(), request); !errors.Is(err, ErrWorkflowCoordinator) {
		t.Fatalf("foreign clone reused active ownership: %v", err)
	}

	if !reflect.DeepEqual(before, nativeQueueImages(t, layout.Root())) {
		t.Fatal("foreign clone rejection changed copied authority")
	}
}
