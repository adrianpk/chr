package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

// Admission and switching share exclusion across their entire configuration
// observation: a paused resolver cannot mix old policy with new role choices.
func TestHarnessBindingSelectionOrder(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	project := store.layout.ProjectDir()
	before := "harnesses:\n  worker: alpha\n  generation: alpha\n"
	writeWorkspaceConfig(t, project, before)

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	entered, release, switching := make(chan struct{}), make(chan struct{}), make(chan struct{})
	resolver := harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))
	request.ResolveHarness = func(ctx context.Context, config WorkspaceConfig) (WorkflowHarnessResolution, error) {
		close(entered)

		select {
		case <-ctx.Done():
			return WorkflowHarnessResolution{}, ctx.Err()
		case <-release:
			return resolver(ctx, config)
		}
	}
	admitted := make(chan harnessBindingAdmissionResult, 1)

	go func() {
		result, err := store.StartWorkflow(ctx, request)
		admitted <- harnessBindingAdmissionResult{result, err}
	}()

	waitHarnessBindingSignal(t, ctx, entered)

	options := harnessSelectionTestOptions(harnessSelectionTestRequest("switch-after-admission", contenthash.Sum([]byte(before)), "beta"))
	options.Request.ProjectRef = store.projectID
	once := sync.Once{}
	options.Authorize = func(context.Context) error { once.Do(func() { close(switching) }); return nil }
	changed := make(chan harnessBindingChangeResult, 1)

	go func() {
		result, err := ChangeHarnessSelections(ctx, project, options)
		changed <- harnessBindingChangeResult{result, err}
	}()

	waitHarnessBindingSignal(t, ctx, switching)
	close(release)

	admission, change := <-admitted, <-changed
	if admission.err != nil || change.err != nil {
		t.Fatal("ordered operations failed", admission.err, change.err)
	}

	checkHarnessBindingAdmissionImage(t, admission.result.Invocation, before)

	binding := admission.result.Invocation.HarnessBinding

	if change.result.Worker != "beta" || change.result.Generation != "beta" || change.result.Revision == binding.SelectionRevision {
		t.Fatal("later switch did not publish both new defaults")
	}
}

func checkHarnessBindingAdmissionImage(t *testing.T, run WorkflowInvocation, before string) {
	t.Helper()

	binding := run.HarnessBinding
	if binding == nil || binding.Profile != "alpha" || binding.SelectionRevision != contenthash.Sum([]byte(before)) || run.Policy.ConfigRevision != binding.SelectionRevision {
		t.Fatal("admission combined different configuration images")
	}
}

type harnessBindingAdmissionResult struct {
	result WorkflowStartResult
	err    error
}
type harnessBindingChangeResult struct {
	result harnesscore.SelectionSnapshot
	err    error
}

func waitHarnessBindingSignal(t *testing.T, ctx context.Context, signal <-chan struct{}) {
	t.Helper()

	select {
	case <-signal:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

// Manual handoff retains mandatory identity without turning an omitted command
// into an automatic model call. Legacy manual behavior stays explicit.
func TestHarnessBindingManual(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.ResolveHarness = harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))
	started := workflowTestStart(t, store, request)
	claim := workflowTestClaim(t, store, started.Invocation.ID)

	task, err := store.PrepareWorkflowManualTask(t.Context(), claim, nil)
	if err != nil || task.Input.Schema != "chr-task-input/v2" || len(task.Input.Worker) != 0 || !sameHarnessBinding(task.Input.HarnessBinding, started.Invocation.HarnessBinding) {
		t.Fatal("manual handoff lost its explicit mode", err)
	}

	if _, err := store.ExecuteWorkflowManualTask(t.Context(), claim); err == nil {
		t.Fatal("empty manual command launched a worker")
	}
}
