package workflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// TestWorkflowConfigAdmission freezes project limits at real admission and keeps
// them across config changes, while rejecting forged or relaxed restrictions.
func TestWorkflowConfigAdmission(t *testing.T) {
	store, hierarchy, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	body := "workflow:\n  version: 1\nwork:\n  worker_timeout_seconds: 60\n  validation_timeout_seconds: 10\n  checkpoint_step_limit: 2\n  max_worker_attempts: 1\n  max_integration_attempts: 2\n  include_later_tickets: true\n"
	writeWorkspaceConfig(t, store.layout.ProjectDir(), body)

	request.Selection.Selector = workflowClaimSelector(hierarchy, "slice")
	later := false
	request.IncludeLaterTickets = &later
	run := workflowTestStart(t, store, request).Invocation

	policy := run.Policy
	checkWorkflowConfigFrozen(t, policy, body)

	progress, err := store.ContinueWorkflow(t.Context(), run.ID, ContinuationRequest{Worker: workflowTaskCommand(t, "worker")}, ContinuationPorts{})
	if err != nil || progress.Steps != 2 || progress.Invocation.Cursor != 2 {
		t.Fatal("configured checkpoint limit was not enforced", progress, err)
	}

	checkWorkflowConfigRetry(t, store, request, policy)
}

// TestWorkflowConfigPublication resolves the owned project file rather than an
// incidental ops file. Human restrictions cannot enable automatic publication.
func TestWorkflowConfigPublication(t *testing.T) {
	store, _ := workflowMaterializationFixture(t)
	if err := os.WriteFile(filepath.Join(store.layout.Root(), "config.yml"), []byte("review:\n  mode: trusted_auto_merge\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "workflow:\n  version: 1\nreview:\n  mode: human_review\n")

	policy, err := store.workflowReviewPolicy(workflowInvocationFile{})
	if err != nil || policy.Mode != workflowstate.ReviewHuman || policy.Source != "project" {
		t.Fatal("publication read policy from the wrong owner", policy, err)
	}

	writeWorkspaceConfig(t, store.layout.ProjectDir(), "workflow:\n  version: 1\n")

	policy, err = store.workflowReviewPolicy(workflowInvocationFile{})
	if err != nil || policy.Mode != workflowstate.ReviewTrustedAutoMerge || policy.Source != "version_default" {
		t.Fatal("publication missed version defaults", policy, err)
	}

	checkWorkflowConfigRestriction(t, store.layout.ProjectDir())
}

// TestWorkflowConfigDeadline runs real gates and confirms the frozen project
// deadline stops ordered validation before its later command can run.
func TestWorkflowConfigDeadline(t *testing.T) {
	for _, extent := range []string{"ticket", "slice"} {
		t.Run(extent, func(t *testing.T) { checkWorkflowConfigDeadline(t, extent) })
	}
}

func checkWorkflowConfigDeadline(t *testing.T, extent string) {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "later-command")
	gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "sleep 4"}, {"touch", marker}}}
	store, request := workflowConfigDeadlineFixture(t, extent, gate)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  validation_timeout_seconds: 1\n")

	run := workflowTestStart(t, store, request).Invocation
	// A later project edit must not extend this invocation's gate deadline.
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  validation_timeout_seconds: 30\n")

	if err := executeWorkflowConfigDeadline(t, extent, store, run.ID); err == nil {
		t.Fatal("deadline recorded success")
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("later validation command ran after expiry", err)
	}
}

func workflowConfigDeadlineFixture(t *testing.T, extent string, gate workflowstate.Validation) (*Store, StartRequest) {
	t.Helper()

	if extent == "slice" {
		store, opts := workflowMaterializationFixture(t)

		opts.Plan.Slices[0].Validation = gate
		for i := range opts.Plan.Slices[0].Tasks {
			opts.Plan.Slices[0].Tasks[i].Validation = workflowstate.Validation{Required: true, Commands: [][]string{{"git", "diff", "--exit-code", "HEAD"}}, TimeoutSeconds: 30}
		}

		hierarchy, err := store.MaterializePlan(t.Context(), opts)
		if err != nil {
			t.Fatal(err)
		}

		commit, _ := workflowSelectionCommit(t, store)
		request := StartRequest{StartInput: workflowstate.StartInput{
			RequestKey: "config-slice-deadline",
			Selection:  workflowstate.PreviewRequest{Selector: workflowClaimSelector(hierarchy, "slice"), Target: workflowstate.PreviewTarget{Repository: "https://forge.example/test/repo", Commit: commit}},
		}}

		return store, request
	}

	store, _, request := workflowExecutionGateFixture(t, gate)

	return store, request
}

