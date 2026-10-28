package workflow

import (
	"context"
	"slices"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ControlWorkflowDecision replays a completed request before checking mutable
// revisions. New effects and their receipt share the existing native journal.
func (s *Store) ControlWorkflowDecision(ctx context.Context, request workflowstate.ControlMutationRequest, provider PullRequestProvider, check AdmissionCheck) (result workflowstate.ControlResult, err error) {
	if !slices.Contains([]string{"approve", "grant", "answer_decision"}, request.Action) {
		return result, workflowstate.ErrConflict
	}

	frozen, err := freezeWorkflowControlMutation(request)

	request, digest := frozen.Request, frozen.Digest

	if err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err = s.withLock(ctx, func() error {
		invocationLocation, err := s.locateWorkflowInvocation(ctx, request.ID)
		if err != nil {
			return err
		}

		prior, err := invocationLocation.Owner.readWorkflowControlReceipt(request, digest)
		if prior.Found || err != nil {
			result = prior.Result

			return err
		}

		if err := workflowCheckAuthorization(ctx, check); err != nil {
			return err
		}

		if err := invocationLocation.Owner.reserveWorkflowControlBinding(request, digest); err != nil {
			return err
		}

		guard, err := invocationLocation.Owner.workflowDecisionControlGuard(ctx, invocationLocation.Run, request)
		if err != nil {
			return err
		}

		commit := invocationLocation.Owner.newWorkflowControlCommit(request, digest, check)

		if err := invocationLocation.Owner.applyWorkflowDecisionControl(ctx, invocationLocation.Run, provider, workflowDecisionControl{Request: request, Guard: guard, Commit: commit}); err != nil {
			return err
		}

		result = commit.receipt.Result

		return nil
	})

	return result, err
}

type workflowDecisionControl struct {
	Request workflowstate.ControlMutationRequest
	Guard   IntegrationBudgetGuard
	Commit  *workflowControlCommit
}

func (s *Store) applyWorkflowDecisionControl(ctx context.Context, run workflowInvocationFile, provider PullRequestProvider, accepted workflowDecisionControl) error {
	request, guard, commit := accepted.Request, accepted.Guard, accepted.Commit

	switch request.Action {
	case "approve":
		_, err := s.approveWorkflowSlice(ctx, run, request.ReviewHash, provider, commit)

		return err
	case "grant":
		_, err := s.grantWorkflowIntegration(ctx, run, IntegrationGrantRequest{IntegrationBudgetGuard: guard, Grant: *request.Grant}, commit)

		return err
	case "answer_decision":
		_, err := s.answerWorkflowDecision(ctx, run, IntegrationAnswerRequest{IntegrationBudgetGuard: guard, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash, Answer: *request.Answer}, provider, commit)

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
