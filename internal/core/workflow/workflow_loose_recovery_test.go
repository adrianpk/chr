package workflow

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

// TestWorkflowLooseCrashProcess is the owned process-death fixture for the
// pending display-write, immutable delivery-frontier and queue-state journals.
func TestWorkflowLooseCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_LOOSE_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")

	index, err := strconv.Atoi(os.Getenv("CHR_LOOSE_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store.afterReplace = func(position int) error {
		if position == index {
			os.Exit(17)
		}

		return nil
	}

	id := os.Getenv("CHR_LOOSE_CRASH_ID")
	switch os.Getenv("CHR_LOOSE_CRASH_ACTION") {
	case "write":
		workflowLooseClassify(t, store, workflowLooseRecord(t, store, "PRJ-91"), "critical", string(ticketcore.StatusReadyForDev))
	case "select":
		_, err = store.ReconcileWorkflowMembership(t.Context(), id)
	case "block":
		_, err = store.ClaimWorkflow(t.Context(), id)
	case "resume":
		run := workflowTestRead(t, store, id)
		remote := os.Getenv("CHR_LOOSE_CRASH_REMOTE")
		git := NewWorkflowIntegrationGit(run.Selection.Target.Repository, remote)
		provider := workflowFrontierProvider{&workflowPRProviderFixture{pr: PullRequestSnapshot{Repository: run.Selection.Target.Repository}}, remote}
		_, err = store.ResumeWorkflow(t.Context(), id, ResumePorts{Provider: provider, Target: git, TargetBranch: "dev"})
	case "frontier":
		err = workflowLooseCrashFrontier(t, store, id)
	}

	t.Fatalf("owned queue writer did not exit: %v", err)
}

func workflowLooseCrashFrontier(t *testing.T, store *Store, id string) error {
	t.Helper()

	run, err := store.readWorkflowInvocation(id)
	if err != nil {
		return err
	}

	proof, err := store.readWorkflowDeliveryProof(run)
	if err != nil {
		return err
	}

	remote := os.Getenv("CHR_LOOSE_CRASH_REMOTE")
	git := NewWorkflowIntegrationGit(proof.Git.Repository, remote)
	provider := workflowFrontierProvider{&workflowPRProviderFixture{pr: proof.PullRequest}, remote}
	_, err = store.AdvanceWorkflowFrontier(t.Context(), id, FrontierPorts{Provider: provider, Observer: git, Target: git})

	return err
}

// TestWorkflowLooseStopRecovery kills only the owned journal writer. Replay
// retains a dependency stop, or both its verified resume image and target,
// without acquiring a worker or repeating the already integrated prefix.
func TestWorkflowLooseStopRecovery(t *testing.T) {
	for _, tc := range []struct {
		action string
		index  int
	}{{"block", 0}, {"resume", 0}, {"resume", 1}} {
		t.Run(tc.action+strconv.Itoa(tc.index), func(t *testing.T) {
			f := workflowLooseBlockedFixture(t)

			target := f.target
			if tc.action == "resume" {
				workflowLoosePauseDependency(t, f)

				target = workflowIntegrationCommit(t, f.actor, "dependency.txt", "Verified outside prerequisite")
				workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")
				workflowLooseConfirmDependency(t, f, target)
			} else if _, err := f.store.ReconcileWorkflowMembership(t.Context(), f.run.ID); err != nil {
				t.Fatal(err)
			}

			crashWorkflowLoose(t, f, tc.action, tc.index)

			restarted := NewWorkflowStore(f.store.layout, f.store.projectID)
			run := workflowTestRead(t, restarted, f.run.ID)
			checkWorkflowLooseStopRecovery(t, run, tc.action, target)
		})
	}
}

func checkWorkflowLooseStopRecovery(t *testing.T, run Invocation, action, target string) {
	t.Helper()

	if len(run.Attempts) != 1 || run.Cursor != 1 || run.ConfirmedCursor != 1 || run.Lane != nil || len(run.Deliveries) != 1 {
		t.Fatal("recovery duplicated or lost accepted work")
	}

	checkWorkflowLooseStopState(t, run, action, target)
}

func checkWorkflowLooseStopState(t *testing.T, run Invocation, action, target string) {
	t.Helper()

	if action == "block" {
		if run.Status != "paused" || run.Blocker == nil || run.Blocker.Dependency.ID != "PRJ-89" {
			t.Fatal("recovery lost the dependency stop")
		}

		return
	}

	if run.Status != "pending" || run.Blocker != nil || len(run.DependencyTargets) != 1 || run.ExecutionSelection().Target.Commit != target {
		t.Fatal("recovery lost verified resume evidence or its target")
	}
}

// TestWorkflowLooseRecovery proves a real host death cannot expose a pending
// task write without its expected image, or lose/duplicate a verified frontier.
func TestWorkflowLooseRecovery(t *testing.T) {
	for _, tc := range []struct {
		action string
		index  int
	}{{"write", 0}, {"write", 1}, {"write", 2}, {"frontier", 0}, {"frontier", 1}, {"select", 0}} {
		t.Run(tc.action+strconv.Itoa(tc.index), func(t *testing.T) {
			f := workflowLooseQueueFixture(t, false)
			prepareWorkflowLooseCrash(t, f, tc.action)
			crashWorkflowLoose(t, f, tc.action, tc.index)

			restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

			run, err := restarted.ReconcileWorkflowMembership(t.Context(), f.run.ID)
			if err != nil || len(run.Attempts) != 1 {
				t.Fatal("journal recovery lost an accepted image or replayed a task", err)
			}

			if tc.action != "write" && (len(run.Deliveries) != 1 || run.ConfirmedCursor != 1 || run.Lane != nil) {
				t.Fatal("journal recovery lost or duplicated the delivered unit")
			}

			if tc.action == "select" && run.Members[1].Admission.Key != "PRJ-92" {
				t.Fatal("queue recovery lost accepted ordering")
			}
		})
	}
}

func prepareWorkflowLooseCrash(t *testing.T, f workflowIntegrationFixture, action string) {
	t.Helper()

	if action == "write" {
		return
	}

	workflowFinalizeFrontier(t, f)

	if action == "select" {
		workflowAdvanceFrontier(t, f)
		workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, "PRJ-92"), "critical", string(ticketcore.StatusReadyForDev))
	}
}

func crashWorkflowLoose(t *testing.T, f workflowIntegrationFixture, action string, index int) {
	t.Helper()

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowLooseCrashProcess$")

	command.Env = append(os.Environ(), "CHR_LOOSE_CRASH_DIR="+f.store.layout.ProjectDir(), "CHR_LOOSE_CRASH_ID="+f.run.ID, "CHR_LOOSE_CRASH_ACTION="+action, "CHR_LOOSE_CRASH_INDEX="+strconv.Itoa(index), "CHR_LOOSE_CRASH_REMOTE="+f.remote)
	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned queue crash: %v: %s", err, output)
	}
}
