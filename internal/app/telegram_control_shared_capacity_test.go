package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramControlWorker struct {
	release  string
	signal   *os.File
	done     chan error
	finalize func(context.Context) error
}

// TestTelegramControlSharedCapacity holds actual source workers behind barriers.
// Two Telegram conversations, daemon activation and manual claims share the
// default serial pool and the configured maximum, including overflow and reuse.
func TestTelegramControlSharedCapacity(t *testing.T) {
	for _, capacity := range []int{1, 8} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) { checkTelegramControlSharedCapacity(t, capacity) })
	}
}

func checkTelegramControlSharedCapacity(t *testing.T, capacity int) {
	t.Helper()

	f, operations, record := telegramWorkAdmissionFixture(t, "ticket")
	if capacity > 1 {
		writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, fmt.Sprintf("ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  max_active_workers: %d\n", capacity))
	}

	first := telegramControlCapacityAdmission(t, f, record.Task.Key)
	runs, waiting := telegramControlCapacityRuns(t, f, operations, first, capacity)

	workers := []*telegramControlWorker{telegramControlSharedWorker(t, f.receiver, operations, first.Invocation.ID, &first, false)}
	for index, run := range runs {
		workers = append(workers, telegramControlSharedWorker(t, f.receiver, operations, run.ID, nil, index == 0))
	}

	checkTelegramControlPool(t, operations, capacity, 0)
	checkTelegramControlCapacityWait(t, f, operations, waiting)
	checkTelegramControlDuplicate(t, f, operations, first)

	finishTelegramControlWorker(t, workers[0])
	checkTelegramControlPool(t, operations, capacity, 1)

	workers[0] = telegramControlSharedWorker(t, f.receiver, operations, waiting.Invocation.ID, &waiting, false)
	checkTelegramControlPool(t, operations, capacity, 0)

	finishTelegramControlWorkers(t, workers)

	checkTelegramControlPool(t, operations, capacity, capacity)

	checkTelegramControlResults(t, operations, append([]string{first.Invocation.ID, waiting.Invocation.ID}, telegramControlRunIDs(runs)...))
}

func telegramControlCapacityAdmission(t *testing.T, f *telegramConversationAppFixture, key string) telegramcore.TelegramWorkReference {
	t.Helper()

	action, _ := telegramWorkOffer(t, f, "/work ticket "+key)
	f.send(t, "/answer "+action.Token+" start", 0)

	refs, err := f.receiver.store.TelegramWorkReferences(t.Context(), f.receiver.botID, "", 16)
	if err != nil {
		t.Fatal(err)
	}

	for _, ref := range refs {
		preview, err := telegramRetainedWorkPreview(ref.Request)
		if err != nil {
			t.Fatal(err)
		}

		if preview.Request.Selector.Identity == key {
			return ref
		}
	}

	t.Fatal("native conversation admission missing", key)

	return telegramcore.TelegramWorkReference{}
}

func telegramControlRunIDs(runs []core.WorkflowInvocation) []string {
	ids := make([]string, len(runs))
	for index, run := range runs {
		ids[index] = run.ID
	}

	return ids
}

func telegramControlCapacityRuns(t *testing.T, f *telegramConversationAppFixture, operations service, first telegramcore.TelegramWorkReference, capacity int) ([]core.WorkflowInvocation, telegramcore.TelegramWorkReference) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	native, err := operations.ReadWorkflow(t.Context(), first.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}

	original, err := store.Read(t.Context(), native.Members[0].Admission.Path)
	if err != nil {
		t.Fatal(err)
	}

	preview, err := telegramRetainedWorkPreview(first.Request)
	if err != nil {
		t.Fatal(err)
	}

	runs := []core.WorkflowInvocation{}
	waiting := telegramcore.TelegramWorkReference{}

	for index := 1; index <= capacity; index++ {
		task := *original.Task

		task.Key = fmt.Sprintf("PRJ-%d", 90+index)
		if _, err := store.Create(t.Context(), "Independent capacity task", core.WorkflowRecord{Kind: "ticket", Task: &task, Domain: original.Domain, Body: original.Body}); err != nil {
			t.Fatal(err)
		}

		if index == 1 {
			waiting = telegramControlCapacityAdmission(t, f, task.Key)

			continue
		}

		selection := preview.Request
		selection.Selector.Identity = task.Key

		selection.SnapshotHash = ""

		started, err := operations.StartWorkflow(t.Context(), core.WorkflowStartRequest{RequestKey: task.Key, Selection: selection})
		if err != nil {
			t.Fatal("admit independent capacity task", task.Key, err)
		}

		runs = append(runs, started.Invocation)
	}

	return runs, waiting
}

