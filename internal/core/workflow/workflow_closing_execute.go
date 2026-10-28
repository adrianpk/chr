package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// workflowClosingHost owns process lifetime and checkout IO, outside persistence.
type workflowClosingHost struct {
	store *Store
}

func (host workflowClosingHost) execute(ctx context.Context, claim *workflowClosingClaim) (workflowstate.ClosingReceipt, error) {
	ctx, cancel := context.WithDeadline(ctx, claim.state.LeaseUntil)
	defer cancel()

	if claim.state.InvocationID != "" {
		watched, stop := context.WithCancel(ctx)

		watchWorkflowStop(watched, host.store, workflowStopWatch{Claim: &Claim{InvocationID: claim.state.InvocationID}, Cancel: stop})
		defer stop()

		ctx = watched
	}

	workspace := workflowClosingWorkspace{host: host, claim: claim}

	return workflowcloseout.Execute(ctx, workflowcloseout.Execution{Input: claim.input, Evidence: claim.state.Input}, workspace, workspace)
}

func (host workflowClosingHost) prepareCheckout(ctx context.Context, claim *workflowClosingClaim) (string, error) {
	common, err := projectcore.ResolveRepositoryDirectory(ctx, host.store.layout.ProjectDir())
	if err != nil {
		return "", err
	}

	parent := filepath.Join(common, "chr-workflow-closing", claim.worker.InvocationID)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", err
	}

	worktree := filepath.Join(parent, fmt.Sprintf("%04d", claim.state.Generation))
	if err := os.Mkdir(worktree, 0o700); err != nil {
		return "", err
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: worktree}}
	for _, argv := range [][]string{{"init", "--quiet"}, {"fetch", "--quiet", "--no-tags", "--", host.store.layout.ProjectDir(), claim.input.Target.Commit}, {"checkout", "--quiet", "--detach", claim.input.Target.Commit}} {
		if _, err := reader.Output(ctx, argv...); err != nil {
			return "", err
		}
	}

	return worktree, checkWorkflowClosingCheckout(ctx, worktree, claim.input)
}

func checkWorkflowClosingCheckout(ctx context.Context, directory string, input workflowstate.ClosingInput) error {
	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: directory}}
	for expression, expected := range map[string]string{"HEAD^{commit}": input.Target.Commit, "HEAD^{tree}": input.Tree} {
		actual, err := reader.Output(ctx, "rev-parse", "--verify", expression)
		if err != nil || actual != expected {
			return errorsOrConflict(err)
		}
	}

	status, err := reader.Output(ctx, "status", "--porcelain", "--untracked-files=all")
	if err != nil || status != "" {
		return errorsOrConflict(err)
	}

	return nil
}

func workflowClosingPrepared(claim *workflowClosingClaim, directory string) workflowPreparedTask {
	return workflowPreparedTask{input: TaskInput{Lane: workflowstate.Lane{Worktree: directory}}, inputRef: claim.state.Input.Ref, inputHash: claim.state.Input.Hash, data: claim.data, leaseUntil: claim.state.LeaseUntil}
}
