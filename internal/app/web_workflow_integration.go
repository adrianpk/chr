package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

func webWorkflowIntegrationFields(status *WorkflowIntegrationStatus) []webui.WorkflowFieldView {
	fields := workflowIntegrationFields(status)
	result := make([]webui.WorkflowFieldView, 0, len(fields))

	for _, field := range fields {
		result = append(result, webui.WorkflowFieldView{Label: field.label, Value: field.value})
	}

	return result
}

func webWorkflowIntegrationControls(status WorkflowWorkStatus) *webui.WorkflowIntegrationControlView {
	current := status.Integration
	if current == nil {
		return nil
	}

	view := &webui.WorkflowIntegrationControlView{EntryID: current.EntryID, EntryRevision: strconv.FormatUint(current.EntryRevision, 10), BudgetRevision: strconv.FormatUint(current.Budget.Revision, 10), CandidateHash: current.CandidateHash, GrantID: workflowIntegrationResponseID("grant", status.ID, status.Revision, ""), GrantKinds: []string{"executor", "reviewer", "closing", "merger", "drift", "submission"}}
	if current.Budget.TokenLimit != nil {
		view.GrantKinds = append(view.GrantKinds, "tokens")
	}

	decision := current.Decision
	if decision == nil || decision.State != "waiting" {
		return view
	}

	view.DecisionRevision, view.DecisionHash = strconv.FormatUint(decision.Revision, 10), decision.Hash
	view.FindingCategories = []string{workflowstate.AgenticReviewFindingCorrectness, workflowstate.AgenticReviewFindingScope, workflowstate.AgenticReviewFindingEvidence, workflowstate.AgenticReviewFindingContract, workflowstate.AgenticReviewFindingSafety, workflowstate.AgenticReviewFindingMergeability}

	for _, question := range decision.Requests {
		view.Questions = append(view.Questions, webui.WorkflowIntegrationQuestionView{ID: question.ID, Issue: question.Issue, Authority: question.Authority, Options: question.Options, ReplayOwnership: question.ID == "replay-owner", AnswerID: workflowIntegrationResponseID("answer", status.ID, status.Revision, question.ID)})
	}

	return view
}

func workflowIntegrationResponseID(action, id string, revision uint64, question string) string {
	input := fmt.Sprintf("%s\n%s\n%d\n%s", action, id, revision, question)

	return fmt.Sprintf("%s-%x", action, sha256.Sum256([]byte(input)))
}

func webWorkflowIntegrationControl(ctx context.Context, operations service, request WorkflowControlRequest, fields map[string]string) (WorkflowControlRequest, error) {
	if request.Action == "grant" || request.Action == "answer_decision" || fields["budget_revision"] != "" {
		return webWorkflowGuardedControl(request, fields)
	}

	if request.Action == "approve" {
		status, err := operations.WorkflowStatus(ctx, request.ID, WorkflowStatusRequest{Limit: 1})
		if err != nil {
			return request, err
		}

		if status.Integration != nil {
			return request, fmt.Errorf("%w: native approval requires displayed integration guards", workflowstate.ErrConflict)
		}
	}

	return request, nil
}

func webWorkflowGuardedControl(request WorkflowControlRequest, fields map[string]string) (WorkflowControlRequest, error) {
	if request.Action == "approve" || request.Action == "grant" || request.Action == "answer_decision" {
		parsed, err := workflowIntegrationControlFields(request.Action, fields)
		parsed.ID, parsed.ExpectedRevision, parsed.BindingHash, parsed.StepLimit = request.ID, request.ExpectedRevision, request.BindingHash, request.StepLimit

		return parsed, err
	}

	guard, err := workflowIntegrationGuardFields(fields)
	request.Integration = &guard

	return request, err
}
