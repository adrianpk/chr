package core

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeCandidateRepairFixture(t *testing.T, method, reason string) (nativePreparationFixture, WorkflowMergerRequest, WorkflowMergerHarness) {
	t.Helper()

	f, candidate, gate := nativeCandidateRepairStart(t, method, reason)

	run := workflowTestRead(t, f.f.store, f.f.run.ID)
	input := nativeCandidateRepairInput(t, f, run, candidate, gate)
	state := nativeCandidateRepairCheckpoint(t, f, candidate)
	request := WorkflowMergerRequest{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, f.f.store, run.ID), Input: input}

	ref, err := f.f.store.RetainWorkflowMergerCheckpoint(t.Context(), request.WorkflowIntegrationBudgetGuard, state, f.f.git)
	if err != nil {
		t.Fatal("candidate repair checkpoint was rejected", err)
	}

	request.Input.RepairCheckpointRef = &ref

	if method == "rebase" {
		checkpoint := nativeMergerReplayCheckpoint(t, state)
		payload := *checkpoint
		payload.Evidence = workflowstate.EvidenceRef{}
		checkpoint.Evidence = retainNativeRecordEvidence(t, f.f.store, run.ID, "candidate-replay", payload)
		request.Input.HistoryCheckpoint = checkpoint
	}

	request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	return f, request, nativeMergerTestHarness(t, f.f.store, run.ID)
}

func nativeCandidateRepairInput(t *testing.T, f nativePreparationFixture, run WorkflowInvocation, candidate WorkflowIntegrationCandidate, gate workflowstate.Validation) WorkflowMergerInput {
	t.Helper()

	repair, err := f.f.store.ReadWorkflowCandidateRepair(t.Context(), run.ID)
	if err != nil {
		t.Fatal("host findings unavailable", err)
	}

	entry := nativeReviewEntry(t, f.f.store, run.ID)
	now := time.Now().UTC()
	// Real Git and durable fixture IO need CI allowance; expiry has separate coverage.
	deadline := now.Add(min(2*time.Minute, time.Duration(run.Policy.WorkerTimeoutSeconds)*time.Second))
	input := WorkflowMergerInput{WorkflowMergerIdentity: WorkflowMergerIdentity{ProtocolVersion: 1, ID: "candidate-repair", InvocationID: run.ID, EntryID: entry.ID, Unit: entry.Unit, CandidateRevision: entry.StateRevision, CandidateHash: run.CandidateRepair.CandidateHash, ClaimGeneration: entry.ClaimGeneration, ScopeHash: entry.FrozenScopeHash, Reason: repair.Reason}, WorkflowMergerGit: WorkflowMergerGit{Repository: entry.Repository, SourceRef: run.Lane.Branch, TargetRef: entry.TargetRef, Source: run.Lane.HeadCommit, Base: f.f.base, Method: candidate.Input.Git.Method}, WorkflowMergerContext: WorkflowMergerContext{FindingRefs: repair.FindingRefs}, WorkflowMergerScope: WorkflowMergerScope{AllowedPaths: []string{"result.txt"}, ProtectedPaths: []string{"Makefile"}}, WorkflowMergerValidation: WorkflowMergerValidation{ValidationCommands: gate.Commands}, WorkflowMergerLimits: WorkflowMergerLimits{CapacityReservationID: "pending-capacity", BudgetReservationID: "candidate-repair", RemainingAttempts: 3, WorkerTimeoutSeconds: run.Policy.WorkerTimeoutSeconds, AdmittedAt: now, Deadline: deadline}}
	input.OwnedWorkspaceID = workflowMergerWorkspaceID(input)
	manifest := WorkflowNativeSourceManifest{Schema: "chr-native-source/v1", Repository: input.Repository, SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: strings.Fields(workflowIntegrationCommand(t, f.f.run.Lane.Worktree, "rev-list", "--reverse", f.f.run.Lane.BaseCommit+".."+input.Source))}
	input.SourceManifestRef = retainNativeRecordEvidence(t, f.f.store, run.ID, "candidate-source", manifest)
	input.ValidationContractRef = retainNativeRecordEvidence(t, f.f.store, run.ID, "candidate-gate", gate)
	source := retainNativeRecordEvidence(t, f.f.store, run.ID, "candidate-intent", candidate.Input.Source)
	target := retainNativeRecordEvidence(t, f.f.store, run.ID, "candidate-target", candidate.Input.Git)
	input.IntentRefs, input.TargetChangeRefs = []workflowstate.EvidenceRef{source}, []workflowstate.EvidenceRef{target}
	input.Requirements = []WorkflowMergerRequirement{{ID: "source-api", Origin: "source", Evidence: source}, {ID: "target-validation", Origin: "target", Evidence: target}}
	input.Conflicts = []WorkflowMergerConflict{{ID: "semantic-result", Path: "result.txt", Kind: "semantic", Evidence: repair.FindingRefs[0], RequirementIDs: []string{"source-api", "target-validation"}}}

	return input
}

func nativeCandidateRepairCheckpoint(t *testing.T, f nativePreparationFixture, candidate WorkflowIntegrationCandidate) WorkflowMergerRepairState {
	t.Helper()

	proof := candidate.Input.Git
	query := WorkflowPreparationQuery{WorkflowIntegrationQuery: proof.WorkflowIntegrationQuery, BaseCommit: proof.BaseCommit}
	query.SourceDirectory, query.ResultCommit = candidate.Lane.Worktree, ""
	query.Lane = WorkflowLane{Unit: f.entry.Unit, Branch: "refs/heads/chr/workflow/semantic", Worktree: filepath.Join(t.TempDir(), "semantic"), BaseCommit: f.f.base, HeadCommit: f.f.base}

	var (
		state WorkflowMergerRepairState
		err   error
	)

	if proof.Method == "rebase" {
		state, err = f.f.git.BeginWorkflowMergerSemanticRepair(t.Context(), query, proof, proof.SourceCommit)
	} else {
		state, err = f.f.git.BeginWorkflowMergerCombinedRepair(t.Context(), query, proof)
	}

	if err != nil {
		t.Fatal("semantic baseline failed", err)
	}

	return state
}

func nativeCandidateRepairStart(t *testing.T, method, reason string) (nativePreparationFixture, WorkflowIntegrationCandidate, workflowstate.Validation) {
	t.Helper()

	gate := workflowPreparationGate("test ! -f target.txt || grep -q 'source and target' result.txt")
	if reason == "actionable_review_finding" {
		gate = workflowPreparationGate("test -f result.txt")
	}

	f := newNativePreparationFixture(t, method, gate)
	if reason == "actionable_review_finding" {
		writeWorkflowTestConfigValue(t, f.f.store.layout.ProjectDir(), "review", "mode", string(workflowstate.ReviewAgentic))
	}

	candidate, err := f.f.store.PrepareWorkflowSliceIntegration(t.Context(), f.f.run.ID, f.provider, f.f.git)
	if reason == "combined_gate_regression" && err == nil {
		t.Fatal("semantic regression did not fail its exact host gate")
	}

	if reason == "actionable_review_finding" {
		if err != nil {
			t.Fatal(err)
		}

		key := f.f.run.Members[len(f.f.run.Members)-1].Admission.Key
		if _, err := f.f.store.ReviewWorkflowSlice(t.Context(), f.f.run.ID, f.provider, workflowReviewWorker("needs_changes", key, "")); err != nil {
			t.Fatal("review repair failed", err)
		}
	}

	return f, candidate, gate
}
