package review

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ParseDecision decodes one strict decision and checks its immutable candidate binding.
func ParseDecision(stdout string, input *workflowstate.ReviewInput) (workflowstate.ReviewDecision, error) {
	var decision workflowstate.ReviewDecision
	if input == nil {
		return decision, workflowstate.ErrConflict
	}

	decoder := json.NewDecoder(bytes.NewBufferString(stdout))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&decision); err != nil {
		return decision, fmt.Errorf("decode workflow review decision: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return decision, fmt.Errorf("workflow review contains trailing output")
	}

	return decision, CheckDecision(decision, *input)
}

// CheckDecision checks outcome, blocking findings and the retained repair task.
func CheckDecision(decision workflowstate.ReviewDecision, input workflowstate.ReviewInput) error {
	hash, err := InputHash(input)
	if err != nil || hash != decision.BindingHash {
		return fmt.Errorf("%w: review decision names another candidate", workflowstate.ErrConflict)
	}

	if err := workflowstate.ValidateAgenticReviewResult(decision.Result); err != nil {
		return err
	}

	if decision.Result.Decision == workflowstate.IntegrationReviewDecisionPass && len(decision.Result.BlockingFindings) != 0 {
		return fmt.Errorf("passing review contains blocking findings")
	}

	if decision.Result.Decision != workflowstate.IntegrationReviewDecisionNeedsChanges && decision.RepairTaskKey != "" {
		return fmt.Errorf("review repair key requires needs_changes")
	}

	if decision.Result.Decision != workflowstate.IntegrationReviewDecisionNeedsChanges {
		return nil
	}

	return checkRepairDecision(decision, input)
}

func checkRepairDecision(decision workflowstate.ReviewDecision, input workflowstate.ReviewInput) error {
	if !slices.ContainsFunc(input.Slice.Tasks, func(task workflowstate.SliceTaskProof) bool { return task.Key == decision.RepairTaskKey }) {
		return fmt.Errorf("review repair task is outside the retained slice")
	}

	for _, finding := range decision.Result.BlockingFindings {
		if finding.Category == workflowstate.AgenticReviewFindingCosmetic {
			return fmt.Errorf("cosmetic findings cannot request automatic repair")
		}
	}

	return nil
}
