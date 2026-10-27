package core

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

func workflowTicketValidated(t *testing.T) (*WorkflowStore, WorkflowRecord, WorkflowInvocation) {
	t.Helper()

	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	run := workflowTestStart(t, store, request).Invocation

	claim := workflowTestClaim(t, store, run.ID)
	if _, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	return store, record, workflowTestRead(t, store, run.ID)
}

// TestWorkflowTicketCandidateGuard rejects altered host inputs/results, task
// projections and lane history before PR creation or delivery can consume them.
func TestWorkflowTicketCandidateGuard(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "global.gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	cases := []struct {
		name   string
		mutate func(*testing.T, *WorkflowStore, WorkflowInvocation)
	}{
		{"result", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation) {
			workflowTicketCorrupt(t, s, r.Attempts[0].ResultRef)
		}},
		{"input", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation) {
			workflowTicketCorrupt(t, s, r.Attempts[0].InputRef)
		}},
		{"candidate", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation) {
			workflowTicketCorrupt(t, s, r.CandidateRef)
		}},
		{"task image", func(t *testing.T, s *WorkflowStore, r WorkflowInvocation) {
			workflowTicketCorrupt(t, s, r.Members[0].Admission.Path)
		}},
		{"lane", func(t *testing.T, _ *WorkflowStore, r WorkflowInvocation) {
			workflowIntegrationCommit(t, r.Lane.Worktree, "extra.txt", "Unaccepted change")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, _, run := workflowTicketValidated(t)
			if _, err := store.PrepareWorkflowTicketCandidate(t.Context(), run.ID); err != nil {
				t.Fatal(err)
			}

			run = workflowTestRead(t, store, run.ID)
			tc.mutate(t, store, run)

			if _, err := store.ReadWorkflowSliceCandidate(t.Context(), run.ID); err == nil {
				t.Fatal("changed candidate evidence was accepted")
			}
		})
	}
}

func workflowTicketCorrupt(t *testing.T, store *WorkflowStore, ref string) {
	t.Helper()

	path := filepath.Join(store.layout.Root(), ref)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowTicketCrashProcess exits an owned writer at either replacement
// in the ticket-candidate journal. The parent observes and replays durable intent.
func TestWorkflowTicketCrashProcess(t *testing.T) {
	dir := os.Getenv("CHR_TICKET_CRASH_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir})
	if err != nil {
		t.Fatal(err)
	}

	index, err := strconv.Atoi(os.Getenv("CHR_TICKET_CRASH_INDEX"))
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	store.afterReplace = func(replaced int) error {
		if replaced == index {
			os.Exit(17)
		}

		return nil
	}

	_, err = store.PrepareWorkflowTicketCandidate(t.Context(), os.Getenv("CHR_TICKET_CRASH_ID"))
	t.Fatalf("owned writer did not exit: %v", err)
}

// TestWorkflowTicketCandidateRecovery verifies both real process-death boundaries
// and keeps the accepted worker proof and immutable candidate identity unchanged.
func TestWorkflowTicketCandidateRecovery(t *testing.T) {
	for index := range 2 {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			store, _, run := workflowTicketValidated(t)
			command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowTicketCrashProcess$")

			command.Env = append(os.Environ(), "CHR_TICKET_CRASH_DIR="+store.layout.ProjectDir(), "CHR_TICKET_CRASH_ID="+run.ID, "CHR_TICKET_CRASH_INDEX="+strconv.Itoa(index))
			output, err := command.CombinedOutput()

			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 17 {
				t.Fatalf("writer: %v: %s", err, output)
			}

			restarted := NewWorkflowStore(store.layout, store.projectID)

			candidate, err := restarted.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
			if err != nil || workflowTestRead(t, restarted, run.ID).Attempts[0].ResultHash != run.Attempts[0].ResultHash {
				t.Fatal("journal recovery lost accepted task proof", err)
			}

			checkWorkflowTicketCandidateRetry(t, restarted, run.ID, candidate)
		})
	}
}

// TestWorkflowTicketClosedPR keeps the recorded canonical identity blocked when
// the provider closes it without delivery. Retry cannot create a replacement.
func TestWorkflowTicketClosedPR(t *testing.T) {
	f, _ := workflowTicketPreparationFixture(t, "rebase", "ticket")

	f.provider.pr.State = "closed"
	if _, err := f.store.EnsureWorkflowSlicePullRequest(t.Context(), f.run.ID, workflowPRContent(), f.provider); err == nil {
		t.Fatal("closed independent PR was accepted")
	}

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Status != "blocked" || run.PullRequest.URL != f.provider.pr.URL || f.provider.creates != 1 || run.DeliveryRef != "" {
		t.Fatal("closed PR lost identity or created delivery")
	}
}

// TestWorkflowTicketFinalizationRecovery replays receipt and task images after
// real host death at every write. Canonical delivery is verified before the
// journal; replay neither needs a worker nor repeats external publication.
func TestWorkflowTicketFinalizationRecovery(t *testing.T) {
	for _, index := range []int{0, 1, 3} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			fixture := workflowTicketExternalDelivery(t)
			crashWorkflowFinalization(t, fixture, index)

			restarted := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

			receipt, err := restarted.FinalizeWorkflowSlice(t.Context(), fixture.run.ID, nil, nil)
			if err != nil || receipt.Receipt.Unit.Kind != "ticket" {
				t.Fatal("receipt journal did not replay", err)
			}

			record, err := restarted.Read(t.Context(), fixture.hierarchy.Tasks[0].Path)
			if err != nil || record.Task.Status != TicketStatusDone || record.Task.DeliveryRef != receipt.Path {
				t.Fatal("ticket receipt projection did not replay", err)
			}

			before := workflowTestRead(t, restarted, fixture.run.ID)
			if _, err := restarted.FinalizeWorkflowSlice(t.Context(), before.ID, nil, nil); err != nil || workflowTestRead(t, restarted, before.ID).Revision != before.Revision {
				t.Fatal("receipt replay finalized twice", err)
			}
		})
	}
}