func executeWorkflowConfigDeadline(t *testing.T, extent string, store *Store, id string) error {
	t.Helper()

	if extent == "ticket" {
		claim := workflowTestClaim(t, store, id)

		result, err := store.ExecuteWorkflowTask(t.Context(), claim, workflowTaskCommand(t, "worker"))
		if result.Successful || len(result.Validation) != 1 {
			t.Fatal("task gate exceeded its frozen shared deadline", result)
		}

		return err
	}

	result, err := store.ExecuteWorkflowSlice(t.Context(), id, workflowTaskCommand(t, "worker"))
	if result.Result.Successful || len(result.Result.Validation) != 1 {
		t.Fatal("slice gate exceeded its frozen shared deadline", result, err)
	}

	return err
}

func checkWorkflowConfigFrozen(t *testing.T, policy workflowstate.ExecutionPolicy, body string) {
	t.Helper()

	if policy.WorkerTimeoutSeconds != 60 || policy.ValidationTimeoutSeconds != 10 || policy.CheckpointStepLimit != 2 || policy.MaxWorkerAttempts != 1 || policy.MaxIntegrationAttempts != 2 || policy.IncludeLaterTickets || policy.ConfigRevision != contenthash.Sum([]byte(body)) {
		t.Fatal("admission did not freeze resolved restrictions", policy)
	}
}

func checkWorkflowConfigRetry(t *testing.T, store *Store, request StartRequest, policy workflowstate.ExecutionPolicy) {
	t.Helper()
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "workflow:\n  version: 1\nwork:\n  max_worker_attempts: 1\n  worker_timeout_seconds: 120\n")

	entry, found, err := store.FindWorkflowEntry(t.Context(), request)
	if err != nil || !found || entry.Policy != policy {
		t.Fatal("entry lookup replaced frozen project policy", entry, err)
	}

	retry, err := store.StartWorkflow(t.Context(), request)
	if err != nil || !retry.Attached || retry.Invocation.Policy != policy || len(retry.Invocation.Attempts) != 2 {
		t.Fatal("retry changed frozen limits or consumed work", retry, err)
	}

	request.Policy = policy

	request.RequestKey = "forged-config"
	if _, err := store.StartWorkflow(t.Context(), request); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("caller forged a retained configuration snapshot", err)
	}

	request.Policy = workflowstate.ExecutionPolicy{MaxWorkerAttempts: 2}
	if _, err := store.StartWorkflow(t.Context(), request); err == nil {
		t.Fatal("invocation relaxed explicit project attempts")
	}
}

func checkWorkflowConfigRestriction(t *testing.T, project string) {
	t.Helper()

	restricted, err := workspaceconfig.ResolveWorkflowConfiguration(project, workflowstate.ReviewHuman)

	mode := workflowConfigTestValue(t, restricted, "review.mode")
	if err != nil || mode.Value != "human_review" || mode.Source != "invocation_restriction" {
		t.Fatal("human restriction did not take precedence", restricted, err)
	}

	if _, err := workspaceconfig.ResolveWorkflowConfiguration(project, workflowstate.ReviewTrustedAutoMerge); err == nil {
		t.Fatal("a restriction enabled automatic policy")
	}
}

// TestWorkflowConfigClosing applies the project deadline to real ordered closing
// commands without rewriting the approved contract or confirming a failed gate.
func TestWorkflowConfigClosing(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "later-closing-command")
	gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "sleep 4"}, {"touch", marker}}}
	fixture := workflowDeliveryFixture(t, "rebase")
	plan := workflowPlanFixture()

	writeWorkspaceConfig(t, fixture.store.layout.ProjectDir(), "work:\n  validation_timeout_seconds: 1\n")

	set, err := fixture.store.Create(t.Context(), "Deadline closing set", Record{Kind: "plan", Set: &workflowstate.WorkflowSet{Schema: WorkflowSchema, Status: "ready", PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: gate}})
	if err != nil {
		t.Fatal(err)
	}

	request := SetClosureRequest{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Repository: fixture.provider.pr.Repository, Branch: "dev"}

	closed, err := fixture.store.CloseWorkflowSet(t.Context(), request, workflowFrontierPorts(fixture))
	if err == nil || closed.Outcome != "closing_validation_failed" || len(closed.Receipt.Validation) != 1 {
		t.Fatal("closing deadline did not retain one failed gate", closed, err)
	}

	if closed.Receipt.ClosingInput.Validation.TimeoutSeconds != 30 || closed.Receipt.ValidationTimeoutSeconds != 1 {
		t.Fatal("closing rewrote the contract or lost its effective deadline", closed.Receipt)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("closing ran a command after the shared deadline", err)
	}
}
