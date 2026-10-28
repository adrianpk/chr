package workflow

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeAgentFaultMode(t *testing.T, mode string) {
	t.Helper()

	if mode != "crash" && mode != "timeout" && mode != "hold" {
		return
	}

	marker := os.Getenv("CHR_NATIVE_AGENT_STARTED")
	if marker != "" {
		if err := os.WriteFile(marker, []byte("driver started"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	switch mode {
	case "crash":
		os.Exit(23)
	case "timeout":
		time.Sleep(30 * time.Second)
		os.Exit(24)
	case "hold":
		waitWorkflowClaimFile(t, os.Getenv("CHR_NATIVE_AGENT_RELEASE"))
		os.Exit(0)
	}
}

// TestNativeAgentFaults uses actual driver exit, deadline and malformed usage.
// Restart retains stopped birth evidence and the entire unknown grant.
func TestNativeAgentFaults(t *testing.T) {
	for _, mode := range []string{"crash", "timeout", "missing"} {
		t.Run(mode, func(t *testing.T) {
			store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)

			request.Policy.AgentTokenBudget = 100
			if mode == "timeout" {
				request.Policy.WorkerTimeoutSeconds = 3
			}

			run := workflowTestStart(t, store, request)
			harness := nativeAgentHarness(t, store, run.Invocation.ID, mode)
			claim := workflowTestClaim(t, store, run.Invocation.ID)

			result, err := store.ExecuteWorkflowTaskWithHarness(t.Context(), claim, harness)
			if err == nil || result.Successful {
				t.Fatal("fault produced successful task", err)
			}

			if err := claim.Close(); err != nil {
				t.Fatal(err)
			}

			restarted := NewWorkflowStore(store.layout, store.projectID)
			nativeAgentStoppedCharge(t, restarted, run.Invocation.ID)
			nativeAgentNoRefund(t, restarted, run.Invocation.ID, harness)
		})
	}
}

func nativeAgentStoppedCharge(t *testing.T, store *Store, id string) {
	t.Helper()

	run := workflowTestRead(t, store, id)
	nativeAgentStoppedEvidence(t, run.Attempts[0])

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || budget.TokensReserved != 100 || budget.TokensSettled != 0 || budget.Usage.WorkerCount != 1 || budget.Reservations[0].State != "unknown" {
		t.Fatal("failed call was refunded", err)
	}
}

func nativeAgentStoppedEvidence(t *testing.T, attempt workflowstate.Attempt) {
	t.Helper()

	if len(attempt.Processes) != 1 || workflowAttemptQuiescent(attempt) != nil || attempt.Processes[0].State != workflowstate.ProcessStopped {
		t.Fatal("actual driver birth did not retain quiescent ownership")
	}

	if attempt.Processes[0].StdoutHash == "" || attempt.Processes[0].StderrHash == "" {
		t.Fatal("stopped process evidence missing")
	}
}

func nativeAgentNoRefund(t *testing.T, store *Store, id string, harness AgentHarness) {
	t.Helper()

	before, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ResumeWorkflow(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	claim := workflowTestClaim(t, store, id)
	defer func() { _ = claim.Close() }() //nolint:errcheck

	if _, err := store.ExecuteWorkflowTaskWithHarness(t.Context(), claim, harness); !errors.Is(err, ErrWorkflowIntegrationBudget) {
		t.Fatal("restart launched an uncharged replacement", err)
	}

	after, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || after.TokensReserved != before.TokensReserved || after.TokensSettled != before.TokensSettled || !reflect.DeepEqual(after.Reservations, before.Reservations) {
		t.Fatal("rejected replacement reset consumption", err)
	}
}

// TestNativeAgentHost is an owned host-loss fixture. It exits before a permit
// or after the actual certified driver starts, without stopping that driver.
func TestNativeAgentHost(t *testing.T) {
	dir := os.Getenv("CHR_NATIVE_AGENT_HOST")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	id := os.Getenv("CHR_NATIVE_AGENT_INVOCATION")
	harness := nativeAgentHarness(t, store, id, "hold")
	claim := workflowTestClaim(t, store, id)
	nativeAgentHostCrash(t, claim.store, id)

	if _, err := store.ExecuteWorkflowTaskWithHarness(t.Context(), claim, harness); err != nil {
		t.Fatal(err)
	}

	t.Fatal("host-loss boundary was not reached")
}

func nativeAgentHostCrash(t *testing.T, store *Store, id string) {
	t.Helper()

	if os.Getenv("CHR_NATIVE_AGENT_BOUNDARY") == "before permit" {
		store.afterReplace = func(_ int) error {
			run, err := store.readWorkflowInvocation(id)
			if err != nil {
				return err
			}

			if len(run.Attempts[0].Processes) > 0 {
				os.Exit(29)
			}

			return nil
		}

		return
	}
	go func() { waitWorkflowClaimFile(t, os.Getenv("CHR_NATIVE_AGENT_STARTED")); os.Exit(29) }()
}

func nativeAgentCrashHost(t *testing.T, store *Store, id, boundary, started, release string) {
	t.Helper()

	host := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNativeAgentHost$")

	host.Env = append(os.Environ(), "CHR_NATIVE_AGENT_HOST="+store.layout.ProjectDir(), "CHR_NATIVE_AGENT_INVOCATION="+id, "CHR_NATIVE_AGENT_BOUNDARY="+boundary, "CHR_NATIVE_AGENT_STARTED="+started, "CHR_NATIVE_AGENT_RELEASE="+release)
	output, err := host.CombinedOutput()

	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 29 {
		t.Fatalf("owned host crash: %v: %s", err, output)
	}
}

// TestNativeAgentBirthCrash replays actual process-birth and launch-charge writes
// interrupted before permission. No driver runs and the uncertainty stays charged.
func TestNativeAgentBirthCrash(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.AgentTokenBudget = 100
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")
	nativeAgentCrashHost(t, store, run.Invocation.ID, "before permit", started, release)

	restarted := NewWorkflowStore(store.layout, store.projectID)

	recovered := waitWorkflowRecovery(t, restarted, run.Invocation.ID)
	if len(recovered.Attempts) != 1 || len(recovered.Attempts[0].Processes) != 1 {
		t.Fatal("birth journal discarded attempt ownership")
	}

	if _, err := os.Stat(started); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("driver ran without a persisted launch permit", err)
	}

	nativeAgentRecoveredCharge(t, restarted, run.Invocation.ID)
}

func nativeAgentRecoveredCharge(t *testing.T, store *Store, id string) {
	t.Helper()

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), id)
	if err != nil || len(budget.Reservations) != 1 || budget.TokensReserved != 100 || budget.TokensSettled != 0 || budget.Usage.WorkerCount != 1 {
		t.Fatal("host recovery refunded or relaunched", err)
	}

	if budget.Reservations[0].State != "launched" && budget.Reservations[0].State != "unknown" {
		t.Fatal("possible dispatch lost its retained charge")
	}
}