func telegramControlSharedWorker(t *testing.T, receiver telegramReceiver, operations service, id string, ref *telegramcore.TelegramWorkReference, daemon bool) *telegramControlWorker {
	t.Helper()

	root := t.TempDir()
	marker := filepath.Join(root, "started")
	worker := &telegramControlWorker{release: filepath.Join(root, "release"), done: make(chan error, 1)}
	worker.signal = telegramControlSourceSignal(t, worker.release)
	operations.workflowWorker = []string{"sh", "-c", `printf ready > "$1"; IFS= read -r permit < "$2"; [ "$permit" = release ] || exit 7; printf 'Delivered\n' > result.txt; git add result.txt; git -c user.name=Test -c user.email=test@example.com commit -qm 'Shared capacity result'`, "worker", marker, worker.release}
	receiver.operations = operations

	ctx, cancel := context.WithCancel(t.Context())

	run, finalize := telegramControlWorkerCall(t, receiver, operations, id, ref, daemon)
	worker.finalize = finalize

	t.Cleanup(func() { finishTelegramControlWorker(t, worker); cancel() })

	go func() { worker.done <- run(ctx); close(worker.done) }()

	telegramWorkWaitFile(t, marker)

	return worker
}

func telegramControlWorkerCall(t *testing.T, receiver telegramReceiver, operations service, id string, ref *telegramcore.TelegramWorkReference, daemon bool) (func(context.Context) error, func(context.Context) error) {
	t.Helper()

	if ref != nil {
		return func(ctx context.Context) error {
			result := receiver.activateWork(ctx, *ref)
			if result.Failure == nil && result.Progress.Steps != 1 {
				return fmt.Errorf("Telegram checkpoint steps: %d", result.Progress.Steps)
			}

			return result.Failure
		}, nil
	}

	if daemon {
		return telegramControlDaemonCall(t, operations, id), nil
	}

	return telegramControlManualCall(t, operations, id)
}

func telegramControlDaemonCall(t *testing.T, operations service, id string) func(context.Context) error {
	t.Helper()

	runner := daemonWorkflowFixture(t, &bytes.Buffer{})

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	return func(ctx context.Context) error {
		result := daemonForegroundLoopResult{}

		_, err := runner.workflow(ctx, operations, status, &result)
		if err == nil && result.WorkflowSteps != 1 {
			return fmt.Errorf("daemon checkpoint steps: %d", result.WorkflowSteps)
		}

		return err
	}
}

func telegramControlManualCall(t *testing.T, operations service, id string) (func(context.Context) error, func(context.Context) error) {
	t.Helper()

	claim, err := operations.ClaimWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	})

	task, err := operations.PrepareWorkflowManualTask(t.Context(), claim, operations.workflowWorker)
	if err != nil {
		t.Fatal(err)
	}

	execute := func(ctx context.Context) error {
		result, err := operations.ExecuteWorkflowManualTask(ctx, claim)
		if err == nil && result.ExitCode != 0 {
			return errors.New("manual source did not exit successfully")
		}

		return err
	}

	finalize := func(ctx context.Context) error {
		return reconcileTelegramControlManual(ctx, operations, claim, task)
	}

	return execute, finalize
}

func reconcileTelegramControlManual(ctx context.Context, operations service, claim *core.WorkflowClaim, task core.WorkflowManualTask) (err error) {
	defer func() { err = errors.Join(err, claim.Close()) }()

	output, err := exec.CommandContext(ctx, "git", "-C", task.Input.Lane.Worktree, "rev-parse", "HEAD", "HEAD^{tree}").Output()
	if err != nil {
		return err
	}

	identities := strings.Fields(string(output))
	if len(identities) != 2 {
		return errors.New("manual candidate commit/tree missing")
	}

	candidate := core.WorkflowManualCandidate{InputHash: task.InputHash, Branch: task.Input.Lane.Branch, Commit: identities[0], Tree: identities[1]}

	result, err := operations.ReconcileWorkflowTask(ctx, claim, candidate)
	if err == nil && (!result.Successful || result.ExecutionMode != "manual") {
		return errors.New("manual source did not validate")
	}

	return err
}

func finishTelegramControlWorker(t *testing.T, worker *telegramControlWorker) {
	t.Helper()

	releaseTelegramControlWorker(t, worker)
	waitTelegramControlWorker(t, worker)
	reconcileTelegramControlWorker(t, worker)
}

func waitTelegramControlWorker(t *testing.T, worker *telegramControlWorker) {
	t.Helper()

	select {
	case err := <-worker.done:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(30 * time.Second):
		t.Error("owned shared-capacity worker did not finish")
	}
}