func workflowTicketExternalDelivery(t *testing.T) workflowIntegrationFixture {
	t.Helper()

	f, _ := workflowTicketPreparationFixture(t, "rebase", "ticket")
	workflowIntegrationCommand(t, f.actor, "branch", "source", "origin/"+f.provider.pr.HeadBranch)
	workflowFixtureIntegrate(t, f.actor, "rebase")

	result := workflowIntegrationCommand(t, f.actor, "rev-parse", "HEAD")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	f.provider.pr.State, f.provider.pr.Merged, f.provider.pr.ResultCommit = "closed", true, result
	f.provider.pr.BaseCommit, f.provider.target = result, result

	git := NewWorkflowIntegrationGit(f.provider.pr.Repository, f.remote)
	if _, err := f.store.VerifyWorkflowSliceIntegration(t.Context(), f.run.ID, "rebase", f.provider, git); err != nil {
		t.Fatal(err)
	}

	return workflowIntegrationFixture{store: f.store, hierarchy: f.hierarchy, run: workflowTestRead(t, f.store, f.run.ID), provider: f.provider.workflowPRProviderFixture, git: git, actor: f.actor, remote: f.remote, base: f.base, result: result, target: result}
}

// TestWorkflowTicketBounded drives source execution and candidate preparation in
// separate calls, retaining the original scope and validated proof at restart.
func TestWorkflowTicketBounded(t *testing.T) {
	store, record, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	request.Policy.CheckpointStepLimit = 1
	run := workflowTestStart(t, store, request).Invocation
	continuation := WorkflowContinuationRequest{Worker: workflowPRWorker(), StepLimit: 1}

	first, err := store.ContinueWorkflow(t.Context(), run.ID, continuation, WorkflowContinuationPorts{})
	if err != nil || first.Steps != 1 || first.Reason != "step_limit" || first.Invocation.Cursor != 1 || first.Invocation.CandidateRef != "" {
		t.Fatal("first bounded call skipped task validation or candidate boundary", err)
	}

	checkWorkflowTicketBoundedRestart(t, store, record, first, continuation)
}

func checkWorkflowTicketBoundedRestart(t *testing.T, store *WorkflowStore, record WorkflowRecord, first WorkflowContinuationProgress, continuation WorkflowContinuationRequest) {
	t.Helper()

	restarted := NewWorkflowStore(store.layout, store.projectID)

	second, err := restarted.ContinueWorkflow(t.Context(), first.Invocation.ID, continuation, WorkflowContinuationPorts{})
	if err != nil || second.Steps != 1 || second.Invocation.CandidateRef == "" || second.Invocation.ConfirmedCursor != 0 || len(second.Invocation.Attempts) != 1 {
		t.Fatal("bounded restart reran the worker or claimed delivery", err)
	}

	candidate, err := restarted.ReadWorkflowSliceCandidate(t.Context(), first.Invocation.ID)
	if err != nil || candidate.Ticket.Unit.ID != record.Task.Key || second.Invocation.RequestHash != first.Invocation.RequestHash {
		t.Fatal("bounded restart broadened the retained task", err)
	}
}

// TestWorkflowTicketRepair invalidates only the independent task after a blocking
// review and retains its PR and consumed repair budget for the next candidate.
func TestWorkflowTicketRepair(t *testing.T) {
	f, _ := workflowTicketPreparationFixture(t, "rebase", "ticket")
	workflowTicketReviewConfig(t, f.store, "rebase", ReviewWorkflowModeAgenticReview)

	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git); err != nil {
		t.Fatal(err)
	}

	binding, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, f.provider, workflowReviewWorker("needs_changes", "PRJ-90", ""))
	if err != nil || binding.State != "needs_repair" {
		t.Fatal("independent task review did not request repair", err)
	}

	checkWorkflowTicketRepair(t, f)
}

func checkWorkflowTicketRepair(t *testing.T, f workflowPreparationFixture) {
	t.Helper()

	run := workflowTestRead(t, f.store, f.run.ID)
	if run.Cursor != 0 || run.CandidateRef != "" || run.Status != "pending" || run.PullRequest.URL != f.provider.pr.URL || len(run.ReviewRepairs) != 1 {
		t.Fatal("repair broadened scope or lost canonical PR")
	}

	if run.ReviewRepairs[0].Unit.Kind != "ticket" || run.ReviewRepairs[0].Consumed != 1 {
		t.Fatal("independent repair lost its bounded unit budget")
	}

	checkWorkflowTicketRepairCandidate(t, f, run)
}

func checkWorkflowTicketRepairCandidate(t *testing.T, f workflowPreparationFixture, run WorkflowInvocation) {
	t.Helper()

	claim := workflowTestClaim(t, f.store, run.ID)
	if _, err := f.store.ExecuteWorkflowTask(t.Context(), claim, workflowPRWorker()); err != nil {
		t.Fatal(err)
	}

	candidate, err := f.store.PrepareWorkflowTicketCandidate(t.Context(), run.ID)
	if err != nil || candidate.ID == f.run.PullRequest.CandidateID || f.provider.creates != 1 {
		t.Fatal("repair reused stale acceptance or created a replacement PR", err)
	}
}
