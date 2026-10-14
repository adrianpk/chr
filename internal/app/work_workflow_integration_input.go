package app

import (
	"fmt"
	"strconv"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

// Transport fields contain displayed fences and explicit operator input only.
// The shared service derives claim generations and owns every mutation.
func workflowIntegrationControlFields(action string, fields map[string]string) (WorkflowControlRequest, error) {
	request := WorkflowControlRequest{Action: strings.ReplaceAll(action, "-", "_")}

	guard, err := workflowIntegrationGuardFields(fields)
	if err != nil {
		return request, err
	}

	request.Integration = &guard
	switch request.Action {
	case "approve":
		return request, nil
	case "grant":
		grant, err := workflowIntegrationGrantFields(fields)
		request.BudgetGrant = &grant

		return request, err
	case "answer_decision":
		return workflowIntegrationAnswerFields(request, fields)
	default:
		return request, fmt.Errorf("unsupported integration control")
	}
}

func workflowIntegrationGuardFields(fields map[string]string) (WorkflowIntegrationControlGuard, error) {
	guard := WorkflowIntegrationControlGuard{EntryID: fields["entry_id"], CandidateHash: fields["candidate_hash"]}

	budget, err := workflowIntegrationRevision(fields["budget_revision"], false)
	if err != nil {
		return guard, err
	}

	entry, err := workflowIntegrationRevision(fields["entry_revision"], guard.EntryID == "")
	if err != nil {
		return guard, err
	}

	if guard.EntryID != "" && !validWorkflowControlIdentity(guard.EntryID) {
		return guard, fmt.Errorf("invalid displayed integration entry")
	}

	if guard.CandidateHash != "" {
		if err := workWorkflowHashInput("integration candidate", guard.CandidateHash); err != nil {
			return guard, err
		}
	}

	guard.EntryRevision, guard.BudgetRevision = entry, budget

	return guard, nil
}

func workflowIntegrationRevision(value string, allowZero bool) (uint64, error) {
	if len(value) > 20 {
		return 0, fmt.Errorf("displayed integration revision bounds exceeded")
	}

	revision, err := strconv.ParseUint(value, 10, 64)
	if err != nil || (revision == 0 && !allowZero) {
		return 0, fmt.Errorf("integration control requires displayed revisions")
	}

	return revision, nil
}

func workflowIntegrationGrantFields(fields map[string]string) (core.WorkflowIntegrationGrant, error) {
	grant := core.WorkflowIntegrationGrant{ID: fields["record_id"], Kind: fields["kind"], Actor: fields["actor"], Reason: fields["reason"]}
	if !validWorkflowControlIdentity(grant.ID) || !validWorkflowControlIdentity(grant.Actor) || !validWorkflowControlIdentity(grant.Reason) {
		return grant, fmt.Errorf("grant requires an explicit record identity, actor and reason within 128 bytes")
	}

	amount, err := strconv.ParseInt(fields["amount"], 10, 64)
	if err != nil || amount <= 0 || len(fields["kind"]) > 16 {
		return grant, fmt.Errorf("grant requires a positive bounded allowance")
	}

	grant.Amount = amount

	return grant, nil
}

func workflowIntegrationAnswerFields(request WorkflowControlRequest, fields map[string]string) (WorkflowControlRequest, error) {
	revision, err := workflowIntegrationRevision(fields["decision_revision"], false)
	if err != nil {
		return request, err
	}

	if err := workWorkflowHashInput("decision", fields["decision_hash"]); err != nil {
		return request, err
	}

	answer, err := workflowIntegrationAnswerPayload(fields)
	if err != nil {
		return request, err
	}

	request.DecisionRevision, request.DecisionHash, request.DecisionAnswer = revision, fields["decision_hash"], &answer

	return request, nil
}

func workflowIntegrationAnswerPayload(fields map[string]string) (core.WorkflowIntegrationAnswer, error) {
	answer := core.WorkflowIntegrationAnswer{ID: fields["record_id"], DecisionID: fields["decision_id"], Choice: fields["choice"], Action: fields["answer_action"], Actor: fields["actor"], Authority: fields["authority"], Reason: fields["reason"], OriginalCommit: fields["original_commit"]}
	for _, value := range []string{answer.ID, answer.DecisionID, answer.Actor, answer.Authority} {
		if !validWorkflowControlIdentity(value) {
			return answer, fmt.Errorf("answer requires the displayed question, actor and authority")
		}
	}

	if err := workflowIntegrationAnswerBounds(fields, answer); err != nil {
		return answer, err
	}

	if fields["fix_category"] != "" || fields["fix_title"] != "" || fields["fix_detail"] != "" {
		answer.IntegrationFix = &core.AgenticReviewFinding{Category: fields["fix_category"], Title: fields["fix_title"], Detail: fields["fix_detail"]}
	}

	return answer, nil
}

func workflowIntegrationAnswerBounds(fields map[string]string, answer core.WorkflowIntegrationAnswer) error {
	for _, value := range []string{answer.Choice, answer.Reason, fields["fix_category"], fields["fix_title"], fields["fix_detail"]} {
		if len(value) > 8192 || strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("decision answer bounds exceeded")
		}
	}

	if len(answer.OriginalCommit) > 64 || (answer.Action != "resolve" && answer.Action != "defer") {
		return fmt.Errorf("answer requires an explicit resolve or defer action")
	}

	return nil
}
