package core

import (
	"context"
	"encoding/json"
	"slices"
	"time"
)

// WorkflowControlIntegrationGuard is the displayed integration image. Private
// claim credentials are derived by the native owner after these fields match.
type WorkflowControlIntegrationGuard struct {
	EntryID        string
	EntryRevision  uint64
	BudgetRevision uint64
	CandidateHash  string
}

// WorkflowDecisionControlRequest binds an attributable human decision to the
// displayed invocation, candidate, review, budget and question images.
type WorkflowDecisionControlRequest struct {
	Binding          WorkflowMutationBinding
	Integration      *WorkflowControlIntegrationGuard
	Grant            *WorkflowIntegrationGrant
	Answer           *WorkflowIntegrationAnswer
	Action           string
	ID               string
	ExpectedRevision uint64
	ReviewHash       string
	DecisionRevision uint64
	DecisionHash     string
}

// WorkflowControlResult is the original confirmed mutation receipt. Its revision
// describes acceptance; callers observe current workflow state separately.
type WorkflowControlResult struct {
	RequestID      string `json:"request_id"`
	Action         string `json:"action"`
	InvocationID   string `json:"invocation_id"`
	ArgumentDigest string `json:"argument_digest"`
	ReceiptRef     string `json:"receipt_ref"`
	PriorRevision  uint64 `json:"prior_revision"`
	Revision       uint64 `json:"revision"`
}

// ControlWorkflowDecision replays a completed request before checking mutable
// revisions. New effects and their receipt share the existing native journal.
func (s *WorkflowStore) ControlWorkflowDecision(ctx context.Context, request WorkflowDecisionControlRequest, provider WorkflowPullRequestProvider, check WorkflowAdmissionCheck) (result WorkflowControlResult, err error) {
	request, digest, err := freezeWorkflowDecisionControl(request)
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

		result = WorkflowControlResult{RequestID: request.Binding.RequestID, Action: request.Action, InvocationID: run.ID, ArgumentDigest: digest, ReceiptRef: workflowMutationReceiptPath(request.Binding.RequestID), PriorRevision: run.Revision, Revision: run.Revision + 1}
		commit := &workflowDecisionControlCommit{receipt: workflowControlReceipt{Schema: "chr-workflow-control/v1", State: "confirmed", ProjectID: owner.projectID, Namespace: owner.layout.Namespace(), Operator: request.Binding.Operator, Result: result}, check: check}

		return owner.applyWorkflowDecisionControl(ctx, run, request, guard, provider, commit)
	})

	return result, err
}

func freezeWorkflowDecisionControl(request WorkflowDecisionControlRequest) (WorkflowDecisionControlRequest, string, error) {
	if err := request.Binding.Validate(); err != nil {
		return request, "", err
	}

	if !validWorkflowRequestKey(request.ID) || request.ExpectedRevision == 0 || !slices.Contains([]string{"approve", "grant", "answer_decision"}, request.Action) {
		return request, "", ErrWorkflowConflict
	}

	data, err := json.Marshal(request)
	if err != nil || len(data) > 32768 {
		return request, "", ErrWorkflowConflict
	}

	var frozen WorkflowDecisionControlRequest
	if err := json.Unmarshal(data, &frozen); err != nil {
		return frozen, "", err
	}

	if err := validateWorkflowDecisionControlPayload(frozen); err != nil {
		return frozen, "", err
	}

	return frozen, workflowHash(data), nil
}

func validateWorkflowDecisionControlPayload(request WorkflowDecisionControlRequest) error {
	switch request.Action {
	case "approve":
		if request.ReviewHash == "" || request.Grant != nil || request.Answer != nil {
			return ErrWorkflowConflict
		}
	case "grant":
		return validateWorkflowControlGrant(request)
	case "answer_decision":
		return validateWorkflowControlAnswer(request)
	}

	return nil
}

func (s *WorkflowStore) applyWorkflowDecisionControl(ctx context.Context, run workflowInvocationFile, request WorkflowDecisionControlRequest, guard WorkflowIntegrationBudgetGuard, provider WorkflowPullRequestProvider, commit *workflowDecisionControlCommit) error {
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
		return ErrWorkflowConflict
	}
}

func validateWorkflowControlGrant(request WorkflowDecisionControlRequest) error {
	if request.Integration == nil || request.Grant == nil || request.Answer != nil {
		return ErrWorkflowConflict
	}

	if request.Grant.ID != request.Binding.RequestID || request.Grant.Actor != request.Binding.Operator {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowControlAnswer(request WorkflowDecisionControlRequest) error {
	if request.Integration == nil || request.Answer == nil || request.Grant != nil {
		return ErrWorkflowConflict
	}

	if request.Answer.ID != request.Binding.RequestID || request.Answer.Actor != request.Binding.Operator {
		return ErrWorkflowConflict
	}

	return nil
}
