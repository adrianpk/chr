package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func serviceWorkflowDependencyFixture(t *testing.T) (service, core.WorkflowRecord, core.WorkflowStartRequest) {
	t.Helper()

	operations, selected, request := serviceWorkflowTicketFixture(t)

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	task := *selected.Task
	task.Key, task.Status = "PRJ-89", core.TicketStatusBacklog

	outside, err := store.Create(t.Context(), "Outside prerequisite", core.WorkflowRecord{Kind: "ticket", Task: &task, Body: selected.Body, Domain: maps.Clone(selected.Domain)})
	if err != nil {
		t.Fatal(err)
	}

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(selected.Domain["execution_contract"], &contract); err != nil {
		t.Fatal(err)
	}

	contract.DependsOn = []workflowstate.Ref{{Kind: "ticket", ID: outside.Task.Key}}

	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}

	hash := sha256.Sum256(append([]byte("\x00\x00"), data...))
	selected.Task.DependsOn, selected.Task.ContractHash = contract.DependsOn, hex.EncodeToString(hash[:])

	selected.Domain["execution_contract"] = data
	if _, err := store.Write(t.Context(), selected, selected.Revision); err != nil {
		t.Fatal(err)
	}

	request.Selection.Selector = workflowstate.Selector{Extent: "loose"}

	return operations, outside, request
}

// TestServiceWorkflowDependency resumes through the shared service after actual
// canonical target movement. Configuration supplies the target branch; the
// invocation retains one admitted ticket and never executes its prerequisite.
func TestServiceWorkflowDependency(t *testing.T) {
	operations, outside, request := serviceWorkflowDependencyFixture(t)

	run, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.ClaimWorkflow(t.Context(), run.Invocation.ID); err == nil {
		t.Fatal("service started an unverified dependency")
	}

	serviceWorkflowConfirmDependency(t, &operations, outside, request)

	resumed, err := operations.ResumeWorkflow(t.Context(), run.Invocation.ID)
	checkServiceWorkflowResumed(t, resumed, request, err)

	claim, err := operations.ClaimWorkflow(t.Context(), resumed.ID)
	if err != nil {
		t.Fatal(err)
	}

	if err := operations.CheckpointWorkflow(t.Context(), claim); err != nil {
		t.Fatal(err)
	}
}

func checkServiceWorkflowResumed(t *testing.T, resumed core.WorkflowInvocation, request core.WorkflowStartRequest, failure error) {
	t.Helper()

	if failure != nil || resumed.Status != "pending" || resumed.Blocker != nil || len(resumed.DependencyTargets) != 1 || len(resumed.Members) != 1 || resumed.Members[0].Admission.Key != "PRJ-90" || resumed.Selection.Target != request.Selection.Target {
		t.Fatal("service changed scope or failed to retain canonical resume proof", failure)
	}
}

func serviceWorkflowConfirmDependency(t *testing.T, operations *service, outside core.WorkflowRecord, request core.WorkflowStartRequest) {
	t.Helper()

	project := operations.projectContext.ProjectDir
	remote := filepath.Join(t.TempDir(), "canonical.git")
	serviceWorkflowDeliveryGit(t, project, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, project, "push", remote, request.Selection.Target.Commit+":refs/heads/dev")
	writeServiceWorkspaceFile(t, filepath.Join(project, "dependency.txt"), "Verified prerequisite\n")
	serviceWorkflowDeliveryGit(t, project, "add", "dependency.txt")
	serviceWorkflowDeliveryGit(t, project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "Prerequisite")
	serviceWorkflowDeliveryGit(t, project, "push", remote, "HEAD:refs/heads/dev")

	target := serviceWorkflowDependencyGit(t, project, "rev-parse", "HEAD")
	tree := serviceWorkflowDependencyGit(t, project, "rev-parse", "HEAD^{tree}")
	serviceWorkflowDependencyReceipt(t, *operations, outside, request.Selection.Target.Repository, target, tree)

	operations.workflowIntegrationGit = core.NewWorkflowIntegrationGit(request.Selection.Target.Repository, remote)
	operations.workflowPullRequests = &serviceWorkflowReviewFixture{&serviceWorkflowPRFixture{pr: core.WorkflowPullRequestSnapshot{BaseBranch: "dev", BaseCommit: target}}}
}

func serviceWorkflowDependencyGit(t *testing.T, project string, args ...string) string {
	t.Helper()

	output, err := exec.CommandContext(t.Context(), "git", append([]string{"-C", project}, args...)...).Output()
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(output))
}

func serviceWorkflowDependencyReceipt(t *testing.T, operations service, outside core.WorkflowRecord, repository, commit, tree string) {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	identity := core.WorkflowIdentity{ID: "dependency-receipt", ProjectID: outside.ProjectID, Namespace: outside.Namespace, Revision: 1, CreatedAt: now, UpdatedAt: now}

	receipt, err := store.CreateReceipt(t.Context(), core.WorkflowRecord{WorkflowIdentity: identity, Kind: "receipt", Path: "run/dependency/evidence/receipt.json", Receipt: &core.WorkflowReceipt{Schema: core.WorkflowSchema, Unit: workflowstate.Ref{Kind: "ticket", ID: outside.Task.Key}, CandidateID: "candidate", SubmissionID: "submission", Provider: "forgejo", Repository: repository, PullRequest: repository + "/pulls/1", Method: "rebase", SourceCommit: commit, BaseCommit: commit, PreparedTree: tree, HistoryProof: "host-verification", ResultCommit: commit, VerifiedTree: tree, ObservedTarget: commit, VerificationRef: "host-proof", ValidationRefs: []string{"host-validation"}, ConfirmedAt: now}})
	if err != nil {
		t.Fatal(err)
	}

	outside.Task.Phase, outside.Task.Status, outside.Task.DeliveryRef = "integrated", core.TicketStatusDone, receipt.Path
	if _, err := store.Write(t.Context(), outside, outside.Revision); err != nil {
		t.Fatal(err)
	}
}
