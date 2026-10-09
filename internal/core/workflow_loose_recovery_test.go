package core

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

// TestWorkflowLooseCrashProcess is the owned process-death fixture for the
// pending display-write, immutable delivery-frontier and queue-state journals.
func TestWorkflowLooseCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_LOOSE_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir})
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
		workflowLooseClassify(t, store, workflowLooseRecord(t, store, "PRJ-91"), "critical", TicketStatusReadyForDev)
	case "select":
		_, err = store.ReconcileWorkflowMembership(t.Context(), id)
	case "frontier":
		run, readErr := store.readWorkflowInvocation(id)
		if readErr != nil {
			t.Fatal(readErr)
		}

		proof, readErr := store.readWorkflowDeliveryProof(run)
		if readErr != nil {
			t.Fatal(readErr)
		}

		remote := os.Getenv("CHR_LOOSE_CRASH_REMOTE")
		git := NewWorkflowIntegrationGit(proof.Git.Repository, remote)
		provider := workflowFrontierProvider{&workflowPRProviderFixture{pr: proof.PullRequest}, remote}
		_, err = store.AdvanceWorkflowFrontier(t.Context(), id, WorkflowFrontierPorts{Provider: provider, Observer: git, Target: git})
	}

	t.Fatalf("owned queue writer did not exit: %v", err)
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
		workflowLooseClassify(t, f.store, workflowLooseRecord(t, f.store, "PRJ-92"), "critical", TicketStatusReadyForDev)
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
