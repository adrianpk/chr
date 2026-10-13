package core

import (
	"slices"
	"strings"
)

func validateWorkflowDecisionRecord(record WorkflowIntegrationDecisions) error {
	if !validWorkflowDecisionHeader(record) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowMergerDecisions(record.Requests, record.Input); err != nil {
		return err
	}

	if err := validateWorkflowMergerInput(record.Input, record.Input.AdmittedAt); err != nil {
		return err
	}

	if (record.State == "resolved") != workflowDecisionsResolved(record) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowDecisionHistory(record); err != nil {
		return err
	}

	if record.Previous != nil {
		return validateWorkflowEvidenceRef(record.Input.InvocationID, *record.Previous)
	}

	return nil
}

func validateWorkflowDecisionAnswer(record WorkflowIntegrationDecisions, answer WorkflowIntegrationAnswer) error {
	index := slices.IndexFunc(record.Requests, func(d WorkflowMergerDecision) bool { return d.ID == answer.DecisionID })
	if index < 0 || !validWorkflowDecisionAnswerActor(answer) {
		return ErrWorkflowConflict
	}

	question := record.Requests[index]
	if answer.Authority != question.Authority || !slices.Contains(question.Options, answer.Choice) {
		return ErrWorkflowConflict
	}

	if !validWorkflowDecisionInterpretation(question, answer) {
		return ErrWorkflowConflict
	}

	return validateWorkflowDecisionPayload(answer)
}

func validateWorkflowDecisionPayload(answer WorkflowIntegrationAnswer) error {
	if answer.DecisionID != "replay-owner" || answer.Action != "resolve" {
		if answer.OriginalCommit != "" || answer.IntegrationFix != nil {
			return ErrWorkflowConflict
		}

		return nil
	}

	return validateWorkflowReplayPayload(answer)
}

func validateWorkflowReplayPayload(answer WorkflowIntegrationAnswer) error {
	switch answer.Choice {
	case "Identify the original source step":
		if !workflowCommitID(answer.OriginalCommit) || answer.IntegrationFix != nil {
			return ErrWorkflowConflict
		}
	case "Declare an integration fix with an actionable finding":
		if answer.OriginalCommit != "" || !validWorkflowDecisionFix(answer.IntegrationFix) {
			return ErrWorkflowConflict
		}
	default:
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowDecisionHeader(record WorkflowIntegrationDecisions) bool {
	return record.SchemaVersion == 1 && validWorkflowRequestKey(record.ID) && record.Revision >= 1 && record.Revision <= WorkflowPageLimit && record.CandidateRevision >= 1 && slices.Contains([]string{"waiting", "resolved"}, record.State) && len(record.Requests) != 0 && len(record.Answers) <= WorkflowPageLimit
}

func validateWorkflowDecisionHistory(record WorkflowIntegrationDecisions) error {
	seen := map[string]bool{}
	for _, answer := range record.Answers {
		if seen[answer.ID] || !validWorkflowDecisionAnswerSeal(record, answer) {
			return ErrWorkflowConflict
		}

		if err := validateWorkflowDecisionAnswer(record, answer); err != nil {
			return err
		}

		seen[answer.ID] = true
	}

	return nil
}

func validWorkflowDecisionAnswerSeal(record WorkflowIntegrationDecisions, answer WorkflowIntegrationAnswer) bool {
	return answer.PriorRevision > 0 && answer.PriorRevision < record.Revision && validWorkflowNativeFingerprint(answer.QuestionHash) && !answer.RecordedAt.IsZero()
}

func validWorkflowDecisionAnswerActor(answer WorkflowIntegrationAnswer) bool {
	return validWorkflowRequestKey(answer.ID) && validWorkflowRequestKey(answer.Actor) && validWorkflowMergerText(answer.Reason) && slices.Contains([]string{"resolve", "defer"}, answer.Action)
}

func validWorkflowDecisionInterpretation(question WorkflowMergerDecision, answer WorkflowIntegrationAnswer) bool {
	if answer.Action != "resolve" {
		return true
	}

	if strings.HasPrefix(question.ID, "procedure-") {
		return false
	}

	return !strings.HasPrefix(question.ID, "intent-") || answer.Choice == "Provide a compatible requirement interpretation"
}

func validWorkflowDecisionFix(finding *AgenticReviewFinding) bool {
	return finding != nil && validateAgenticReviewFinding(*finding) == nil && finding.Category != AgenticReviewFindingCosmetic
}
