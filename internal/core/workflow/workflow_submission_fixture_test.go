package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowSubmissionProviderFixture struct {
	*workflowPreparationProviderFixture
	remote  string
	mu      sync.Mutex
	calls   int
	lost    bool
	unknown bool
	sealed  map[string]bool
	before  func(context.Context, workflowstate.Submission)
	after   func(workflowstate.Submission)
}

func newWorkflowSubmissionFixture(t *testing.T, method string) (workflowPreparationFixture, *workflowSubmissionProviderFixture) {
	t.Helper()

	f := newWorkflowPreparationFixture(t, method, workflowstate.Validation{WaiverReason: "Submission fixture validates real Git publication"})

	p := &workflowSubmissionProviderFixture{workflowPreparationProviderFixture: f.provider, remote: f.remote, sealed: map[string]bool{}}
	if _, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, p, f.git); err != nil {
		t.Fatal(err)
	}

	if _, err := f.store.ReviewWorkflowSlice(t.Context(), f.run.ID, p, nil); err != nil {
		t.Fatal(err)
	}

	return f, p
}

func (p *workflowSubmissionProviderFixture) ReadWorkflowPullRequest(ctx context.Context, _ string) (workflowstate.PullRequestSnapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pr := p.pr

	data, err := os.ReadFile(filepath.Join(p.remote, "workflow-provider-pr.json"))
	if err == nil {
		err = json.Unmarshal(data, &pr)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return pr, err
	}

	pr.BaseCommit, err = p.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)

	return pr, err
}

func (p *workflowSubmissionProviderFixture) ReadWorkflowBranch(ctx context.Context, _, branch string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", p.remote, "rev-parse", "refs/heads/"+branch).Output()

	return strings.TrimSpace(string(output)), err
}

func (p *workflowSubmissionProviderFixture) SubmitWorkflowIntegration(ctx context.Context, submission workflowstate.Submission) (workflowstate.SubmissionResponse, error) {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()

	if p.before != nil {
		p.before(ctx, submission)
	}

	if p.unknown {
		return workflowstate.SubmissionResponse{}, context.DeadlineExceeded
	}

	if err := p.publish(ctx, submission); err != nil {
		target, readErr := p.ReadWorkflowBranch(ctx, submission.Guard.PullRequest.Repository, submission.Guard.Git.TargetBranch)
		if ctx.Err() == nil && readErr == nil && target != submission.Guard.Git.ResultCommit {
			p.seal(submission.ID)
		}

		return workflowstate.SubmissionResponse{}, err
	}

	if p.after != nil {
		p.after(submission)
	}

	response := workflowstate.SubmissionResponse{Accepted: true, ResultCommit: submission.Guard.Git.ResultCommit}
	if p.lost {
		return response, context.DeadlineExceeded
	}

	return response, nil
}

func (p *workflowSubmissionProviderFixture) publish(ctx context.Context, submission workflowstate.Submission) error {
	guard := submission.Guard

	command := exec.CommandContext(ctx, "git", "-C", p.remote, "fetch", "--quiet", "--no-tags", "--", submission.PreparedLane.Worktree, guard.Git.ResultCommit)
	if _, err := command.CombinedOutput(); err != nil {
		return err
	}

	command = exec.CommandContext(ctx, "git", "-C", p.remote, "update-ref", "--stdin")

	command.Stdin = strings.NewReader("start\nverify refs/heads/" + guard.PullRequest.HeadBranch + " " + guard.Git.SourceCommit + "\nupdate refs/heads/" + guard.Git.TargetBranch + " " + guard.Git.ResultCommit + " " + guard.Git.BaseCommit + "\nprepare\ncommit\n")
	if _, err := command.CombinedOutput(); err != nil {
		return err
	}

	pr := guard.PullRequest
	pr.State, pr.Merged, pr.ResultCommit = "closed", true, guard.Git.ResultCommit

	data, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	return opsfile.WriteAtomic(filepath.Join(p.remote, "workflow-provider-pr.json"), data)
}

func (p *workflowSubmissionProviderFixture) ObserveWorkflowSubmission(_ context.Context, submission workflowstate.Submission) (workflowstate.SubmissionObservation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return workflowstate.SubmissionObservation{SubmissionID: submission.ID, Unpublished: p.sealed[submission.ID]}, nil
}

func (p *workflowSubmissionProviderFixture) seal(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sealed[id] = true
}

func workflowSubmissionObserver(f workflowPreparationFixture) OSWorkflowIntegrationGit {
	return NewWorkflowIntegrationGit(f.provider.pr.Repository, f.remote)
}
