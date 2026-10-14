package app

import (
	"fmt"
	"strconv"
	"strings"
)

type workflowIntegrationField struct{ label, value string }

// Every display consumes the public retained projection. Rendering cannot
// claim a lane, advance a queue entry, or infer publication from approval.
func workflowIntegrationFields(status *WorkflowIntegrationStatus) []workflowIntegrationField {
	if status == nil {
		return nil
	}

	fields := []workflowIntegrationField{{"Integration phase", status.Phase}, {"Publication", status.Publication}}
	if status.EntryID != "" {
		fields = append(fields,
			workflowIntegrationField{"Integration unit", status.Unit.Kind + " " + status.Unit.ID},
			workflowIntegrationField{"Integration entry", fmt.Sprintf("%s revision %d", status.EntryID, status.EntryRevision)},
			workflowIntegrationField{"Target queue", status.Repository + " " + status.TargetRef},
			workflowIntegrationField{"Queue position", strconv.Itoa(status.QueuePosition)},
		)
	}

	fields = append(fields, workflowIntegrationCandidateFields(status)...)
	fields = append(fields, workflowIntegrationRoleFields(status)...)
	fields = append(fields, workflowIntegrationBudgetFields(status.Budget)...)

	if status.Blocker != nil {
		fields = append(fields, workflowIntegrationField{"Integration blocker", status.Blocker.Reason}, workflowIntegrationField{"Required action", status.Blocker.RequiredAction})
	}

	return append(fields, workflowIntegrationDecisionFields(status.Decision)...)
}

func workflowIntegrationLines(status *WorkflowIntegrationStatus) []string {
	fields := workflowIntegrationFields(status)
	lines := make([]string, 0, len(fields))

	for _, field := range fields {
		lines = append(lines, field.label+": "+field.value)
	}

	return lines
}

func workflowIntegrationCandidateFields(status *WorkflowIntegrationStatus) []workflowIntegrationField {
	fields := []workflowIntegrationField{}

	for _, field := range []workflowIntegrationField{
		{"Integration candidate", status.CandidateHash},
		{"Original source", status.OriginalSource},
		{"Validated base", status.ValidatedBase},
		{"Prepared result", status.PreparedResult},
		{"Integration method", status.Method},
	} {
		if field.value != "" {
			fields = append(fields, field)
		}
	}

	return fields
}

func workflowIntegrationRoleFields(status *WorkflowIntegrationStatus) []workflowIntegrationField {
	mode := "serial"
	if status.MaximumRoles > 1 {
		mode = "parallel"
	}

	fields := []workflowIntegrationField{
		{"Agent capacity", fmt.Sprintf("%s; limit %d; available %d", mode, status.MaximumRoles, status.AvailableRoles)},
		{"Active agents", strconv.Itoa(len(status.Roles))},
	}
	for _, role := range status.Roles {
		label := "Active role"
		if role.Slot > 0 {
			label = "Agent lane " + strconv.Itoa(role.Slot)
		}

		fields = append(fields, workflowIntegrationField{label, role.Role + "; " + role.State + "; attempt " + role.AttemptID})
	}

	return fields
}

func workflowIntegrationBudgetFields(budget WorkflowIntegrationBudgetStatus) []workflowIntegrationField {
	fields := []workflowIntegrationField{{"Budget revision", strconv.FormatUint(budget.Revision, 10)}}
	for _, allowance := range []struct {
		kind                          string
		admitted, limit, used, remain int64
	}{
		{"Worker", budget.AdmissionLimits.Worker, budget.Limits.Worker, budget.Consumed.Worker, budget.Remaining.Worker},
		{"Review", budget.AdmissionLimits.Review, budget.Limits.Review, budget.Consumed.Review, budget.Remaining.Review},
		{"Closing", budget.AdmissionLimits.Closing, budget.Limits.Closing, budget.Consumed.Closing, budget.Remaining.Closing},
		{"Merger", budget.AdmissionLimits.Merger, budget.Limits.Merger, budget.Consumed.Merger, budget.Remaining.Merger},
		{"Drift", budget.AdmissionLimits.Drift, budget.Limits.Drift, budget.Consumed.Drift, budget.Remaining.Drift},
		{"Submission", budget.AdmissionLimits.Submission, budget.Limits.Submission, budget.Consumed.Submission, budget.Remaining.Submission},
	} {
		fields = append(fields, workflowIntegrationField{allowance.kind + " budget", fmt.Sprintf("admitted %d; limit %d; consumed %d; remaining %d", allowance.admitted, allowance.limit, allowance.used, allowance.remain)})
	}

	fields = append(fields, workflowIntegrationField{"Token budget", workflowIntegrationTokenValue(budget)})
	if budget.AdmissionLimit != nil {
		fields = append(fields, workflowIntegrationField{"Admission token limit", strconv.FormatInt(*budget.AdmissionLimit, 10)})
	}

	return fields
}

func workflowIntegrationTokenValue(budget WorkflowIntegrationBudgetStatus) string {
	usage := fmt.Sprintf("settled %d; reserved %d", budget.TokensSettled, budget.TokensReserved)
	if budget.TokenLimit == nil {
		return "unmetered; " + usage
	}

	available := "unknown"
	if budget.TokensAvailable != nil {
		available = strconv.FormatInt(*budget.TokensAvailable, 10)
	}

	return fmt.Sprintf("limit %d; %s; available %s", *budget.TokenLimit, usage, available)
}

func workflowIntegrationDecisionFields(decision *WorkflowIntegrationDecisionStatus) []workflowIntegrationField {
	if decision == nil {
		return nil
	}

	fields := []workflowIntegrationField{{"Decision binding", fmt.Sprintf("%s revision %d; %s; hash %s", decision.ID, decision.Revision, decision.State, decision.Hash)}}
	for _, question := range decision.Requests {
		fields = append(fields, workflowIntegrationField{"Question " + question.ID, question.Issue + "; authority " + question.Authority + "; options " + strings.Join(question.Options, "; ")})
	}

	for _, answer := range decision.Answers {
		fields = append(fields, workflowIntegrationField{"Answer " + answer.DecisionID, answer.Action + "; " + answer.Choice + "; actor " + answer.Actor + "; " + answer.Reason})
	}

	return fields
}