func reconcileTelegramControlWorker(t *testing.T, worker *telegramControlWorker) {
	t.Helper()

	if worker.finalize == nil {
		return
	}

	finalize := worker.finalize
	worker.finalize = nil

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 30*time.Second)
	defer cancel()

	if err := finalize(ctx); err != nil {
		t.Error(err)
	}
}

func checkTelegramControlPool(t *testing.T, operations service, maximum, available int) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	capacity, err := store.QueryWorkflowIntegrationCapacity(t.Context())
	if err != nil || capacity.Maximum != maximum || capacity.Available != available {
		t.Fatal("adapters did not share canonical capacity", capacity, err)
	}
}

func checkTelegramControlCapacityWait(t *testing.T, f *telegramConversationAppFixture, operations service, ref telegramcore.TelegramWorkReference) {
	t.Helper()

	result := f.receiver.activateWork(t.Context(), ref)
	if result.Failure != nil || result.Progress.Attempted || result.Progress.Steps != 0 {
		t.Fatal("another conversation bypassed full capacity", result)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), ref.Invocation.ID)
	if claim != nil {
		t.Error(claim.Close())
	}

	if !errors.Is(err, core.ErrWorkflowCapacity) {
		t.Fatal("manual adapter bypassed full capacity", err)
	}

	checkTelegramControlDaemonWait(t, operations, ref.Invocation.ID)

	run, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil || len(run.Attempts) != 0 || run.Continuation != nil {
		t.Fatal("waiting consumed an activation or lease", run, err)
	}
}

func checkTelegramControlDuplicate(t *testing.T, f *telegramConversationAppFixture, operations service, ref telegramcore.TelegramWorkReference) {
	t.Helper()

	result := f.receiver.activateWork(t.Context(), ref)
	if (!errors.Is(result.Failure, core.ErrWorkflowOwned) && !errors.Is(result.Failure, core.ErrWorkflowCapacity)) || result.Progress.Steps != 0 {
		t.Fatal("duplicate Telegram activation bypassed native owner", result.Progress.Steps, result.Failure)
	}

	run, err := operations.ReadWorkflow(t.Context(), ref.Invocation.ID)
	if err != nil || len(run.Attempts) != 1 || run.Cursor != 0 {
		t.Fatal("duplicate activation created another attempt", run, err)
	}
}

func checkTelegramControlDaemonWait(t *testing.T, operations service, id string) {
	t.Helper()

	runner := daemonWorkflowFixture(t, &bytes.Buffer{})

	status, err := operations.WorkflowStatus(t.Context(), id, WorkflowStatusRequest{Limit: 8})
	if err != nil {
		t.Fatal(err)
	}

	observed := daemonForegroundLoopResult{}
	if _, err := runner.workflow(t.Context(), operations, status, &observed); err != nil || observed.WorkflowSteps != 0 {
		t.Fatal("daemon bypassed full capacity", observed, err)
	}
}

func releaseTelegramControlWorker(t *testing.T, worker *telegramControlWorker) {
	t.Helper()

	if _, err := worker.signal.WriteString("release\n"); err != nil {
		t.Error(err)
	}
}

func finishTelegramControlWorkers(t *testing.T, workers []*telegramControlWorker) {
	t.Helper()

	// The manual adapter retains its claim after source execution. Let all
	// source groups exit before reconciling their exact candidates through the
	// unchanged host gates. Source overlap and every final proof remain asserted.
	for _, worker := range workers {
		releaseTelegramControlWorker(t, worker)
	}

	for _, worker := range workers {
		waitTelegramControlWorker(t, worker)
	}

	for _, worker := range workers {
		reconcileTelegramControlWorker(t, worker)
	}
}

func telegramControlSourceSignal(t *testing.T, path string) *os.File {
	t.Helper()

	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}

	// Keep one host endpoint open until cleanup so an early permit remains
	// buffered. Native sources block in read rather than spawning sleep loops.
	file, err := os.OpenFile(path, os.O_RDWR|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	})

	return file
}

func checkTelegramControlResults(t *testing.T, operations service, ids []string) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	runs, next, err := store.InspectWorkflowPage(t.Context(), 0, 16)
	if err != nil || next != -1 || len(runs) != len(ids) {
		t.Fatal("shared activation lost an invocation", len(runs), next, err)
	}

	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}

	for _, run := range runs {
		if !wanted[run.ID] {
			t.Fatal("shared activation duplicated or replaced an invocation", run.ID)
		}

		checkTelegramControlResult(t, run)
		delete(wanted, run.ID)
	}
}

func checkTelegramControlResult(t *testing.T, run core.WorkflowInvocation) {
	t.Helper()

	if run.Cursor != 1 || len(run.Attempts) != 1 || run.Attempts[0].Outcome != "validated" {
		t.Fatal("shared activation lost or duplicated its checkpoint", run.ID)
	}
}