// TestNativeAgentOrphan retains the serial pool across host loss and aliases.
// After the surviving driver exits, the slot frees but unknown usage cannot refund.
func TestNativeAgentOrphan(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	request.Policy.AgentTokenBudget = 100
	run := workflowTestStart(t, store, request)
	fixture := t.TempDir()
	started, release := filepath.Join(fixture, "started"), filepath.Join(fixture, "release")

	t.Cleanup(func() { stopWorkflowRecoveryChildren(t, store, run.Invocation.ID, release) })
	nativeAgentCrashHost(t, store, run.Invocation.ID, "after permit", started, release)

	alias := workflowRecoveryAlias(t, store, request.Selection.Target.Commit)
	nativeAgentOrphanCapacity(t, alias, run.Invocation.ID, fixture, release)

	if err := os.WriteFile(release, []byte("finish owned fixture"), 0o600); err != nil {
		t.Fatal(err)
	}

	recovered := waitWorkflowRecovery(t, alias, run.Invocation.ID)
	if len(recovered.Attempts) != 1 || workflowAttemptQuiescent(recovered.Attempts[0]) != nil {
		t.Fatal("host recovery discarded stopped attempt")
	}

	nativeAgentRecoveredCharge(t, alias, run.Invocation.ID)

	slot, guard := workflowTestSlot(t, alias)
	if err := errors.Join(slot.Close(), guard.Close()); err != nil {
		t.Fatal(err)
	}
}

func nativeAgentOrphanCapacity(t *testing.T, store *Store, id, fixture, release string) {
	t.Helper()

	run := workflowTestRead(t, store, id)

	alive, err := workflowProcessGroupAlive(run.Attempts[0].Processes[0])
	if err != nil || !alive {
		t.Fatal("fixture did not retain a live child", err)
	}

	assertWorkflowRecoveryCapacity(t, store, id)

	probe := workflowCapacityProcessFixture(t, store, filepath.Join(fixture, "capacity"), release)
	if output, err := probe.CombinedOutput(); err != nil {
		t.Fatalf("cross-process capacity probe: %v: %s", err, output)
	}

	data, err := os.ReadFile(filepath.Join(fixture, "capacity"))
	if err != nil || string(data) != "waiting" {
		t.Fatal("another process bypassed serial agent capacity", err)
	}

	nativeAgentRecoveredCharge(t, store, id)
}
