package workflow

import (
	"encoding/json"
	"fmt"
	"slices"

	workflowreview "forge.adrianpk.com/chr/internal/core/workflow/review"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowReviewBinding(run workflowInvocationFile) error {
	if run.ReviewCount < 0 || run.ReviewCount > workflowstate.PageLimit || len(run.ReviewRepairs) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow review bounds exceeded")
	}

	if err := validateWorkflowReviewBudgets(run.ReviewRepairs); err != nil {
		return err
	}

	if run.Review == nil {
		retained := workflowRetainedRoleCounts(run)
		if run.ReviewCount != retained.Review {
			return workflowstate.ErrConflict
		}

		return nil
	}

	return validateWorkflowCurrentReviewBinding(run)
}

func validateWorkflowReviewBudgets(budgets []workflowstate.ReviewRepairBudget) error {
	seen := map[workflowstate.Ref]bool{}
	for _, budget := range budgets {
		if !slices.Contains([]string{"slice", "ticket"}, budget.Unit.Kind) || budget.Unit.ID == "" || seen[budget.Unit] || budget.Consumed < 1 || budget.Consumed > workflowReviewRepairLimit {
			return workflowstate.ErrConflict
		}

		seen[budget.Unit] = true
	}

	return nil
}

func validateWorkflowCurrentReviewBinding(run workflowInvocationFile) error {
	binding := run.Review
	if run.Lane == nil || binding.Unit != run.Lane.Unit || binding.Round < 1 || binding.Round != run.ReviewCount || binding.InputRef != workflowReviewPath(run.ID, binding.Round, "input.json") || len(binding.InputHash) != 64 {
		return workflowstate.ErrConflict
	}

	if !slices.Contains([]workflowstate.ReviewState{workflowstate.ReviewPending, workflowstate.ReviewRunning, workflowstate.ReviewWaitingApproval, workflowstate.ReviewAuthorized, workflowstate.ReviewNeedsRepair, workflowstate.ReviewInsufficientEvidence, workflowstate.ReviewHumanRequired, workflowstate.ReviewStale, workflowstate.ReviewRepairLimitExceeded}, binding.State) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowReviewDecisionRefs(run.ID, *binding)
}

func validateWorkflowReviewDecisionRefs(id string, binding workflowstate.ReviewBinding) error {
	if binding.DecisionCount == 0 {
		if !slices.Contains([]workflowstate.ReviewState{workflowstate.ReviewPending, workflowstate.ReviewRunning, workflowstate.ReviewWaitingApproval, workflowstate.ReviewStale}, binding.State) || binding.ResultRef != "" || binding.ResultHash != "" {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if binding.DecisionCount < 1 || binding.DecisionCount > 2 || len(binding.ResultHash) != 64 || binding.ResultRef != workflowReviewPath(id, binding.Round, fmt.Sprintf("decision-%d.json", binding.DecisionCount)) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) readWorkflowReviewDecisions(run workflowInvocationFile, input workflowstate.ReviewInput) (workflowstate.ReviewDecision, error) {
	ref, hash := run.Review.ResultRef, run.Review.ResultHash

	var newest workflowstate.ReviewDecision

	for count := run.Review.DecisionCount; count > 0; count-- {
		evidence, err := s.readWorkflowReviewEvidence(run, input, count, workflowstate.EvidenceRef{Path: ref, Hash: hash})
		if err != nil {
			return newest, err
		}

		if count == run.Review.DecisionCount {
			newest = evidence.Decision
		}

		ref, hash = evidence.PreviousRef, evidence.PreviousHash
	}

	if ref != "" || hash != "" {
		return newest, workflowstate.ErrConflict
	}

	if run.Review.State == workflowstate.ReviewAuthorized && newest.Result.Decision != workflowstate.IntegrationReviewDecisionPass {
		return newest, workflowstate.ErrConflict
	}

	if newest.RepairTaskKey != run.Review.RepairTaskKey {
		return newest, workflowstate.ErrConflict
	}

	return newest, nil
}

func (s *Store) readWorkflowReviewEvidence(run workflowInvocationFile, input workflowstate.ReviewInput, count int, image workflowstate.EvidenceRef) (workflowReviewEvidence, error) {
	ref, hash := image.Path, image.Hash

	var evidence workflowReviewEvidence
	if ref != workflowReviewPath(run.ID, run.Review.Round, fmt.Sprintf("decision-%d.json", count)) {
		return evidence, workflowstate.ErrConflict
	}

	data, err := s.readWorkflowReviewImage(ref, hash)
	if err != nil {
		return evidence, err
	}

	if err := json.Unmarshal(data, &evidence); err != nil {
		return evidence, err
	}

	if err := validateWorkflowReviewEvidenceHeader(run, count, evidence); err != nil {
		return evidence, err
	}

	if err := validateWorkflowReviewOrigin(input.Policy.Mode, count, evidence); err != nil {
		return evidence, err
	}

	if err := workflowreview.CheckDecision(evidence.Decision, input); err != nil {
		return evidence, err
	}

	return evidence, nil
}

func validateWorkflowReviewOrigin(mode workflowstate.ReviewMode, count int, evidence workflowReviewEvidence) error {
	if count == 2 {
		if mode != workflowstate.ReviewAgentic || evidence.Origin != "human" || evidence.Decision.Result.Decision != workflowstate.IntegrationReviewDecisionPass {
			return workflowstate.ErrConflict
		}

		return nil
	}

	expected := "agentic"
	if mode == workflowstate.ReviewHuman {
		expected = "human"
	}

	if mode == workflowstate.ReviewTrustedAutoMerge {
		expected = "trusted"
	}

	if evidence.Origin != expected {
		return workflowstate.ErrConflict
	}

	if expected != "agentic" && evidence.Decision.Result.Decision != workflowstate.IntegrationReviewDecisionPass {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowReviewEvidenceHeader(run workflowInvocationFile, count int, evidence workflowReviewEvidence) error {
	if evidence.Schema != "chr-workflow-review/v1" || evidence.RecordedAt.IsZero() {
		return workflowstate.ErrConflict
	}

	if run.Review.DecisionCount == 2 && count == 1 && evidence.Decision.Result.Decision != workflowstate.IntegrationReviewDecisionHumanRequired {
		return workflowstate.ErrConflict
	}

	return nil
}
