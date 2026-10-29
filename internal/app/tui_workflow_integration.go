package app

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/tui"
)

func mapTUIIntegrationControls(view *tui.WorkflowView, current *WorkflowIntegrationStatus) {
	if current == nil {
		return
	}

	view.Integration = &workflowstate.ControlIntegrationGuard{EntryID: current.EntryID, EntryRevision: current.EntryRevision, BudgetRevision: current.Budget.Revision, CandidateHash: current.CandidateHash}

	view.GrantKinds = []string{"executor", "reviewer", "closing", "merger", "drift", "submission"}
	if current.Budget.TokenLimit != nil {
		view.GrantKinds = append(view.GrantKinds, "tokens")
	}

	if decision := current.Decision; decision != nil && decision.State == "waiting" {
		view.DecisionRevision, view.DecisionHash = decision.Revision, decision.Hash
		view.Questions = decision.Requests
	}
}

func tuiWorkflowControlRequest(request tui.WorkActionRequest) (WorkflowControlRequest, error) {
	native := request.Workflow

	control := WorkflowControlRequest{Action: strings.TrimPrefix(string(request.Action), "workflow_"), ID: native.ID, ExpectedRevision: native.ExpectedRevision, BindingHash: native.BindingHash, StepLimit: 1, Integration: native.Integration}
	if control.Action != "grant" && control.Action != "answer_decision" {
		return control, nil
	}

	if native.Integration == nil {
		return control, fmt.Errorf("inspect the retained integration guards first")
	}

	fields := maps.Clone(native.Fields)
	if fields == nil {
		fields = map[string]string{}
	}

	guard := native.Integration
	fields["entry_id"], fields["entry_revision"] = guard.EntryID, strconv.FormatUint(guard.EntryRevision, 10)
	fields["budget_revision"], fields["candidate_hash"] = strconv.FormatUint(guard.BudgetRevision, 10), guard.CandidateHash
	fields["decision_revision"], fields["decision_hash"] = strconv.FormatUint(native.DecisionRevision, 10), native.DecisionHash
	fields["record_id"] = native.RequestKey
	parsed, err := workflowIntegrationControlFields(control.Action, fields)
	parsed.ID, parsed.ExpectedRevision, parsed.BindingHash, parsed.StepLimit = control.ID, control.ExpectedRevision, control.BindingHash, control.StepLimit

	return parsed, err
}

func (e tuiWorkActionExecutor) validateTUIIntegrationControl(request tui.WorkActionRequest, status WorkflowWorkStatus) error {
	control, err := tuiWorkflowControlRequest(request)
	if err != nil {
		return err
	}

	switch control.Action {
	case "grant", "answer_decision", "approve":
		return e.validateTUIIntegrationGuard(control, status)
	default:
		return nil
	}
}

func (e tuiWorkActionExecutor) validateTUIIntegrationGuard(control WorkflowControlRequest, status WorkflowWorkStatus) error {
	if status.Integration != nil && status.Integration.EntryID != "" && control.Integration == nil {
		return workflowstate.ErrConflict
	}

	if control.Integration != nil {
		if _, err := e.operations.workflowIntegrationControlGuard(e.operations.ctx, control); err != nil {
			return err
		}
	}

	if control.Action == "answer_decision" {
		return validateTUIDecisionGuard(control, status.Integration)
	}

	return nil
}

func validateTUIDecisionGuard(control WorkflowControlRequest, integration *WorkflowIntegrationStatus) error {
	if integration == nil || integration.Decision == nil {
		return workflowstate.ErrConflict
	}

	decision := integration.Decision
	if decision.Revision != control.DecisionRevision || decision.Hash != control.DecisionHash {
		return workflowstate.ErrConflict
	}

	return nil
}
