package app

import (
	"slices"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func telegramControlRequest(request telegramcore.OperationRequest, args telegramControlArguments, status WorkflowWorkStatus) (WorkflowControlRequest, error) {
	control := WorkflowControlRequest{Action: args.Action, ID: args.ID, ExpectedRevision: status.Revision, StepLimit: 1}
	if status.Revision == 0 || !validTelegramControlAction(args.Action) {
		return control, workflowstate.ErrConflict
	}

	if err := validateTelegramControlParameters(args.Action, args.Parameters); err != nil {
		return control, err
	}

	if !workflowDecisionControlAction(args.Action) {
		return control, nil
	}

	return telegramBoundDecisionControlRequest(request, control, args.Parameters, status)
}

func telegramBoundDecisionControlRequest(request telegramcore.OperationRequest, control WorkflowControlRequest, parameters *telegramControlParameters, status WorkflowWorkStatus) (WorkflowControlRequest, error) {
	control.BindingHash = status.ApprovalHash
	if current := status.Integration; current != nil && current.EntryID != "" {
		control.Integration = &workflowstate.ControlIntegrationGuard{EntryID: current.EntryID, EntryRevision: current.EntryRevision, BudgetRevision: current.Budget.Revision, CandidateHash: current.CandidateHash}
	}

	if control.Action == "approve" {
		if control.BindingHash == "" || !slices.Contains([]string{"waiting_approval", "human_required"}, status.ReviewState) {
			return control, workflowstate.ErrConflict
		}

		return control, nil
	}

	if control.Integration == nil {
		return control, workflowstate.ErrConflict
	}

	p := parameters
	if control.Action == "grant" {
		control.BudgetGrant = &workflowstate.IntegrationGrant{ID: request.ID, Actor: request.Operator, Kind: p.Kind, Amount: p.Amount, Reason: p.Reason}

		return control, nil
	}

	return telegramDecisionAnswerRequest(request, control, *p, status.Integration.Decision)
}

func telegramDecisionAnswerRequest(request telegramcore.OperationRequest, control WorkflowControlRequest, p telegramControlParameters, decision *WorkflowIntegrationDecisionStatus) (WorkflowControlRequest, error) {
	if decision == nil || decision.State != "waiting" {
		return control, workflowstate.ErrConflict
	}

	for _, question := range decision.Requests {
		if question.ID != p.QuestionID {
			continue
		}

		if p.Option > len(question.Options) {
			return control, workflowstate.ErrConflict
		}

		control.DecisionRevision, control.DecisionHash = decision.Revision, decision.Hash
		control.DecisionAnswer = &workflowstate.IntegrationAnswer{ID: request.ID, Actor: request.Operator, DecisionID: question.ID, Authority: question.Authority, Choice: question.Options[p.Option-1], Action: p.AnswerAction, Reason: p.Reason}

		return control, nil
	}

	return control, workflowstate.ErrConflict
}
