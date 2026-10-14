package app

import (
	"context"

	"forge.adrianpk.com/chr/internal/core"
)

// WorkflowIntegrationControlGuard contains displayed revisions and candidate
// identity. The service derives the private claim fence from retained ownership.
type WorkflowIntegrationControlGuard struct {
	EntryID        string
	EntryRevision  uint64
	BudgetRevision uint64
	CandidateHash  string
}

func (s service) executeWorkflowIntegrationControl(ctx context.Context, request WorkflowControlRequest) error {
	guard, err := s.workflowIntegrationControlGuard(ctx, request)
	if err != nil {
		return err
	}

	switch request.Action {
	case "grant":
		if request.BudgetGrant == nil {
			return core.ErrWorkflowConflict
		}

		_, err = s.GrantWorkflowIntegration(ctx, core.WorkflowIntegrationGrantRequest{WorkflowIntegrationBudgetGuard: guard, Grant: *request.BudgetGrant})
	case "answer_decision":
		if request.DecisionAnswer == nil {
			return core.ErrWorkflowConflict
		}

		_, err = s.AnswerWorkflowIntegrationDecision(ctx, core.WorkflowIntegrationAnswerRequest{WorkflowIntegrationBudgetGuard: guard, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash, Answer: *request.DecisionAnswer})
	default:
		return core.ErrWorkflowConflict
	}

	return err
}

func (s service) workflowIntegrationControlGuard(ctx context.Context, request WorkflowControlRequest) (core.WorkflowIntegrationBudgetGuard, error) {
	store, err := s.workflowStore()
	if err != nil {
		return core.WorkflowIntegrationBudgetGuard{}, err
	}

	inspection, err := store.InspectWorkflow(ctx, request.ID)
	if err != nil {
		return core.WorkflowIntegrationBudgetGuard{}, err
	}

	if request.Integration == nil || inspection.Invocation.Revision != request.ExpectedRevision || inspection.Integration == nil {
		return core.WorkflowIntegrationBudgetGuard{}, core.ErrWorkflowConflict
	}

	shown, expected := workflowIntegrationStatus(inspection.Integration), request.Integration
	if !workflowIntegrationDisplayedGuardMatches(*expected, *shown) {
		return core.WorkflowIntegrationBudgetGuard{}, core.ErrWorkflowConflict
	}

	guard := core.WorkflowIntegrationBudgetGuard{InvocationID: request.ID, InvocationRevision: request.ExpectedRevision, BudgetRevision: expected.BudgetRevision, EntryID: expected.EntryID, EntryRevision: expected.EntryRevision}
	if inspection.Integration.Entry != nil {
		guard.ClaimGeneration = inspection.Integration.Entry.ClaimGeneration
	}

	return guard, nil
}

func workflowIntegrationDisplayedGuardMatches(expected WorkflowIntegrationControlGuard, shown WorkflowIntegrationStatus) bool {
	return expected.EntryID == shown.EntryID && expected.EntryRevision == shown.EntryRevision && expected.BudgetRevision == shown.Budget.Revision && expected.CandidateHash == shown.CandidateHash
}
