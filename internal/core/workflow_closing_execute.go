package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) executeWorkflowClosing(ctx context.Context, claim *workflowClosingClaim) (WorkflowClosingReceipt, error) {
	ctx, cancel := context.WithDeadline(ctx, claim.state.LeaseUntil)
	defer cancel()

	if claim.state.InvocationID != "" {
		watched, stop := s.watchWorkflowStop(ctx, &WorkflowClaim{InvocationID: claim.state.InvocationID})
		defer stop()

		ctx = watched
	}

	receipt := WorkflowClosingReceipt{WorkflowClosingInput: claim.input, Input: claim.state.Input, Waiver: claim.input.Validation.WaiverReason}
	receipt.Schema, receipt.Outcome = "chr-set-closure/v1", "closing_validation_failed"

	worktree, err := s.prepareWorkflowClosingCheckout(ctx, claim)
	if err != nil {
		return receipt, err
	}

	for _, command := range claim.input.Validation.Commands {
		result, err := s.runWorkflowClosingCommand(ctx, claim, worktree, command)

		receipt.Validation = append(receipt.Validation, result)
		if err != nil || result.ExitCode != 0 {
			return receipt, errorsOrConflict(err)
		}
	}

	if err := checkWorkflowClosingCheckout(ctx, worktree, claim.input); err != nil {
		return receipt, err
	}

	receipt.Outcome = "completed"

	return receipt, ctx.Err()
}

func (s *WorkflowStore) prepareWorkflowClosingCheckout(ctx context.Context, claim *workflowClosingClaim) (string, error) {
	common, err := projectcore.ResolveRepositoryDirectory(ctx, s.layout.ProjectDir())
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
	for _, argv := range [][]string{{"init", "--quiet"}, {"fetch", "--quiet", "--no-tags", "--", s.layout.ProjectDir(), claim.input.Target.Commit}, {"checkout", "--quiet", "--detach", claim.input.Target.Commit}} {
		if _, err := reader.Output(ctx, argv...); err != nil {
			return "", err
		}
	}

	return worktree, checkWorkflowClosingCheckout(ctx, worktree, claim.input)
}

func checkWorkflowClosingCheckout(ctx context.Context, directory string, input WorkflowClosingInput) error {
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
	return workflowPreparedTask{input: WorkflowTaskInput{Lane: WorkflowLane{Worktree: directory}}, inputRef: claim.state.Input.Ref, inputHash: claim.state.Input.Hash, data: claim.data, leaseUntil: claim.state.LeaseUntil}
}

func workflowClosingLeaseSeconds(validation workflowstate.Validation) time.Duration {
	if !validation.Required {
		return time.Minute
	}

	return time.Duration(validation.TimeoutSeconds) * time.Second
}
