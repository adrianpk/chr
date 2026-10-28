package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceIntegrationDecisionReady(t *testing.T) (service, string) {
	t.Helper()

	operations, id, original := serviceWorkflowPreparationReady(t)
	operations.workflowPullRequests = &serviceNativeReviewFixture{original, "policy-v1"}
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "review:\n  mode: human_review\nwork:\n  merge_strategy: rebase\n  integration:\n    publication_profile: git_atomic_dual_ref\n")
	serviceNativeReviewAdmit(t, operations, id)

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	page, err := store.QueryWorkflowIntegrationQueue(t.Context(), workflowcore.IntegrationQueueQuery{InvocationID: id, Limit: 1})
	if err != nil || len(page.Entries) != 1 {
		t.Fatal("missing actual admitted integration unit", err)
	}

	run, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	entry := page.Entries[0].Entry

	entry, err = store.TransitionWorkflowIntegration(t.Context(), workflowcore.IntegrationTransition{InvocationID: id, EntryID: entry.ID, InvocationRevision: run.Revision, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: "merging"})
	if err != nil {
		t.Fatal(err)
	}

	retainServiceIntegrationDecision(t, operations, entry)

	return operations, id
}

func retainServiceIntegrationDecision(t *testing.T, operations service, entry workflowcore.IntegrationEntry) {
	t.Helper()

	run, err := operations.ReadWorkflow(t.Context(), entry.InvocationID)
	if err != nil {
		t.Fatal(err)
	}

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	budget, err := store.ReadWorkflowIntegrationBudget(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}

	input := serviceIntegrationDecisionInput(t, operations, run, entry)

	questions, err := workflowcore.MergerRequiredDecisions(input)
	if err != nil || len(questions) == 0 {
		t.Fatal("sealed requirements did not retain an operator choice", err)
	}

	guard := workflowcore.IntegrationBudgetGuard{InvocationID: run.ID, InvocationRevision: run.Revision, EntryID: entry.ID, EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, BudgetRevision: budget.Revision}
	if _, err := operations.RecordWorkflowIntegrationDecisions(t.Context(), workflowcore.IntegrationDecisionRequest{IntegrationBudgetGuard: guard, ID: "requirements-choice", Input: input, Decisions: questions}); err != nil {
		t.Fatal(err)
	}
}

func serviceIntegrationDecisionInput(t *testing.T, operations service, run workflowcore.Invocation, entry workflowcore.IntegrationEntry) workflowcore.MergerInput {
	t.Helper()

	now := time.Now().UTC()
	gate := serviceIntegrationDecisionGate(t, operations, entry.Unit.ID)
	input := workflowcore.MergerInput{MergerIdentity: workflowcore.MergerIdentity{ProtocolVersion: 1, ID: "surface-decision", InvocationID: run.ID, EntryID: entry.ID, Unit: entry.Unit, CandidateRevision: entry.StateRevision, CandidateHash: run.CandidateHash, ClaimGeneration: entry.ClaimGeneration, ScopeHash: entry.FrozenScopeHash, Reason: "actionable_review_finding"}, MergerGit: workflowcore.MergerGit{Repository: entry.Repository, SourceRef: run.Lane.Branch, TargetRef: entry.TargetRef, Source: run.Lane.HeadCommit, Base: run.PullRequest.TargetCommit, Method: "rebase"}, MergerScope: workflowcore.MergerScope{AllowedPaths: []string{"result.txt"}, ProtectedPaths: []string{"Makefile"}}, MergerValidation: workflowcore.MergerValidation{ValidationCommands: gate.Commands}, MergerLimits: workflowcore.MergerLimits{CapacityReservationID: "pending-capacity", BudgetReservationID: "surface-decision", RemainingAttempts: 3, WorkerTimeoutSeconds: run.Policy.WorkerTimeoutSeconds, AdmittedAt: now, Deadline: now.Add(time.Duration(run.Policy.WorkerTimeoutSeconds) * time.Second)}}
	input.OwnedWorkspaceID = fmt.Sprintf("%x", sha256.Sum256([]byte(run.ID+"\x00"+entry.ID+"\x00"+input.ID)))
	commits := strings.Fields(serviceWorkflowDependencyGit(t, run.Lane.Worktree, "rev-list", "--reverse", run.Lane.BaseCommit+".."+run.Lane.HeadCommit))
	input.SourceManifestRef = serviceIntegrationDecisionEvidence(t, operations, run.ID, "source", workflowcore.NativeSourceManifest{Schema: "chr-native-source/v1", Repository: entry.Repository, SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: commits})
	input.ValidationContractRef = serviceIntegrationDecisionEvidence(t, operations, run.ID, "validation", gate)
	source := serviceIntegrationDecisionEvidence(t, operations, run.ID, "source-intent", map[string]string{"intent": "Preserve the original source behavior"})
	target := serviceIntegrationDecisionEvidence(t, operations, run.ID, "target-intent", map[string]string{"intent": "Preserve the target validation requirements"})
	finding := serviceIntegrationDecisionEvidence(t, operations, run.ID, "finding", workflowstate.ReviewFinding{Category: workflowstate.AgenticReviewFindingCorrectness, Title: "Reconcile declared requirements", Detail: "The source and target intent require an explicit compatible interpretation"})
	input.IntentRefs, input.TargetChangeRefs, input.FindingRefs = []workflowstate.EvidenceRef{source}, []workflowstate.EvidenceRef{target}, []workflowstate.EvidenceRef{finding}
	input.Requirements = []workflowcore.MergerRequirement{{ID: "source-api", Origin: "source", Evidence: source, ConflictsWith: []string{"target-validation"}}, {ID: "target-validation", Origin: "target", Evidence: target}}

	sealed, err := workflowcore.SealWorkflowMergerInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	return sealed
}

func serviceIntegrationDecisionEvidence(t *testing.T, operations service, id, name string, value any) workflowstate.EvidenceRef {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		t.Fatal(err)
	}

	ref := workflowstate.EvidenceRef{Path: "run/" + id + "/evidence/surface-" + name + ".json", Hash: fmt.Sprintf("%x", sha256.Sum256(data))}
	path := filepath.Join(layout.Root(), ref.Path)

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	return ref
}

func serviceIntegrationDecisionGate(t *testing.T, operations service, sliceID string) workflowstate.Validation {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		t.Fatal(err)
	}

	descriptor := serviceIntegrationSliceDescriptor(t, snapshot, sliceID)

	data, err := os.ReadFile(filepath.Join(layout.Root(), descriptor.ContractRef))
	if err != nil || authoringcore.WorkflowContentHash(data) != descriptor.ContractHash {
		t.Fatal("retained slice contract changed", err)
	}

	var contract struct {
		Validation workflowstate.Validation `json:"validation"`
	}
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}

	return contract.Validation
}

func serviceIntegrationSliceDescriptor(t *testing.T, snapshot workflowcore.Snapshot, sliceID string) workflowstate.WorkflowSlice {
	t.Helper()

	for _, record := range snapshot.Records {
		if record.Set == nil {
			continue
		}

		for _, descriptor := range record.Set.Slices {
			if descriptor.ID == sliceID {
				return descriptor
			}
		}
	}

	t.Fatal("missing retained slice validation contract")

	return workflowstate.WorkflowSlice{}
}
