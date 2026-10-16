package app

import (
	"fmt"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
)

func telegramControlRequest(request core.TelegramOperationRequest, args telegramControlArguments, status WorkflowWorkStatus) (WorkflowControlRequest, error) {
	control := WorkflowControlRequest{Action: args.Action, ID: args.ID, ExpectedRevision: status.Revision, StepLimit: 1}
	if status.Revision == 0 || !validTelegramControlAction(args.Action) {
		return control, core.ErrWorkflowConflict
	}

	if err := validateTelegramControlParameters(args.Action, args.Parameters); err != nil {
		return control, err
	}

	if !workflowDecisionControlAction(args.Action) {
		return control, nil
	}

	return telegramBoundDecisionControlRequest(request, control, args.Parameters, status)
}

func telegramBoundDecisionControlRequest(request core.TelegramOperationRequest, control WorkflowControlRequest, parameters *telegramControlParameters, status WorkflowWorkStatus) (WorkflowControlRequest, error) {
	control.BindingHash = status.ApprovalHash
	if current := status.Integration; current != nil && current.EntryID != "" {
		control.Integration = &WorkflowIntegrationControlGuard{EntryID: current.EntryID, EntryRevision: current.EntryRevision, BudgetRevision: current.Budget.Revision, CandidateHash: current.CandidateHash}
	}

	if control.Action == "approve" {
		if control.BindingHash == "" || !slices.Contains([]string{"waiting_approval", "human_required"}, status.ReviewState) {
			return control, core.ErrWorkflowConflict
		}

		return control, nil
	}

	if control.Integration == nil {
		return control, core.ErrWorkflowConflict
	}

	p := parameters
	if control.Action == "grant" {
		control.BudgetGrant = &core.WorkflowIntegrationGrant{ID: request.ID, Actor: request.Operator, Kind: p.Kind, Amount: p.Amount, Reason: p.Reason}

		return control, nil
	}

	return telegramDecisionAnswerRequest(request, control, *p, status.Integration.Decision)
}

func telegramDecisionAnswerRequest(request core.TelegramOperationRequest, control WorkflowControlRequest, p telegramControlParameters, decision *WorkflowIntegrationDecisionStatus) (WorkflowControlRequest, error) {
	if decision == nil || decision.State != "waiting" {
		return control, core.ErrWorkflowConflict
	}

	for _, question := range decision.Requests {
		if question.ID != p.QuestionID {
			continue
		}

		if p.Option > len(question.Options) {
			return control, core.ErrWorkflowConflict
		}

		control.DecisionRevision, control.DecisionHash = decision.Revision, decision.Hash
		control.DecisionAnswer = &core.WorkflowIntegrationAnswer{ID: request.ID, Actor: request.Operator, DecisionID: question.ID, Authority: question.Authority, Choice: question.Options[p.Option-1], Action: p.AnswerAction, Reason: p.Reason}

		return control, nil
	}

	return control, core.ErrWorkflowConflict
}

func telegramControlDecisionText(control WorkflowControlRequest) string {
	text := ""

	if grant := control.BudgetGrant; grant != nil {
		text += fmt.Sprintf("\nGrant: %d %s; reason: %s", grant.Amount, grant.Kind, grant.Reason)
	}

	if answer := control.DecisionAnswer; answer != nil {
		text += fmt.Sprintf("\nQuestion: %s; revision: %d; hash: %s\nAnswer: %s; choice: %s\nReason: %s", answer.DecisionID, control.DecisionRevision, control.DecisionHash, answer.Action, answer.Choice, answer.Reason)
	}

	return text
}
