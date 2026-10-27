package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"

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
		_, retained := workflowRetainedRoleCounts(run)
		if run.ReviewCount != retained {
			return workflowstate.ErrConflict
		}

		return nil
	}

	return validateWorkflowCurrentReviewBinding(run)
}

func validateWorkflowReviewBudgets(budgets []WorkflowReviewRepairBudget) error {
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

	if !slices.Contains([]string{"pending", "running", "waiting_approval", "authorized", "needs_repair", "insufficient_evidence", "human_required", "stale", "repair_limit_exceeded"}, binding.State) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowReviewDecisionRefs(run.ID, *binding)
}

func validateWorkflowReviewDecisionRefs(id string, binding WorkflowReviewBinding) error {
	if binding.DecisionCount == 0 {
		if !slices.Contains([]string{"pending", "running", "waiting_approval", "stale"}, binding.State) || binding.ResultRef != "" || binding.ResultHash != "" {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if binding.DecisionCount < 1 || binding.DecisionCount > 2 || len(binding.ResultHash) != 64 || binding.ResultRef != workflowReviewPath(id, binding.Round, fmt.Sprintf("decision-%d.json", binding.DecisionCount)) {
		return workflowstate.ErrConflict
	}

	return nil
}

func parseWorkflowReviewDecision(stdout string, input *WorkflowReviewInput) (WorkflowReviewDecision, error) {
	var decision WorkflowReviewDecision
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

	return decision, validateWorkflowReviewDecision(decision, *input)
}

func validateWorkflowReviewDecision(decision WorkflowReviewDecision, input WorkflowReviewInput) error {
	hash, err := workflowReviewInputHash(input)
	if err != nil || hash != decision.BindingHash {
		return fmt.Errorf("%w: review decision names another candidate", workflowstate.ErrConflict)
	}

	if err := ValidateAgenticReviewResult(decision.Result); err != nil {
		return err
	}

	if decision.Result.Decision == IntegrationReviewDecisionPass && len(decision.Result.BlockingFindings) != 0 {
		return fmt.Errorf("passing review contains blocking findings")
	}

	if decision.Result.Decision != IntegrationReviewDecisionNeedsChanges && decision.RepairTaskKey != "" {
		return fmt.Errorf("review repair key requires needs_changes")
	}

	if decision.Result.Decision != IntegrationReviewDecisionNeedsChanges {
		return nil
	}

	return validateWorkflowReviewRepairDecision(decision, input)
}

func validateWorkflowReviewRepairDecision(decision WorkflowReviewDecision, input WorkflowReviewInput) error {
	if !slices.ContainsFunc(input.Slice.Tasks, func(task WorkflowSliceTaskProof) bool { return task.Key == decision.RepairTaskKey }) {
		return fmt.Errorf("review repair task is outside the retained slice")
	}

	for _, finding := range decision.Result.BlockingFindings {
		if finding.Category == AgenticReviewFindingCosmetic {
			return fmt.Errorf("cosmetic findings cannot request automatic repair")
		}
	}

	return nil
}

func (s *WorkflowStore) readWorkflowReviewDecisions(run workflowInvocationFile, input WorkflowReviewInput) (WorkflowReviewDecision, error) {
	ref, hash := run.Review.ResultRef, run.Review.ResultHash

	var newest WorkflowReviewDecision

	for count := run.Review.DecisionCount; count > 0; count-- {
		evidence, err := s.readWorkflowReviewEvidence(run, input, count, ref, hash)
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

	if run.Review.State == "authorized" && newest.Result.Decision != IntegrationReviewDecisionPass {
		return newest, workflowstate.ErrConflict
	}

	if newest.RepairTaskKey != run.Review.RepairTaskKey {
		return newest, workflowstate.ErrConflict
	}

	return newest, nil
}

func (s *WorkflowStore) readWorkflowReviewEvidence(run workflowInvocationFile, input WorkflowReviewInput, count int, ref, hash string) (workflowReviewEvidence, error) {
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

	if err := validateWorkflowReviewDecision(evidence.Decision, input); err != nil {
		return evidence, err
	}

	return evidence, nil
}

func validateWorkflowReviewOrigin(mode ReviewWorkflowMode, count int, evidence workflowReviewEvidence) error {
	if count == 2 {
		if mode != ReviewWorkflowModeAgenticReview || evidence.Origin != "human" || evidence.Decision.Result.Decision != IntegrationReviewDecisionPass {
			return workflowstate.ErrConflict
		}

		return nil
	}

	expected := "agentic"
	if mode == ReviewWorkflowModeHumanReview {
		expected = "human"
	}

	if mode == ReviewWorkflowModeTrustedAutoMerge {
		expected = "trusted"
	}

	if evidence.Origin != expected {
		return workflowstate.ErrConflict
	}

	if expected != "agentic" && evidence.Decision.Result.Decision != IntegrationReviewDecisionPass {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowReviewEvidenceHeader(run workflowInvocationFile, count int, evidence workflowReviewEvidence) error {
	if evidence.Schema != "chr-workflow-review/v1" || evidence.RecordedAt.IsZero() {
		return workflowstate.ErrConflict
	}

	if run.Review.DecisionCount == 2 && count == 1 && evidence.Decision.Result.Decision != IntegrationReviewDecisionHumanRequired {
		return workflowstate.ErrConflict
	}

	return nil
}
