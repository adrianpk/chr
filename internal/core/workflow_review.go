package core

import (
	"context"
	"errors"
	"time"
)

const (
	workflowSliceReviewPurpose = "slice_review"
	workflowReviewRepairLimit  = 3
)

// ErrWorkflowReviewRepairBudget requires an explicit operator action.
var ErrWorkflowReviewRepairBudget = errors.New("automatic review repair budget exhausted")

// WorkflowReviewPolicy identifies the current effective publication policy.
type WorkflowReviewPolicy struct {
	Mode     ReviewWorkflowMode `json:"mode"`
	Method   string             `json:"method"`
	Source   string             `json:"source"`
	Revision string             `json:"revision"`
}

// WorkflowReviewInput binds a decision to host validation and canonical inputs.
type WorkflowReviewInput struct {
	Schema        string                      `json:"schema"`
	InvocationID  string                      `json:"invocation_id"`
	Unit          WorkflowRef                 `json:"unit"`
	CandidateID   string                      `json:"candidate_id"`
	CandidateRef  string                      `json:"candidate_ref"`
	CandidateHash string                      `json:"candidate_hash"`
	Slice         WorkflowSliceInput          `json:"slice"`
	PullRequest   WorkflowPullRequestSnapshot `json:"pull_request"`
	TargetCommit  string                      `json:"target_commit"`
	Policy        WorkflowReviewPolicy        `json:"policy"`
}

// WorkflowReviewDecision is returned by the distinct supervised review role.
// Its binding hash must name the exact input supplied on stdin.
type WorkflowReviewDecision struct {
	BindingHash   string              `json:"binding_hash"`
	Result        AgenticReviewResult `json:"result"`
	RepairTaskKey string              `json:"repair_task_key,omitempty"`
}

// WorkflowReviewBinding retains one current decision and immutable evidence.
type WorkflowReviewBinding struct {
	Round         int         `json:"round"`
	Unit          WorkflowRef `json:"unit"`
	State         string      `json:"state"`
	InputRef      string      `json:"input_ref"`
	InputHash     string      `json:"input_hash"`
	ResultRef     string      `json:"result_ref"`
	ResultHash    string      `json:"result_hash"`
	DecisionCount int         `json:"decision_count"`
	RepairTaskKey string      `json:"repair_task_key"`
}

// WorkflowReviewRepairBudget survives candidate replacement and continuation.
type WorkflowReviewRepairBudget struct {
	Unit     WorkflowRef `json:"unit"`
	Consumed int         `json:"consumed"`
}

// ReviewWorkflowSlice applies current policy. Human review retains a wait;
// trusted mode grants validated inputs; agentic mode runs a separate owned role.
// It never publishes a target or completes a task.
func (s *WorkflowStore) ReviewWorkflowSlice(ctx context.Context, id string, provider WorkflowPullRequestProvider, reviewer []string) (result WorkflowReviewBinding, err error) {
	claim, prepared, binding, err := s.prepareWorkflowReview(ctx, id, provider, reviewer)

	if err != nil || claim == nil {
		if err == nil && binding.State == "needs_repair" {
			_, err = s.RepairWorkflowSlice(ctx, id, binding.RepairTaskKey)
		}

		return binding, err
	}
	defer func() { err = errors.Join(err, closeWorkflowSliceCapacity(claim)) }()

	if err := beginWorkflowClaimExecution(claim); err != nil {
		return result, err
	}
	defer endWorkflowClaimExecution(claim)

	watched, stop := s.watchWorkflowStop(ctx, claim)
	defer stop()

	deadline, cancel := context.WithDeadline(watched, prepared.leaseUntil)
	defer cancel()

	command, executionErr := workflowRunCommand(deadline, claim, prepared, reviewer)
	if command.Truncated {
		executionErr = errors.Join(executionErr, errors.New("workflow review output exceeded its bounds"))
	}

	if errors.Is(executionErr, ErrWorkflowOwned) {
		return result, executionErr
	}

	decision, parseErr := parseWorkflowReviewDecision(command.Stdout, prepared.input.Review)

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	result, err = s.finishWorkflowReview(cleanup, claim, prepared, provider, decision, errors.Join(executionErr, parseErr), watched.Err() != nil)
	if err == nil && result.State == "needs_repair" {
		_, err = s.RepairWorkflowSlice(ctx, id, decision.RepairTaskKey)
	}

	return result, err
}

// ApproveWorkflowSlice authorizes only the displayed immutable binding. The
// current policy, validated candidate, canonical PR and target are read again.
func (s *WorkflowStore) ApproveWorkflowSlice(ctx context.Context, id, bindingHash string, provider WorkflowPullRequestProvider) (result WorkflowReviewBinding, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		result, readErr = owner.approveWorkflowSlice(ctx, run, bindingHash, provider)

		return readErr
	})

	return result, err
}

// CheckWorkflowReviewGrant re-observes the exact authorized inputs. Publication
// must call this immediately before recording a guarded submission.
func (s *WorkflowStore) CheckWorkflowReviewGrant(ctx context.Context, id string, provider WorkflowPullRequestProvider) (input WorkflowReviewInput, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		input, readErr = owner.checkWorkflowReviewGrant(ctx, run, provider)

		return readErr
	})

	return input, err
}
