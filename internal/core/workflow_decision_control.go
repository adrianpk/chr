package core

import (
	"context"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ControlWorkflowDecision replays a completed request before checking mutable
// revisions. New effects and their receipt share the existing native journal.
func (s *WorkflowStore) ControlWorkflowDecision(ctx context.Context, request workflowstate.ControlMutationRequest, provider WorkflowPullRequestProvider, check WorkflowAdmissionCheck) (result workflowstate.ControlResult, err error) {
	if !slices.Contains([]string{"approve", "grant", "answer_decision"}, request.Action) {
		return result, workflowstate.ErrConflict
	}

	request, digest, err := freezeWorkflowControlMutation(request)
	if err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		run, owner, err := s.locateWorkflowInvocation(ctx, request.ID)
		if err != nil {
			return err
		}

		prior, found, err := owner.readWorkflowControlReceipt(request, digest)
		if found || err != nil {
			result = prior

			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		if err := owner.reserveWorkflowControlBinding(request, digest); err != nil {
			return err
		}

		guard, err := owner.workflowDecisionControlGuard(ctx, run, request)
		if err != nil {
			return err
		}

		commit := owner.newWorkflowControlCommit(request, digest, check)

		if err := owner.applyWorkflowDecisionControl(ctx, run, request, guard, provider, commit); err != nil {
			return err
		}

		result = commit.receipt.Result

		return nil
	})

	return result, err
}

func (s *WorkflowStore) applyWorkflowDecisionControl(ctx context.Context, run workflowInvocationFile, request workflowstate.ControlMutationRequest, guard WorkflowIntegrationBudgetGuard, provider WorkflowPullRequestProvider, commit *workflowControlCommit) error {
	switch request.Action {
	case "approve":
		_, err := s.approveWorkflowSlice(ctx, run, request.ReviewHash, provider, commit)

		return err
	case "grant":
		_, err := s.grantWorkflowIntegration(ctx, run, WorkflowIntegrationGrantRequest{WorkflowIntegrationBudgetGuard: guard, Grant: *request.Grant}, commit)

		return err
	case "answer_decision":
		_, err := s.answerWorkflowDecision(ctx, run, WorkflowIntegrationAnswerRequest{WorkflowIntegrationBudgetGuard: guard, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash, Answer: *request.Answer}, provider, commit)

		return err
	default:
		return workflowstate.ErrConflict
	}
}

func validateWorkflowControlGrant(request workflowstate.ControlMutationRequest) error {
	if request.Integration == nil || request.Grant == nil || request.Answer != nil {
		return workflowstate.ErrConflict
	}

	if request.Grant.ID != request.Binding.RequestID || request.Grant.Actor != request.Binding.Operator {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowControlAnswer(request workflowstate.ControlMutationRequest) error {
	if request.Integration == nil || request.Answer == nil || request.Grant != nil {
		return workflowstate.ErrConflict
	}

	if request.Answer.ID != request.Binding.RequestID || request.Answer.Actor != request.Binding.Operator {
		return workflowstate.ErrConflict
	}

	return nil
}
