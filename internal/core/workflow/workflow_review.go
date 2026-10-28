package workflow

import (
	"context"
	"errors"
	"time"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const (
	workflowSliceReviewPurpose = "slice_review"
	workflowReviewRepairLimit  = 3
)

// ErrWorkflowReviewRepairBudget requires an explicit operator action.
var ErrWorkflowReviewRepairBudget = errors.New("automatic review repair budget exhausted")

// ReviewRepairBudget survives candidate replacement and continuation.
type ReviewRepairBudget struct {
	Unit     workflowstate.Ref `json:"unit"`
	Consumed int               `json:"consumed"`
}

// ReviewWorkflowSlice applies current policy. Human review retains a wait;
// trusted mode grants validated inputs; agentic mode runs a separate owned role.
// It never publishes a target or completes a task.
func (s *Store) ReviewWorkflowSlice(ctx context.Context, id string, provider PullRequestProvider, reviewer []string) (result workflowstate.ReviewBinding, err error) {
	return s.reviewWorkflowSlice(ctx, id, provider, reviewer, nil)
}

// ReviewWorkflowSliceWithHarness shares the invocation ceiling with executor and merger calls.
func (s *Store) ReviewWorkflowSliceWithHarness(ctx context.Context, id string, provider PullRequestProvider, harness AgentHarness) (workflowstate.ReviewBinding, error) {
	return s.reviewWorkflowSlice(ctx, id, provider, harness.Command, &harness)
}

func (s *Store) reviewWorkflowSlice(ctx context.Context, id string, provider PullRequestProvider, reviewer []string, harness *AgentHarness) (result workflowstate.ReviewBinding, err error) {
	claim, prepared, binding, err := s.prepareWorkflowReview(ctx, id, provider, reviewer, harness)

	if err != nil || claim == nil {
		if err == nil && binding.State == "needs_repair" {
			err = s.repairWorkflowReviewedSlice(ctx, id, binding.RepairTaskKey, provider)
		}

		return binding, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return result, err
	}
	defer endWorkflowClaimExecution(claim)

	watched, stop := context.WithCancel(ctx)

	s.watchWorkflowStop(watched, workflowStopWatch{Claim: claim, Cancel: stop})
	defer stop()

	deadline, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	command, executionErr := workflowRunAgentCommand(deadline, claim, prepared, reviewer)
	if command.Truncated {
		executionErr = errors.Join(executionErr, errors.New("workflow review output exceeded its bounds"))
	}

	if errors.Is(executionErr, ErrWorkflowOwned) {
		return result, executionErr
	}

	decisions := workflowReviewDecisions{store: s, claim: claim, prepared: prepared, provider: provider}
	completed, err := workflowreview.Complete(ctx, watched, workflowreview.Execution{Input: prepared.input.Review, Output: command, Failure: executionErr}, decisions)

	result = completed.Binding
	if err == nil && result.State == "needs_repair" {
		err = s.repairWorkflowReviewedSlice(ctx, id, completed.RepairTaskKey, provider)
	}

	return result, err
}

// ApproveWorkflowSlice authorizes only the displayed immutable binding. The
// current policy, validated candidate, canonical PR and target are read again.
func (s *Store) ApproveWorkflowSlice(ctx context.Context, id, bindingHash string, provider PullRequestProvider, expected ...uint64) (result workflowstate.ReviewBinding, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if err := checkWorkflowRevision(invocationLocation.Run.Invocation, expected); err != nil {
			return err
		}

		result, readErr = invocationLocation.Owner.approveWorkflowSlice(ctx, invocationLocation.Run, bindingHash, provider, nil)

		return readErr
	})

	return result, err
}

// CheckWorkflowReviewGrant re-observes the exact authorized inputs. Publication
// must call this immediately before recording a guarded submission.
func (s *Store) CheckWorkflowReviewGrant(ctx context.Context, id string, provider PullRequestProvider) (input workflowstate.ReviewInput, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation2, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		input, readErr = invocationLocation2.Owner.checkWorkflowReviewGrant(ctx, invocationLocation2.Run, provider)

		return readErr
	})

	return input, err
}
