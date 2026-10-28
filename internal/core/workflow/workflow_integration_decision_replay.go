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

func workflowMergerAnsweredPlan(plan MergerRepairPlan, input workflowstate.MergerInput, owners []string) (workflowAnsweredRepair, error) {
	if input.OperatorDecisionRef == nil {
		return workflowAnsweredRepair{Plan: plan, Found: false}, nil
	}

	for i := len(input.OperatorAnswers) - 1; i >= 0; i-- {
		answer := input.OperatorAnswers[i]
		if answer.DecisionID != "replay-owner" {
			continue
		}

		if answer.Action != "resolve" {
			return workflowAnsweredRepair{Plan: plan, Found: false}, nil
		}

		if answer.OriginalCommit != "" {
			if !slices.Contains(owners, answer.OriginalCommit) {
				return workflowAnsweredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
			}

			plan.Outcome, plan.OriginalCommit = "replay_step", answer.OriginalCommit

			return workflowAnsweredRepair{Plan: plan, Found: true}, nil
		}

		if answer.IntegrationFix == nil || workflowstate.ValidateReviewFinding(*answer.IntegrationFix) != nil || answer.IntegrationFix.Category == workflowstate.AgenticReviewFindingCosmetic {
			return workflowAnsweredRepair{Plan: plan, Found: false}, workflowstate.ErrConflict
		}

		finding := *answer.IntegrationFix
		plan.Outcome, plan.IntegrationFix = "integration_fix", &finding

		return workflowAnsweredRepair{Plan: plan, Found: true}, nil
	}

	return workflowAnsweredRepair{Plan: plan, Found: false}, nil
}
