package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
)

// ExecuteWorkflowMerger runs one certified broker driver in a private owned
// checkout. It retains a proposal; it never accepts, commits or publishes it.
func (s *Store) ExecuteWorkflowMerger(ctx context.Context, request MergerRequest, harness MergerHarness, provider PullRequestProvider) (attempt MergerAttempt, err error) {
	return s.executeWorkflowMerger(ctx, request, harness, provider, nil)
}

func (s *Store) executeWorkflowMerger(ctx context.Context, request MergerRequest, harness MergerHarness, provider PullRequestProvider, git MergerRepairGit) (attempt MergerAttempt, err error) {
	harness.Command = slices.Clone(harness.Command)

	claim, prepared, err := s.prepareWorkflowMerger(ctx, request, harness, provider)
	if err != nil {
		return attempt, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return attempt, err
	}
	defer endWorkflowClaimExecution(claim)

	watched, stop := context.WithCancel(ctx)

	s.watchWorkflowStop(watched, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	deadline, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	harness = MergerHarness{Command: prepared.input.Worker, Certificate: *prepared.input.MergerHarnessRef}
	attempt = MergerAttempt{SchemaVersion: 1, Input: *prepared.input.Merger, HarnessRef: harness.Certificate}
	result := s.runWorkflowMerger(deadline, claim, prepared, provider, harness)
	attempt.Harness, attempt.TrustedUsage = result.command, result.usage

	attempt.Checks = result.checks
	if result.err == nil {
		proposal, parseErr := workflowmerge.ParseResult(result.data, attempt.Input, workflowmerge.InputFence(attempt.Input), time.Now().UTC())

		result.err = parseErr
		if parseErr == nil {
			attempt.Proposal = &proposal
		}
	}

	if result.err == nil && git != nil {
		result.err = s.acceptWorkflowMerger(deadline, claim, prepared, provider, git, &attempt)
	}

	if result.err != nil {
		attempt.Failure = workflowFailureText(result.err)
	}

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	err = s.finishWorkflowMerger(cleanup, claim, prepared, provider, &attempt, result.err)

	return attempt, errors.Join(result.err, err)
}

func (s *Store) runWorkflowMerger(ctx context.Context, claim *Claim, prepared workflowPreparedTask, provider PullRequestProvider, harness MergerHarness) workflowMergerDriverResult {
	if err := s.checkWorkflowMergerPrepared(ctx, claim, prepared, provider); err != nil {
		return workflowMergerDriverResult{err: fmt.Errorf("merger execution admission: %w", err)}
	}

	certificate, err := claim.store.checkWorkflowMergerHarness(ctx, *prepared.input.Merger, harness)
	if err != nil {
		return workflowMergerDriverResult{err: err}
	}

	var evidence workflowmerge.Evidence = claim.store
	if err := evidence.CheckWorkflowMergerEvidence(ctx, *prepared.input.Merger, time.Now().UTC()); err != nil {
		return workflowMergerDriverResult{err: err}
	}

	call, err := claim.store.workflowMergerCall(*prepared.input.Merger)
	if err != nil {
		return workflowMergerDriverResult{err: fmt.Errorf("merger evidence input: %w", err)}
	}

	root, err := os.OpenRoot(prepared.input.Lane.Worktree)
	if err != nil {
		return workflowMergerDriverResult{err: fmt.Errorf("merger checkout access: %w", err)}
	}
	defer func() { _ = root.Close() }() //nolint:errcheck

	tools := workflowmerge.NewTools(*prepared.input.Merger, root, func(ctx context.Context) error { return s.checkWorkflowMergerPrepared(ctx, claim, prepared, provider) })
	result := runWorkflowMergerDriver(ctx, workflowMergerDriverInput{Claim: claim, Prepared: prepared, Tools: tools, Call: call})

	result.checks = tools.Checks()
	if err := checkWorkflowMergerHarnessUsage(*prepared.input.Merger, certificate, result.usage); err != nil {
		result.usage, result.err = nil, errors.Join(result.err, err)
	}

	return result
}
