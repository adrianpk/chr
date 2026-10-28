package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type serviceWorkEntryProvider struct {
	*serviceWorkflowSubmissionFixture
	remote, repository string
}

func (p *serviceWorkEntryProvider) ReadWorkflowBranch(ctx context.Context, repository, branch string) (string, error) {
	if repository != p.repository {
		return "", workflowstate.ErrConflict
	}

	output, err := exec.CommandContext(ctx, "git", "--git-dir", p.remote, "rev-parse", "refs/heads/"+branch).Output()

	return strings.TrimSpace(string(output)), err
}

func (p *serviceWorkEntryProvider) CreateWorkflowPullRequest(ctx context.Context, request workflowstate.PullRequestRequest) (workflowstate.PullRequestSnapshot, error) {
	head, err := p.ReadWorkflowBranch(ctx, request.Repository, request.HeadBranch)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	base, err := p.ReadWorkflowBranch(ctx, request.Repository, request.BaseBranch)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	p.creates++
	p.pr = workflowstate.PullRequestSnapshot{URL: p.repository + "/pulls/7", Repository: p.repository, HeadRepository: p.repository, HeadBranch: request.HeadBranch, BaseBranch: request.BaseBranch, HeadCommit: head, BaseCommit: base, State: "open"}

	return p.pr, nil
}

func (p *serviceWorkEntryProvider) ReadWorkflowPullRequest(ctx context.Context, url string) (workflowstate.PullRequestSnapshot, error) {
	if url != p.pr.URL {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	head, err := p.ReadWorkflowBranch(ctx, p.repository, p.pr.HeadBranch)
	if err != nil {
		return workflowstate.PullRequestSnapshot{}, err
	}

	base, err := p.ReadWorkflowBranch(ctx, p.repository, p.pr.BaseBranch)
	p.pr.HeadCommit, p.pr.BaseCommit = head, base

	return p.pr, err
}

func serviceWorkNativeFixture(t *testing.T, plans ...generationcore.GeneratedWorkflowPlan) (service, workflowcore.Record, workflowcore.StartRequest, *serviceWorkEntryProvider) {
	t.Helper()

	operations, record, request := serviceWorkflowTicketFixture(t, plans...)
	project := operations.projectContext.ProjectDir
	writeCLIWorkspaceConfig(t, project, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n")

	remote := filepath.Join(t.TempDir(), "canonical.git")
	serviceWorkflowDeliveryGit(t, project, "init", "--bare", remote)
	serviceWorkflowDeliveryGit(t, project, "push", remote, request.Selection.Target.Commit+":refs/heads/dev")
	serviceWorkflowDeliveryGit(t, project, "remote", "add", "origin", request.Selection.Target.Repository+".git")

	provider := &serviceWorkEntryProvider{serviceWorkflowSubmissionFixture: &serviceWorkflowSubmissionFixture{serviceWorkflowPreparationFixture: &serviceWorkflowPreparationFixture{&serviceWorkflowReviewFixture{&serviceWorkflowPRFixture{}}}}, remote: remote, repository: request.Selection.Target.Repository}
	operations.workflowPullRequests = provider
	operations.workflowIntegrationGit = workflowcore.NewWorkflowIntegrationGit(request.Selection.Target.Repository, remote)
	operations.workflowPreparationGit = workflowcore.NewWorkflowPreparationGit(request.Selection.Target.Repository, remote)
	operations.workflowWorker = []string{"sh", "-c", "printf 'Delivered\\n' > result.txt && git add result.txt && git -c user.name=Test -c user.email=test@example.com commit -qm 'Task result'"}

	return operations, record, request, provider
}

// TestServiceWorkNative maps preview/run-all onto the shared loose lifecycle,
// excludes grouped work and uses actual source publication before one PR.
func TestServiceWorkNative(t *testing.T) {
	operations, _, _, provider := serviceWorkNativeFixture(t)

	checkServiceWorkNativePreview(t, operations)
	checkServiceWorkNativeSteps(t, operations)

	result, err := operations.StartAllReadyAgentWork()
	if err != nil || result.Workflow == nil || result.Workflow.Review == nil || result.Workflow.Review.State != "waiting_approval" || provider.creates != 1 || serviceWorkTaskAttempts(*result.Workflow) != 1 || len(result.Workflow.Members) != 1 {
		t.Fatalf("legacy run-all result: %+v, creates=%d, err=%v", result.Workflow, provider.creates, err)
	}

	checkServiceWorkNativeDisplay(t, result)
	checkServiceWorkNativeAttachment(t, operations, *result.Workflow, provider)
}

func checkServiceWorkNativeDisplay(t *testing.T, result AgentWorkResult) {
	t.Helper()

	var output bytes.Buffer
	if err := printAgentWorkRunAllExecution(&output, result); err != nil {
		t.Fatal(err)
	}

	for _, line := range []string{"authority: tickets", "scope: loose", "excluded_slice_owned: 6", "invocation: " + result.Workflow.ID} {
		if !strings.Contains(output.String(), line) {
			t.Fatalf("missing authority output %q: %s", line, output.String())
		}
	}
}

func checkServiceWorkNativeAttachment(t *testing.T, operations service, run workflowcore.Invocation, provider *serviceWorkEntryProvider) {
	t.Helper()

	if _, err := operations.StopWorkflow(t.Context(), run.ID, "pause"); err != nil {
		t.Fatal(err)
	}

	project := operations.projectContext.ProjectDir
	writeServiceWorkspaceFile(t, filepath.Join(project, "external.txt"), "Canonical movement\n")
	serviceWorkflowDeliveryGit(t, project, "add", "external.txt")
	serviceWorkflowDeliveryGit(t, project, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "-qm", "External target")
	serviceWorkflowDeliveryGit(t, project, "push", provider.remote, "HEAD:refs/heads/dev")

	attached, err := operations.StartAllReadyAgentWork()
	if err != nil || attached.Workflow == nil || attached.Workflow.ID != run.ID || attached.Workflow.Selection.Target != run.Selection.Target || attached.Workflow.Status != workflowstate.InvocationPaused || serviceWorkTaskAttempts(*attached.Workflow) != 1 || provider.creates != 1 {
		t.Fatal("entry replaced an unfinished target or replayed work", err)
	}
}

// TestServiceWorkTicket maps exact independent/grouped tickets without selecting
// their parent extent. Tracker compatibility refuses both before any worker.
func TestServiceWorkTicket(t *testing.T) {
	for _, kind := range []string{"independent", "grouped"} {
		t.Run(kind, func(t *testing.T) {
			operations, _, _, _ := serviceWorkNativeFixture(t)

			key := "PRJ-90"
			if kind == "grouped" {
				key = serviceWorkGroupedKey(t, operations)
			}

			result, err := operations.StartAgentWorkTicket(key)
			if err != nil || result.Workflow == nil || len(result.Workflow.Members) != 1 || result.Workflow.Members[0].Admission.Key != key || serviceWorkTaskAttempts(*result.Workflow) != 1 {
				t.Fatalf("single-ticket result: %+v, err=%v", result.Workflow, err)
			}

			writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "work:\n  state_authority: tracker_compat\n")

			if _, err := operations.StartAgentWorkTicket(key); !errors.Is(err, workflowcore.ErrWorkflowAuthority) {
				t.Fatal("compatibility duplicated adopted execution", err)
			}
		})
	}
}

// TestServiceWorkAuthority rejects invalid authority and excessive native limits
// before runtime initialization, invocation allocation or worker creation.
func TestServiceWorkAuthority(t *testing.T) {
	cases := []struct {
		name, config string
		limit        int
	}{
		{"invalid authority", "work:\n  state_authority: both\n", 0},
		{"oversized native call", "work:\n  state_authority: tickets\n", 129},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := newTestWorkspace(t)
			writeCLIWorkspaceConfig(t, project, tc.config)

			operations := newTestService(t, project)
			if _, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: tc.limit}); err == nil {
				t.Fatal("invalid authority request accepted")
			}

			if _, err := os.Stat(filepath.Join(project, "ops", "default", "run")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rejected entry allocated work", err)
			}
		})
	}
}

