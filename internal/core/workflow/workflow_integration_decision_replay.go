package workflow

import (
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowDecisionReplayAnswer(record IntegrationDecisions, answer workflowstate.IntegrationAnswer) error {
	if answer.OriginalCommit == "" {
		return nil
	}

	var manifest NativeSourceManifest
	if err := s.readWorkflowNativeEvidence(record.Input.InvocationID, record.Input.SourceManifestRef, &manifest); err != nil {
		return err
	}

	if !slices.Contains(manifest.Commits, answer.OriginalCommit) {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowMergerAnsweredPlan(plan MergerRepairPlan, input MergerInput, owners []string) (MergerRepairPlan, bool, error) {
	if input.OperatorDecisionRef == nil {
		return plan, false, nil
	}

	for i := len(input.OperatorAnswers) - 1; i >= 0; i-- {
		answer := input.OperatorAnswers[i]
		if answer.DecisionID != "replay-owner" {
			continue
		}

		if answer.Action != "resolve" {
			return plan, false, nil
		}

		if answer.OriginalCommit != "" {
			if !slices.Contains(owners, answer.OriginalCommit) {
				return plan, false, workflowstate.ErrConflict
			}

			plan.Outcome, plan.OriginalCommit = "replay_step", answer.OriginalCommit

			return plan, true, nil
		}

		if answer.IntegrationFix == nil || workflowstate.ValidateReviewFinding(*answer.IntegrationFix) != nil || answer.IntegrationFix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return plan, false, workflowstate.ErrConflict
		}

		finding := *answer.IntegrationFix
		plan.Outcome, plan.IntegrationFix = "integration_fix", &finding

		return plan, true, nil
	}

	return plan, false, nil
}
