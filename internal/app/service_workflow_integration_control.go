package app

import (
	"context"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s service) executeWorkflowIntegrationControl(ctx context.Context, request WorkflowControlRequest) error {
	guard, err := s.workflowIntegrationControlGuard(ctx, request)
	if err != nil {
		return err
	}

	switch request.Action {
	case "grant":
		if request.BudgetGrant == nil {
			return workflowstate.ErrConflict
		}

		_, err = s.GrantWorkflowIntegration(ctx, workflowcore.IntegrationGrantRequest{IntegrationBudgetGuard: guard, Grant: *request.BudgetGrant})
	case "answer_decision":
		if request.DecisionAnswer == nil {
			return workflowstate.ErrConflict
		}

		_, err = s.AnswerWorkflowIntegrationDecision(ctx, workflowcore.IntegrationAnswerRequest{IntegrationBudgetGuard: guard, DecisionRevision: request.DecisionRevision, DecisionHash: request.DecisionHash, Answer: *request.DecisionAnswer})
	default:
		return workflowstate.ErrConflict
	}

	return err
}

func (s service) workflowIntegrationControlGuard(ctx context.Context, request WorkflowControlRequest) (workflowcore.IntegrationBudgetGuard, error) {
	store, err := s.workflowStore()
	if err != nil {
		return workflowcore.IntegrationBudgetGuard{}, err
	}

	inspection, err := store.InspectWorkflow(ctx, request.ID)
	if err != nil {
		return workflowcore.IntegrationBudgetGuard{}, err
	}

	if request.Integration == nil || inspection.Invocation.Revision != request.ExpectedRevision || inspection.Integration == nil {
		return workflowcore.IntegrationBudgetGuard{}, workflowstate.ErrConflict
	}

	shown, expected := workflowIntegrationStatus(inspection.Integration), request.Integration
	if !workflowIntegrationDisplayedGuardMatches(*expected, *shown) {
		return workflowcore.IntegrationBudgetGuard{}, workflowstate.ErrConflict
	}

	guard := workflowcore.IntegrationBudgetGuard{InvocationID: request.ID, InvocationRevision: request.ExpectedRevision, BudgetRevision: expected.BudgetRevision, EntryID: expected.EntryID, EntryRevision: expected.EntryRevision}
	if inspection.Integration.Entry != nil {
		guard.ClaimGeneration = inspection.Integration.Entry.ClaimGeneration
	}

	return guard, nil
}

func workflowIntegrationDisplayedGuardMatches(expected workflowstate.ControlIntegrationGuard, shown WorkflowIntegrationStatus) bool {
	return expected.EntryID == shown.EntryID && expected.EntryRevision == shown.EntryRevision && expected.BudgetRevision == shown.Budget.Revision && expected.CandidateHash == shown.CandidateHash
}