func serviceWorkGroupedKey(t *testing.T, operations service) string {
	t.Helper()

	store, err := operations.workflowStore()
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := store.Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && record.Task.SetID != "" && record.Task.Order == 1 {
			return record.Task.Key
		}
	}

	t.Fatal("grouped fixture has no first task")

	return ""
}

func serviceWorkTaskAttempts(run workflowcore.Invocation) int {
	count := 0

	for _, attempt := range run.Attempts {
		if attempt.Purpose == "" || attempt.Purpose == "task" {
			count++
		}
	}

	return count
}

func checkServiceWorkNativePreview(t *testing.T, operations service) {
	t.Helper()

	preview, err := operations.PreviewAllReadyAgentWork()
	if err != nil || preview.Authority != workspaceconfig.WorkStateAuthorityTickets || preview.EligibleCount != 1 || preview.Excluded["slice_owned"] != 6 || len(preview.SelectedQueue) != 1 || preview.SelectedQueue[0].Key != "PRJ-90" {
		t.Fatal("native preview broadened scope", preview, err)
	}
}

func checkServiceWorkNativeSteps(t *testing.T, operations service) {
	t.Helper()

	for range 2 {
		result, err := operations.StartAgentWorkLoop(t.Context(), AgentWorkSelectionOptions{MaxIterations: 1})
		if err != nil || result.Workflow == nil || serviceWorkTaskAttempts(*result.Workflow) != 1 || result.Workflow.ConfirmedCursor != 0 {
			t.Fatal("bounded native step replayed a worker", err)
		}
	}
}
