package workflow

import (
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowDecisionRecord(record IntegrationDecisions) error {
	if !validWorkflowDecisionHeader(record) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowMergerDecisions(record.Requests, record.Input); err != nil {
		return err
	}

	if err := validateWorkflowMergerInput(record.Input, record.Input.AdmittedAt); err != nil {
		return err
	}

	if (record.State == "resolved") != workflowDecisionsResolved(record) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowDecisionHistory(record); err != nil {
		return err
	}

	if record.Previous != nil {
		return validateWorkflowEvidenceRef(record.Input.InvocationID, *record.Previous)
	}

	return nil
}

func validateWorkflowDecisionAnswer(record IntegrationDecisions, answer workflowstate.IntegrationAnswer) error {
	index := slices.IndexFunc(record.Requests, func(d MergerDecision) bool { return d.ID == answer.DecisionID })
	if index < 0 || !validWorkflowDecisionAnswerActor(answer) {
		return workflowstate.ErrConflict
	}

	question := record.Requests[index]
	if answer.Authority != question.Authority || !slices.Contains(question.Options, answer.Choice) {
		return workflowstate.ErrConflict
	}

	if !validWorkflowDecisionInterpretation(question, answer) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowDecisionPayload(answer)
}

func validateWorkflowDecisionPayload(answer workflowstate.IntegrationAnswer) error {
	if answer.DecisionID != "replay-owner" || answer.Action != "resolve" {
		if answer.OriginalCommit != "" || answer.IntegrationFix != nil {
			return workflowstate.ErrConflict
		}

		return nil
	}

	return validateWorkflowReplayPayload(answer)
}

func validateWorkflowReplayPayload(answer workflowstate.IntegrationAnswer) error {
	switch answer.Choice {
	case "Identify the original source step":
		if !workflowCommitID(answer.OriginalCommit) || answer.IntegrationFix != nil {
			return workflowstate.ErrConflict
		}
	case "Declare an integration fix with an actionable finding":
		if answer.OriginalCommit != "" || !validWorkflowDecisionFix(answer.IntegrationFix) {
			return workflowstate.ErrConflict
		}
	default:
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowDecisionHeader(record IntegrationDecisions) bool {
	return record.SchemaVersion == 1 && workflowstate.ValidRequestKey(record.ID) && record.Revision >= 1 && record.Revision <= workflowstate.PageLimit && record.CandidateRevision >= 1 && slices.Contains([]string{"waiting", "resolved"}, record.State) && len(record.Requests) != 0 && len(record.Answers) <= workflowstate.PageLimit
}

func validateWorkflowDecisionHistory(record IntegrationDecisions) error {
	seen := map[string]bool{}
	for _, answer := range record.Answers {
		if seen[answer.ID] || !validWorkflowDecisionAnswerSeal(record, answer) {
			return workflowstate.ErrConflict
		}

		if err := validateWorkflowDecisionAnswer(record, answer); err != nil {
			return err
		}

		seen[answer.ID] = true
	}

	return nil
}

func validWorkflowDecisionAnswerSeal(record IntegrationDecisions, answer workflowstate.IntegrationAnswer) bool {
	return answer.PriorRevision > 0 && answer.PriorRevision < record.Revision && validWorkflowNativeFingerprint(answer.QuestionHash) && !answer.RecordedAt.IsZero()
}

func validWorkflowDecisionAnswerActor(answer workflowstate.IntegrationAnswer) bool {
	return workflowstate.ValidRequestKey(answer.ID) && workflowstate.ValidRequestKey(answer.Actor) && validWorkflowMergerText(answer.Reason) && slices.Contains([]string{"resolve", "defer"}, answer.Action)
}

func validWorkflowDecisionInterpretation(question MergerDecision, answer workflowstate.IntegrationAnswer) bool {
	if answer.Action != "resolve" {
		return true
	}

	if strings.HasPrefix(question.ID, "procedure-") {
		return false
	}

	return !strings.HasPrefix(question.ID, "intent-") || answer.Choice == "Provide a compatible requirement interpretation"
}

func validWorkflowDecisionFix(finding *workflowstate.ReviewFinding) bool {
	return finding != nil && workflowstate.ValidateReviewFinding(*finding) == nil && finding.Category != workflowstate.AgenticReviewFindingCosmetic
}
